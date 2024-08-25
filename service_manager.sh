#!/bin/bash

# Set the Go project root directory
PROJECT_ROOT=$(pwd)
BIN_DIR="$PROJECT_ROOT/bin"
LOG_DIR="$PROJECT_ROOT/logs"
SERVICES=("app1" "app2" "app3" "app4")

# Function to build all services
build_services() {
    echo "Building services..."
    mkdir -p $BIN_DIR
    for service in "${SERVICES[@]}"; do
        echo "Building $service..."
        go build -o $BIN_DIR/$service $PROJECT_ROOT/cmd/$service
        if [ $? -ne 0 ]; then
            echo "Build failed for $service!"
            exit 1
        fi
    done
    echo "All services built successfully."
}

# Function to start all services
start_services() {
    echo "Starting services..."
    mkdir -p $LOG_DIR
    for service in "${SERVICES[@]}"; do
        if pgrep -f $BIN_DIR/$service > /dev/null; then
            echo "$service is already running."
        else
            echo "Starting $service..."
            $BIN_DIR/$service > $LOG_DIR/$service.log 2>&1 &
            echo "$service started."
        fi
    done
}

# Function to stop all services
stop_services() {
    echo "Stopping services..."
    for service in "${SERVICES[@]}"; do
        pkill -f $BIN_DIR/$service
        echo "$service stopped."
    done
}

# Function to check the status of all services
status_services() {
    echo "Checking status of services..."
    for service in "${SERVICES[@]}"; do
        if pgrep -f $BIN_DIR/$service > /dev/null; then
            echo "$service is running."
        else
            echo "$service is not running."
        fi
    done
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
