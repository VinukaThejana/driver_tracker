#!/bin/bash

if [ -z "$1" ]; then
  echo "Please provide the appropriate booking ID"
  exit 1
fi

while true; do
  printf '{
    "location": {
        "lat": %.6f,
        "lon": %.6f,
        "status": 2,
        "heading": %.2f,
        "accuracy": %.2f
    }
}' $(awk 'BEGIN{srand(); print rand()*180-90}') \
    $(awk 'BEGIN{srand(); print rand()*360-180}') \
    $(awk 'BEGIN{srand(); print rand()*360}') \
    $(awk 'BEGIN{srand(); print rand()*100}') |
    http --follow --timeout 3600 POST 'http://localhost:8080/stream/add' \
      Content-Type:'application/json' \
      Authorization:"Bearer $1"

  sleep 1
done
