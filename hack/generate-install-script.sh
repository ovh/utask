#!/bin/bash

set -e

dst=$1
if [ -z "$1" ]
then
    dst="install-utask.sh"
fi

version=`git describe --tags $(git rev-list --tags --max-count=1)`

write_block() {
    echo "cat <<'EOF' >$1" >> $dst
    sed "s/DOCKER_TAG/$version/" $2 >> $dst
    echo "EOF" >> $dst
    echo "" >> $dst
    echo "" >> $dst
}

touch $dst

cat <<EOF >$dst
#!/bin/bash

set -e

mkdir -p templates functions sql plugins init
touch functions/.gitkeep

### DOCKER
EOF

write_block "docker-compose.yaml"  docker-compose.yaml
write_block "Dockerfile"           hack/Dockerfile-child

echo "### SQL" >> $dst
write_block "sql/schema.sql"       sql/schema.sql

echo "### TEMPLATES" >> $dst
write_block "templates/hello-world-now.yaml" ./examples/templates/hello-world-now.yaml

echo "install script saved at $dst"
