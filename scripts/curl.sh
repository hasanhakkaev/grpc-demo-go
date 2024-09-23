#!/bin/bash

# Loop 10000 times
for i in {1..10000}
do
  # Generate a random type between 0 and 9
  type=$(( RANDOM % 10 ))

  # Generate a random value between 0 and 99
  value=$(( RANDOM % 100 ))

  # Run grpcurl with the random values
  grpcurl -d "{\"task\":{\"type\":\"$type\",\"value\":\"$value\",\"state\":\"RECEIVED\"}}" \
  -plaintext 0.0.0.0:3030 api.tasks.v1.TaskService.CreateTask

  # Optionally add a small sleep time between iterations to avoid overwhelming the server
  # sleep 0.1
done
