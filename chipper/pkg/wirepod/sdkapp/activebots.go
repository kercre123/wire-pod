package sdkapp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type activeBot struct {
	ESN         string `json:"esn"`
	Name        string `json:"name"`
	IP          string `json:"ip_address"`
	Active      bool   `json:"active"`
	SecondsSeen int64  `json:"seconds_since_seen"`
}

func activeBotsHandler(w http.ResponseWriter) {
	bots := []activeBot{}
	for _, robot := range vars.GetBotInfo().Robots {
		b := activeBot{
			ESN:         robot.Esn,
			Name:        vars.RobotNameForESN(robot.Esn),
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
