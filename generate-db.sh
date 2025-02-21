#!/bin/bash

# Install BobGen
go install github.com/stephenafamo/bob/gen/bobgen-psql@latest
go get github.com/stephenafamo/bob

# Read the content of init.sql
INIT_SQL_CONTENT=$(<"${PWD}/db/init.sql")

# Create init_wrapper.sql to run the script in a transaction
cat << EOF > ./db/init_wrapper.sql
BEGIN;

$INIT_SQL_CONTENT

COMMIT;
EOF

# Start PostgreSQL container
docker run -d \
    --name "app_bob_code_generation" \
    --platform "linux/amd64" \
    -e "POSTGRES_PASSWORD=secret" \
    -p "6001:5432" \
    -v "${PWD}/db/init_wrapper.sql:/docker-entrypoint-initdb.d/init_wrapper.sql" \
    debezium/postgres:15

# Wait for PostgreSQL to initialize
sleep 5

# Run Bob code generation
bobgen-psql \
    -c db/bobgen.yaml

# Stop and remove the database container
docker stop "app_bob_code_generation" > /dev/null

docker logs "app_bob_code_generation" 2>&1 | grep "ERROR" || {
    echo "All good."
}

docker rm "app_bob_code_generation" > /dev/null

# Cleanup
rm ./db/init_wrapper.sql

go mod tidy
