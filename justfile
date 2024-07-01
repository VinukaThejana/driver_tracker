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
  #!/bin/bash
  bash scripts/kafka_ui.sh
  docker run -p 9090:8080 -v ./.kafka.yml:/application.yml provectuslabs/kafka-ui:latest

stream_add_stg booking_id:
  go run tests/stream/generator.go stg {{booking_id}}
  bash tests/stream/add.sh $STG_URL $BOOKING_TOKEN

stream_add_prd booking_id:
  go run tests/stream/generator.go prd {{booking_id}}
  bash tests/stream/add.sh $PRD_URL $BOOKING_TOKEN

stream_add_local booking_id:
  go run tests/stream/generator.go dev {{booking_id}}
  bash tests/stream/add.sh http://localhost:8080 $BOOKING_TOKEN
