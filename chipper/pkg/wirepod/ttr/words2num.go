package wirepod_ttr

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	lcztn "github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization" 
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

//initialize by default in english during chipper compilation
var textToNumber = map[string]int{
	lcztn.GetText(lcztn.STR_ZERO)		: 0, 
	lcztn.GetText(lcztn.STR_ONE)		: 1, 
	lcztn.GetText(lcztn.STR_TWO)		: 2, 
	lcztn.GetText(lcztn.STR_THREE)		: 3, 
	lcztn.GetText(lcztn.STR_FOUR)		: 4, 
	lcztn.GetText(lcztn.STR_FIVE)		: 5,
	lcztn.GetText(lcztn.STR_SIX)		: 6, 
	lcztn.GetText(lcztn.STR_SEVEN)		: 7, 
	lcztn.GetText(lcztn.STR_EIGHT)		: 8, 
	lcztn.GetText(lcztn.STR_NINE)		: 9, 
	lcztn.GetText(lcztn.STR_TEN)		: 10,
	lcztn.GetText(lcztn.STR_ELEVEN)		: 11, 
	lcztn.GetText(lcztn.STR_TWELVE)		: 12, 
	lcztn.GetText(lcztn.STR_THIRTEEN)	: 13, 
	lcztn.GetText(lcztn.STR_FOURTEEN)	: 14, 
	lcztn.GetText(lcztn.STR_FIFTEEN)	: 15,
	lcztn.GetText(lcztn.STR_SIXTEEN)	: 16, 
	lcztn.GetText(lcztn.STR_SEVENTEEN)	: 17, 
	lcztn.GetText(lcztn.STR_EIGHTEEN)	: 18, 
	lcztn.GetText(lcztn.STR_NINETEEN)	: 19, 
	lcztn.GetText(lcztn.STR_TWENTY)		: 20,
	lcztn.GetText(lcztn.STR_THIRTY)		: 30, 
	lcztn.GetText(lcztn.STR_FOURTY)		: 40, 
	lcztn.GetText(lcztn.STR_FIFTY)		: 50, 
	lcztn.GetText(lcztn.STR_SIXTY)		: 60,
	lcztn.GetText(lcztn.STR_SEVENTY)	: 70,
	lcztn.GetText(lcztn.STR_EIGHTY)		: 80,
	lcztn.GetText(lcztn.STR_NINETY)		: 90,
	lcztn.GetText(lcztn.STR_ONE_HUNDRED): 100,

}

func words2num(input string) string {

	initializeTextToNumberwithCurrentLocalization()

	containsNum, _ := regexp.MatchString(`\b\d+\b`, input)
	if os.Getenv("STT_SERVICE") == "whisper.cpp" && containsNum {
		return whisperSpeechtoNum(input)
	}
	totalSeconds := 0

	input = strings.ToLower(input)
	if strings.Contains(input, lcztn.GetText(lcztn.STR_ONE_HOUR)) || strings.Contains(input, lcztn.GetText(lcztn.STR_ONE_HOUR_ALT)) {
		return "3600"
	}

	str_regex_time_pattern := `(\d+|\w+(?:-\w+)?)\s*(`+lcztn.GetText(lcztn.STR_MINUTE)+`|`+lcztn.GetText(lcztn.STR_SECOND)+`|`+lcztn.GetText(lcztn.STR_HOUR)+`)s?`

	// timePattern := regexp.MustCompile(`(\d+|\w+(?:-\w+)?)\s*(minute|second|hour)s?`)
	timePattern := regexp.MustCompile(str_regex_time_pattern)

	matches := timePattern.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		unit := match[2]
		number := match[1]

		value, err := strconv.Atoi(number)
		if err != nil {
			value = mapTextToNumber(number)
		}

		switch unit {
		// minute	
		case lcztn.GetText(lcztn.STR_MINUTE):
			totalSeconds += value * 60
		// second	
		case lcztn.GetText(lcztn.STR_SECOND):
			totalSeconds += value
		// hour	
		case lcztn.GetText(lcztn.STR_HOUR):
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

func initializeTextToNumberwithCurrentLocalization () {
	textToNumber = map[string]int{
		lcztn.GetText(lcztn.STR_ZERO)		: 0, 
		lcztn.GetText(lcztn.STR_ONE)		: 1, 
		lcztn.GetText(lcztn.STR_TWO)		: 2, 
		lcztn.GetText(lcztn.STR_THREE)		: 3, 
		lcztn.GetText(lcztn.STR_FOUR)		: 4, 
		lcztn.GetText(lcztn.STR_FIVE)		: 5,
		lcztn.GetText(lcztn.STR_SIX)		: 6, 
		lcztn.GetText(lcztn.STR_SEVEN)		: 7, 
		lcztn.GetText(lcztn.STR_EIGHT)		: 8, 
		lcztn.GetText(lcztn.STR_NINE)		: 9, 
		lcztn.GetText(lcztn.STR_TEN)		: 10,
		lcztn.GetText(lcztn.STR_ELEVEN)		: 11, 
		lcztn.GetText(lcztn.STR_TWELVE)		: 12, 
		lcztn.GetText(lcztn.STR_THIRTEEN)	: 13, 
		lcztn.GetText(lcztn.STR_FOURTEEN)	: 14, 
		lcztn.GetText(lcztn.STR_FIFTEEN)	: 15,
		lcztn.GetText(lcztn.STR_SIXTEEN)	: 16, 
		lcztn.GetText(lcztn.STR_SEVENTEEN)	: 17, 
		lcztn.GetText(lcztn.STR_EIGHTEEN)	: 18, 
		lcztn.GetText(lcztn.STR_NINETEEN)	: 19, 
		lcztn.GetText(lcztn.STR_TWENTY)		: 20,
		lcztn.GetText(lcztn.STR_THIRTY)		: 30, 
		lcztn.GetText(lcztn.STR_FOURTY)		: 40, 
		lcztn.GetText(lcztn.STR_FIFTY)		: 50, 
		lcztn.GetText(lcztn.STR_SIXTY)		: 60,
		lcztn.GetText(lcztn.STR_SEVENTY)	: 70,
		lcztn.GetText(lcztn.STR_EIGHTY)		: 80,
		lcztn.GetText(lcztn.STR_NINETY)		: 90,
		lcztn.GetText(lcztn.STR_ONE_HUNDRED): 100,
	
	}
}
