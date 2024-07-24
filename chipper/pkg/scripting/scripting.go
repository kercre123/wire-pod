package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	lualibs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
)

/*

assumeBehaviorControl(priority int)
	-	10,20,30 (10 highest priority, overriding behaviors. 30 lowest)
releaseBehaviorControl()

<these require behavior control>
<goroutine determines whether the function blocks or not>
sayText(text string, goroutine bool)
playAnimation(animation string, goroutine bool)

*/

type ExternalLuaRequest struct {
	ESN    string `json:"esn"`
	Script string `json:"script"`
}

type Bot struct {
	ESN   string
	Robot *vector.Vector
}

func sayText(L *lua.LState) int {
	textToSay := L.ToString(1)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.SayText(L.Context(), &vectorpb.SayTextRequest{Text: textToSay, UseVectorVoice: true, DurationScalar: 1.0})
		return err
	})
	return 0
}

func playAnimation(L *lua.LState) int {
	animToPlay := L.ToString(1)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.PlayAnimation(L.Context(), &vectorpb.PlayAnimationRequest{Animation: &vectorpb.Animation{Name: animToPlay}, Loops: 1})
		return err
	})
	return 0
}

// get robot from LState
func gRfLS(L *lua.LState) *vector.Vector {
	ud := L.GetGlobal("bot").(*lua.LUserData)
	bot := ud.Value.(*Bot)
	return bot.Robot
}

func MakeLuaState(esn string, validating bool) (*lua.LState, error) {
	L := lua.NewState()
	lualibs.Preload(L)
	L.SetContext(context.Background())
	L.SetGlobal("sayText", L.NewFunction(sayText))
	L.SetGlobal("playAnimation", L.NewFunction(playAnimation))
	SetBControlFunctions(L)
	ud := L.NewUserData()
	if !validating {
		rob, err := vars.GetRobot(esn)
		if err != nil {
			return nil, err
		}
		ctx, can := context.WithTimeout(context.Background(), time.Second*3)
		defer can()
		_, err = rob.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			return nil, err
		}
		ud.Value = &Bot{ESN: esn, Robot: rob}
		L.SetGlobal("bot", ud)
	}
	return L, nil
}

func executeWithGoroutine(L *lua.LState, fn func(L *lua.LState) error) {
	goroutine := L.ToBool(2)
	if goroutine {
		go func() {
			err := fn(L)
			if err != nil {
				logger.Println("LUA: failure: " + err.Error())
			}
		}()
	} else {
		err := fn(L)
		if err != nil {
			logger.Println("LUA: failure: " + err.Error())
		}
	}
}

func RunLuaScript(esn string, luaScript string) error {
	L, err := MakeLuaState(esn, false)
	if err != nil {
		return err
	}
	defer L.Close()

	if err := L.DoString(luaScript); err != nil {
		return err
	}
	L.DoString("releaseBehaviorControl()")
	return nil
}

func ValidateLuaScript(luaScript string) error {
	L, _ := MakeLuaState("", true)
	defer L.Close()

	err := L.DoString(fmt.Sprintf("return function() %s end", luaScript))
	if err != nil {
		return err
	}
	return nil
}

func ScriptingAPI(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api-lua/run_script":
		fBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request body couldn't be read: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var scriptReq ExternalLuaRequest
		err = json.Unmarshal(fBody, &scriptReq)
		if err != nil {
			http.Error(w, "request body couldn't be unmarshalled: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = RunLuaScript(scriptReq.ESN, scriptReq.Script)
		if err != nil {
			logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func RegisterScriptingAPI() {
	http.HandleFunc("/api-lua/", ScriptingAPI)
}
