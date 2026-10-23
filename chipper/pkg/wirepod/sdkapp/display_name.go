package sdkapp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const maxRobotDisplayNameRunes = 48

var robotDisplayNameMu sync.Mutex

func normalizedRobotDisplayName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) > maxRobotDisplayNameRunes {
		return "", fmt.Errorf("display name must be %d characters or fewer", maxRobotDisplayNameRunes)
	}
	if strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("display name cannot contain control characters")
	}
	return name, nil
}

func robotDisplayName(serial string) (string, bool) {
	robotDisplayNameMu.Lock()
	defer robotDisplayNameMu.Unlock()
	for _, robot := range vars.BotInfo.Robots {
		if strings.EqualFold(strings.TrimSpace(serial), robot.Esn) {
			return robot.DisplayName, true
		}
	}
	return "", false
}

func setRobotDisplayName(serial, name string) error {
	name, err := normalizedRobotDisplayName(name)
	if err != nil {
		return err
	}
	robotDisplayNameMu.Lock()
	defer robotDisplayNameMu.Unlock()
	for i := range vars.BotInfo.Robots {
		if !strings.EqualFold(strings.TrimSpace(serial), vars.BotInfo.Robots[i].Esn) {
			continue
		}
		vars.BotInfo.Robots[i].DisplayName = name
		contents, err := json.Marshal(vars.BotInfo)
		if err != nil {
			return fmt.Errorf("encode robot display name: %w", err)
		}
		if err := os.WriteFile(vars.BotInfoPath, contents, 0600); err != nil {
			return fmt.Errorf("save robot display name: %w", err)
		}
		return nil
	}
	return fmt.Errorf("robot not found")
}
