#!/bin/bash

export PG_USER="test"
export PG_PASSWORD="test"
export PG_PORT="5433"
export PG_HOST="127.0.0.1"
export PG_DATABASENAME="postgres"

PG_DOCKER_ID=$(docker run -e POSTGRES_USER=$PG_USER -e POSTGRES_PASSWORD=$PG_PASSWORD -p$PG_PORT:5432 -d postgres:9.5.3)

echo "Spawned Postgres docker: $PG_DOCKER_ID"

export PSQL_BIN="docker exec -i $PG_DOCKER_ID su postgres -c psql"

echo "Wait for init..."
sleep 10

(bash hack/test.sh $*)
result=$?

echo "Killing Postgres docker"
docker kill $PG_DOCKER_ID
docker rm -v $PG_DOCKER_ID

exit $result
