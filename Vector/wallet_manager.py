#!/home/connorbailey/vector-venv/bin/python3

import anki_vector
import grpc
from grpc._channel import _MultiThreadedRendezvous
import json

class WalletManager():
    def __init__(self):
        self.robot_config_path: str = "../wire-pod/chipper/webroot/robot_config.json"

    def read_json_file(self, filename):
        with open(filename, 'r') as file:
            data = json.load(file)
        return data
    
    def querry_balance(self):
        print("Reading robot config...")
        robot_data = self.read_json_file(self.robot_config_path)
        print(f"{robot_data['robot_wallet']} coins")
        try:
            with anki_vector.Robot(ip=robot_data["ip_address"], escape_pod=True) as robot:
                robot.behavior.say_text(f"I have {robot_data['robot_wallet']} coins in my wallet!")
        except grpc._channel._InactiveRpcError as e:
            if "Maximum auth rate exceeded" in str(e):
                print("Authentication rate limit exceeded. Please wait and try again later.")
            else:
                print(f"An unexpected error occurred: {e}")
        except _MultiThreadedRendezvous as e:
            if e.code() == grpc.StatusCode.DEADLINE_EXCEEDED:
                print("The request timed out.")
            else:
                print(f"An unexpected GRPC error occurred: {e.details()}")

if __name__ == "__main__":
    vector = WalletManager()
    vector.querry_balance()