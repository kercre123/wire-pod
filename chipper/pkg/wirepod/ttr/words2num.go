package wirepod_ttr

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// This file contains words2num. It is given the spoken text and returns a string which contains the true number.

func whisperSpeechtoNum(input string) string {
	// whisper returns actual numbers in its response
	// ex. "set a timer for 10 minutes and 11 seconds"
	totalSeconds := 0

	minutePattern := regexp.MustCompile(`(\d+)\s*minute`)
	secondPattern := regexp.MustCompile(`(\d+)\s*second`)

	minutesMatches := minutePattern.FindStringSubmatch(input)
	secondsMatches := secondPattern.FindStringSubmatch(input)

	if len(minutesMatches) > 1 {
		minutes, err := strconv.Atoi(minutesMatches[1])
		if err == nil {
			totalSeconds += minutes * 60
		}
	}
	if len(secondsMatches) > 1 {
		seconds, err := strconv.Atoi(secondsMatches[1])
		if err == nil {
			totalSeconds += seconds
		}
	}

	return strconv.Itoa(totalSeconds)
}

var textToNumber = map[string]int{
	"zero": 0, "one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
	"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	"eleven": 11, "twelve": 12, "thirteen": 13, "fourteen": 14, "fifteen": 15,
	"sixteen": 16, "seventeen": 17, "eighteen": 18, "nineteen": 19, "twenty": 20,
	"thirty": 30, "forty": 40, "fifty": 50, "sixty": 60,
}

func words2num(input string) string {
	containsNum, _ := regexp.MatchString(`\b\d+\b`, input)
	if os.Getenv("STT_SERVICE") == "whisper.cpp" && containsNum {
		return whisperSpeechtoNum(input)
	}
	totalSeconds := 0

	input = strings.ToLower(input)
	if strings.Contains(input, "one hour") || strings.Contains(input, "an hour") {
		return "3600"
	}

	timePattern := regexp.MustCompile(`(\d+|\w+(?:-\w+)?)\s*(minute|second|hour)s?`)

	matches := timePattern.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		unit := match[2]
		number := match[1]

		value, err := strconv.Atoi(number)
		if err != nil {
			value = mapTextToNumber(number)
		}

		switch unit {
		case "minute":
			totalSeconds += value * 60
		case "second":
			totalSeconds += value
		case "hour":
			totalSeconds += value * 3600
		}
	}

	return strconv.Itoa(totalSeconds)
}

func mapTextToNumber(text string) int {
	if val, ok := textToNumber[text]; ok {
		return val
	}
	parts := strings.Split(text, "-")
	sum := 0
	for _, part := range parts {
		if val, ok := textToNumber[part]; ok {
			sum += val
		}
	}
	return sum
}
