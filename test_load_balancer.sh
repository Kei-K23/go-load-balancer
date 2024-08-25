#!/bin/bash

# URL of your load balancer
url="http://localhost:8080"

# Number of requests to send
requests=100

# Log the start time
start_time=$(date +%s)

echo "Sending $requests requests to $url..."

# Send requests in parallel using &
for i in $(seq 1 $requests); do
    curl -s $url &
done

# Wait for all background jobs to finish
wait

# Log the end time
end_time=$(date +%s)

# Calculate the total time taken
total_time=$((end_time - start_time))

echo "Sent $requests requests to $url"
echo "Total time: $total_time seconds"
