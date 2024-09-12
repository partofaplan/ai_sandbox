import requests
import json
import sys

# Variables
HOST = "localhost"
PORT = "11434"
MODEL = "llama2"

# Check if a prompt is provided as an argument
if len(sys.argv) < 2:
    print("Usage: python script.py \"<prompt>\"")
    sys.exit(1)

# Assign the prompt argument
PROMPT = sys.argv[1]

# API URL
url = f"http://{HOST}:{PORT}/api/generate"

# Data to send in the request
data = {
    "model": MODEL,
    "prompt": PROMPT,
    "stream": False
}

# Make the request
response = requests.post(url, json=data)

# Parse the response
if response.status_code == 200:
    result = response.json().get("response", "No response found.")
    print(result)
else:
    print(f"Error: {response.status_code} - {response.text}")
