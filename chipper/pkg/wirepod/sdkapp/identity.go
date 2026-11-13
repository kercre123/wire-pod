package sdkapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
				err = writeFileAtomic(vars.BotInfoPath, jsonBytes, 0644)
			}
			if err != nil {
				vars.BotInfo.Robots[i].Name, vars.BotInfo.Robots[i].Personality = oldName, oldPersonality
			}
			return err
		}
	}
	return errRobotNotFound
}

// writeFileAtomic writes to a temp file in the same folder and renames it into place, so a failed or
// interrupted save can't leave a truncated botSdkInfo.json behind.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmpName, perm)
	}
	if err == nil {
		err = os.Rename(tmpName, path)
	}
	if err != nil {
		os.Remove(tmpName)
	}
	return err
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
		// set_robot_identity only changes the fields that are present in the request (an empty value clears
		// one), so posting just "name" leaves the personality alone and vice versa.
		r.ParseForm()
		_, hasName := r.Form["name"]
		_, hasPersonality := r.Form["personality"]
		if r.URL.Path == "/api-sdk/set_robot_name" {
			hasName, hasPersonality = true, false
		}
		var name, personality *string
		if hasName {
			n := strings.Join(strings.Fields(r.FormValue("name")), " ")
			if utf8.RuneCountInString(n) > 32 {
				http.Error(w, "error: name must be 32 characters or fewer", http.StatusBadRequest)
				return true
			}
			name = &n
		}
		if hasPersonality {
			p := strings.TrimSpace(r.FormValue("personality"))
			if utf8.RuneCountInString(p) > 1000 {
				http.Error(w, "error: personality must be 1000 characters or fewer", http.StatusBadRequest)
				return true
			}
			personality = &p
		}
		if err := setRobotIdentity(strings.TrimSpace(r.FormValue("serial")), name, personality); err != nil {
			identityError(w, err)
			return true
		}
		fmt.Fprint(w, "done")
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
