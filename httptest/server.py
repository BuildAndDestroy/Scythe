#!/usr/bin/env python3

from flask import Flask, request, jsonify, send_from_directory, abort
import os

app = Flask(__name__)

# Set the base directory for serving files
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

@app.route('/receive', methods=['POST'])
def receive_data():
    """Handles POST requests to receive command execution results."""
    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "No JSON data received"}), 400
        
        command = data.get("command", "N/A")
        output = data.get("output", "N/A")

        print(f"[+] Received command execution result:\nCommand: {command}\nOutput:\n{output}")

        return jsonify({"status": "success", "message": "Data received successfully"}), 200

    except Exception as e:
        return jsonify({"error": str(e)}), 500


@app.route('/<path:filename>', methods=['GET'])
def serve_static_files(filename):
    """Serves all static files while protecting sensitive files."""
    # Normalize path to prevent directory traversal
    safe_path = os.path.normpath(os.path.join(BASE_DIR, filename))

    # Check if request is trying to escape BASE_DIR
    if not safe_path.startswith(BASE_DIR) or filename == 'server.py':
        abort(404)  # Pretend the file does not exist

    if not safe_path.startswith(BASE_DIR) or filename == 'requirements.py':
        abort(404)  # Pretend the file does not exist

    # Serve the file if it exists
    if os.path.isfile(safe_path):
        return send_from_directory(BASE_DIR, filename, as_attachment=False)
    else:
        return abort(404)  # Standard "not found" response


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)

