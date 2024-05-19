package main

import (
	"crypto/tls"
	"crypto/x509"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"metrics/server/pb/pv"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	pathOfConfigFile = "./server/config.yaml"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	pv.UnimplementedVersionServiceServer
	lastSuccessfulRequests *CircularQueue
	lastErrorRequests      *CircularQueue
	cacheMutex             sync.Mutex
	logger                 *zap.Logger
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
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Human-readable timestamps
	cfg.OutputPaths = []string{"stdout", "./logs/server.log"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}

	logger, err := cfg.Build(zap.AddCallerSkip(1)) // Skip the zap library's frames in the call stack
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func getServerCertAndPool() (tls.Certificate, *x509.CertPool) {
	// load certs
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

	return serverCert, certPool
}

func configureLogger() *zap.Logger {
	// Ensure the logs directory exists
	err := os.MkdirAll("./logs", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Initialize logger based on configuration
	loggerConfig, _ := loadConfig(pathOfConfigFile)
	logger, err := initLogger(loggerConfig)
	if err != nil {
		panic(err)
	}

	return logger
}

func main() {
	// setup logger
	logger := configureLogger()
	defer logger.Sync()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		//log.Fatalf("Failed to listen: %v", err)
		logger.Error("Failed to listen %v", zap.Error(err))
	}

	serverCert, certPool := getServerCertAndPool()
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
		grpc.ChainUnaryInterceptor(UnaryInterceptorPrometheus,
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(),
			)),
	)
	// Initialize the server struct with the logger
	srv := &server{
		logger:                 logger,
		lastErrorRequests:      NewCircularQueue(10),
		lastSuccessfulRequests: NewCircularQueue(10),
	}

	pb.RegisterMetricsServiceServer(s, srv)
	pv.RegisterVersionServiceServer(s, srv)
	reflection.Register(s)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":9091", nil)
	}()

	//log.Printf("Server is listening on port 8080...")
	logger.Info("Server is listening on port 8080...")
	if err := s.Serve(listener); err != nil {
		//log.Fatalf("Failed to serve: %v", err)
		logger.Error("Failed to serve: %v", zap.Error(err))
	}
}
