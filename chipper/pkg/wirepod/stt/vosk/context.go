package wirepod_vosk

import (
	"strings"

	"github.com/kercre123/chipper/pkg/vars"
	"github.com/kercre123/chipper/pkg/wirepod/localization"
)

var NumbersEN_US []string = []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety", "hundred", "seconds", "minutes", "hours", "minute", "second", "hour"}

func removeDuplicates(strings []string) []string {
	occurred := map[string]bool{}
	var result []string
	for _, str := range strings {
		if !occurred[str] {
			result = append(result, str)
			occurred[str] = true
		}
	}
	return result
}
func GetGrammerList(lang string) string {
	var wordsList []string
	var grammer string
	// add words in intent json
	for _, words := range vars.MatchListList {
		for _, word := range words {
			wors := strings.Split(word, " ")
			for _, wor := range wors {
				found := model.FindWord(wor)
				if found != -1 {
					wordsList = append(wordsList, wor)
				}
			}
		}
	}
	// add words in localization
	for _, str := range localization.ALL_STR {
		text := localization.GetText(str)
		wors := strings.Split(text, " ")
		for _, wor := range wors {
			found := model.FindWord(wor)
			if found != -1 {
				wordsList = append(wordsList, wor)
			}
		}
	}
	// add custom intent matches
	for _, intent := range vars.CustomIntents {
		for _, utterance := range intent.Utterances {
			wors := strings.Split(utterance, " ")
			for _, wor := range wors {
				found := model.FindWord(wor)
				if found != -1 {
					wordsList = append(wordsList, wor)
				}
			}
		}
	}
	// add numbers
	for _, wor := range NumbersEN_US {
		found := model.FindWord(wor)
		if found != -1 {
			wordsList = append(wordsList, wor)
		}
	}

	wordsList = removeDuplicates(wordsList)
	for i, word := range wordsList {
		if i == len(wordsList)-1 {
			grammer = grammer + `"` + word + `"`
		} else {
			grammer = grammer + `"` + word + `"` + ", "
		}
	}
	grammer = "[" + grammer + "]"
	return grammer
}
