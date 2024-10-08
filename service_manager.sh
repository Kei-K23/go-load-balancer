#!/bin/bash

# Set the Go project root directory
PROJECT_ROOT=$(pwd)
BIN_DIR="$PROJECT_ROOT/bin"
LOG_DIR="$PROJECT_ROOT/logs"
SERVICES_FILE="$PROJECT_ROOT/services.list"

# Function to build all services
build_services() {
    echo "Building services..."
    mkdir -p "$BIN_DIR"
    while IFS= read -r service; do
        if [ "$service" == "loadbalancer" ]; then
            # Build the load balancer which is located in the root directory
            echo "Building load balancer..."
            go build -o "$BIN_DIR/$service" "$PROJECT_ROOT/main.go"
        else
            # Build the servers which are located in their respective folders
            echo "Building $service..."
            go build -o "$BIN_DIR/$service" "$PROJECT_ROOT/$service/main.go"
        fi
        
        if [ $? -ne 0 ]; then
            echo "Build failed for $service!"
            exit 1
        fi
    done < "$SERVICES_FILE"
    echo "All services built successfully."
}

# Function to start all services
start_services() {
    echo "Starting services..."
    mkdir -p "$LOG_DIR"
    while IFS= read -r service; do
        if pgrep -f "$BIN_DIR/$service" > /dev/null; then
            echo "$service is already running."
        else
            echo "Starting $service..."
            "$BIN_DIR/$service" > "$LOG_DIR/$service.log" 2>&1 &
            echo "$service started."
        fi
    done < "$SERVICES_FILE"
}

# Function to stop all services
stop_services() {
    echo "Stopping services..."
    while IFS= read -r service; do
        pkill -f "$BIN_DIR/$service"
        echo "$service stopped."
    done < "$SERVICES_FILE"
}

# Function to check the status of all services
status_services() {
    echo "Checking status of services..."
    while IFS= read -r service; do
        if pgrep -f "$BIN_DIR/$service" > /dev/null; then
            echo "$service is running."
        else
            echo "$service is not running."
        fi
    done < "$SERVICES_FILE"
}

# Display usage information
usage() {
    echo "Usage: $0 {build|start|stop|status}"
    exit 1
}

# Check the first argument to the script
case "$1" in
    build)
        build_services
        ;;
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    status)
        status_services
        ;;
    *)
        usage
        ;;
esac
