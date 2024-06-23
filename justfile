set dotenv-load

default:
  @just --choose

run:
  go run cmd/main.go

redis:
  iredis --url $(echo $REDIS_DB_URL)

env:
  doppler secrets download --no-file --format=env > .env

prd_env:
  bash ./scripts/secrets.sh $DOPPLER_PRD_SERVICE_ACCOUNT $SERVICE_NAME $SERVICE_REGION

stg_env:
  bash ./scripts/secrets.sh $DOPPLER_STG_SERVICE_ACCOUNT $SERVICE_STG_NAME $SERVICE_STG_REGION

kafka_ui:
  docker run -p 9090:8080 -v ./.kafka.yml:/application.yml provectuslabs/kafka-ui:latest
