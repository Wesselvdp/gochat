# scripts/create_account.sh
source .env
URL=$1
ID=$2

echo "$URL/patron/account/$ID"

curl -X GET "$URL/patron/user/$ID" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY"