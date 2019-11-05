#!/bin/bash

export PGUSER=test
export PGPASSWORD=test
export PGPORT="5433"
export PGHOST="127.0.0.1"

PG_DOCKER_ID=$(docker run -e POSTGRES_USER=$PGUSER -e POSTGRES_PASSWORD=$PGPASSWORD -p$PGPORT:5432 -d postgres:9.5.3)

echo "Spawned Postgres docker: $PG_DOCKER_ID"

export PSQL_BIN="docker exec -i $PG_DOCKER_ID su postgres -c psql"

echo "Wait for init..."
sleep 10

bash test.sh $*

echo "Killing Postgres docker"
docker kill $PG_DOCKER_ID
docker rm -v $PG_DOCKER_ID
