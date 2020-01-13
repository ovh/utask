#!/bin/bash

export PG_USER="test"
export PG_PASSWORD="test"
export PG_PORT="5432"
export PG_DATABASENAME="postgres"

PG_DOCKER_ID=$(docker run -e POSTGRES_USER=$PG_USER -e POSTGRES_PASSWORD=$PG_PASSWORD -d postgres:9.5.3)

echo "Spawned Postgres docker: $PG_DOCKER_ID"

export PG_HOST=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $PG_DOCKER_ID)
export PSQL_BIN="docker exec -i $PG_DOCKER_ID su postgres -c psql"

echo "Wait for database init..."
hack/wait-for-it/wait-for-it.sh -t 120 ${PG_HOST}:${PG_PORT}

(bash hack/test.sh $*)
result=$?

echo "Killing Postgres docker"
docker kill $PG_DOCKER_ID
docker rm -v $PG_DOCKER_ID

exit $result
