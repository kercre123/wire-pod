#!/home/connorbailey/vector-venv/bin/python3

import anki_vector
from anki_vector.util import degrees, distance_mm, speed_mmps
import grpc
from grpc._channel import _MultiThreadedRendezvous
import time
import random
from datetime import datetime, timedelta
from vector_helpers import VectorHelpers

class VectorGoForJog():
    def __init__(self):
        self.required_energy: int = 10
        self.cool_down: int = 300
        self.helpers = VectorHelpers()
    
    def find_money(self, data, robot):
        chance = random.randint(1, 100)
        print(f"Chance: {chance}")
        if chance > 95:
            found_money = random.randint(2, 4)
            robot.behavior.say_text(f"Oh hey look, I found {found_money} coins on the ground. Let me put them in my wallet.")
            
            count = 0
            for count in range(found_money):
                print(f"Count: {count}")
                print("Lift down...")
                robot.motors.set_lift_motor(-5)
                time.sleep(1.0)
                print("Lift up")
                robot.motors.set_lift_motor(5)
                time.sleep(1.0)
                count += 1
            
            data["robot_wallet"] = data["robot_wallet"] + found_money
            self.helpers.write_json_file(data)
    
    def do_action(self, robot_data, robot):
        self.helpers.update_energy_level(self.required_energy)
        robot.behavior.drive_off_charger()
        robot.motors.set_lift_motor(5)
        for _ in range(8):
            print("Drive Vector straight...")
            robot.behavior.drive_straight(distance_mm(125), speed_mmps(100))
            robot_data["robot_total_jog_dist"] = robot_data["robot_total_jog_dist"] + 125
            
            self.find_money(data=robot_data, robot=robot)

            print("Turn Vector in place...")
            robot.behavior.turn_in_place(degrees(90))
        
        self.helpers.write_json_file(robot_data)
        robot.motors.set_lift_motor(-5)
        robot.anim.play_animation_trigger('GreetAfterLongTime', ignore_body_track=True)

        earned_xp = self.helpers.generate_xp(2, 5)
        robot.behavior.say_text(f"Wow! That was great! I earned {earned_xp} experience points")
        print(f"Robot XP: {robot_data['robot_xp']}")
        print(f"Earned XP: {earned_xp}")
        self.helpers.update_xp_and_check_level(earned_xp, robot)


    def main(self):
        print("Reading robot config...")
        robot_data = self.helpers.read_json_file()

        print("Connecting to the robot...")
        try:
            with anki_vector.Robot(ip=robot_data["ip_address"], escape_pod=True) as robot:
                print("Checking time since last jog...")
                elapsed_time = self.helpers.time_elapsed_since_last_log()
                
                print(f"Elapsed Time: {elapsed_time}")
                print(f"Cooldown Period: {timedelta(seconds=self.cool_down)}")

                if elapsed_time > timedelta(seconds=self.cool_down):
                    if self.helpers.check_energy_level(self.required_energy):
                        robot.behavior.say_text("Lets go jogging!")
                        robot_data["last_jog"] = datetime.now().strftime('%Y-%m-%dT%H:%M:%S')
                        self.helpers.write_json_file(robot_data)

                        self.do_action(robot_data=robot_data, robot=robot)
                    else:
                        robot.behavior.say_text(f"I'm tired, I need {self.required_energy} energy to do that")
                else:
                    remaining_time = self.cool_down - elapsed_time.total_seconds()

                    hours, remainder = divmod(remaining_time, 3600)
                    minutes, seconds = divmod(remainder, 60)

                    time_string = ""
                    if hours:
                        time_string += f"{int(hours)} hours "
                    if minutes:
                        time_string += f"{int(minutes)} minutes "
                    if seconds:
                        time_string += f"{int(seconds)} seconds"

                    robot.behavior.say_text(f"I can't go jogging yet. I need to wait for {time_string} more.")

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
    vector = VectorGoForJog()
    vector.main()