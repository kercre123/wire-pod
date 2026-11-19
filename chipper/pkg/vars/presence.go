package vars

import (
	"strings"
	"sync"
	"time"
)

const BotActiveWindow = 30 * time.Second

var (
	seenMu   sync.Mutex
	lastSeen = map[string]time.Time{}
)

func MarkBotSeen(esn string) {
	esn = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(esn), "vic:"))
	if !ValidESN(esn) {
		return
	}
	seenMu.Lock()
	lastSeen[esn] = time.Now()
	seenMu.Unlock()
}

func MarkBotSeenByIP(ip string) {
	if ip == "" {
		return
	}
	for _, bot := range GetBotInfo().Robots {
		if bot.IPAddress == ip {
			MarkBotSeen(bot.Esn)
			return
		}
	}
}

func BotLastSeen(esn string) (time.Time, bool) {
	seenMu.Lock()
	defer seenMu.Unlock()
	t, ok := lastSeen[strings.ToLower(esn)]
	return t, ok
}

func BotIsActive(esn string) bool {
	t, ok := BotLastSeen(esn)
	return ok && time.Since(t) < BotActiveWindow
}

func RobotNameForESN(esn string) string {
	rinfoMu.Lock()
	defer rinfoMu.Unlock()
	for _, rinf := range recurringInfo {
		if strings.EqualFold(rinf.ESN, esn) {
			return rinf.ID
		}
	}
	return ""
}
