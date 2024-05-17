package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v11 "go.opentelemetry.io/proto/otlp/common/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	v12 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"log"
	"metrics/server/pb/pv"
	"sync"
)

const (
	numTotalRequests      = 1
	numConcurrentRequests = 1
	serverAddress         = "localhost:8080"
)

func getVersion(client pv.VersionServiceClient) {
	resp, err := client.GetVersion(context.Background(), &emptypb.Empty{})
	log.Printf("Response: %v", resp)
	if err != nil {
		log.Printf("Failed to get response: %v", err)
	}
}

func getGoodRequest() *pb.ExportMetricsServiceRequest {
	req := &pb.ExportMetricsServiceRequest{
		ResourceMetrics: []*v1.ResourceMetrics{
			{
				Resource: &v12.Resource{
					Attributes: []*v11.KeyValue{
						{
							Key:   "service.name",
							Value: &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: "my-service"}},
						},
					},
					DroppedAttributesCount: 0,
				},
				ScopeMetrics: []*v1.ScopeMetrics{
					{
						Scope: &v11.InstrumentationScope{
							Name: "scope-name",
						},
						Metrics: []*v1.Metric{
							{
								Name:        "metric_name",
								Description: "Description of the metric",
								Unit:        "unit",
								Data: &v1.Metric_Gauge{
									Gauge: &v1.Gauge{},
								},
								Metadata: []*v11.KeyValue{
									{
										Key:   "metadata_key",
										Value: &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: "metadata_value"}},
									},
								},
							},
						},
						SchemaUrl: "schema_url",
					},
				},
				SchemaUrl: "schema_url",
			},
		},
	}
	return req
}
func getInvalidRequest() *pb.ExportMetricsServiceRequest {
	req := &pb.ExportMetricsServiceRequest{
		ResourceMetrics: []*v1.ResourceMetrics{
			{
				Resource: &v12.Resource{
					Attributes: []*v11.KeyValue{
						{
							Key:   "service.name",
							Value: &v11.AnyValue{Value: &v11.AnyValue_BoolValue{BoolValue: true}}, // Invalid type
						},
					},
					DroppedAttributesCount: 0,
				},
				ScopeMetrics: []*v1.ScopeMetrics{
					{
						Scope: &v11.InstrumentationScope{
							Name: "scope-name",
						},
						Metrics: []*v1.Metric{
							{
								Name:        "metric_name",
								Description: "Description of the metric",
								Unit:        "unit",
								Data: &v1.Metric_Gauge{
									Gauge: &v1.Gauge{},
								},
								Metadata: []*v11.KeyValue{
									{
										Key:   "metadata_key",
										Value: &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: "metadata_value"}},
									},
								},
							},
						},
						SchemaUrl: "schema_url",
					},
				},
				SchemaUrl: "schema_url",
			},
		},
	}

	return req
}

func sendRequests(client pb.MetricsServiceClient, wg *sync.WaitGroup, requestJson string) {
	defer wg.Done()

	for i := 0; i < numTotalRequests/numConcurrentRequests; i++ {
		req := getGoodRequest()
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

	// Send invalid requests
	for _, invalidRequestFile := range invalidRequestFiles {
		invalidRequest, err := readJSONFile(invalidRequestFile)

		if err != nil {
			log.Fatalf("Failed to read invalid request file (%s): %v", invalidRequestFile, err)
		}

		for i := 0; i < numConcurrentRequests; i++ {
			go sendRequests(client, &wg, invalidRequest)
		}
	}

	wg.Wait()
}
