#!/usr/bin/env python3

# Copyright (c) 2018 Anki, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License in the file LICENSE.txt or at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Display an image on Vector's face
"""

import os
import sys
import time

try:
    from PIL import Image
except ImportError:
    sys.exit("Cannot import from PIL: Do `pip3 install --user Pillow` to install")

import anki_vector
from anki_vector.util import degrees


def main():
    args = anki_vector.util.parse_command_args()

    with anki_vector.Robot(args.serial) as robot:
        # If necessary, move Vector's Head and Lift to make it easy to see his face
        robot.behavior.set_head_angle(degrees(45.0))
        robot.behavior.set_lift_height(0.0)

        # target_path = "/home/connorbailey/VectorConfig/face_images/cozmo_image.jpg"
        target_path = "/home/connorbailey/VectorConfig/face_images/numbers/font-1.png"

        # Load an image
        image_file = Image.open(target_path)

        # Convert the image to the format used by the Screen
        print("Display image on Vector's face...")
        screen_data = anki_vector.screen.convert_image_to_screen_data(image_file)

        duration_s = 4.0
        robot.screen.set_screen_with_image_data(screen_data, duration_s)
        time.sleep(duration_s)


if __name__ == "__main__":
    main()