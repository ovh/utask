#!/usr/bin/bash
set -e

dst=$1
if [ -z "$1" ]
then
    dst="install-utask.sh"
fi

write_block() {
    echo "cat <<EOF >$1" >> $dst
    cat $2 >> $dst
    echo "EOF" >> $dst
    echo "" >> $dst
    echo "" >> $dst
}

touch $dst

cat <<EOF >$dst
#!/usr/bin/bash
set -e 

mkdir -p templates sql plugins init

### DOCKER
EOF

write_block "docker-compose.yaml"  docker-compose.yaml
write_block "Dockerfile"           hack/Dockerfile-child

echo "### SQL" >> $dst
write_block "sql/schema.sql"       sql/schema.sql

echo "### TEMPLATES" >> $dst
write_block "templates/hello-world-now.yaml" ./examples/templates/hello-world-now.yaml

echo "install script saved at $dst"
