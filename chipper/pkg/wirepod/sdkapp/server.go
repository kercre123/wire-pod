package sdkapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
)

var serverFiles string = "./webroot/sdkapp"

func SdkapiHandler(w http.ResponseWriter, r *http.Request) {
	robotObj, robotIndex, err := getRobot(r.FormValue("serial"))
	robot := robotObj.Vector
	ctx := robotObj.Ctx
	if r.URL.Path != "/api-sdk/get_sdk_info" && r.URL.Path != "/api-sdk/debug" {
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		robots[robotIndex].ConnTimer = 0
	}
	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	case r.URL.Path == "/api-sdk/conn_test":
		// getRobot does connection check and will return error if failed
		fmt.Fprint(w, "success")
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
		robot.Conn.AppIntent(ctx,
			&vectorpb.AppIntentRequest{
				Intent: intent,
			},
		)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/eye_color":
		eye_color := r.FormValue("color")
		setPresetEyeColor(robotObj, eye_color)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/custom_eye_color":
		hue := r.FormValue("hue")
		sat := r.FormValue("sat")
		setCustomEyeColor(robotObj, hue, sat)
		fmt.Fprintf(w, hue+sat)
		return
	case r.URL.Path == "/api-sdk/volume":
		volume := r.FormValue("volume")
		setSettingSDKintbool(robotObj, "master_volume", volume)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/locale":
		locale := r.FormValue("locale")
		setSettingSDKstring(robotObj, "locale", locale)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/location":
		location := r.FormValue("location")
		setSettingSDKstring(robotObj, "default_location", location)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/timezone":
		timezone := r.FormValue("timezone")
		setSettingSDKstring(robotObj, "time_zone", timezone)
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/get_sdk_info":
		jsonBytes, err := json.Marshal(vars.BotInfo)
		if err != nil {
			fmt.Fprintf(w, "error marshaling json")
			return
		}
		fmt.Fprint(w, string(jsonBytes))
		return
	case r.URL.Path == "/api-sdk/get_sdk_settings":
		resp, err := robot.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
			JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_SETTINGS},
		})
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		json := resp.NamedJdocs[0].Doc.JsonDoc
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(json))
		return
	case r.URL.Path == "/api-sdk/time_format_12":
		setSettingSDKintbool(robotObj, "clock_24_hour", "false")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/time_format_24":
		setSettingSDKintbool(robotObj, "clock_24_hour", "true")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/temp_c":
		setSettingSDKintbool(robotObj, "temp_is_fahrenheit", "false")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/temp_f":
		setSettingSDKintbool(robotObj, "temp_is_fahrenheit", "true")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/button_hey_vector":
		setSettingSDKintbool(robotObj, "button_wakeword", "0")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/button_alexa":
		setSettingSDKintbool(robotObj, "button_wakeword", "1")
		fmt.Fprintf(w, "done")
		return
	case r.URL.Path == "/api-sdk/assume_behavior_control":
		fmt.Fprintf(w, "success")
		assumeBehaviorControl(robotObj, robotIndex, r.FormValue("priority"))
		return
	case r.URL.Path == "/api-sdk/release_behavior_control":
		robots[robotIndex].BcAssumption = false
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/say_text":
		if len([]rune(r.FormValue("text"))) >= 600 {
			fmt.Fprint(w, "error: text is too long")
		}
		robot.Conn.SayText(
			ctx,
			&vectorpb.SayTextRequest{
				DurationScalar: 1,
				UseVectorVoice: true,
				Text:           r.FormValue("text"),
			},
		)
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/move_wheels":
		lw, _ := strconv.Atoi(r.FormValue("lw"))
		rw, _ := strconv.Atoi(r.FormValue("rw"))
		robot.Conn.DriveWheels(ctx,
			&vectorpb.DriveWheelsRequest{
				LeftWheelMmps:   float32(lw),
				RightWheelMmps:  float32(rw),
				LeftWheelMmps2:  float32(lw),
				RightWheelMmps2: float32(rw),
			},
		)
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/move_lift":
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		robot.Conn.MoveLift(
			ctx,
			&vectorpb.MoveLiftRequest{
				SpeedRadPerSec: float32(speed),
			},
		)
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/move_head":
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		robot.Conn.MoveHead(
			ctx,
			&vectorpb.MoveHeadRequest{
				SpeedRadPerSec: float32(speed),
			},
		)
		fmt.Fprintf(w, "")
		return
	case r.URL.Path == "/api-sdk/get_faces":
		resp, err := robot.Conn.RequestEnrolledNames(
			ctx,
			&vectorpb.RequestEnrolledNamesRequest{})
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		bytes, _ := json.Marshal(resp.Faces)
		fmt.Fprint(w, string(bytes))
		return
	case r.URL.Path == "/api-sdk/rename_face":
		id := r.FormValue("id")
		oldname := r.FormValue("oldname")
		newname := r.FormValue("newname")
		idInt, _ := strconv.Atoi(id)
		idInt32 := int32(idInt)
		_, err := robot.Conn.UpdateEnrolledFaceByID(
			ctx,
			&vectorpb.UpdateEnrolledFaceByIDRequest{
				FaceId:  idInt32,
				OldName: oldname,
				NewName: newname,
			})
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/delete_face":
		id := r.FormValue("id")
		idInt, _ := strconv.Atoi(id)
		idInt32 := int32(idInt)
		_, err := robot.Conn.EraseEnrolledFaceByID(
			ctx,
			&vectorpb.EraseEnrolledFaceByIDRequest{
				FaceId: idInt32,
			})
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		fmt.Fprintf(w, "success")
		return
	case r.URL.Path == "/api-sdk/add_face":
		name := r.FormValue("name")
		_, err := robot.Conn.AppIntent(
			ctx,
			&vectorpb.AppIntentRequest{
				Intent: "intent_meet_victor",
				Param:  name,
			},
		)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		fmt.Fprintf(w, "success")
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
	case r.URL.Path == "/api-sdk/begin_event_stream":
		// setup websocket
		robots[robotIndex].EventsStreaming = true
		go func() {
			client, err := robot.Conn.EventStream(
				ctx,
				&vectorpb.EventRequest{
					ListType: &vectorpb.EventRequest_WhiteList{
						WhiteList: &vectorpb.FilterList{
							List: []string{"stimulation_info"},
						},
					},
					ConnectionId: "wirepod",
				},
			)
			if err != nil {
				fmt.Fprint(w, err.Error())
			}
			for {
				if robots[robotIndex].EventsStreaming {
					resp, err := client.Recv()
					if err != nil {
						fmt.Fprint(w, err.Error())
						robots[robotIndex].EventsStreaming = false
						return
					}
					stimInfo := resp.Event.GetStimulationInfo()
					stimInfoString := fmt.Sprint(stimInfo)
					if strings.Contains(stimInfoString, "velocity") {
						// velocity in the string means there is a value
						robots[robotIndex].StimState = stimInfo.Value
					}
				} else {
					return
				}
			}
		}()
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-sdk/stop_event_stream":
		robots[robotIndex].EventsStreaming = false
		robots[robotIndex].StimState = 0
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-sdk/get_stim_status":
		if robots[robotIndex].EventsStreaming {
			fmt.Fprint(w, robots[robotIndex].StimState)
			return
		}
		fmt.Fprint(w, "error: must start event stream")
		return
	case r.URL.Path == "/api-sdk/begin_cam_stream":
		//robots[robotIndex].CamStreaming = true
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-sdk/stop_cam_stream":
		robots[robotIndex].CamStreaming = false
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-sdk/get_image_ids":
		var photoIds []uint32
		resp, _ := robot.Conn.PhotosInfo(
			ctx,
			&vectorpb.PhotosInfoRequest{},
		)
		for _, photo := range resp.PhotoInfos {
			photoIds = append(photoIds, photo.PhotoId)
		}
		writeBytes, _ := json.Marshal(photoIds)
		w.Write(writeBytes)
		return
	case r.URL.Path == "/api-sdk/get_image":
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		resp, err := robot.Conn.Photo(
			ctx,
			&vectorpb.PhotoRequest{
				PhotoId: uint32(id),
			},
		)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		w.Write(resp.Image)
		return
	case r.URL.Path == "/api-sdk/get_image_thumb":
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		resp, err := robot.Conn.Thumbnail(
			ctx,
			&vectorpb.ThumbnailRequest{
				PhotoId: uint32(id),
			},
		)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		w.Write(resp.Image)
		return
	case r.URL.Path == "/api-sdk/delete_image":
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		_, err = robot.Conn.DeletePhoto(
			ctx,
			&vectorpb.DeletePhotoRequest{
				PhotoId: uint32(id),
			},
		)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-sdk/print_robot_info":
		fmt.Fprint(w, robot)
		return
	case r.URL.Path == "/api-sdk/disconnect":
		removeRobot(robotObj.ESN, "server")
		fmt.Fprint(w, "done")
		return
	}
}

func camStreamHandler(w http.ResponseWriter, r *http.Request) {
	robotObj, robotIndex, err := getRobot(r.FormValue("serial"))
	if err != nil {
		fmt.Fprint(w, "error: "+err.Error())
		return
	}
	if robots[robotIndex].CamStreaming {
		robots[robotIndex].CamStreaming = false
		time.Sleep(time.Second / 2)
	}
	robotObj.Vector.Conn.EnableImageStreaming(
		robotObj.Ctx,
		&vectorpb.EnableImageStreamingRequest{
			Enable: true,
		},
	)
	var client vectorpb.ExternalInterface_CameraFeedClient
	client, err = robotObj.Vector.Conn.CameraFeed(
		robotObj.Ctx,
		&vectorpb.CameraFeedRequest{},
	)
	if err != nil {
		fmt.Fprint(w, "error: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--boundary")
	multi := io.MultiWriter(w)
	robots[robotIndex].CamStreaming = true
	for {
		select {
		case <-r.Context().Done():
			robotObj.Vector.Conn.EnableImageStreaming(
				robotObj.Ctx,
				&vectorpb.EnableImageStreamingRequest{
					Enable: false,
				},
			)
			robots[robotIndex].CamStreaming = false
			return
		default:
			if robots[robotIndex].CamStreaming {
				response, err := client.Recv()
				if err == nil {
					imageBytes := response.GetData()
					img, _, _ := image.Decode(bytes.NewReader(imageBytes))
					fmt.Fprintf(multi, "--boundary\r\nContent-Type: image/jpeg\r\n\r\n")
					jpeg.Encode(multi, img, &jpeg.Options{
						Quality: 50,
					})
				}
			} else {
				robotObj.Vector.Conn.EnableImageStreaming(
					robotObj.Ctx,
					&vectorpb.EnableImageStreamingRequest{
						Enable: false,
					},
				)
				return
			}
		}
	}
}

func BeginServer() {
	if os.Getenv("JDOCS_PINGER_ENABLED") == "false" {
		PingerEnabled = false
		logger.Println("Jdocs pinger has been disabled")
	}
	http.HandleFunc("/api-sdk/", SdkapiHandler)
	if runtime.GOOS == "android" {
		serverFiles = filepath.Join(vars.AndroidPath, "/static/webroot")
	}
	fileServer := http.FileServer(http.Dir(serverFiles))
	http.Handle("/sdk-app", fileServer)
	// in jdocspinger.go
	http.HandleFunc("/ok:80", connCheck)
	http.HandleFunc("/ok", connCheck)
	InitJdocsPinger()
	// camstream
	http.HandleFunc("/cam-stream", camStreamHandler)
	logger.Println("Starting SDK app")
	fmt.Printf("Starting server at port 80 for connCheck\n")
	ipAddr := botsetup.GetOutboundIP().String()
	logger.Println("\033[1;36mConfiguration page: http://" + ipAddr + ":" + vars.WebPort + "\033[0m")
	if runtime.GOOS != "android" {
		if err := http.ListenAndServe(":80", nil); err != nil {
			if vars.Packaged {
				logger.WarnMsg("A process is using port 80. Wire-pod will keep running, but connCheck functionality will not work, so your bot may not always stay connected to your wire-pod instance.")
			}
			logger.Println("A process is already using port 80 - connCheck functionality will not work")
		}
	}
}
