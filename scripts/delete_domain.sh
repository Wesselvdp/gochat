# scripts/create_account.sh
source .env
URL=$1
DOMAIN=$2

curl -X GET "$URL/patron/account/accountdomains/delete/$DOMAIN" \
     -H "Content-Type: application/json" \
     -H "X-Admin-Key: $ADMIN_API_KEY" \
