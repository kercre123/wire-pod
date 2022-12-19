#!/bin/bash

systemctl stop anki-robot.target
echo "Resetting Vector to Onboarding mode (user data is NOT being reset, all faces and photos and stuff will persist), screen will be blank for a minute"
sleep 2
chmod +rwx /anki/data/assets/cozmo_resources/config/server_config.json
rm -f /data/data/com.anki.victor/persistent/token/token.jwt
rm -f /data/data/com.anki.victor/persistent/onboarding/onboardingState.json
rm -f /data/etc/robot.pem
rm -rf /data/protected
mkdir -p /data/etc
openssl genrsa -out /data/etc/robot.pem 2048
chown net:anki /data/etc/robot.pem
chmod 440 /data/etc/robot.pem
rm -rf /data/vic-gateway/*
sync
/etc/initscripts/ankiinit
sync
vic-gateway-cert
sync
systemctl start anki-robot.target
