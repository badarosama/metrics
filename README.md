# Metrics Server and Client

This repository contains a gRPC-based Metrics Server and Client for exporting metrics data. The server receives metrics data and processes it, while the client sends requests to the server to export metrics data.

## Components

1. **Server**: The Metrics Server implemented in Go listens for incoming gRPC requests, processes the metrics data, and returns appropriate responses. It includes features like TLS encryption, Prometheus metrics endpoint, and logging.

2. **Client**: The Metrics Client implemented in Go sends requests to the Metrics Server to export metrics data. It includes features like concurrent request handling, TLS encryption, and JSON request file support.

## Requirements

To run the Metrics Server and Client, ensure you have the following installed:

- Go (Programming Language)
- gRPC (Google's Remote Procedure Call framework)
- Prometheus (Metrics monitoring and alerting toolkit)

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/metrics-server-client.git
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

## Server

### Configuration

Before running the server, ensure that you have the necessary configuration files:

- `./server/config.yaml`: Configuration file for server logging.

### Running the Server

1. Navigate to the server directory:

    ```bash
    cd server
    ```

2. Build and run the server:

    ```bash
    go run server.go
    ```

3. The server will start listening on port `8080`.

## Client

### Configuration

Before running the client, ensure that you have the necessary configuration files:

- `./client/certs/ca.crt`: CA certificate file for TLS encryption.
- `./client/certs/client.crt`: Client certificate file for TLS encryption.
- `./client/certs/client.key`: Client private key file for TLS encryption.

### Running the Client

1. Navigate to the client directory:

    ```bash
    cd client
    ```

2. Build and run the client:

    ```bash
    go run client.go -filename <path_to_request_json> -duration <load_test_duration_minutes> -concurrent <num_concurrent_requests>
    ```

    Replace `<path_to_request_json>` with the path to the JSON file containing the request data, `<load_test_duration_minutes>` with the duration of the load test in minutes, and `<num_concurrent_requests>` with the number of concurrent requests to be made.

3. The client will send requests to the server and display statistics after the test duration.

## Performance Metrics (NA)

Performance metrics for load testing are not available at the moment.

## License

This project is licensed under the [MIT License](LICENSE).
