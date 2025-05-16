# scripts/create_account.sh
URL=$1
NAME="$2"

echo "ADMIN_API_KEY: $ADMIN_API_KEY"
echo "Name: $NAME"

curl -X POST "$URL/patron/account/create" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY" \
     -d '{"name": "'"$NAME"'"}'