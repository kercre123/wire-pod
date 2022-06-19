#!/bin/bash

if [[ ! -d ./chipper ]]; then
  echo "This must be run in the jank-escape-pod/ directory."
  exit 0
fi

git pull
#cd chipper
#./build.sh
#cd ..
echo
echo "Updated!"
echo
