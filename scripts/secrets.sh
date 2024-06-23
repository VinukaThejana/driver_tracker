#Get the production secrets from Doppler and update the cloud run account environmental varables
#!/bin/bash

if [ -z "$1" ]; then
	echo "Please provide the Doppler PRD service account secret to update the env variables"
	exit 1
fi
if [ -z "$2" ]; then
	echo "Please provide the Google run service name"
	exit 1
fi
if [ -z "$3" ]; then
	echo "Please provide the Google run service region"
	exit 1
fi

secret=$1
service_name=$2
service_regioin=$3

env=$(
	curl --request GET \
		--url 'https://api.doppler.com/v3/configs/config/secrets/download?format=env' \
		--header "Authorization: Bearer $secret"
)

escape_value() {
	echo "$1" | awk '{gsub(/"/, "\\\""); gsub(/\n/, "\\n")}1'
}

result=""

while IFS= read -r line; do
	key=$(echo "$line" | cut -d'=' -f1)
	value=$(echo "$line" | cut -d'=' -f2- | tr -d '"')
	escaped_value=$(escape_value "$value")

	if [ -n "$key" ]; then
		if [ -z "$result" ]; then
			result="${key}=${escaped_value}"
		else
			result="${result},${key}=${escaped_value}"
		fi
	fi
done <<<"$env"

gcloud run services update $service_name --update-env-vars $result --region $service_regioin
