#!/bin/bash

export PG_USER="test"
export PG_PASSWORD="test"
export PG_PORT="5432"
export PG_DATABASENAME="test"
export PG_SCHEMA=${PG_SCHEMA:-${PWD}/sql/schema.sql}

ENV_FILE=$(readlink -f $(dirname ${0}))/test.env

if [ -f ${ENV_FILE} ]; then
    source ${ENV_FILE}
fi

DOCKER_OPTS=$(cat <<EOF
-v ${PG_SCHEMA}:/docker-entrypoint-initdb.d/v0001.sql:ro
-e POSTGRES_USER=${PG_USER}
-e POSTGRES_PASSWORD=${PG_PASSWORD}
-e POSTGRES_DBNAME=${PG_DATABASENAME}
EOF
)

if [ ! -z "${PG_NETWORK}" ]; then
    DOCKER_OPTS="${DOCKER_OPTS} --network ${PG_NETWORK}"
fi

PG_DOCKER_ID=$(docker run ${DOCKER_OPTS} -d postgres:14-alpine)

echo "Spawned Postgres docker: $PG_DOCKER_ID"

if [ -z "${PG_NETWORK}" ]; then
    export PG_HOST=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $PG_DOCKER_ID)
else
    export PG_HOST=$(docker inspect --format "{{ (index .NetworkSettings.Networks \"${PG_NETWORK}\").IPAddress }}" $PG_DOCKER_ID)
fi

echo "Wait for database init..."
hack/wait-for-it/wait-for-it.sh -t 120 ${PG_HOST}:${PG_PORT}

(bash hack/test.sh $*)
result=$?

echo "Killing Postgres docker"
docker kill $PG_DOCKER_ID
docker rm -v $PG_DOCKER_ID

exit $result
