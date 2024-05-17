package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/spf13/viper"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"metrics/server/pb/pv"
	"net"
	"sync"
	"time"
)

const (
	pathOfConfig = "./server/config.yaml"
)

var logger *zap.Logger

type cachedRequest struct {
	Request   *pb.ExportMetricsServiceRequest
	Timestamp time.Time
}

type server struct {
	pb.UnimplementedMetricsServiceServer
	pv.UnimplementedVersionServiceServer
	lastSuccessfulRequests []*cachedRequest
	lastErrorRequests      []*cachedRequest
	cacheMutex             sync.Mutex
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

func loadConfig(path string) (LoggerConfig, error) {
	// Load configuration from file
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	// Unmarshal configuration into struct
	var loggerConfig LoggerConfig
	if err := viper.Unmarshal(&loggerConfig); err != nil {
		panic(err)
	}

	return loggerConfig, nil
}

func initLogger(config LoggerConfig) (*zap.Logger, error) {
	var level zap.AtomicLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func main() {
	// Initialize logger based on configuration
	loggerConfig, _ := loadConfig(pathOfConfig)
	logger, err := initLogger(loggerConfig)
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	caPem, err := ioutil.ReadFile("certs/ca.crt")
	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal(err)
	}
	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")

	if err != nil {
		log.Fatal(err)
	}

	// configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	tlsCredentials := credentials.NewTLS(conf)
	s := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)

	pb.RegisterMetricsServiceServer(s, &server{})
	pv.RegisterVersionServiceServer(s, &server{})
	reflection.Register(s)

	log.Printf("Server is listening on port 8080...")
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
