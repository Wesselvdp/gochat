#!/bin/sh

# Run migrations
migrate -database "sqlite3://${DB_PATH}" -path /db/migrations up

# Start the server
exec /server