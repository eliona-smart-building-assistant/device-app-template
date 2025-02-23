go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc generate -f db/sqlc.yaml

docker logs "app_sql_boiler_code_generation" 2>&1 | grep "ERROR" || {
    echo "All good."
}

go mod tidy
