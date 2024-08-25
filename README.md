# Go load balancer

Creating a full-featured load balancer in Go that can distribute HTTP requests to multiple backend servers using weighted round-robin algorithm, Circuit Breaker and other. This project uses only `Go standard library`.

## Features

- **Server Pool**: Maintains a list of backend servers.
- **Weighted Round-Robin Algorithm**: Selects servers based on dynamically adjusted weights.
- **Health Checks**: Periodically checks the health of each server and updates their status.
- **Circuit Breaker**: Protects failing servers from being overloaded by temporarily removing them from the selection pool.
- **Dynamic Weight Adjustment**: Adjusts server weights based on performance, including response time and failure count.

## Project Structure

- `main.go`: Entry point for the load balancer server.
- `load-balancer/main.go`: Contains the core load balancer logic.
- `logger/main.go`: Custom logger implementation for consistent logging throughout the application.
- `server*/` : Backend servers

## Getting Started

### Prerequisites

- Go 1.23.0 (1.20+)
- A set of backend servers to balance. For this example, servers are expected to be running on `http://localhost:8081`, `http://localhost:8082`, and `http://localhost:8083` (You can add more servers depending on what your need).

### Installation

1. **Clone the repository:**

```sh
git clone https://github.com/Kei-K23/go-load-balancer
cd go-load-balancer
```

2. **Run the Load Balancer:**

```sh
go run main.go
```

3. **Run backend servers**

```sh
# navigate to the server folder
cd server1
# run Go file
go run main.go
# You can run backend servers as you need
```

4. **Stop the Load Balancer:**

Press `Ctrl+C` in the terminal or send a termination signal to gracefully shut down the load balancer.

### Adding Backend servers

I write ready to use shell script to generate `Go backend servers` in the project root path. All you need to do is adding the backend server address to main entry point for load balancer.

1. **Run add server shell script in the root terminal to create new server**

```sh
# Make sure server name should be in the format of name+number e.g(server5, server6 up to server9)
./add_server.sh server4
```

2. **Add newly added server address to load balancer**

```go
// /main.go

// This is predefine backend servers
serverPool.AddServer(loadbalancer.NewServer("http://localhost:8081"))
serverPool.AddServer(loadbalancer.NewServer("http://localhost:8082"))
serverPool.AddServer(loadbalancer.NewServer("http://localhost:8083"))
// e.g. serverPool.AddServer(loadbalancer.NewServer("http://localhost:8084"))
// add more backend servers here
```

### Testing the load balancer and backend servers

To perform load-testing, you can use any load-testing tool that you want. In my case, I use `JMeter` and prepare the test plan.

```sh
# Run this command in the root terminal
jmeter -n -t /Your-Path/go-load-balancer/load-testing.jmx -l /Your-Path/go-load-balancer/results.jtl -Jproperty_name=value

# e.g for my case jmeter -n -t /Users/user/Desktop/projects/go-projects/go-load-balancer/load-testing.jmx -l /Users/user/Desktop/projects/go-projects/go-load-balancer/results.jtl -Jproperty_name=value
```

## Todo

- [ ] Publish load-balancer as `Go package`. (I will publish as package only when the code meet the requirement as actual real-world load balancer)
- [ ] Add request filtering and banning (To be as like API Gateway but API Gateway and Load balancer are different topic)
- [ ] Add failure recovery
- [ ] Add testing
- [ ] Add more features, improve performance

## Contributing

Contributions are welcome! Please fork the repository and open a pull request with your changes.

## License

This project is licensed under the [Apache-2.0 License](/LICENSE). See the LICENSE file for details.
