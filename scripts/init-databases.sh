#!/bin/bash
# Script de inicialização para criar múltiplos bancos de dados PostgreSQL
# Usado pelo Docker Compose para setup inicial

set -e
set -u

function create_user_and_database() {
    local database=$1
    echo "  Creating user and database '$database'"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        CREATE USER ${database}_user WITH PASSWORD '${database}_pass';
        CREATE DATABASE $database;
        GRANT ALL PRIVILEGES ON DATABASE $database TO ${database}_user;
EOSQL
}

if [ -n "${POSTGRES_MULTIPLE_DATABASES:-}" ]; then
    echo "Multiple database creation requested: ${POSTGRES_MULTIPLE_DATABASES:-}"
    for db in $(echo ${POSTGRES_MULTIPLE_DATABASES:-} | tr ',' ' '); do
        create_user_and_database "$db"
    done
    echo "Multiple databases created"
fi
