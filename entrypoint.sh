#!/bin/sh

echo "Waiting for PostgreSQL..."
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
    sleep 1
done

echo "PostgreSQL is ready!"
echo "Running migrations..."

goose -dir /app/migrations postgres "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSL_MODE" up

if [ $? -ne 0 ]; then
    echo "Migration failed!"
    exit 1
fi

echo "Migrations applied!"
echo "Starting server..."

exec ./server