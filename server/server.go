package main

import (
	"crypto/tls"
	"crypto/x509"
	pb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"metrics/server/pb/pv"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
)

var (
	buildTime int
	gitCommit string
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	pv.UnimplementedVersionServiceServer
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

// init initializes the build time and git commit variables.
func init() {
	buildTime = time.Now().UTC().Second()
	gitCommit, _ = getGitCommit()
}

func main() {
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

// getGitCommit retrieves the git commit hash of the current HEAD.
func getGitCommit() (string, error) {
	// Run git command to get the commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = getModuleRoot()
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// getModuleRoot retrieves the root directory of the Go module.
func getModuleRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for !isModuleRoot(cwd) {
		cwd = filepath.Dir(cwd)
	}
	return cwd
}

// isModuleRoot checks if the current directory is the root of the Go module.
func isModuleRoot(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}
