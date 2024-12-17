# scripts/create_account.sh
source .env
URL=$1
NAME=$2
DOMAIN=$3

curl -X POST "$URL/patron/account/accountdomains/create" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY" \
     -d '{
         "accountId": "'$ACCOUNT'",
         "domain": "'$DOMAIN'"
     }'