# Key

## speechrequest
-	Contains code for dealing with the intent types, and has functions to help convert stream bytes to ones which are stt-engine-friendly
-	speechrequest refers to the type which every intent type gets turned into

## preqs
-	Process request functions, the first which get launched in the chain of code

## stintent
-   Speech-to-intent, for services like Picovoice Rhino

## stt
-	Speech-to-text functions, where you would put your own STT engine implementation

## ttr
-	Text-to-response, takes text from wirepod-stt, turns it into a response, and sends it to the bot. Does all intent parsing and stuff as well

## config-ws
-	Webserver for custom intents and such

## sdkapp
-   App for configuring bot settings and controlling bots
