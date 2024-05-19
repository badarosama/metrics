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
	"sync"
	"time"
)

const (
	serverAddress         = "localhost:8080"
	numConcurrentRequests = 1
)

var (
	totalRequests             = 0
	totalFailedRequests       = 0
	totalSuccessfulRequests   = 0
	firstFailedRequestDetails string
	firstSuccessfulRequest    string
)

func getVersion(client pv.VersionServiceClient) {
	resp, err := client.GetVersion(context.Background(), &emptypb.Empty{})
	log.Printf("Response: %s", resp.GitCommitSha)
	log.Printf("Response: %v", resp.BuildTimestamp)
	if err != nil {
		log.Printf("Failed to get response: %v", err)
	}
}

func sendRequests(client pb.MetricsServiceClient, wg *sync.WaitGroup, requestJson string) {
	defer wg.Done()

	req := &pb.ExportMetricsServiceRequest{}
	if err := jsonpb.UnmarshalString(requestJson, req); err != nil {
		log.Printf("Error unmarshalling request: %v", err)
		return
	}

	for i := 0; i < numConcurrentRequests; i++ {
		resp, err := client.Export(context.Background(), req)
		log.Printf("resp: %v", resp)
		totalRequests++
		if err != nil || (resp != nil && resp.GetPartialSuccess() != nil) {
			totalFailedRequests++
			if firstFailedRequestDetails == "" {
				firstFailedRequestDetails = fmt.Sprintf("Failed request details: %v", req)
			}
			log.Printf("Failed to get response: %v", err)
		} else {
			totalSuccessfulRequests++
			if firstSuccessfulRequest == "" {
				firstSuccessfulRequest = fmt.Sprintf("Successful request details: %v", req)
			}
			log.Printf("Response: %v", resp)
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

func main() {
	filename := flag.String("filename", "", "Path to the JSON file containing the request data")
	duration := flag.Int("duration", 5, "Duration of the load test in minutes")

	flag.Parse()

	if *filename == "" {
		log.Fatal("Please provide the filename parameter")
	}

	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal(caCert)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatal(err)
	}

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
	//versionClient := pv.NewVersionServiceClient(conn)
	//getVersion(versionClient)
	var wg sync.WaitGroup

	requestJson, err := readJSONFile(*filename)
	if err != nil {
		log.Fatalf("Failed to read request file: %v", err)
	}

	testDuration := time.Duration(*duration) * time.Minute
	startTime := time.Now()

	for {
		if time.Since(startTime) >= testDuration {
			break
		}

		wg.Add(numConcurrentRequests)
		go sendRequests(client, &wg, requestJson)
		time.Sleep(1 * time.Second) // Adjust the delay between requests as needed
	}

	wg.Wait()

	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Total Successful Requests: %d\n", totalSuccessfulRequests)
	fmt.Printf("Total Failed Requests: %d\n", totalFailedRequests)
	fmt.Println(firstFailedRequestDetails)
	fmt.Println(firstSuccessfulRequest)
}
