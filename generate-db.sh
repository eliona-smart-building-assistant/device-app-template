go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

go get github.com/volatiletech/sqlboiler/v4
go get github.com/volatiletech/null/v8

# Read the content of init.sql
INIT_SQL_CONTENT=$(<"${PWD}/conf/init.sql")

# Create init_wrapper.sql to run the script in a transaction. This is needed for
# COMMIT AND CHAIN to work in the script.
cat << EOF > ./conf/init_wrapper.sql
BEGIN;

$INIT_SQL_CONTENT

COMMIT;
EOF

docker run -d \
    --name "app_sql_boiler_code_generation" \
    --platform "linux/amd64" \
    -e "POSTGRES_PASSWORD=secret" \
    -p "6001:5432" \
    -v "${PWD}/conf/init_wrapper.sql:/docker-entrypoint-initdb.d/init_wrapper.sql" \
    debezium/postgres:12

# Wait for PostgreSQL to initialize
sleep 5

sqlboiler psql \
    -c sqlboiler.toml \
    --wipe --no-tests

docker stop "app_sql_boiler_code_generation" > /dev/null

docker logs "app_sql_boiler_code_generation" 2>&1 | grep "ERROR" || {
    echo "All good."
}

docker rm "app_sql_boiler_code_generation" > /dev/null

rm ./conf/init_wrapper.sql

go mod tidy
