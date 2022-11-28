package sdkapp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/vector-go-sdk/pkg/vector"
	"github.com/kercre123/vector-go-sdk/pkg/vectorpb"
	"hz.tools/mjpeg"
)

const serverFiles string = "./webroot/sdkapp"

var sdkAddress string = "localhost:443"

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
	clientGUID := string("tni1TRsTRTaNSapjo0Y+Sw==")
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
	case r.URL.Path == "/api-sdk/rainbow_on":
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrl", "rainbowon")
		cmd.Run()
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/rainbow_off":
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrl", "rainbowoff")
		cmd.Run()
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/snore_enable":
		fmt.Fprintf(w, "executing")
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrldd", "snore_enable")
		cmd.Run()
		return
	case r.URL.Path == "/api-sdk/snore_disable":
		fmt.Fprintf(w, "executing")
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrldd", "snore_disable")
		cmd.Run()
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
	case r.URL.Path == "/api-sdk/server_escape":
		fmt.Fprintf(w, "executing")
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrldd", "server_escape")
		cmd.Run()
		return
	case r.URL.Path == "/api-sdk/server_prod":
		fmt.Fprintf(w, "executing")
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrldd", "server_prod")
		cmd.Run()
		return
	case r.URL.Path == "/api-sdk/snowglobe":
		fmt.Fprintf(w, "executing")
		cmd := exec.Command("/bin/bash", "/sbin/vector-ctrldd", "snowglobe")
		cmd.Run()
		return
	case r.URL.Path == "/api-sdk/initSDK":
		serial := r.FormValue("serial")
		if serial == "" {
			fmt.Fprintf(w, "no serial given")
			return
		}
		var err error
		robot, err = vector.NewEP(serial)
		if err != nil {
			fmt.Fprintf(w, "failed: "+err.Error())
			return
		}
		jsonBytes, err := os.ReadFile("./jdocs/botSdkInfo.json")
		if err != nil {
			fmt.Fprintf(w, "failed: "+err.Error())
			return
		}
		type RobotSDKInfoStore struct {
			GlobalGUID string `json:"global_guid"`
			Robots     []struct {
				Esn       string `json:"esn"`
				IPAddress string `json:"ip_address"`
			} `json:"robots"`
		}
		var robotSdkInfo RobotSDKInfoStore
		json.Unmarshal(jsonBytes, &robotSdkInfo)
		matched := false
		for num, robot := range robotSdkInfo.Robots {
			if robot.Esn == serial {
				matched = true
				sdkAddress = robotSdkInfo.Robots[num].IPAddress + ":443"
			}
		}
		_, err = robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			fmt.Fprintf(w, "failed to get battery info")
			return
		}
		if !matched {
			fmt.Fprintf(w, "failed to set bot ip")
			return
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
