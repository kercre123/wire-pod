package sdkapp

import (
	"bytes"
	"context"
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
	"github.com/kercre123/wire-pod/chipper/pkg/scripting"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

var serverFiles string = "./webroot/sdkapp"

// SdkapiHandler handles incoming SDK-related routes.
func SdkapiHandler(w http.ResponseWriter, r *http.Request) {
	robotObj, robotIndex, err := getRobot(r.FormValue("serial"))
	if err != nil && r.URL.Path != "/api-sdk/get_sdk_info" && r.URL.Path != "/api-sdk/debug" {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Only proceed if we have a valid robot object (except for get_sdk_info/debug).
	var robot *vectorpb.ExternalInterface = nil
	var ctx context.Context
	if robotObj != nil {
		robot = robotObj.Vector
		ctx = robotObj.Ctx
	}

	// For all routes except get_sdk_info/debug, reset connection timer if no error
	if r.URL.Path != "/api-sdk/get_sdk_info" && r.URL.Path != "/api-sdk/debug" && err == nil {
		robots[robotIndex].ConnTimer = 0
	}

	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return

	case r.URL.Path == "/api-sdk/conn_test":
		// getRobot does connection check
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success")
		return

	case r.URL.Path == "/api-sdk/alexa_sign_in":
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		_, callErr := robot.Conn.AlexaOptIn(ctx, &vectorpb.AlexaOptInRequest{OptIn: true})
		if callErr != nil {
			http.Error(w, "error: "+callErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/alexa_sign_out":
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		_, callErr := robot.Conn.AlexaOptIn(ctx, &vectorpb.AlexaOptInRequest{OptIn: false})
		if callErr != nil {
			http.Error(w, "error: "+callErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/cloud_intent":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		intent := r.FormValue("intent")
		_, callErr := robot.Conn.AppIntent(ctx, &vectorpb.AppIntentRequest{Intent: intent})
		if callErr != nil {
			http.Error(w, "error: "+callErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/eye_color":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		eye_color := r.FormValue("color")
		setPresetEyeColor(robotObj, eye_color)
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/custom_eye_color":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		hue := r.FormValue("hue")
		sat := r.FormValue("sat")
		setCustomEyeColor(robotObj, hue, sat)
		fmt.Fprintf(w, hue+sat)
		return

	case r.URL.Path == "/api-sdk/volume":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		volume := r.FormValue("volume")
		setSettingSDKintbool(robotObj, "master_volume", volume)
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/locale":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		locale := r.FormValue("locale")
		setSettingSDKstring(robotObj, "locale", locale)
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/location":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		location := r.FormValue("location")
		setSettingSDKstring(robotObj, "default_location", location)
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/timezone":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		timezone := r.FormValue("timezone")
		setSettingSDKstring(robotObj, "time_zone", timezone)
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/get_sdk_info":
		if len(vars.BotInfo.Robots) == 0 {
			http.Error(w, "no bots are authenticated", http.StatusInternalServerError)
			return
		}
		jsonBytes, mErr := json.Marshal(vars.BotInfo)
		if mErr != nil {
			fmt.Fprintf(w, "error marshaling json: %v", mErr)
			return
		}
		fmt.Fprint(w, string(jsonBytes))
		return

	case r.URL.Path == "/api-sdk/get_sdk_settings":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		i := 0
		for {
			resp, pullErr := robot.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
				JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_SETTINGS},
			})
			if pullErr != nil {
				http.Error(w, "error: "+pullErr.Error(), http.StatusInternalServerError)
				return
			}
			if len(resp.NamedJdocs) == 0 {
				http.Error(w, "error: no Jdocs returned", http.StatusInternalServerError)
				return
			}
			if strings.Contains(resp.NamedJdocs[0].Doc.JsonDoc, "BStat.ReactedToTriggerWord") {
				time.Sleep(time.Second / 2)
				if i > 3 {
					logger.Println("Bot refuses to return RobotSettings jdoc...")
					logger.Println("Returned Jdoc: ", resp.NamedJdocs[0].Doc.JsonDoc)
					http.Error(w, "error: bot refuses to return robotsettings", http.StatusInternalServerError)
					return
				}
				i++
				continue
			}
			jsonDoc := resp.NamedJdocs[0].Doc.JsonDoc
			var ajdoc vars.AJdoc
			ajdoc.DocVersion = resp.NamedJdocs[0].Doc.DocVersion
			ajdoc.FmtVersion = resp.NamedJdocs[0].Doc.FmtVersion
			ajdoc.JsonDoc = resp.NamedJdocs[0].Doc.JsonDoc
			vars.AddJdoc("vic:"+robotObj.ESN, "vic.RobotSettings", ajdoc)
			logger.Println("Updating vic.RobotSettings (source: sdkapp)")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			_, writeErr := w.Write([]byte(jsonDoc))
			if writeErr != nil {
				logger.Println("Write error:", writeErr)
			}
			return
		}

	case r.URL.Path == "/api-sdk/play_sound":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		file, _, fErr := r.FormFile("sound")
		if fErr != nil {
			logger.Println("Error retrieving file:", fErr)
			http.Error(w, "error: "+fErr.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		pcmFile, readErr := io.ReadAll(file)
		if readErr != nil {
			logger.Println("Error reading file:", readErr)
			http.Error(w, "error: "+readErr.Error(), http.StatusInternalServerError)
			return
		}

		var audioChunks [][]byte
		for len(pcmFile) >= 1024 {
			audioChunks = append(audioChunks, pcmFile[:1024])
			pcmFile = pcmFile[1024:]
		}
		if len(pcmFile) > 0 {
			audioChunks = append(audioChunks, pcmFile)
		}

		audioClient, aErr := robot.Conn.ExternalAudioStreamPlayback(ctx)
		if aErr != nil {
			logger.Println("Audio stream init error:", aErr)
			http.Error(w, "error: "+aErr.Error(), http.StatusInternalServerError)
			return
		}
		// Prepare
		errSend := audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
				AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
					AudioFrameRate: 8000,
					AudioVolume:    uint32(100),
				},
			},
		})
		if errSend != nil {
			logger.Println("Error sending AudioStreamPrepare:", errSend)
			http.Error(w, "error: "+errSend.Error(), http.StatusInternalServerError)
			return
		}
		// Send chunks
		for _, chunk := range audioChunks {
			errChunk := audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: uint32(len(chunk)),
						AudioChunkSamples:   chunk,
					},
				},
			})
			if errChunk != nil {
				logger.Println("Error sending AudioStreamChunk:", errChunk)
				http.Error(w, "error: "+errChunk.Error(), http.StatusInternalServerError)
				return
			}
			time.Sleep(time.Millisecond * 60)
		}
		// Complete
		errComplete := audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})
		if errComplete != nil {
			logger.Println("Error sending AudioStreamComplete:", errComplete)
			http.Error(w, "error: "+errComplete.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "success")
		return

	case r.URL.Path == "/api-sdk/get_battery":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		resp, bErr := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if bErr != nil {
			fmt.Fprint(w, "error: "+bErr.Error())
			return
		}
		jsonBytes, mErr := json.Marshal(resp)
		if mErr != nil {
			fmt.Fprint(w, "error: "+mErr.Error())
			return
		}
		fmt.Fprint(w, string(jsonBytes))
		return

	case r.URL.Path == "/api-sdk/time_format_12":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "clock_24_hour", "false")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/time_format_24":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "clock_24_hour", "true")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/temp_c":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "temp_is_fahrenheit", "false")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/temp_f":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "temp_is_fahrenheit", "true")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/button_hey_vector":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "button_wakeword", "0")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/button_alexa":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		setSettingSDKintbool(robotObj, "button_wakeword", "1")
		fmt.Fprintf(w, "done")
		return

	case r.URL.Path == "/api-sdk/assume_behavior_control":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")
		assumeBehaviorControl(robotObj, robotIndex, r.FormValue("priority"))
		return

	case r.URL.Path == "/api-sdk/release_behavior_control":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		robots[robotIndex].BcAssumption = false
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/say_text":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len([]rune(r.FormValue("text"))) >= 600 {
			fmt.Fprint(w, "error: text is too long")
			return
		}
		_, sayErr := robot.Conn.SayText(ctx, &vectorpb.SayTextRequest{
			DurationScalar: 1,
			UseVectorVoice: true,
			Text:           r.FormValue("text"),
		})
		if sayErr != nil {
			http.Error(w, "error: "+sayErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/move_wheels":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		lw, _ := strconv.Atoi(r.FormValue("lw"))
		rw, _ := strconv.Atoi(r.FormValue("rw"))
		_, wheelsErr := robot.Conn.DriveWheels(ctx, &vectorpb.DriveWheelsRequest{
			LeftWheelMmps:   float32(lw),
			RightWheelMmps:  float32(rw),
			LeftWheelMmps2:  float32(lw),
			RightWheelMmps2: float32(rw),
		})
		if wheelsErr != nil {
			http.Error(w, "error: "+wheelsErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "")
		return

	case r.URL.Path == "/api-sdk/move_lift":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		_, liftErr := robot.Conn.MoveLift(ctx, &vectorpb.MoveLiftRequest{
			SpeedRadPerSec: float32(speed),
		})
		if liftErr != nil {
			http.Error(w, "error: "+liftErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "")
		return

	case r.URL.Path == "/api-sdk/move_head":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		speed, _ := strconv.Atoi(r.FormValue("speed"))
		_, headErr := robot.Conn.MoveHead(ctx, &vectorpb.MoveHeadRequest{
			SpeedRadPerSec: float32(speed),
		})
		if headErr != nil {
			http.Error(w, "error: "+headErr.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "")
		return

	case r.URL.Path == "/api-sdk/get_faces":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp, faceErr := robot.Conn.RequestEnrolledNames(ctx, &vectorpb.RequestEnrolledNamesRequest{})
		if faceErr != nil {
			http.Error(w, "error: "+faceErr.Error(), http.StatusInternalServerError)
			return
		}
		bytes, _ := json.Marshal(resp.Faces)
		fmt.Fprint(w, string(bytes))
		return

	case r.URL.Path == "/api-sdk/rename_face":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id := r.FormValue("id")
		oldname := r.FormValue("oldname")
		newname := r.FormValue("newname")
		idInt, _ := strconv.Atoi(id)
		idInt32 := int32(idInt)
		_, renameErr := robot.Conn.UpdateEnrolledFaceByID(ctx, &vectorpb.UpdateEnrolledFaceByIDRequest{
			FaceId:  idInt32,
			OldName: oldname,
			NewName: newname,
		})
		if renameErr != nil {
			fmt.Fprint(w, "error: "+renameErr.Error())
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/delete_face":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id := r.FormValue("id")
		idInt, _ := strconv.Atoi(id)
		idInt32 := int32(idInt)
		_, delErr := robot.Conn.EraseEnrolledFaceByID(ctx, &vectorpb.EraseEnrolledFaceByIDRequest{
			FaceId: idInt32,
		})
		if delErr != nil {
			fmt.Fprint(w, "error: "+delErr.Error())
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/add_face":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		name := r.FormValue("name")
		_, addErr := robot.Conn.AppIntent(ctx, &vectorpb.AppIntentRequest{
			Intent: "intent_meet_victor",
			Param:  name,
		})
		if addErr != nil {
			fmt.Fprint(w, "error: "+addErr.Error())
			return
		}
		fmt.Fprintf(w, "success")
		return

	case r.URL.Path == "/api-sdk/mirror_mode":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		enable := r.FormValue("enable")
		var boolVal bool
		if enable == "true" {
			boolVal = true
		} else {
			boolVal = false
		}
		_, mirrorErr := robot.Conn.EnableMirrorMode(ctx, &vectorpb.EnableMirrorModeRequest{
			Enable: boolVal,
		})
		if mirrorErr != nil {
			fmt.Fprint(w, "error: "+mirrorErr.Error())
			return
		}
		fmt.Fprint(w, "success")
		return

	case r.URL.Path == "/api-sdk/begin_event_stream":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		robots[robotIndex].EventsStreaming = true
		go func() {
			client, eErr := robot.Conn.EventStream(ctx, &vectorpb.EventRequest{
				ListType: &vectorpb.EventRequest_WhiteList{
					WhiteList: &vectorpb.FilterList{List: []string{"stimulation_info"}},
				},
				ConnectionId: "wirepod",
			})
			if eErr != nil {
				logger.Println("EventStream error:", eErr)
				robots[robotIndex].EventsStreaming = false
				return
			}
			for {
				if !robots[robotIndex].EventsStreaming {
					return
				}
				resp, recvErr := client.Recv()
				if recvErr != nil {
					logger.Println("EventStream recv error:", recvErr)
					robots[robotIndex].EventsStreaming = false
					return
				}
				stimInfo := resp.Event.GetStimulationInfo()
				stimInfoString := fmt.Sprint(stimInfo)
				if strings.Contains(stimInfoString, "velocity") {
					robots[robotIndex].StimState = stimInfo.Value
				}
			}
		}()
		fmt.Fprint(w, "done")
		return

	case r.URL.Path == "/api-sdk/stop_event_stream":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		robots[robotIndex].EventsStreaming = false
		robots[robotIndex].StimState = 0
		fmt.Fprint(w, "done")
		return

	case r.URL.Path == "/api-sdk/get_stim_status":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if robots[robotIndex].EventsStreaming {
			fmt.Fprint(w, robots[robotIndex].StimState)
			return
		}
		fmt.Fprint(w, "error: must start event stream")
		return

	case r.URL.Path == "/api-sdk/begin_cam_stream":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// robots[robotIndex].CamStreaming = true (handled in camStreamHandler)
		fmt.Fprint(w, "done")
		return

	case r.URL.Path == "/api-sdk/stop_cam_stream":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		robots[robotIndex].CamStreaming = false
		fmt.Fprint(w, "done")
		return

	case r.URL.Path == "/api-sdk/get_image_ids":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var photoIds []uint32
		resp, photoErr := robot.Conn.PhotosInfo(ctx, &vectorpb.PhotosInfoRequest{})
		if photoErr != nil {
			http.Error(w, "error: "+photoErr.Error(), http.StatusInternalServerError)
			return
		}
		for _, photo := range resp.PhotoInfos {
			photoIds = append(photoIds, photo.PhotoId)
		}
		writeBytes, _ := json.Marshal(photoIds)
		w.Write(writeBytes)
		return

	case r.URL.Path == "/api-sdk/get_image":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id, convErr := strconv.Atoi(r.FormValue("id"))
		if convErr != nil {
			fmt.Fprint(w, "error: "+convErr.Error())
			return
		}
		resp, imgErr := robot.Conn.Photo(ctx, &vectorpb.PhotoRequest{PhotoId: uint32(id)})
		if imgErr != nil {
			fmt.Fprint(w, "error: "+imgErr.Error())
			return
		}
		w.Write(resp.Image)
		return

	case r.URL.Path == "/api-sdk/get_image_thumb":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id, convErr := strconv.Atoi(r.FormValue("id"))
		if convErr != nil {
			fmt.Fprint(w, "error: "+convErr.Error())
			return
		}
		resp, thumbErr := robot.Conn.Thumbnail(ctx, &vectorpb.ThumbnailRequest{PhotoId: uint32(id)})
		if thumbErr != nil {
			fmt.Fprint(w, "error: "+thumbErr.Error())
			return
		}
		w.Write(resp.Image)
		return

	case r.URL.Path == "/api-sdk/delete_image":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id, convErr := strconv.Atoi(r.FormValue("id"))
		if convErr != nil {
			fmt.Fprint(w, "error: "+convErr.Error())
			return
		}
		_, delErr := robot.Conn.DeletePhoto(ctx, &vectorpb.DeletePhotoRequest{PhotoId: uint32(id)})
		if delErr != nil {
			fmt.Fprint(w, "error: "+delErr.Error())
			return
		}
		fmt.Fprint(w, "done")
		return

	case r.URL.Path == "/api-sdk/get_robot_stats":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp, statsErr := robot.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
			JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_LIFETIME_STATS},
		})
		if statsErr != nil {
			fmt.Fprint(w, "error: "+statsErr.Error())
			return
		}
		if len(resp.GetNamedJdocs()) == 0 {
			fmt.Fprint(w, "error: no robot stats Jdocs found")
			return
		}
		w.Write([]byte(resp.GetNamedJdocs()[0].Doc.JsonDoc))
		return

	case r.URL.Path == "/api-sdk/print_robot_info":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, robot)
		return

	case r.URL.Path == "/api-sdk/disconnect":
		if err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		removeRobot(robotObj.ESN, "server")
		fmt.Fprint(w, "done")
		return
	}
}

// camStreamHandler starts a live camera stream.
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
	_, eErr := robotObj.Vector.Conn.EnableImageStreaming(robotObj.Ctx, &vectorpb.EnableImageStreamingRequest{
		Enable: true,
	})
	if eErr != nil {
		http.Error(w, "error enabling image streaming: "+eErr.Error(), http.StatusInternalServerError)
		return
	}
	client, cErr := robotObj.Vector.Conn.CameraFeed(robotObj.Ctx, &vectorpb.CameraFeedRequest{})
	if cErr != nil {
		http.Error(w, "error starting camera feed: "+cErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--boundary")
	multi := io.MultiWriter(w)
	robots[robotIndex].CamStreaming = true

	for {
		select {
		case <-r.Context().Done():
			_, dErr := robotObj.Vector.Conn.EnableImageStreaming(robotObj.Ctx, &vectorpb.EnableImageStreamingRequest{
				Enable: false,
			})
			if dErr != nil {
				logger.Println("Error disabling image streaming:", dErr)
			}
			robots[robotIndex].CamStreaming = false
			return
		default:
			if robots[robotIndex].CamStreaming {
				response, recvErr := client.Recv()
				if recvErr != nil {
					logger.Println("CameraFeed recv error:", recvErr)
					robots[robotIndex].CamStreaming = false
					return
				}
				imageBytes := response.GetData()
				img, _, decodeErr := image.Decode(bytes.NewReader(imageBytes))
				if decodeErr != nil {
					logger.Println("Image decode error:", decodeErr)
					continue
				}
				fmt.Fprintf(multi, "--boundary\r\nContent-Type: image/jpeg\r\n\r\n")
				jErr := jpeg.Encode(multi, img, &jpeg.Options{Quality: 50})
				if jErr != nil {
					logger.Println("JPEG encode error:", jErr)
				}
			} else {
				_, disErr := robotObj.Vector.Conn.EnableImageStreaming(robotObj.Ctx, &vectorpb.EnableImageStreamingRequest{
					Enable: false,
				})
				if disErr != nil {
					logger.Println("Error disabling image streaming:", disErr)
				}
				return
			}
		}
	}
}

// DisableCachingAndSniffing sets common security headers on responses.
func DisableCachingAndSniffing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// BeginServer starts the HTTP server for the SDK app.
func BeginServer() {
	scripting.RegisterScriptingAPI()
	if os.Getenv("JDOCS_PINGER_ENABLED") == "false" {
		PingerEnabled = false
		logger.Println("Jdocs pinger has been disabled")
	}
	http.HandleFunc("/api-sdk/", SdkapiHandler)
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		serverFiles = filepath.Join(vars.AndroidPath, "/static/webroot")
	}
	fileServer := http.FileServer(http.Dir(serverFiles))
	http.Handle("/sdk-app", DisableCachingAndSniffing(fileServer))
	http.HandleFunc("/ok:80", connCheck)
	http.HandleFunc("/ok", connCheck)
	InitJdocsPinger()
	http.HandleFunc("/cam-stream", camStreamHandler)
	logger.Println("Starting SDK app")
	fmt.Printf("Starting server at port 80 for connCheck\n")
	ipAddr := vars.GetOutboundIP().String()
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

// rgbToBytes converts a 2D array of RGB pixels into a []byte.
func rgbToBytes(rgbValues [][][3]uint8) ([]byte, error) {
	var buffer bytes.Buffer
	for _, row := range rgbValues {
		for _, pixel := range row {
			buffer.WriteByte(pixel[0]) // R
			buffer.WriteByte(pixel[1]) // G
			buffer.WriteByte(pixel[2]) // B
		}
	}
	return buffer.Bytes(), nil
}

// imageToBytes converts an image.Image to RGB bytes.
func imageToBytes(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	var buffer bytes.Buffer
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			buffer.WriteByte(uint8(r >> 8))
			buffer.WriteByte(uint8(g >> 8))
			buffer.WriteByte(uint8(b >> 8))
		}
	}
	return buffer.Bytes(), nil
}

// resizeImage performs a simple nearest-neighbor resize on an image.
func resizeImage(original image.Image, width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return original
	}
	newImage := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(original.Bounds().Dx()) / float64(width)
	scaleY := float64(original.Bounds().Dy()) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			newImage.Set(x, y, original.At(srcX, srcY))
		}
	}
	return newImage
}
