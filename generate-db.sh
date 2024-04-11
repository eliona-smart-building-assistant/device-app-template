go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

go get github.com/volatiletech/sqlboiler/v4
go get github.com/volatiletech/null/v8

docker run -d \
    --name "app_sql_boiler_code_generation" \
    --platform "linux/amd64" \
    -e "POSTGRES_PASSWORD=secret" \
    -p "6001:5432" \
    -v "${PWD}/conf/init.sql:/docker-entrypoint-initdb.d/init.sql" \
    debezium/postgres:12

sleep 5

sqlboiler psql \
    -c sqlboiler.toml \
    --wipe --no-tests

docker stop "app_sql_boiler_code_generation" > /dev/null

docker logs "app_sql_boiler_code_generation" 2>&1 | grep "ERROR" || {
    echo "All good."
}

docker rm "app_sql_boiler_code_generation" > /dev/null

go mod tidy
