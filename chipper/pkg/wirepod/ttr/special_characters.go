package wirepod_ttr

//
//special_characters.go
//

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

    "github.com/kercre123/wire-pod/chipper/pkg/logger"
)


//
// Special Character Handling section
//

// special characters replacements
var specialCharactersReplacements = strings.NewReplacer(
    "‘", "'", "’", "'", "“", "\"", "”", "\"",
    "—", " - ", "–", " - ", "…", ",",
    "\u00A0", " ", "  ", " ",
    "¼", "1/4", "½", "1/2", "¾", "3/4", "×", "x", "@", "(a)",
    "facepalm", "face-palm", "chatbot", "chat-bot", "bedsheet", "bed-sheet", " **", " ,",
    "&", "and", "^", "caret", "#", "hashtag",
    "%20", " ", "%21", "!", "%22", "\"", "%23", "#", "%24", "$", "%25", "%",
    "%26", "&", "%27", "'", "%28", "(", "%29", ")", "%2A", "*", "%2B", "+",
    "%2C", ",", "%2D", "-", "%2E", ".", "%2F", "/", "%3A", ":", "%3B", ";",
    "%3C", "<", "%3D", "=", "%3E", ">", "%3F", "?", "%40", "@", "%5B", "[",
    "%5C", "\\", "%5D", "]", "%5E", "^", "%5F", "_", "%60", "", "%7B", "{",
    "%7C", "|", "%7D", "}", "%7E", "~",
    "±", "+/-", "÷", "/", "√", "sqrt", "∞", "infinity",
    "≈", "~", "≠", "!=", "≡", "==", "≤", "<=", "≥", ">=", "°", "deg",
    "π", "pi", "∆", "delta", "∑", "sum", "∏", "prod",
    "€", "EUR", "£", "GBP", "¥", "JPY", "₹", "INR", "$", "USD",
    "©", "(c)", "®", "(r)", "™", "(tm)", "†", "+", "‡", "++", "§", "SS",
    "•", "*", "‣", "*", "◦", "*",
    "₀", "0", "₁", "1", "₂", "2", "₃", "3", "₄", "4",
    "₅", "5", "₆", "6", "₇", "7", "₈", "8", "₉", "9",
)

// emoji patterns
var emojiPattern = `[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F004}]|[\x{1F0CF}]|[\x{1F18E}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]`

// phonetic map for Vector 
var phoneticReplacements = map[string]string{
    "AIs": "AY-EYES", "AI": "AY-EYE", "CrushN8r": "Crushinayter",
    "SpaceX": "Space-X", "H2O": "H-2-O", "Communicate": "Cum-Uni-Kate",
    "Aha": "Ah-ha", "A-ha": "Ah-ha",
}

// normalize Text
func normalizeText(str string) (string, error) {
    t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
    normalized, _, err := transform.String(t, str)
    return normalized, err
}

// checks if a rune is a non-spacing mark
func isMn(r rune) bool {
    return unicode.Is(unicode.Mn, r)
}

// case-insensitively replaces words/phrases
func replaceWords(str string) string {
    for find, replace := range phoneticReplacements {
        regex := regexp.MustCompile(`\b(?i)` + regexp.QuoteMeta(find) + `\b`)
        str = regex.ReplaceAllString(str, replace)
    }
    return str
}

// remove emojis using emojiPattern
func removeEmojis(str string) string {
    return regexp.MustCompile(emojiPattern).ReplaceAllString(str, "")
}

// ensures decimals are pronounced correctly
func replaceDecimals(str string) string {
    decimalPattern := regexp.MustCompile(`(\d+)\.(\d+)`)
    return decimalPattern.ReplaceAllStringFunc(str, func(match string) string {
        parts := decimalPattern.FindStringSubmatch(match)
        if len(parts) == 3 {
            return parts[1] + " point " + parts[2]
        }
        logger.LogUI("\n\n"+"replaceDecimals() match = " + match+"\n\n")
        return match
    })
}

// removeSpecialCharacters main removal and replacements
func removeSpecialCharacters(str string) (string, error) {
    normalized, err := normalizeText(str)
    if err != nil {
        return str, err // Returning original string for clarity on error
    }

    result := specialCharactersReplacements.Replace(normalized)
    result = replaceDecimals(result)
    result = replaceWords(result)
    result = removeEmojis(result)

    return result, nil
}

//
// //special_characters.go - END
//

