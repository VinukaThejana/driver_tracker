#!/bin/bash

MIN_LATITUDE=34.0
MAX_LATITUDE=36.0
MIN_LONGITUDE=-118.0
MAX_LONGITUDE=-116.0

if [ -z "$1" ]; then
	echo "Please provide the appropriate URL"
	exit 1
fi
if [ -z "$2" ]; then
	echo "Please provide the booking token"
	exit 1
fi

# URL
URL="$1/stream/add"
# Booking Token
BEARER_TOKEN="$2"
BASE_DIR="tests/stream/data"

declare -a latitudes
declare -a longitudes
declare -a headings

while IFS= read -r line; do
	latitudes+=("$line")
done <"$BASE_DIR/lat"
while IFS= read -r line; do
	longitudes+=("$line")
done <"$BASE_DIR/lon"
while IFS= read -r line; do
	headings+=("$line")
done <"$BASE_DIR/headings"

post_coordinate() {
	local lat=$1
	local lon=$2
	local heading=$3
	local body=$(printf '{"lat": %.12f, "lon": %.15f, "heading": %.8f}' "$lat" "$lon" "$heading")
	RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$URL" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $BEARER_TOKEN" \
		-d "$body")

	if [ "$RESPONSE" -eq 200 ]; then
		echo "$body"
	else
		echo "Failed to post coordinate: $RESPONSE"
	fi
}

for i in "${!latitudes[@]}"; do
	post_coordinate "${latitudes[$i]}" "${longitudes[$i]}" "${headings[$i]}"
	sleep 1
done
