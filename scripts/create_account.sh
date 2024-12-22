# scripts/create_account.sh
source .env
URL=$1
NAME="$2"

curl -X POST "$URL/patron/account/create" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY" \
     -d '{"name": "'"$NAME"'"}'