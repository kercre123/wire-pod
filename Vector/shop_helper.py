import json
from vector_helpers import VectorHelpers

class ShopHelper:
    def __init__(self):
        self.shop_items_path = "../wire-pod/chipper/webroot/shop_items.json"
        self.helpers = VectorHelpers()
    

    def buy_item(self, item_id):
        """
            Accepts item_id as the id from the list of items in shop_items.json. 
            Validates the item and the transaction. 

            Returns: Bool, String
        """
        robot_data = self.helpers.read_json_file()
        shop_items = self.helpers.read_json_file(self.shop_items_path)

        # Get item from shop
        item = next((item for item in shop_items if item["id"] == item_id), None)
        
        # Verify the item
        if item is None:
            return False, f"Error: Item ID: {item_id} does not exist."
        
        # Validate if the robot has enough coins to afford this item
        if robot_data["robot_wallet"] < item["cost"]:
            amount_needed = item["cost"] - robot_data["robot_wallet"]
            return False, f"Error: Invalid transaction. You need {amount_needed} more coins to afford this."
        
        # Deduct the cost of the item and add it to the robot's inventory
        robot_data["robot_wallet"] -= item["cost"]
        robot_data["items"].append(item_id)

        self.helpers.write_json_file(robot_data)
        return True, f"Successfully purchased {item['title']} for {item["cost"]} coins."


