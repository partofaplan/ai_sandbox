from flask import Flask, request, jsonify, render_template
import requests
from requests.exceptions import RequestException, Timeout
import logging

app = Flask(__name__)

# $HOST is ollama.ollama in RKE and https://localhost when locally deployed
HOST = "ollama.ollama"
# Port is 80 on RKE and 11434 when deployed locally
PORT = "80"
MODEL = "prospector:latest"
TIMEOUT = 300  # seconds for API calls

# API URL to reload the model
CREATE_API_URL = f"http://{HOST}:{PORT}/api/create"
GENERATE_API_URL = f"http://{HOST}:{PORT}/api/generate"

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Route for the home page
@app.route('/')
def home():
    return render_template('index.html')

# Route for handling chatbot requests
@app.route('/chat', methods=['POST'])
def chat():
    user_input = request.json.get('prompt', '')

    if not user_input:
        return jsonify({"error": "Prompt is required"}), 400

    # Prepare payload for Ollama API
    data = {
        "model": MODEL,
        "prompt": user_input,
        "stream": False
    }

    try:
        # Make the request to the Ollama API with timeout
        logger.info(f"Sending chat request to Ollama: {GENERATE_API_URL}, Payload: {data}")
        response = requests.post(GENERATE_API_URL, json=data, timeout=TIMEOUT)

        if response.status_code == 200:
            response_data = response.json()
            logger.info(f"Ollama API responded successfully: {response_data}")
            result = response_data.get("response", "No response found.")
            return jsonify({"response": result})
        else:
            logger.error(f"Ollama API error: {response.status_code} - {response.text}")
            return jsonify({"error": f"Ollama error: {response.status_code} - {response.text}"}), response.status_code
    except Timeout:
        logger.error("Timeout occurred while communicating with Ollama API")
        return jsonify({"error": "Timeout occurred while communicating with Ollama API"}), 504
    except RequestException as e:
        logger.error(f"RequestException occurred: {e}")
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500

# Route to reload the model
@app.route('/reload-model', methods=['POST'])
def reload_model():
    # Prepare payload for reloading the model
    model_data = {
        "model": "prospector",
        "path": "/root/models/Modelfile"
    }

    try:
        # Make the request to reload the model with timeout
        logger.info(f"Sending model reload request to Ollama: {CREATE_API_URL}, Payload: {model_data}")
        response = requests.post(CREATE_API_URL, json=model_data, timeout=TIMEOUT)

        if response.status_code == 200:
            logger.info("Model reloaded successfully.")
            return jsonify({"message": "Model reloaded successfully."})
        else:
            logger.error(f"Failed to reload model: {response.status_code} - {response.text}")
            return jsonify({"error": f"Failed to reload model: {response.status_code} - {response.text}"}), response.status_code
    except Timeout:
        logger.error("Timeout occurred while reloading the model")
        return jsonify({"error": "Timeout occurred while reloading the model"}), 504
    except RequestException as e:
        logger.error(f"RequestException occurred during model reload: {e}")
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=6500)
