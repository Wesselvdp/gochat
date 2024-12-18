# scripts/create_account.sh
source .env
URL=$1
ACCOUNT=$2

curl -X POST "$URL/patron/account/change-user-account" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY" \
     -d '{
         "userID": "a50029fc-f6ce-4ec9-a8a5-8a7481d56886",
         "accountID": "'$ACCOUNT'"
     }'