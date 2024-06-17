#!/home/connorbailey/vector-venv/bin/python3

import anki_vector
from anki_vector.connection import ControlPriorityLevel
import json
import grpc
from grpc._channel import _MultiThreadedRendezvous

class BatteryManager():
    def __init__(self):
        self.robot_config_path: str = "../wire-pod/chipper/webroot/robot_config.json"
    
    def read_json_file(self, filename):
        with open(filename, 'r') as file:
            data = json.load(file)
        return data
    
    def estimate_soc(self, voltage):
        # Define the voltage-to-SoC mapping
        # NOTE: This is a basic linear curve for a 3.7V LiPo battery with 320 mAh battery
        # discharge which is as close to the original battery curve as I could find.
        #
        # This also isn't very good. This at the end of the day may not be possible as the battery
        # sits at 4.0V for most of its runtime, then quickly drops to the voltage threshold of 
        # 3.6V
        #
        # TODO: Add parameters for other batteries with different parameters
        #
        # TODO: Find out if the discharge curve in any other documentation
        #
        voltages = [4.20, 4.00, 3.80, 3.70, 3.60, 3.40, 3.20, 3.0]
        soc_values = [100, 85, 70, 50, 35, 20, 10, 5]

        # Linearly interpolate between the known values
        for i in range(len(voltages) - 1):
            if voltages[i] >= voltage >= voltages[i + 1]:
                soc_range = soc_values[i] - soc_values[i + 1]
                voltage_range = voltages[i] - voltages[i + 1]
                battery_soc = round(soc_values[i] - ((voltage - voltages[i]) * soc_range / voltage_range), 2)

                # Clamp the battery_soc between 0 and 100
                battery_soc = max(0, min(100, battery_soc))
                
                print(f"Battery SOC: {battery_soc}%")
                print(f"Voltage: {voltage} V")
                return battery_soc

        # If the voltage is outside the range of the provided voltages,
        # we can simply return 0 or 100 based on whether it's below or above the range, respectively.
        return 0 if voltage < voltages[-1] else 100

    def querry_battery_state(self):
        print("Reading robot config...")
        robot_data = self.read_json_file(self.robot_config_path)
        try:
            with anki_vector.Robot(ip=robot_data["ip_address"], escape_pod=True, behavior_control_level=ControlPriorityLevel.OVERRIDE_BEHAVIORS_PRIORITY) as robo:
                battery_state = robo.world.robot.get_battery_state()
                rounded_volts = round(battery_state.battery_volts, 3)
                battery_soc = self.estimate_soc(rounded_volts)
                
                if battery_state.battery_level == 3:
                    robo.behavior.say_text(f"My battery is full at {battery_soc} percent! I'm reading {rounded_volts} volts.")
                
                if battery_state.battery_level == 2:
                    robo.behavior.say_text(f"My battery is nominal at {battery_soc} percent! I'm reading {rounded_volts} volts.")
                
                if battery_state.battery_level == 1:
                    robo.behavior.say_text(f"My battery is low at {battery_soc} percent! I'm reading {rounded_volts} volts!")
                
                if battery_state.battery_level == 0:
                    robo.behavior.say_text(f"Critical battery state of charge, {battery_soc} percent! I'm reading {rounded_volts} volts!")

                if battery_state.is_on_charger_platform:
                    robo.behavior.say_text(f"I'm sitting on my charger!")
                else:
                    robo.behavior.say_text(f"I'm not sitting on my charger!")

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
    vector = BatteryManager()
    vector.querry_battery_state()