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
    cd metrics
    ```

2. Build and run the server:

    ```bash
    go run ./server/server
    ```

3. The server will start listening on port `8080`.

### Prometheus Instrumentation

The server is instrumented with Prometheus metrics to track request counts and durations. These metrics can be scraped by Prometheus and visualized using Grafana. To view the metrics:

- Install Prometheus.
- Ensure Prometheus is configured to scrape metrics from the Metrics Server. This can typically be done by adding a scrape configuration in Prometheus configuration file.
- Access Prometheus dashboard on `localhost:9090` to view and query the collected metrics.

## Client

### Configuration

Before running the client, ensure that you have the necessary configuration files:

- `./client/certs/ca.crt`: CA certificate file for TLS encryption.
- `./client/certs/client.crt`: Client certificate file for TLS encryption.
- `./client/certs/client.key`: Client private key file for TLS encryption.

### Running the Client

1. Navigate to the client directory:

    ```bash
    cd metrics
    ```

2. Build and run the client:

    ```bash
    go run ./client/client.go -filename <path_to_request_json> -duration <load_test_duration_seconds> -concurrent <num_concurrent_requests>
    ```

    Replace `<path_to_request_json>` with the path to the JSON file containing the request data, `<load_test_duration_minutes>` with the duration of the load test in minutes, and `<num_concurrent_requests>` with the number of concurrent requests to be made.

3. The client will send requests to the server and display statistics after the test duration.

## Cached Requests

The Metrics Server employs a cache mechanism to store the last 10 successful and error responses. This cache is implemented using a circular queue data structure, providing efficient memory utilization and optimized access to the most recent responses.

### Circular Queue Implementation

The circular queue is implemented in the server codebase to efficiently manage the cache. Here's how it works:

- **Initialization**: The circular queue is initialized with a fixed size, which in this case is set to 10 to store the last 10 responses.

- **Enqueue Operation**: When a new request is received and processed successfully or results in an error, the corresponding response along with its timestamp is enqueued into the circular queue.

- **Efficient Memory Utilization**: The circular queue ensures efficient memory utilization by reusing memory slots. As new responses are enqueued, if the queue is full, it overwrites the oldest response, thus maintaining a fixed memory footprint.

- **Head and Tail Pointers**: The circular queue maintains two pointers, namely the head and tail pointers. These pointers keep track of the starting and ending positions of the queue, respectively.

- **Optimized Access**: The circular queue allows for optimized access to the most recent responses. By keeping track of the head and tail pointers, it facilitates constant-time access to the first and last elements of the queue.

### Benefits

The use of a circular queue for caching offers several advantages:

1. **Constant-Time Operations**: Enqueueing and dequeueing operations have constant time complexity, ensuring efficient processing of requests.

2. **Fixed Memory Footprint**: The queue size remains constant, resulting in predictable memory consumption, which is crucial for resource-constrained environments.

3. **Efficient Memory Utilization**: Memory slots are reused, preventing memory fragmentation and optimizing memory usage.

4. **Optimized Access**: Access to the most recent responses is optimized through the head and tail pointers, facilitating quick retrieval of cached data.

### Code Reference

Below is a snippet of the circular queue implementation used in the server codebase:

```go
// CircularQueue represents a circular queue data structure
type CircularQueue struct {
    queue []CachedRequest
    size  int
    head  int
    tail  int
    mutex sync.Mutex
}

// NewCircularQueue initializes a new circular queue with the specified size
func NewCircularQueue(size int) *CircularQueue {
    return &CircularQueue{
        queue: make([]CachedRequest, size),
        size:  size,
    }
}

// Enqueue adds a new request to the circular queue
func (q *CircularQueue) Enqueue(request CachedRequest) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    // Enqueue operation implementation
    // ...
}

## License

This project is licensed under the [MIT License](LICENSE).
