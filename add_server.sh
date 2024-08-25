#!/bin/bash

# Check if the user provided a server name
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <server_name>"
    exit 1
fi

SERVER_NAME=$1

# Define a base port number
BASE_PORT=8080

# Extract numeric part from SERVER_NAME using regex
if [[ $SERVER_NAME =~ [0-9]+ ]]; then
    PORT_SUFFIX=${BASH_REMATCH[0]}
else
    echo "Error: SERVER_NAME does not contain a numeric part."
    exit 1
fi

# Calculate the port number dynamically
PORT=$((BASE_PORT + PORT_SUFFIX))

# Set the Go project root directory
PROJECT_ROOT=$(pwd)
BIN_DIR="$PROJECT_ROOT/bin"
SERVICES_FILE="$PROJECT_ROOT/services.list"

# Create the server directory and main.go file
echo "Creating directory for $SERVER_NAME..."
mkdir -p "$PROJECT_ROOT/$SERVER_NAME"

echo "Creating main.go for $SERVER_NAME..."
cat <<EOL > "$PROJECT_ROOT/$SERVER_NAME/main.go"
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "fmt"
)

const (
	PORT        = ":$PORT"
	SERVER_NAME = "$SERVER_NAME"
)

func main() {

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]string{
            "message": "This is health response from " + SERVER_NAME,
            "address": PORT,
        }

        // Set the content type to application/json
        w.Header().Set("Content-Type", "application/json")

        // Encode the response as JSON and write it to the response writer
        if err := json.NewEncoder(w).Encode(response); err != nil {
            http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
            return
        }

        fmt.Println(":::Calling health check route:::")
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Define a response structure with dynamic server name
        response := map[string]string{
            "message": "Hello from " + SERVER_NAME,
            "address": PORT,
        }

        // Set the content type to application/json
        w.Header().Set("Content-Type", "application/json")

        // Encode the response as JSON and write it to the response writer
        if err := json.NewEncoder(w).Encode(response); err != nil {
            http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
            return
        }
    })

    log.Printf("Server name $SERVER_NAME started on %s", PORT)
    log.Fatal(http.ListenAndServe(PORT, nil))
}
EOL

# Add the new server to the SERVICES array in the script
echo "Updating services list..."
if ! grep -q "$SERVER_NAME" "$SERVICES_FILE"; then
    echo "$SERVER_NAME" >> "$SERVICES_FILE"
    echo "$SERVER_NAME added to the services list."
else
    echo "$SERVER_NAME already exists in the services list."
fi

# Ensure the loadbalancer service is present in the SERVICES_FILE
if ! grep -q "loadbalancer" "$SERVICES_FILE"; then
    echo "loadbalancer" >> "$SERVICES_FILE"
    echo "loadbalancer added to the services list."
fi

echo "Setup complete. You can now build and start $SERVER_NAME."
