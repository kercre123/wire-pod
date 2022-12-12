package sdkapp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digital-dream-labs/hugh/grpc/client"
	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"hz.tools/mjpeg"
)

const serverFiles string = "./webroot/sdkapp"

var sdkAddress string = "localhost:443"
var robotGUID string = "tni1TRsTRTaNSapjo0Y+Sw=="
var globalGUID string = "tni1TRsTRTaNSapjo0Y+Sw=="

var robot *vector.Vector
var bcAssumption bool = false
var ctx context.Context
var camStreamEnable bool = false
var camStreamClient vectorpb.ExternalInterface_CameraFeedClient

var transCfg = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore SSL warnings
}

func assumeBehaviorControl(priority string) {
	var controlRequest *vectorpb.BehaviorControlRequest
	if priority == "high" {
		controlRequest = &vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{
					Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
				},
			},
		}
	} else {
		controlRequest = &vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{
					Priority: vectorpb.ControlRequest_DEFAULT,
				},
			},
		}
	}
	go func() {
		start := make(chan bool)
		stop := make(chan bool)
		bcAssumption = true
		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Conn.BehaviorControl(
				ctx,
			)
			if err != nil {
				log.Println(err)
				return
			}

			if err := r.Send(controlRequest); err != nil {
				log.Println(err)
				return
			}

			for {
				ctrlresp, err := r.Recv()
				if err != nil {
					log.Println(err)
					return
				}
				if ctrlresp.GetControlGrantedResponse() != nil {
					start <- true
					break
				}
			}

			for {
				select {
				case <-stop:
					if err := r.Send(
						&vectorpb.BehaviorControlRequest{
							RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
								ControlRelease: &vectorpb.ControlRelease{},
							},
						},
					); err != nil {
						log.Println(err)
						return
					}
					return
				default:
					continue
				}
			}
			// * end - modified from official vector-go-sdk
		}()
		for range start {
			for {
				if bcAssumption {
					time.Sleep(time.Millisecond * 500)
				} else {
					break
				}
			}
			stop <- true
		}
	}()
}

func sayText(text string) {
	_, _ = robot.Conn.SayText(
		ctx,
		&vectorpb.SayTextRequest{
			Text:           text,
			UseVectorVoice: true,
			DurationScalar: 1.0,
		},
	)
}

func driveWheelsForward(lw float32, rw float32, lwtwo float32, rwtwo float32) {
	_, _ = robot.Conn.DriveWheels(
		ctx,
		&vectorpb.DriveWheelsRequest{
			LeftWheelMmps:   lw,
			RightWheelMmps:  rw,
			LeftWheelMmps2:  lwtwo,
			RightWheelMmps2: rwtwo,
		},
	)
}

func moveLift(speed float32) {
	_, _ = robot.Conn.MoveLift(
		ctx,
		&vectorpb.MoveLiftRequest{
			SpeedRadPerSec: speed,
		},
	)
}

func moveHead(speed float32) {
	_, _ = robot.Conn.MoveHead(
		ctx,
		&vectorpb.MoveHeadRequest{
			SpeedRadPerSec: speed,
		},
	)
}

func releaseBehaviorControl() {
	bcAssumption = false
}

func sendAppIntent(intent string, param string) {
	if param == "" {
		_, _ = robot.Conn.AppIntent(
			ctx,
			&vectorpb.AppIntentRequest{
				Intent: intent,
			},
		)
	} else {
		_, _ = robot.Conn.AppIntent(
			ctx,
			&vectorpb.AppIntentRequest{
				Intent: intent,
				Param:  param,
			},
		)
	}
}

func getGUID() string {
	clientGUID := string(robotGUID)
	return clientGUID
}

func setCustomEyeColor(hue string, sat string) {
	clientGUID := getGUID()
	if !strings.Contains(clientGUID, "error") {
		url := "https://" + sdkAddress + "/v1/update_settings"
		var updateJSON = []byte(`{"update_settings": true, "settings": {"custom_eye_color": {"enabled": true, "hue": ` + hue + `, "saturation": ` + sat + `} } }`)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(updateJSON))
		req.Header.Set("Authorization", "Bearer "+clientGUID)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Transport: transCfg}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	} else {
		log.Println("GUID not there")
	}
}

func getSDKSettings() []byte {
	resp, err := robot.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
		JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_SETTINGS},
	})
	if err != nil {
		return []byte(err.Error())
	}
	json := resp.NamedJdocs[0].Doc.JsonDoc
	return []byte(json)
}

func setPresetEyeColor(value string) {
	clientGUID := getGUID()
	if !strings.Contains(clientGUID, "error") {
		url := "https://" + sdkAddress + "/v1/update_settings"
		var updateJSON = []byte(`{"update_settings": true, "settings": {"custom_eye_color": {"enabled": false}, "eye_color": ` + value + `} }`)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(updateJSON))
		req.Header.Set("Authorization", "Bearer "+clientGUID)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Transport: transCfg}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	} else {
		log.Println("GUID not there")
	}
}

func setSettingSDKstring(setting string, value string) {
	clientGUID := getGUID()
	if !strings.Contains(clientGUID, "error") {
		url := "https://" + sdkAddress + "/v1/update_settings"
		var updateJSON = []byte(`{"update_settings": true, "settings": {"` + setting + `": "` + value + `" } }`)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(updateJSON))
		req.Header.Set("Authorization", "Bearer "+clientGUID)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Transport: transCfg}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	} else {
		log.Println("GUID not there")
	}
}

func setSettingSDKintbool(setting string, value string) {
	clientGUID := getGUID()
	if !strings.Contains(clientGUID, "error") {
		url := "https://" + sdkAddress + "/v1/update_settings"
		var updateJSON = []byte(`{"update_settings": true, "settings": {"` + setting + `": ` + value + ` } }`)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(updateJSON))
		req.Header.Set("Authorization", "Bearer "+clientGUID)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Transport: transCfg}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	} else {
		log.Println("GUID not there")
	}
}

type RobotSDKInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
		GUID      string `json:"guid"`
	} `json:"robots"`
}

type options struct {
	SerialNo  string
	RobotName string `ini:"name"`
	CertPath  string `ini:"cert"`
	Target    string `ini:"ip"`
	Token     string `ini:"guid"`
}

func NewWP(serial string, useGlobal bool) (*vector.Vector, error) {
	if serial == "" {
		log.Fatal("please use the -serial argument and set it to your robots serial number")
		return nil, fmt.Errorf("Configuration options missing")
	}

	cfg := options{}
	wirepodPath := os.Getenv("WIREPOD_HOME")
	if len(wirepodPath) == 0 {
		wirepodPath = "."
	}
	jsonBytes, err := os.ReadFile("jdocs/botSdkInfo.json")
	if err != nil {
		log.Println("vector-go-sdk error: Error opening " + "jdocs/botSdkInfo.json" + ", likely doesn't exist")
		return nil, err
	}
	var robotSDKInfo RobotSDKInfoStore
	json.Unmarshal(jsonBytes, &robotSDKInfo)
	matched := false
	for _, robot := range robotSDKInfo.Robots {
		if strings.TrimSpace(strings.ToLower(robot.Esn)) == strings.TrimSpace(strings.ToLower(serial)) {
			cfg.Target = robot.IPAddress + ":443"
			matched = true
			if robot.GUID == "" {
				robot.GUID = robotSDKInfo.GlobalGUID
				cfg.Token = robotSDKInfo.GlobalGUID
			} else {
				cfg.Token = robot.GUID
				fmt.Println("Using " + cfg.Token)
			}
		}
	}
	if !matched {
		log.Println("vector-go-sdk error: serial did not match any bot in bot json")
		return nil, errors.New("vector-go-sdk error: serial did not match any bot in bot json")
	}
	c, err := client.New(
		client.WithTarget(cfg.Target),
		client.WithInsecureSkipVerify(),
	)
	if err != nil {
		return nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, err
	}

	cfg.SerialNo = serial

	if useGlobal {
		cfg.Token = globalGUID
	}

	return vector.New(
		vector.WithTarget(cfg.Target),
		vector.WithSerialNo(cfg.SerialNo),
		vector.WithToken(cfg.Token),
	)
}

func SdkapiHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	case r.URL.Path == "/api-sdk/alexa_sign_in":
		robot.Conn.AlexaOptIn(ctx, &vectorpb.AlexaOptInRequest{
			OptIn: true,
		})
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/alexa_sign_out":
		robot.Conn.AlexaOptIn(ctx, &vectorpb.AlexaOptInRequest{
			OptIn: false,
		})
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/cloud_intent":
		intent := r.FormValue("intent")
		sendAppIntent(intent, "")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/set_timer":
		secs := r.FormValue("secs")
		sendAppIntent("intent_clock_settimer", `{"timer_duration":"`+secs+`","unit":"s"}`)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/eye_color":
		eye_color := r.FormValue("color")
		setPresetEyeColor(eye_color)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/custom_eye_color":
		hue := r.FormValue("hue")
		sat := r.FormValue("sat")
		setCustomEyeColor(hue, sat)
		fmt.Fprintf(w, hue+sat)
		return
	case r.URL.Path == "/api-sdk/volume":
		volume := r.FormValue("volume")
		setSettingSDKintbool("master_volume", volume)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/locale":
		locale := r.FormValue("locale")
		setSettingSDKstring("locale", locale)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/location":
		location := r.FormValue("location")
		setSettingSDKstring("default_location", location)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/timezone":
		timezone := r.FormValue("timezone")
		setSettingSDKstring("time_zone", timezone)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/get_sdk_info":
		jsonBytes, err := os.ReadFile("./jdocs/botSdkInfo.json")
		if err != nil {
			fmt.Fprintf(w, "error reading file")
			return
		}
		fmt.Fprint(w, string(jsonBytes))
		return
	case r.URL.Path == "/api-sdk/get_sdk_settings":
		settings := getSDKSettings()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(settings)
		return
	case r.URL.Path == "/api-sdk/time_format_12":
		setSettingSDKintbool("clock_24_hour", "false")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/time_format_24":
		setSettingSDKintbool("clock_24_hour", "true")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/temp_c":
		setSettingSDKintbool("temp_is_fahrenheit", "false")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/temp_f":
		setSettingSDKintbool("temp_is_fahrenheit", "true")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/button_hey_vector":
		setSettingSDKintbool("button_wakeword", "0")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/button_alexa":
		setSettingSDKintbool("button_wakeword", "1")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/initSDK":
		serial := r.FormValue("serial")
		if serial == "" {
			fmt.Fprintf(w, "no serial given")
			return
		}
		var err error
		robot, err = NewWP(serial, false)
		if err != nil {
			fmt.Fprint(w, "failed: "+err.Error())
			return
		}
		sdkAddress = robot.Cfg.Target
		fmt.Println("sdkApp: Initiating SDK with " + robot.Cfg.SerialNo)
		robotGUID = robot.Cfg.Token
		_, err = robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			fmt.Println("Failed to initiate SDK with normal GUID, trying global GUID")
			robot, err = NewWP(serial, true)
			if err != nil {
				fmt.Fprint(w, "failed: "+err.Error())
				return
			}
			sdkAddress = robot.Cfg.Target
			fmt.Println("sdkApp: Initiating SDK with " + robot.Cfg.SerialNo)
			robotGUID = robot.Cfg.Token
			_, err = robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
			if err != nil {
				fmt.Fprintf(w, "failed to make test request: "+err.Error())
				return
			}
		}
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/assume_behavior_control":
		fmt.Fprintf(w, "success")
		assumeBehaviorControl(r.FormValue("priority"))
		return
	case r.URL.Path == "/api-sdk/release_behavior_control":
		fmt.Fprintf(w, "success")
		releaseBehaviorControl()
		return
	case r.URL.Path == "/api-sdk/say_text":
		sayText(r.FormValue("text"))
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/move_wheels":
		lw, _ := strconv.Atoi(r.FormValue("lw"))
		rw, _ := strconv.Atoi(r.FormValue("rw"))
		driveWheelsForward(float32(lw), float32(rw), float32(lw), float32(rw))
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/begin_cam_stream":
		camStreamClient, _ = robot.Conn.CameraFeed(ctx, &vectorpb.CameraFeedRequest{})
		camStreamEnable = true
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/stop_cam_stream":
		camStreamEnable = false
		camStreamClient = nil
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/move_lift":
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		moveLift(float32(speed))
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/move_head":
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		moveHead(float32(speed))
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/mirror_mode":
		enable := r.FormValue("enable")
		if enable == "true" {
			_, err := robot.Conn.EnableMirrorMode(
				ctx,
				&vectorpb.EnableMirrorModeRequest{
					Enable: true,
				},
			)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		} else {
			_, err := robot.Conn.EnableMirrorMode(
				ctx,
				&vectorpb.EnableMirrorModeRequest{
					Enable: false,
				},
			)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		}
		fmt.Fprint(w, "success")
		return
	}
}

func BeginServer() {
	ctx = context.Background()
	camStream := mjpeg.NewStream()
	i := image.NewGray(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: 640, Y: 360},
	})
	go func() {
		for {
			if camStreamEnable {
				response, _ := camStreamClient.Recv()
				imageBytes := response.GetData()
				img, _, _ := image.Decode(bytes.NewReader(imageBytes))
				camStream.Update(img)
			} else {
				for j := range i.Pix {
					i.Pix[j] = uint8(rand.Uint32())
				}

				time.Sleep(time.Second)
				camStream.Update(i)
			}
		}
	}()
	http.HandleFunc("/api-sdk/", SdkapiHandler)
	fileServer := http.FileServer(http.Dir(serverFiles))
	http.Handle("/sdk-app", fileServer)
	http.Handle("/stream", camStream)
	fmt.Println("Starting SDK app")

	// fmt.Printf("Starting server at port 8081\n")
	// if err := http.ListenAndServe(":8081", nil); err != nil {
	// 	log.Fatal(err)
	// }
}
