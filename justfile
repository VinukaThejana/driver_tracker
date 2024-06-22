set dotenv-load

default:
  @just --choose

# Connect with the Redis database
redis:
  iredis --url $(echo $REDIS_DB_URL)

env:
  doppler secrets download --no-file --format=env > .env

prd_env:
  bash ./scripts/prd_secrets.sh $DOPPLER_PRD_SERVICE_ACCOUNT $SERVICE_NAME $SERVICE_REGION
