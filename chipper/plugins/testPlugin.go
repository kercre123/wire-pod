package main

import "fmt"

// plugin example

// when the voice request matches the strings stored in Utterances, the Action function will be performed
var Utterances = []string{"test plugin"}
var Name = "Test Plugin"

func Action(transcribedText string, botSerial string) string {
	// you are given the full transcription of what was said by the user and the serial number of the bot
	fmt.Println("(in testPlugin) Transcribed text: "+transcribedText, ", serial number: "+botSerial)
	return "intent_imperative_praise"
}
