package main

import "fmt"

// testing the plugin system

// when the voice request matches the strings stored in Utterances, the Action function will be performed
var Utterances = []string{"test plugin"}
var Name = "Test Plugin"

func Action(transcribedText string) string {
	fmt.Println("(in testPlugin) Printing transcribed text: " + transcribedText)
	return "intent_imperative_praise"
}
