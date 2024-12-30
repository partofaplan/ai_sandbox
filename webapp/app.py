from flask import Flask, request, jsonify, render_template
import requests

app = Flask(__name__)

# Variables
HOST = "localhost"
PORT = "11434"
MODEL = "prospector:latest"

# API URL to reload the model
CREATE_API_URL = f"http://{HOST}:{PORT}/api/create"

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

    # API URL for chat generation
    url = f"http://{HOST}:{PORT}/api/generate"
    data = {
        "model": MODEL,
        "prompt": user_input,
        "stream": False
    }

    try:
        # Make the request to the Ollama API
        response = requests.post(url, json=data)

        if response.status_code == 200:
            result = response.json().get("response", "No response found.")
            return jsonify({"response": result})
        else:
            return jsonify({"error": f"Ollama error: {response.status_code} - {response.text}"}), response.status_code
    except Exception as e:
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500

# Route to reload the model
@app.route('/reload-model', methods=['POST'])
def reload_model():
    # Data for reloading the model
    model_data = {
        "model": "prospector",
        "path": "/root/models/Modelfile"
    }

    try:
        # Make the request to reload the model
        response = requests.post(CREATE_API_URL, json=model_data)

        if response.status_code == 200:
            return jsonify({"message": "Model reloaded successfully."})
        else:
            return jsonify({"error": f"Failed to reload model: {response.status_code} - {response.text}"}), response.status_code
    except Exception as e:
        return jsonify({"error": f"An error occurred: {str(e)}"}), 500

if __name__ == '__main__':
    app.run(debug=True)
