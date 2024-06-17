#!/home/connorbailey/vector-venv/bin/python3

import anki_vector
from anki_vector.util import degrees, distance_mm, speed_mmps
import grpc
from grpc._channel import _MultiThreadedRendezvous
import time
import random
from datetime import datetime, timedelta
from vector_helpers import VectorHelpers

try:
    from PIL import Image
except ImportError:
    sys.exit("Cannot import from PIL: Do `pip3 install --user Pillow` to install")

class VectorScratchTicket():
    def __init__(self):
        self.required_energy: int = 5
        self.display_number_timeout = 1.0
        self.lotto_number_count = 3
        self.min_jackpot: int = 100
        self.max_jackpot: int = 500
        self.helpers = VectorHelpers()
    
    def lotto_numbers(self):
        numbers = []
        for _ in range(self.lotto_number_count):
            number = random.randint(1, 5)  # Generate a random number between 1 and 5
            numbers.append(number)
        return numbers
    
    def get_jackpot(self):
        return random.randint(self.min_jackpot, self.max_jackpot)
    
    def do_action(self, robot_data, robot):
        self.helpers.update_energy_level(self.required_energy)
        robot.behavior.drive_off_charger()
        winning_numbers = self.lotto_numbers()
        robot_numbers = self.lotto_numbers()

        print(f"Winning Numbers: {winning_numbers}")
        print(f"Robot Numbers: {robot_numbers}")

        robot.behavior.say_text("Match all three numbers exactly to win the jackpot! The winning numbers are")
        time.sleep(0.25)
        robot.behavior.set_head_angle(degrees(45.0))

        for index, element in enumerate(winning_numbers):
            image_name = "font-" + str(element) + ".png"
            image_path = "/home/connorbailey/VectorConfig/face_images/numbers/" + image_name
            image_file = Image.open(image_path)
            screen_data = anki_vector.screen.convert_image_to_screen_data(image_file)
            robot.screen.set_screen_with_image_data(screen_data, self.display_number_timeout)
            robot.behavior.say_text(f"{element}")
            time.sleep(self.display_number_timeout)
        
        robot.anim.play_animation_trigger('GreetAfterLongTime', ignore_body_track=True)
        robot.behavior.say_text(f"Good luck!")
        robot.behavior.set_lift_height(0.3)

        robot.behavior.say_text(f"Our tickets numbers are")
        time.sleep(0.25)
        robot.behavior.set_head_angle(degrees(45.0))

        for index, element in enumerate(robot_numbers):
            image_name = "font-" + str(element) + ".png"
            image_path = "/home/connorbailey/VectorConfig/face_images/numbers/" + image_name
            image_file = Image.open(image_path)
            screen_data = anki_vector.screen.convert_image_to_screen_data(image_file)
            robot.screen.set_screen_with_image_data(screen_data, self.display_number_timeout)
            robot.behavior.say_text(f"{element}")
            time.sleep(self.display_number_timeout)

        if winning_numbers == robot_numbers:
            jackpot_amount = self.get_jackpot()
            robot.anim.play_animation_trigger("BlackJack_VictorBlackJackLose", ignore_body_track=True)
            robot.behavior.say_text(f"We are rich! Looks like we just won {jackpot_amount} coins! Let me put all these coins in my wallet.")
            for count in range(5):
                print(f"Count: {count}")
                print("Lift down...")
                robot.motors.set_lift_motor(-5)
                time.sleep(0.5)
                print("Lift up")
                robot.motors.set_lift_motor(5)
                time.sleep(0.5)
                count += 1
            robot_data["robot_wallet"] = robot_data["robot_wallet"] + jackpot_amount
            self.helpers.write_json_file(filename=self.robot_config_path, data=robot_data)
            robot.behavior.say_text(f"That is amazing! There was only a 0.8% chance of that happening. I have goosebumps.")
        else:
            robot.behavior.say_text("We lost! Better luck next time.")
            robot.anim.play_animation_trigger("BlackJack_VictorWin", ignore_body_track=True)
            self.helpers.write_json_file(data=robot_data)
        

        earned_xp = self.helpers.generate_xp(1, 1)
        print(f"Earned XP: {earned_xp}")
        robot.behavior.say_text(f"Ah. I also gained {earned_xp} experience point.")
        self.helpers.update_xp_and_check_level(earned_xp, robot)

    def main(self):
        print("Reading robot config...")
        robot_data = self.helpers.read_json_file()
        print("Connecting to the robot...")

        try:
            with anki_vector.Robot(ip=robot_data["ip_address"], escape_pod=True) as robot:
                if self.helpers.check_energy_level(self.required_energy):
                    print("Checking for lottery ticket in inventory...")
                    # TODO: Check robot's inventory
                    robot.behavior.say_text("Let me get my lucky coin!")
                    self.do_action(robot_data=robot_data, robot=robot)
                else:
                    robot.behavior.say_text(f"I'm tired, I need {self.required_energy} energy to do that")  
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
    vector = VectorScratchTicket()
    vector.main()