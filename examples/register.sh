#!/bin/bash

AGENT_NAME=$1

curl -X 'POST' \
  'http://localhost:8007/api/v1/register' \
  -H 'Content-Type: application/json' \
  --data-binary @- << EOF
  {
    "name": "$AGENT_NAME",
    "apikey": "$APIKEY",
    "base64privatekey": "$BASE64PKEY"
  }
EOF


