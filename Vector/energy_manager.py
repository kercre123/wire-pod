import json
import os

"""
    Every hour, Vector gains (10 x [Vector Level]) / 2 energy back. Vectors energy value must remain between min: 0, and max: 95 + ([Vector Level] * 5)
    This script is run every hour in a cronjob :)
"""
# Path to the config file
CONFIG_PATH = os.path.expanduser('~/wire-pod/chipper/webroot/robot_config.json')

def modify_energy_level(data):
    """
        Modify the robot_energy_level based on the robot_level.
    """
    
    # Calculate energy increase based on robot's level
    energy_increase = (10 * data['robot_level']) / 2
    print(f"Energy Increase: {energy_increase}")
    
    # Calculate the maximum allowable energy level
    max_energy = 95 + (data['robot_level'] * 5)
    print(f"Max Energy: {max_energy}")
    
    # Ensure energy level does not go above the calculated maximum
    new_energy_level = data['robot_energy_level'] + energy_increase
    data['robot_energy_level'] = min(max_energy, new_energy_level)

    return data

def main():
    # Load the JSON data
    with open(CONFIG_PATH, 'r') as file:
        data = json.load(file)
    
    # Modify the data
    modified_data = modify_energy_level(data)
    
    # Save the modified data back to the file
    with open(CONFIG_PATH, 'w') as file:
        json.dump(modified_data, file, indent=4)

if __name__ == '__main__':
    main()
