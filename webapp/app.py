from flask import Flask, request, jsonify, render_template
import requests

app = Flask(__name__)

# Variables
HOST = "localhost"
PORT = "11434"
MODEL = "llama3.2:prospector"

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

    # API URL
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

if __name__ == '__main__':
    app.run(debug=True)
