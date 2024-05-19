package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"log"
	"metrics/client/pb/pv"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const serverAddress = "localhost:8080"
const logFileName = "client/run_log"

var (
	totalRequests             int64
	totalFailedRequests       int64
	totalSuccessfulRequests   int64
	firstFailedRequestDetails string
	lastFailedRequestDetails  string
	firstFailedRequestOnce    sync.Once
	lastFailedRequestMutex    sync.Mutex
	gitCommitSha              string
	buildTimeStamp            int64
)

func getVersion(client pv.VersionServiceClient) {
	resp, err := client.GetVersion(context.Background(), &emptypb.Empty{})
	log.Printf("Response from version api: %v", resp)
	if err != nil {
		log.Printf("Failed to get response: %v", err)
	}
	gitCommitSha = resp.GitCommitSha
	buildTimeStamp = resp.BuildTimestamp
}

func sendRequests(client pb.MetricsServiceClient, requestJson string, wg *sync.WaitGroup, ctx context.Context, numConcurrentRequests int) {
	defer wg.Done()

	req := &pb.ExportMetricsServiceRequest{}
	if err := jsonpb.UnmarshalString(requestJson, req); err != nil {
		log.Printf("Error unmarshalling request: %v", err)
		return
	}

	requestCh := make(chan struct{}, numConcurrentRequests)
	defer close(requestCh)

	for i := 0; i < numConcurrentRequests; i++ {
		requestCh <- struct{}{}
	}

	for {
		select {
		case <-ctx.Done():
			// If context is canceled, return immediately
			return
		case <-requestCh:
			go func() {
				defer func() {
					requestCh <- struct{}{}
				}()
				resp, err := client.Export(context.Background(), req)
				atomic.AddInt64(&totalRequests, 1)

				if err != nil {
					// Case 1: Error returned by the Export call
					atomic.AddInt64(&totalFailedRequests, 1)
					firstFailedRequestOnce.Do(func() {
						firstFailedRequestDetails = fmt.Sprintf("Failed request details: %v", req)
					})
					lastFailedRequestMutex.Lock()
					lastFailedRequestDetails = fmt.Sprintf("Failed request details: %v", req)
					lastFailedRequestMutex.Unlock()
					//log.Printf("Failed to get response: %v", err)
				} else if resp.GetPartialSuccess() != nil {
					// Case 2: Partial success (some data points were rejected)
					rejectedDataPoints := resp.GetPartialSuccess().GetRejectedDataPoints()
					// Assuming each Metric contains exactly one data point, which might not be true
					totalSentDataPoints := int64(len(req.ResourceMetrics))
					successfulDataPoints := totalSentDataPoints - rejectedDataPoints
					atomic.AddInt64(&totalFailedRequests, rejectedDataPoints)
					atomic.AddInt64(&totalSuccessfulRequests, successfulDataPoints)
					firstFailedRequestOnce.Do(func() {
						firstFailedRequestDetails = fmt.Sprintf("Partial error request details: %v", req)
					})
					lastFailedRequestMutex.Lock()
					lastFailedRequestDetails = fmt.Sprintf("Failed request details: %v", req)
					lastFailedRequestMutex.Unlock()
				} else {
					// Case 3: Successful response
					atomic.AddInt64(&totalSuccessfulRequests, 1)
					//log.Printf("Successful response received")
				}
			}()
		}
	}
}

func readJSONFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getClientCertAndPool() (tls.Certificate, *x509.CertPool) {
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatal(err)
	}

	return clientCert, certPool
}

func main() {
	filename := flag.String("filename", "", "Path to the JSON file containing the request data")
	duration := flag.Int("duration", 0, "Duration of the load test in seconds")
	numConcurrentRequests := flag.Int("numConcurrentRequests", 1, "Number of concurrent requests to send")

	flag.Parse()

	if *filename == "" {
		log.Fatal("Please provide the filename parameter")
	}

	clientCert, certPool := getClientCertAndPool()
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	tlsCredential := credentials.NewTLS(config)
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)

	versionClient := pv.NewVersionServiceClient(conn)
	getVersion(versionClient)
	requestJson, err := readJSONFile(*filename)
	if err != nil {
		log.Fatalf("Failed to read request file: %v", err)
	}

	var outputWriter *os.File

	outputWriter, err = os.Create(logFileName)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputWriter.Close()
	log.SetOutput(outputWriter)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	if *duration > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(*duration)*time.Second)
		defer cancel()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	for i := 0; i < *numConcurrentRequests; i++ {
		wg.Add(1)
		go sendRequests(client, requestJson, &wg, ctx, *numConcurrentRequests)
	}

	wg.Wait()

	// Print parameters used for the run
	if outputWriter != nil {
		fmt.Fprintf(outputWriter, "Parameters:\n")
		fmt.Fprintf(outputWriter, "Filename: %s\n", *filename)
		fmt.Fprintf(outputWriter, "Duration: %d\n", *duration)
		fmt.Fprintf(outputWriter, "Number of Concurrent Requests: %d\n", *numConcurrentRequests)
		fmt.Fprintf(outputWriter, "GitCommitSha: %s\n", gitCommitSha)
		buildTimeStampStr := strconv.FormatInt(buildTimeStamp, 10)
		fmt.Fprintf(outputWriter, "BuildTimestamp: %s\n", buildTimeStampStr)

		// Print summary at the end
		// Print summary at the end
		fmt.Fprintf(outputWriter, "Total Requests: %d\n", totalRequests)
		fmt.Fprintf(outputWriter, "Total Successful Requests: %d\n", totalSuccessfulRequests)
		fmt.Fprintf(outputWriter, "Total Failed Requests: %d\n", totalFailedRequests)
		fmt.Fprintf(outputWriter, "First Failed Request: %s\n", firstFailedRequestDetails)
		fmt.Fprintf(outputWriter, "Last Failed Request: %s\n", lastFailedRequestDetails)
	}
}
