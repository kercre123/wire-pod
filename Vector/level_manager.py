#!/home/connorbailey/vector-venv/bin/python3

import anki_vector
import grpc
from grpc._channel import _MultiThreadedRendezvous
import json

class LevelManager():
    def __init__(self):
        self.robot_config_path: str = "../wire-pod/chipper/webroot/robot_config.json"
        self.levels: [int] = [0, 1, 40, 82, 168, 334, 684, 1402, 2874, 5891, 12076]

    def read_json_file(self, filename):
        with open(filename, 'r') as file:
            data = json.load(file)
        return data
    
    def check_if_level_up(self, robot_data):
        current_level = robot_data["robot_level"]
        xp = robot_data["robot_xp"]

        if self.levels[current_level + 1] > xp:
            return False
        else:
            return True
    
    def xp_to_level_up(self, robot_data):
        return self.levels[robot_data["robot_level"] + 1] - robot_data["robot_xp"]

    def querry_level(self):
        print("Reading robot config...")
        robot_data = self.read_json_file(self.robot_config_path)
        print(f"Level: {robot_data['robot_level']}")
        next_level = self.xp_to_level_up(robot_data=robot_data)
        print(f"XP: {robot_data['robot_xp']}/{robot_data['robot_xp'] + next_level}")
        try:
            with anki_vector.Robot(ip=robot_data["ip_address"], escape_pod=True) as robot:
                robot.behavior.say_text(f"I am currently level {robot_data['robot_level']} with {robot_data['robot_xp']} experience points.")
                robot.behavior.say_text(f"I need {next_level} more experience points to reach Level {robot_data['robot_level'] + 1}")
        
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
    vector = LevelManager()
    vector.querry_level()