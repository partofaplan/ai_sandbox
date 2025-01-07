from flask import Flask, request, jsonify, render_template
import requests
from requests.exceptions import RequestException, Timeout
import logging
import json
import sys

app = Flask(__name__)

# Environment variables (replace or load dynamically as needed)
HOST = "ollama.ollama"  # $HOST is "ollama.ollama" in RKE and "https://localhost" when locally deployed
PORT = "80"  # Port is 80 on RKE and 11434 when deployed locally
MODEL = "prospector:latest"
TIMEOUT = 300  # seconds for API calls

# API URLs
CREATE_API_URL = f"http://{HOST}:{PORT}/api/create"
GENERATE_API_URL = f"http://{HOST}:{PORT}/api/generate"

# JSON logging configuration
class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_record = {
            "level": record.levelname,
            "message": record.getMessage(),
            "timestamp": self.formatTime(record, self.datefmt),
        }
        if record.exc_info:
            log_record["exception"] = self.formatException(record.exc_info)
        return json.dumps(log_record)

# Configure structured JSON logging
json_formatter = JSONFormatter()
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(json_formatter)
logger = logging.getLogger(__name__)
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

logger.info(f"Generated API URL: {GENERATE_API_URL}")

@app.before_request
def log_request():
    logger.debug(f"Incoming request headers: {request.headers}")
    logger.debug(f"Raw incoming request data: {request.data}")
    content_type = request.headers.get('Content-Type', '')
    if content_type == 'application/json':
        logger.info(f"Incoming request: {request.method} {request.path}, Body: {request.get_json()}")
    else:
        logger.info(f"Incoming request: {request.method} {request.path}, Content-Type: {content_type}, Body not logged due to unsupported content type")

@app.after_request
def log_response(response):
    logger.info(f"Outgoing response: {response.status_code}, Body: {response.get_json()}")
    return response

@app.errorhandler(Exception)
def handle_unexpected_error(e):
    logger.exception("Unhandled exception occurred")
    return jsonify({"error": "An unexpected error occurred"}), 500

@app.route('/')
def home():
    return render_template('index.html')

@app.route('/chat', methods=['POST'])
def chat():
    if not request.is_json:
        logger.error("Request failed: Content-Type is not application/json")
        return jsonify({"error": "Content-Type must be application/json"}), 400

    try:
        payload = request.get_json()
        logger.debug(f"Parsed JSON payload: {payload}")

        if not payload:
            logger.error("Empty JSON payload received.")
            return jsonify({"error": "Empty JSON payload"}), 400

        user_input = payload.get('prompt')
        if not user_input:
            logger.error("'prompt' key missing in payload.")
            return jsonify({"error": "Prompt is required"}), 400

        # Prepare payload for Ollama API
        data = {
            "model": MODEL,
            "prompt": user_input,
            "stream": False
        }

        logger.info(f"Sending chat request to Ollama: URL: {GENERATE_API_URL}, Payload: {data}")
        response = requests.post(GENERATE_API_URL, json=data, headers={"Content-Type": "application/json"}, timeout=TIMEOUT)

        logger.debug(f"Ollama API raw response: Status: {response.status_code}, Body: {response.text}")

        if response.status_code == 200:
            try:
                response_data = response.json()
                logger.debug(f"Parsed Ollama API response: {response_data}")
            except ValueError as e:
                logger.error(f"Failed to parse JSON response from Ollama API: {e}")
                return jsonify({"error": "Invalid response from Ollama API"}), 500

            result = response_data.get("response")
            if not result:
                logger.error(f"Missing 'response' key in Ollama API response: {response_data}")
                return jsonify({"error": "Invalid response from Ollama API"}), 500

            return jsonify({"response": result})

        else:
            logger.error(f"Ollama API error: {response.status_code}, Response: {response.text}")
            return jsonify({"error": f"Ollama error: {response.status_code} - {response.text}"}), response.status_code

    except Timeout:
        logger.error("Timeout occurred while communicating with Ollama API")
        return jsonify({"error": "Timeout occurred while communicating with Ollama API"}), 504
    except RequestException as e:
        logger.exception("RequestException occurred while communicating with Ollama API")
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500
    except Exception as e:
        logger.exception(f"Unexpected error in /chat route: {e}")
        return jsonify({"error": "An unexpected error occurred"}), 500

@app.route('/reload-model', methods=['POST'])
def reload_model():
    model_data = {
        "model": "prospector",
        "path": "/root/models/Modelfile"
    }

    try:
        logger.info(f"Sending model reload request to Ollama: {CREATE_API_URL}, Payload: {model_data}")
        response = requests.post(CREATE_API_URL, json=model_data, headers={"Content-Type": "application/json"}, timeout=TIMEOUT)

        if response.status_code == 200:
            logger.info("Model reloaded successfully.")
            return jsonify({"message": "Model reloaded successfully."})
        else:
            logger.error(f"Failed to reload model: {response.status_code}, Response: {response.text}")
            return jsonify({"error": f"Failed to reload model: {response.status_code} - {response.text}"}), response.status_code
    except Timeout:
        logger.error("Timeout occurred while reloading the model")
        return jsonify({"error": "Timeout occurred while reloading the model"}), 504
    except RequestException as e:
        logger.exception("RequestException occurred during model reload")
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500

if __name__ == '__main__':
    app.logger.setLevel(logging.DEBUG)
    app.run(debug=True, host='0.0.0.0', port=6500)