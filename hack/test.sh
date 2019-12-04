#!/bin/bash

require() {
    for s in $*
    do
        if [ -z "$(eval echo \$\{$s\})" ]; then
            echo "Missing mandatory env '$s'" 1>&2
            exit 1
        fi
    done
}

require_exec() {
    for x in $*
    do
        if ! [ -x "$x" ]; then
            echo "$x is not an executable" 1>&2
            exit 1
        fi
    done
}

require PG_USER PG_PASSWORD PG_HOST PG_PORT PG_DATABASENAME PSQL_BIN

export CFG_DATABASE="postgres://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DATABASENAME?connect_timeout=5&sslmode=disable"

export SCRIPTS="$GOPATH/src/github.com/ovh/utask/scripts"

mkdir -p $PWD/config
export CONFIGURATION_FROM=filetree:$PWD/config,env:CFG

cat <<EOF >$PWD/config/encryption-key
{
    "identifier":"storage",
    "cipher":"aes-gcm",
    "timestamp":1535627466,
    "key":"e5f45aef9f072e91f735547be63f3434e6de49695b178e3868b23b0e32269800"
}
EOF

cat <<EOF >$PWD/config/utask-cfg
{
    "admin_usernames": ["admin"],
    "resolver_usernames": ["resolver"]
}
EOF

echo "Initializing DB..."

$PSQL_BIN <<EOF
$(cat $PWD/sql/schema.sql)
EOF

echo "Running commands..."

($*)
result=$?

echo "Done, cleaning up..."

exit $result
