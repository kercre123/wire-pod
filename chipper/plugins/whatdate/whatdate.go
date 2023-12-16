package main

import (
	"strconv"
	"strings"
	"time"
)

var Utterances = []string{"what day is it", "date today", "date"}
var Name = "Correct Date"

func stripOutTriggerWords(s string) string {
	result := strings.Replace(s, "simon says", "", 1)
	result = strings.Replace(result, "repeat", "", 1)
	return result
}

func CountWords(s string) int {
	return len(strings.Fields(s))
}

//Example go plugin that give back the correct date

func Action(transcribedText string, botSerial string) (string, string) {
	year, month, day := time.Now().Date()
	yearSring := strconv.FormatInt(int64(year), 10)

	VECTOR_PHRASE := "The date is " + month.String() + " " + strconv.FormatInt(int64(day), 10) + ", " + yearSring + " "

	return "intent_imperative_praise", VECTOR_PHRASE
}
