package main

import (
    "encoding/json"
    "log"
    "net/http"
)

const PORT = ":8085"

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

    log.Printf("Server server5 started on %s", PORT)
    log.Fatal(http.ListenAndServe(PORT, nil))
}
