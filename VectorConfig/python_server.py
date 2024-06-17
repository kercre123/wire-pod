from flask import Flask, request, jsonify
from flask_cors import CORS
import subprocess

app = Flask(__name__)
CORS(app)

def execute_script_in_venv(script_path):
    venv_python_path = "/home/connorbailey/vector-venv/bin/python3"
    
    command = [
        venv_python_path,
        script_path
    ]

    # The environment variable can be set directly in the subprocess call
    env = dict(PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION='python')

    try:
        output = subprocess.check_output(command, env=env)
        return True, output.decode('utf-8')
    except subprocess.CalledProcessError as e:
        return False, str(e)

@app.route('/run_level_manager', methods=['POST'])
def run_level_manager():
    success, message = execute_script_in_venv("../Vector/level_manager.py")
    return jsonify(success=success, message=message), 200 if success else 500

@app.route('/run_go_for_jog', methods=['POST'])
def run_go_for_jog():
    success, message = execute_script_in_venv("../Vector/vector_go_for_jog.py")
    return jsonify(success=success, message=message), 200 if success else 500

@app.route('/run_scratch_ticket', methods=['POST'])
def run_scratch_ticket():
    success, message = execute_script_in_venv("../Vector/vector_scratch_ticket.py")
    return jsonify(success=success, message=message), 200 if success else 500

@app.route('/run_wallet_manager', methods=['POST'])
def run_wallet_manager():
    success, message = execute_script_in_venv("../Vector/wallet_manager.py")
    return jsonify(success=success, message=message), 200 if success else 500

@app.route('/run_battery_manager', methods=['POST'])
def run_battery_manager():
    success, message = execute_script_in_venv("../Vector/battery_manager.py")
    return jsonify(success=success, message=message), 200 if success else 500

@app.route('/buy_item', methods=['POST'])
def shop_buy_item():
    item_id = request.json.get('item_id')
    if item_id is None:
        return jsonify(success=False, message="Item ID is required."), 400
    
    

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8091)
