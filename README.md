Metrics Server and Client
This repository contains a gRPC-based Metrics Server and Client for exporting metrics data. The server receives metrics data and processes it, while the client sends requests to the server to export metrics data.

Components
Server: The Metrics Server implemented in Go listens for incoming gRPC requests, processes the metrics data, and returns appropriate responses. It includes features like TLS encryption, Prometheus metrics endpoint, and logging.

Client: The Metrics Client implemented in Go sends requests to the Metrics Server to export metrics data. It includes features like concurrent request handling, TLS encryption, and JSON request file support.

Requirements
To run the Metrics Server and Client, ensure you have the following installed:

Go (Programming Language)
gRPC (Google's Remote Procedure Call framework)
Prometheus (Metrics monitoring and alerting toolkit)
Installation
Clone the repository:

bash
Copy code
git clone https://github.com/yourusername/metrics-server-client.git
Install dependencies:

bash
Copy code
go mod tidy
Server
Configuration
Before running the server, ensure that you have the necessary configuration files:

./server/config.yaml: Configuration file for server logging.
Running the Server
Navigate to the server directory:

bash
Copy code
cd server
Build and run the server:

bash
Copy code
go run server.go
The server will start listening on port 8080.

Client
Configuration
Before running the client, ensure that you have the necessary configuration files:

./client/certs/ca.crt: CA certificate file for TLS encryption.
./client/certs/client.crt: Client certificate file for TLS encryption.
./client/certs/client.key: Client private key file for TLS encryption.
Running the Client
Navigate to the client directory:

bash
Copy code
cd client
Build and run the client:

bash
Copy code
go run client.go -filename <path_to_request_json> -duration <load_test_duration_minutes> -concurrent <num_concurrent_requests>
Replace <path_to_request_json> with the path to the JSON file containing the request data, <load_test_duration_minutes> with the duration of the load test in minutes, and <num_concurrent_requests> with the number of concurrent requests to be made.

The client will send requests to the server and display statistics after the test duration.

Performance Metrics (NA)
Performance metrics for load testing are not available at the moment.

License
This project is licensed under the MIT License.
