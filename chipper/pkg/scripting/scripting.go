package scripting

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
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
sleep(milliseconds int)
moveLift(radpersecond int)
moveHead(radpersecond int)
// leftWheelmmps2 and rightWheelmmps2 are what you want the wheels to accelerate to. if you want
// the wheels to accelerate immediately, just set leftWheelmmps2 and rightWheelmmps2 to 0
moveWheels(leftWheelmmps, rightWheelmmps, leftWheelmmps2, rightWheelmmps2 int)
// won't block
showImage(filePath string, durationMs int)

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
	}, false)
	return 0
}

func playAnimation(L *lua.LState) int {
	animToPlay := L.ToString(1)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.PlayAnimation(L.Context(), &vectorpb.PlayAnimationRequest{Animation: &vectorpb.Animation{Name: animToPlay}, Loops: 1})
		return err
	}, false)
	return 0
}

func sleep(L *lua.LState) int {
	sleepInMS := L.ToInt(1)
	time.Sleep(time.Millisecond * time.Duration(sleepInMS))
	return 0
}

func moveHead(L *lua.LState) int {
	headSpeed := L.ToInt(1)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.MoveHead(L.Context(), &vectorpb.MoveHeadRequest{SpeedRadPerSec: float32(headSpeed)})
		return err
	}, true)
	return 0
}

func moveLift(L *lua.LState) int {
	liftSpeed := L.ToInt(1)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.MoveLift(L.Context(), &vectorpb.MoveLiftRequest{SpeedRadPerSec: float32(liftSpeed)})
		return err
	}, true)
	return 0
}

func moveWheels(L *lua.LState) int {
	leftWheelSpeed := L.ToInt(1)
	rightWheelSpeed := L.ToInt(2)
	leftWheelSpeed2 := L.ToInt(3)
	rightWheelSpeed2 := L.ToInt(4)
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.DriveWheels(L.Context(),
			&vectorpb.DriveWheelsRequest{
				LeftWheelMmps:   float32(leftWheelSpeed),
				RightWheelMmps:  float32(rightWheelSpeed),
				LeftWheelMmps2:  float32(leftWheelSpeed2),
				RightWheelMmps2: float32(rightWheelSpeed2),
			})
		return err
	}, true)
	return 0
}

// later
func showImageOnScreen(L *lua.LState) int {
	filePath := L.ToString(1)
	duration := L.ToInt(2)
	f, err := os.Open(filePath)
	if err != nil {
		logger.Println("Lua error: unable to open image file:", err)
		return 1
	}
	img, _, err := image.Decode(f)
	if err != nil {
		logger.Println("Lua error: file is not an image:", err)
		return 1
	}
	pixels := ConvertPixelsToRawBitmap(img, 100)
	buf := new(bytes.Buffer)
	for _, ui := range pixels {
		binary.Write(buf, binary.LittleEndian, ui)
	}
	executeWithGoroutine(L, func(L *lua.LState) error {
		_, err := gRfLS(L).Conn.DisplayFaceImageRGB(L.Context(),
			&vectorpb.DisplayFaceImageRGBRequest{
				FaceData:         buf.Bytes(),
				DurationMs:       uint32(duration),
				InterruptRunning: true,
			})
		return err
	}, true)
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
	L.SetGlobal("sleep", L.NewFunction(sleep))
	L.SetGlobal("moveHead", L.NewFunction(moveHead))
	L.SetGlobal("moveLift", L.NewFunction(moveLift))
	L.SetGlobal("moveWheels", L.NewFunction(moveWheels))
	L.SetGlobal("showImage", L.NewFunction(showImageOnScreen))
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

func executeWithGoroutine(L *lua.LState, fn func(L *lua.LState) error, force bool) {
	var goroutine bool
	if force {
		goroutine = true
	} else {
		goroutine = L.ToBool(2)
	}
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
