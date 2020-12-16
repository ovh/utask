#!/bin/bash

export PG_USER="test"
export PG_PASSWORD="test"
export PG_PORT="5432"
export PG_DATABASENAME="test"

PG_DOCKER_ID=$(docker run -v $PWD/sql/schema.sql:/docker-entrypoint-initdb.d/v0001.sql:ro -e POSTGRES_USER=$PG_USER -e POSTGRES_PASSWORD=$PG_PASSWORD -e POSTGRES_DBNAME=$PG_DATABASENAME -d postgres:9.6-alpine)

echo "Spawned Postgres docker: $PG_DOCKER_ID"

export PG_HOST=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $PG_DOCKER_ID)

echo "Wait for database init..."
hack/wait-for-it/wait-for-it.sh -t 120 ${PG_HOST}:${PG_PORT}

(bash hack/test.sh $*)
result=$?

echo "Killing Postgres docker"
docker kill $PG_DOCKER_ID
docker rm -v $PG_DOCKER_ID

exit $result
