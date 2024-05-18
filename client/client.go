package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"log"
	"metrics/server/pb/pv"
	"sync"
)

const (
	numTotalRequests      = 1000000
	numConcurrentRequests = 1
	serverAddress         = "localhost:8080"
)

var totalRequests = 0
var totalFailedRequests = 0
var totalSuccessfulRequests = 0

func getVersion(client pv.VersionServiceClient) {
	resp, err := client.GetVersion(context.Background(), &emptypb.Empty{})
	log.Printf("Response: %v", resp)
	if err != nil {
		log.Printf("Failed to get response: %v", err)
	}
}

func sendRequests(client pb.MetricsServiceClient, wg *sync.WaitGroup, requestJson string) {
	defer wg.Done()

	req := &pb.ExportMetricsServiceRequest{}
	// Unmarshal JSON into the ExportMetricsServiceRequest protobuf struct
	if err := json.Unmarshal([]byte(requestJson), req); err != nil {
		fmt.Println("Error:", err)
		return
	}

	//log.Printf("request json: %v\n", *req)
	for i := 0; i < numTotalRequests/numConcurrentRequests; i++ {
		resp, err := client.Export(context.Background(), req)
		log.Printf("Response: %v", resp)
		if err != nil {
			log.Printf("Failed to get response: %v", err)
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
	// read ca's cert
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal(caCert)
	}
	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}
	//read client cert
	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatal(err)
	}

	// set config of tls credential
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	tlsCredential := credentials.NewTLS(config)
	// Dial the server with TLS credentials, skipping hostname verification
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsServiceClient(conn)

	versionClient := pv.NewVersionServiceClient(conn)

	getVersion(versionClient)
	var wg sync.WaitGroup

	// Read valid request
	validRequest, err := readJSONFile("./client/request_jsons/valid_request.json")
	if err != nil {
		log.Fatalf("Failed to read valid request file: %v", err)
	}

	invalidRequestFiles := []string{
		//"invalid_request.json",
		//"invalid_request_2.json",
		//"invalid_request_3.json",
	}

	wg.Add(numConcurrentRequests * (1 + len(invalidRequestFiles)))

	// Send valid requests
	for i := 0; i < numConcurrentRequests; i++ {
		go sendRequests(client, &wg, validRequest)
	}

	//// Send invalid requests
	//for _, invalidRequestFile := range invalidRequestFiles {
	//	invalidRequest, err := readJSONFile(invalidRequestFile)
	//
	//	if err != nil {
	//		log.Fatalf("Failed to read invalid request file (%s): %v", invalidRequestFile, err)
	//	}
	//
	//	for i := 0; i < numConcurrentRequests; i++ {
	//		go sendRequests(client, &wg, invalidRequest)
	//	}
	//}

	wg.Wait()
	println("Execution finished %s, %s, %s", totalRequests, totalSuccessfulRequests, totalFailedRequests)
}
