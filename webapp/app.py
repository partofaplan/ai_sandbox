from flask import Flask, request, jsonify, render_template
from flask_cors import CORS, cross_origin
import requests
from requests.exceptions import RequestException, Timeout
import logging
import json
import sys
import os

app = Flask(__name__, 
    static_folder='static',  # Define static folder
    template_folder='templates'  # Define templates folder
)

# CORS configuration
cors = CORS(app, resources={
    r"/*": {
        "origins": "*",
        "methods": ["GET", "POST", "OPTIONS"],
        "allow_headers": ["Content-Type", "Authorization"]
    }
})

# Environment variables with fallbacks
HOST = os.getenv("OLLAMA_HOST", "ollama.ollama")
PORT = os.getenv("OLLAMA_PORT", "80")
MODEL = os.getenv("OLLAMA_MODEL", "prospector:latest")
TIMEOUT = int(os.getenv("API_TIMEOUT", "300"))

# API URLs
CREATE_API_URL = f"http://{HOST}:{PORT}/api/create"
GENERATE_API_URL = f"http://{HOST}:{PORT}/api/generate"

# Logging setup
class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_record = {
            "level": record.levelname,
            "message": record.getMessage(),
            "timestamp": self.formatTime(record, self.datefmt),
            "path": request.path if request else None,
            "method": request.method if request else None
        }
        if record.exc_info:
            log_record["exception"] = self.formatException(record.exc_info)
        return json.dumps(log_record)

# Configure logging
json_formatter = JSONFormatter()
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(json_formatter)
logger = logging.getLogger(__name__)
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

@app.route('/')
def home():
    logger.debug("Serving home page")
    return render_template('index.html')

@app.route('/health')
def health_check():
    try:
        response = requests.get(f"http://{HOST}:{PORT}/api/tags", timeout=5)
        if response.ok:
            return jsonify({"status": "healthy", "ollama_connection": "ok"}), 200
        return jsonify({"status": "degraded", "ollama_connection": "failed"}), 503
    except Exception as e:
        return jsonify({"status": "unhealthy", "error": str(e)}), 503

@app.route('/chat', methods=['POST', 'OPTIONS'])
@cross_origin()
def chat():
    if request.method == 'OPTIONS':
        return '', 204
        
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

        data = {
            "model": MODEL,
            "prompt": user_input,
            "stream": False
        }

        logger.info(f"Sending chat request to Ollama: {GENERATE_API_URL}")
        response = requests.post(
            GENERATE_API_URL, 
            json=data, 
            headers={"Content-Type": "application/json"}, 
            timeout=TIMEOUT
        )

        logger.debug(f"Ollama API response status: {response.status_code}")

        if response.ok:
            try:
                response_data = response.json()
                result = response_data.get("response")
                if not result:
                    raise ValueError("Missing 'response' key")
                return jsonify({"response": result})
            except Exception as e:
                logger.error(f"Failed to process Ollama response: {e}")
                return jsonify({"error": "Invalid response from Ollama API"}), 500
        else:
            error_message = f"Ollama error: {response.status_code}"
            try:
                error_message += f" - {response.json().get('error', '')}"
            except:
                error_message += f" - {response.text[:100]}"
            logger.error(error_message)
            return jsonify({"error": error_message}), response.status_code

    except Timeout:
        logger.error("Timeout occurred while communicating with Ollama API")
        return jsonify({"error": "Request timed out"}), 504
    except Exception as e:
        logger.exception(f"Unexpected error in /chat route: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/reload-model', methods=['POST'])
def reload_model():
    try:
        model_data = {
            "model": "prospector",
            "path": "/root/models/Modelfile"
        }
        
        response = requests.post(
            CREATE_API_URL, 
            json=model_data, 
            headers={"Content-Type": "application/json"}, 
            timeout=TIMEOUT
        )

        if response.ok:
            return jsonify({"message": "Model reloaded successfully"}), 200
        else:
            return jsonify({"error": f"Failed to reload model: {response.text}"}), response.status_code
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    # Make sure templates are auto-reloaded during development
    app.config['TEMPLATES_AUTO_RELOAD'] = True
    app.run(debug=True, host='0.0.0.0', port=6500)