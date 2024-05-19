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
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// serverAddress is the address of the server to connect to.
const serverAddress = "localhost:8080"

// Variables to track total requests, failures, successes, and details of the first failed/successful requests.
var (
	// totalRequests keeps track of the total number of requests made.
	totalRequests = 0
	// totalFailedRequests keeps track of the total number of failed requests.
	totalFailedRequests = 0
	// totalSuccessfulRequests keeps track of the total number of successful requests.
	totalSuccessfulRequests = 0
	// firstFailedRequestDetails stores the details of the first failed request.
	firstFailedRequestDetails string
	// firstSuccessfulRequest stores the details of the first successful request.
	firstSuccessfulRequest string
)

func sendRequests(client pb.MetricsServiceClient, requestJson string, duration time.Duration, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	req := &pb.ExportMetricsServiceRequest{}
	if err := jsonpb.UnmarshalString(requestJson, req); err != nil {
		log.Printf("Error unmarshalling request: %v", err)
		return
	}

	timer := time.NewTimer(duration)
	for {
		select {
		case <-ctx.Done():
			// If context is canceled, return immediately
			return
		case <-timer.C:
			// If time is up, return
			return
		default:
			resp, err := client.Export(context.Background(), req)
			totalRequests++

			if err != nil {
				// Case 1: Error returned by the Export call
				totalFailedRequests++
				if firstFailedRequestDetails == "" {
					firstFailedRequestDetails = fmt.Sprintf("Failed request details: %v", req)
				}
				log.Printf("Failed to get response: %v", err)
			} else if resp.GetPartialSuccess() != nil {
				// Case 2: Partial success (some data points were rejected)
				totalFailedRequests += int(resp.GetPartialSuccess().GetRejectedDataPoints())
				if firstFailedRequestDetails == "" {
					firstFailedRequestDetails = fmt.Sprintf("Partial error request details: %v", req)
				}
				log.Printf("Partial success: %v", resp.GetPartialSuccess())
			} else {
				// Case 3: Successful response
				totalSuccessfulRequests++
				if firstSuccessfulRequest == "" {
					firstSuccessfulRequest = fmt.Sprintf("Successful request details: %v", req)
				}
				log.Printf("Successful response: %v", resp)
			}
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
	numConReq := flag.Int("concurrent", 1, "Number of concurrent requests")

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

	requestJson, err := readJSONFile(*filename)
	if err != nil {
		log.Fatalf("Failed to read request file: %v", err)
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigCh:
				// If canceled by user, cancel the context
				cancel()
				return
			default:
				wg.Add(*numConReq)
				go sendRequests(client, requestJson, time.Duration(*duration)*time.Second, &wg, ctx)
				wg.Wait() // Wait for all requests to finish
			}
		}
	}()

	// Wait until the context is canceled
	<-ctx.Done()

	// Print statistics after the test is canceled
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Total Successful Requests: %d\n", totalSuccessfulRequests)
	fmt.Printf("Total Failed Requests: %d\n", totalFailedRequests)
	fmt.Println(firstFailedRequestDetails)
	fmt.Println(firstSuccessfulRequest)
}
