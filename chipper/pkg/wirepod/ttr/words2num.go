package wirepod_ttr

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// This file contains words2num. It is given the spoken text and returns a string which contains the true number.

var number int = 0

func basicspeechText2num(speechText string) int {
	if strings.Contains(speechText, "one") && !strings.Contains(speechText, "one hundred") {
		return 1
	} else if strings.Contains(speechText, "two") && !strings.Contains(speechText, "two hundred") {
		return 2
	} else if strings.Contains(speechText, "three") && !strings.Contains(speechText, "three hundred") {
		return 3
	} else if strings.Contains(speechText, "four") && !strings.Contains(speechText, "four hundred") {
		return 4
	} else if strings.Contains(speechText, "five") && !strings.Contains(speechText, "five hundred") {
		return 5
	} else if strings.Contains(speechText, "six ") && !strings.Contains(speechText, "six hundred") {
		return 6
	} else if strings.Contains(speechText, "seven ") && !strings.Contains(speechText, "seven hundred") {
		return 7
	} else if strings.Contains(speechText, "eight ") && !strings.Contains(speechText, "eight hundred") {
		return 8
	} else if strings.Contains(speechText, "nine ") && !strings.Contains(speechText, "nine hundred") {
		return 9
	}
	return 0
}

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

func words2num(speechText string) string {
	if os.Getenv("STT_SERVICE") == "whisper.cpp" {
		return whisperSpeechtoNum(speechText)
	}
	number = basicspeechText2num(speechText)
	if number == 0 {
		number = 1
	}
	if strings.Contains(speechText, "teen") {
		number = 10
		if strings.Contains(speechText, "thir") {
			number = 14
		} else if strings.Contains(speechText, "four") {
			number = 14
		} else if strings.Contains(speechText, "fif") {
			number = 15
		} else if strings.Contains(speechText, "six") {
			number = 16
		} else if strings.Contains(speechText, "seven") {
			number = 17
		} else if strings.Contains(speechText, "eight") {
			number = 18
		} else if strings.Contains(speechText, "nine") {
			number = 19
		}
	} else if strings.Contains(speechText, "ten") {
		number = 10
	} else if strings.Contains(speechText, "eleven") {
		number = 11
	} else if strings.Contains(speechText, "twelve") {
		number = 12
	}
	if strings.Contains(speechText, "twenty") {
		number = 20 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "thirty") {
		number = 30 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "forty") {
		number = 40 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "fifty") {
		number = 50 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "sixty") {
		number = 60 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "seventy") {
		number = 70 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "eighty") {
		number = 80 + basicspeechText2num(speechText)
	} else if strings.Contains(speechText, "ninety") {
		number = 90 + basicspeechText2num(speechText)
	}
	if strings.Contains(speechText, "hundred") {
		number = number + 100
	}
	if strings.Contains(speechText, "minute") {
		number = number * 60
	}
	return strconv.Itoa(number)
}
