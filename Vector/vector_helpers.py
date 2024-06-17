import json
import random
from datetime import datetime, timedelta

class VectorHelpers:
    def __init__(self):
        self.config_path = "../wire-pod/chipper/webroot/robot_config.json" # Path to default to

    def read_json_file(self, path=None):
        """
            Reads one of the robot's configuration files. Defaults to robot_config.json if no path is provided.

            Returns: dict()
        """
        if path is None:
            with open(self.config_path, 'r') as file:
                return json.load(file)
        else:
            with open(path, 'r') as file:
                return json.load(file)

    def write_json_file(self, data, path=None):
        """
            Writes to one of the robot's configuration files. Defaults to robot_config.json if no path is provided.

            Returns: N/A
        """
        if path is None:
            with open(self.config_path, 'w') as file:
                json.dump(data, file, indent=4)
        else:
            with open(path, 'w') as file:
                json.dump(data, file, indent=4)

    def check_energy_level(self, required_energy):
        """
            Checks if the robot has more than or equal to the required_energy.

            Returns: Bool
        """
        data = self.read_json_file()
        return data["robot_energy_level"] >= required_energy

    def update_energy_level(self, energy_consumed):
        """
            Modifies the robot's energy value, subtracting energy_consumed from robot_energy_level. 

            Returns: N/A
        """
        data = self.read_json_file()
        data["robot_energy_level"] -= energy_consumed
        self.write_json_file(data)

    def time_elapsed_since_last_log(self):
        """
            Reads the robot_config.json file and compares the current time to the log of the last time the robot jogged. 

            Returns: Bool
        """
        data = self.read_json_file()
        last_jog_data = data["last_jog"]
        last_jog = datetime.strptime(last_jog_data, '%Y-%m-%dT%H:%M:%S')
        now = datetime.now()
        return now - last_jog

    def generate_xp(self, min_xp, max_xp):
        """
            Returns a random number between min_xp and max_xp.

            Returns: int
        """
        return random.randint(min_xp, max_xp)

    def update_xp_and_check_level(self, xp_earned, robot):
        """
            Adds the xp_earned to the robot_xp value in robot_config.json. Then calls the check_if_level_up()
            to see if the new XP amount will set the robot at a new level. 

            Returns: N/A
        """
        data = self.read_json_file()
        data["robot_xp"] += xp_earned
        self.write_json_file(data)
        
        leveled_up = self.check_if_level_up()
        if leveled_up:
            data["robot_level"] += 1
            robot.behavior.say_text(f"Oh? I also leveled up to Level {data['robot_level']}!")
        self.write_json_file(data)

    def check_if_level_up(self):
        """
            Compares the robot's current XP with the XP needed for the next level.

            Returns: Bool
        """
        levels: [int] = [0, 1, 40, 82, 168, 334, 684, 1402, 2874, 5891, 12076]
        
        data = self.read_json_file()
        current_level = data["robot_level"]
        xp = data["robot_xp"]

        if levels[current_level + 1] > xp:
            return False
        else:
            return True