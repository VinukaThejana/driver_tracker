set dotenv-load

default:
  @just --choose

# Connect with the Redis database
redis:
  iredis --url $(echo $REDIS_DB_URL)
