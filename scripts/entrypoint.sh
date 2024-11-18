#!/bin/sh

# Run migrations
migrate -database "sqlite3://${DB_PATH}?_foreign_keys=on" -path /db/migrations up -verbose

# Start the server
exec /server