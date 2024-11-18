#!/bin/sh

# Run migrations
migrate -database "sqlite3://${DB_PATH}?_foreign_keys=on"-verbose -path /db/migrations up

# Start the server
exec /server