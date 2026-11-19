package sdkapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type activeBot struct {
	ESN         string `json:"esn"`
	Name        string `json:"name"`         // what to show: display name, else the robot's own name
	DisplayName string `json:"display_name"` // user-set display name only (empty if unset)
	IP          string `json:"ip_address"`
	Active      bool   `json:"active"`
	SecondsSeen int64  `json:"seconds_since_seen"`
}


func activeBotsHandler(w http.ResponseWriter) {
	bots := []activeBot{}
	for _, robot := range vars.GetBotInfo().Robots {
		name := robot.Name
		if name == "" {
			name = vars.RobotNameForESN(robot.Esn)
		}
		b := activeBot{
			ESN:         robot.Esn,
			Name:        name,
			DisplayName: robot.Name,
			IP:          robot.IPAddress,
			Active:      vars.BotIsActive(robot.Esn),
			SecondsSeen: -1,
		}
		if t, ok := vars.BotLastSeen(robot.Esn); ok {
			b.SecondsSeen = int64(time.Since(t) / time.Second)
		}
		bots = append(bots, b)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bots)
}

// setBotNameHandler sets or clears a bot's display name. it doesn't need the bot
// to be reachable, so it is handled before getRobot.
func setBotNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must use POST", http.StatusMethodNotAllowed)
		return
	}
	serial := strings.TrimSpace(r.FormValue("serial"))
	name := strings.TrimSpace(r.FormValue("name"))
	if serial == "" {
		fmt.Fprint(w, "error: serial is required")
		return
	}
	// empty name clears the display name
	if name != "" && !vars.ValidDisplayName(name) {
		fmt.Fprintf(w, "error: name must be 1-%d characters with no control characters", vars.MaxDisplayNameLen)
		return
	}
	if err := vars.SetBotDisplayName(serial, name); err != nil {
		fmt.Fprint(w, "error: "+err.Error())
		return
	}
	fmt.Fprint(w, "success")
}
