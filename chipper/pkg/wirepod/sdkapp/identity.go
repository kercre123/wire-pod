package sdkapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/memories"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

var errRobotNotFound = errors.New("robot not found")

// setRobotIdentity sets the user-set nickname and/or LLM personality of a known robot (nil = leave as is,
// empty string = clear) and persists the bot info store. If saving fails, the change is rolled back.
func setRobotIdentity(esn string, name *string, personality *string) error {
	vars.BotInfoMu.Lock()
	defer vars.BotInfoMu.Unlock()
	for i, robot := range vars.BotInfo.Robots {
		if strings.EqualFold(esn, robot.Esn) {
			oldName, oldPersonality := robot.Name, robot.Personality
			if name != nil {
				vars.BotInfo.Robots[i].Name = *name
			}
			if personality != nil {
				vars.BotInfo.Robots[i].Personality = *personality
			}
			jsonBytes, err := json.Marshal(vars.BotInfo)
			if err == nil {
				err = os.WriteFile(vars.BotInfoPath, jsonBytes, 0644)
			}
			if err != nil {
				vars.BotInfo.Robots[i].Name, vars.BotInfo.Robots[i].Personality = oldName, oldPersonality
			}
			return err
		}
	}
	return errRobotNotFound
}

func robotKnown(esn string) bool {
	vars.BotInfoMu.RLock()
	defer vars.BotInfoMu.RUnlock()
	for _, robot := range vars.BotInfo.Robots {
		if strings.EqualFold(esn, robot.Esn) {
			return true
		}
	}
	return false
}

// ChatLine is one saved line of a robot's recent conversation.
type ChatLine struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// robotChats returns the user and assistant lines wire-pod has saved for a robot (the LLM's recent
// context), oldest first. It is empty unless saving conversations is on.
func robotChats(esn string) []ChatLine {
	lines := []ChatLine{}
	for _, chat := range vars.RememberedChats {
		if !strings.EqualFold(chat.ESN, esn) {
			continue
		}
		for _, m := range chat.Chats {
			if (m.Role == "user" || m.Role == "assistant") && strings.TrimSpace(m.Content) != "" {
				lines = append(lines, ChatLine{Role: m.Role, Content: m.Content})
			}
		}
	}
	return lines
}

func identityError(w http.ResponseWriter, err error) {
	switch {
	case err == errRobotNotFound:
		http.Error(w, "error: "+err.Error(), http.StatusNotFound)
	case err == memories.ErrEmpty || err == memories.ErrTooLong || err == memories.ErrBadESN:
		http.Error(w, "error: "+err.Error(), http.StatusBadRequest)
	default:
		logger.Println("Error saving robot identity/memories: " + err.Error())
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
	}
}

// handleRobotIdentityAPI serves the per-robot name, personality and memory endpoints. These only touch
// wire-pod's own data, so no robot connection is needed. Returns false if the path isn't one of them.
func handleRobotIdentityAPI(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	default:
		return false
	case "/api-sdk/set_robot_name", "/api-sdk/set_robot_identity":
		name := strings.Join(strings.Fields(r.FormValue("name")), " ")
		if utf8.RuneCountInString(name) > 32 {
			http.Error(w, "error: name must be 32 characters or fewer", http.StatusBadRequest)
			return true
		}
		var personality *string
		if r.URL.Path == "/api-sdk/set_robot_identity" {
			p := strings.TrimSpace(r.FormValue("personality"))
			if utf8.RuneCountInString(p) > 1000 {
				http.Error(w, "error: personality must be 1000 characters or fewer", http.StatusBadRequest)
				return true
			}
			personality = &p
		}
		if err := setRobotIdentity(strings.TrimSpace(r.FormValue("serial")), &name, personality); err != nil {
			identityError(w, err)
			return true
		}
		fmt.Fprint(w, "done")
	case "/api-sdk/get_robot_chats":
		serial := strings.TrimSpace(r.FormValue("serial"))
		if !robotKnown(serial) {
			identityError(w, errRobotNotFound)
			return true
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(robotChats(serial))
	case "/api-sdk/get_robot_memories", "/api-sdk/add_robot_memory", "/api-sdk/delete_robot_memory", "/api-sdk/clear_robot_memories":
		serial := strings.TrimSpace(r.FormValue("serial"))
		if !robotKnown(serial) {
			identityError(w, errRobotNotFound)
			return true
		}
		var err error
		switch r.URL.Path {
		case "/api-sdk/get_robot_memories":
			var list []memories.Memory
			if list, err = memories.List(serial); err == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(list)
				return true
			}
		case "/api-sdk/add_robot_memory":
			_, _, err = memories.Add(serial, r.FormValue("text"))
		case "/api-sdk/delete_robot_memory":
			var found bool
			if found, err = memories.Delete(serial, r.FormValue("id")); err == nil && !found {
				http.Error(w, "error: memory not found", http.StatusNotFound)
				return true
			}
		case "/api-sdk/clear_robot_memories":
			err = memories.Clear(serial)
		}
		if err != nil {
			identityError(w, err)
			return true
		}
		fmt.Fprint(w, "done")
	}
	return true
}
