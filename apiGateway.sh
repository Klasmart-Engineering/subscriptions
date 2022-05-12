#!/bin/bash

set -e

echo "Enter number of requests:"
read -r requestCount

echo "Enter User ID:"
read -r GUID
echo

for i in $(seq 1 "$requestCount")
do
   http_response=$(curl -s -o response.json -w "%{http_code}" -X POST 'http://localhost:8010/log' \
                                                              -H 'Content-Type: application/json' \
                                                              -d '{
                                                                  "GUID": "'"$GUID"'",
                                                                  "subscriptionAccountId": 123,
                                                                  "actionType": "API Call",
                                                                  "usageAmount": 1,
                                                                  "product": "Simple Teacher Module"

                                                              }')

#   echo "Status:   "  "$http_response"
   responseBody=$(cat response.json )
   echo "Response: " "$responseBody"
   echo
done



