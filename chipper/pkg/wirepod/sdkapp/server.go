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
		if len(vars.BotInfo.Robots) == 0 {
			http.Error(w, "no bots are authenticated", http.StatusInternalServerError)
			return
		}
		jsonBytes, err := json.Marshal(vars.BotInfo)
		if err != nil {
			fmt.Fprintf(w, "error marshaling json")
			return
		}
		fmt.Fprint(w, string(jsonBytes))
		return
	case r.URL.Path == "/api-sdk/get_sdk_settings":
		i := 0
		for {
			resp, err := robot.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
				JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_SETTINGS},
			})
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			if strings.Contains(resp.NamedJdocs[0].Doc.JsonDoc, "BStat.ReactedToTriggerWord") {
				time.Sleep(time.Second / 2)
				if i > 3 {
					logger.Println("Bot refuses to return RobotSettings jdoc...")
					logger.Println("Returned Jdoc: ", resp.NamedJdocs[0].Doc.JsonDoc)
					w.Write([]byte("error: bot refuses to return robotsettings"))
					return
				}
				i = i + 1
				continue
			}
			json := resp.NamedJdocs[0].Doc.JsonDoc
			var ajdoc vars.AJdoc
			ajdoc.DocVersion = resp.NamedJdocs[0].Doc.DocVersion
			ajdoc.FmtVersion = resp.NamedJdocs[0].Doc.FmtVersion
			ajdoc.JsonDoc = resp.NamedJdocs[0].Doc.JsonDoc
			vars.AddJdoc("vic:"+robotObj.ESN, "vic.RobotSettings", ajdoc)
			logger.Println("Updating vic.RobotSettings (source: sdkapp)")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte(json))
			return
		}

	case r.URL.Path == "/api-sdk/play_sound":
		file, _, err := r.FormFile("sound")
		if err != nil {
			println("Error retrieving the file:", err)
			return
		}
		defer file.Close()

		// Lê o conteúdo do arquivo em um slice de bytes
		pcmFile, err := io.ReadAll(file)
		if err != nil {
			println("Error reading the file:", err)
			return
		}

		var audioChunks [][]byte
		for len(pcmFile) >= 1024 {
			audioChunks = append(audioChunks, pcmFile[:1024])
			pcmFile = pcmFile[1024:]
		}

		var audioClient vectorpb.ExternalInterface_ExternalAudioStreamPlaybackClient
		audioClient, _ = robot.Conn.ExternalAudioStreamPlayback(ctx)
		audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
				AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
					AudioFrameRate: 8000,
					AudioVolume:    uint32(100),
				},
			},
		})

		for _, chunk := range audioChunks {
			audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: uint32(len(chunk)),
						AudioChunkSamples:   chunk,
					},
				},
			})
			time.Sleep(time.Millisecond * 60)
		}

		audioClient.SendMsg(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})

		return

	case r.URL.Path == "/api-sdk/get_battery":
		// Ensure the endpoint times out after 15 seconds
		ctx := r.Context() // Get the request context
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		resp, err := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, string(jsonBytes))
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
	case r.URL.Path == "/api-sdk/get_robot_stats":
		resp, err := robot.Conn.PullJdocs(ctx,
			&vectorpb.PullJdocsRequest{
				JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_LIFETIME_STATS},
			})
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		w.Write([]byte(resp.GetNamedJdocs()[0].Doc.JsonDoc))
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

func DisableCachingAndSniffing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

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
	// in jdocspinger.go
	http.HandleFunc("/ok:80", connCheck)
	http.HandleFunc("/ok", connCheck)
	InitJdocsPinger()
	// camstream
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

func rgbToBytes(rgbValues [][][3]uint8) ([]byte, error) {
	var buffer bytes.Buffer

	for _, row := range rgbValues {
		for _, pixel := range row {
			// Directly add the R, G and B values ​​to the buffer
			buffer.WriteByte(pixel[0]) // R
			buffer.WriteByte(pixel[1]) // G
			buffer.WriteByte(pixel[2]) // B
		}
	}

	return buffer.Bytes(), nil
}
func imageToBytes(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	var buffer bytes.Buffer

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Obtém a cor do pixel
			c := img.At(x, y)
			r, g, b, _ := c.RGBA() // Ignorando o valor Alpha

			// Converte de uint32 para uint8
			buffer.WriteByte(uint8(r >> 8))
			buffer.WriteByte(uint8(g >> 8))
			buffer.WriteByte(uint8(b >> 8))
		}
	}

	return buffer.Bytes(), nil
}

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
