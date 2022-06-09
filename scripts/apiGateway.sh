#!/bin/bash

set -e

echo "Enter number of requests:"
read -r requestCount

echo "Enter Subscription ID:"
read -r SubId
echo

for i in $(seq 1 "$requestCount")
do

  epochTime=`date +%s`
#  TODO update this to point at the gateway instead of the app once Redpanda is running in K8s.
   http_response=$(curl -s -o response.json -w "%{http_code}" -X POST 'http://localhost:8070/log-action' \
                                                              -H 'Content-Type: application/json' \
                                                              -d '{
                                                                  "SubscriptionId": "'"$SubId"'",
                                                                  "actionType": "API Call",
                                                                  "usageAmount": 1,
                                                                  "product": "Simple Teacher Module",
                                                                  "interactionTimeEpoch": "'"$epochTime"'"
                                                              }')

#   echo "Status:   "  "$http_response"
   responseBody=$(cat response.json )
   echo "Response: " "$responseBody"
   echo
done



