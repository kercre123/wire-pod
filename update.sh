#!/bin/bash

if [[ ! -d ./chipper ]]; then
  echo "This must be run in the wire-pod/ directory."
  exit 1
fi

if [[ $EUID != "0" ]]; then
  echo "This must be run as root."
  exit 1
fi

git fetch --all
git reset --hard origin/main
echo
echo "Updated!"
echo
