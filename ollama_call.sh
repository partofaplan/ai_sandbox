#!/bin/bash

# Variables
HOST="localhost"
PORT="11434"
MODEL="orca-mini"

# Check if a prompt is provided as an argument
if [ -z "$1" ]; then
  echo "Usage: $0 \"<prompt>\""
  exit 1
fi

# Assign the prompt argument
PROMPT="$1"

# Make the curl request and parse the response
curl http://$HOST:$PORT/api/generate -d "{\"model\": \"$MODEL\", \"prompt\": \"$PROMPT\", \"stream\": false}" | jq '.["response"]'

