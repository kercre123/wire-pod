package noop

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	opus "github.com/digital-dream-labs/opus-go/opus"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

var debugLoggingKG bool

var NoResult string = "NoResultCommand"
var NoResultSpoken string

var botNumKG int = 0
var matchedKG int = 0

func pcmToWavKG(pcmFile string, wavFile string) {
	in, err := os.Open(pcmFile)
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(wavFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	e := wav.NewEncoder(out, 16000, 16, 1, 1)
	audioBuf, err := newAudioIntBufferKG(in)
	if err != nil {
		log.Fatal(err)
	}
	if err := e.Write(audioBuf); err != nil {
		log.Fatal(err)
	}
	if err := e.Close(); err != nil {
		log.Fatal(err)
	}
}

func newAudioIntBufferKG(r io.Reader) (*audio.IntBuffer, error) {
	buf := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  16000,
		},
	}
	for {
		var sample int16
		err := binary.Read(r, binary.LittleEndian, &sample)
		switch {
		case err == io.EOF:
			return &buf, nil
		case err != nil:
			return nil, err
		}
		buf.Data = append(buf.Data, int(sample))
	}
}

func dumpTimerKG(voiceTimer int) {
	time.Sleep(time.Millisecond * 300)
	for voiceTimer < 7 {
		voiceTimer = voiceTimer + 1
		//log.Println("voiceKG Timer")
		//log.Println(voiceTimer)
		time.Sleep(time.Second * 1)
	}
	return
}

func bytesToSamplesKG(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func bytesToIntKG(stream opus.OggStream, data []byte, numBot int, voiceTimer int, die bool) {
	if die == true {
		return
	}
	f, err := os.Create("/tmp/" + strconv.Itoa(numBot) + "voiceKG.pcm")
	if err != nil {
		log.Println(err)
	}
	n, err := stream.Decode(data)
	f.Write(n)
	if voiceTimer == 1 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voiceKG1.wav"); err == nil {
			//
		} else {
			pcmToWavKG("/tmp/"+strconv.Itoa(numBot)+"voiceKG.pcm", "/tmp/"+strconv.Itoa(numBot)+"voiceKG1.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumpedKG1")
		}
	}
	if voiceTimer == 2 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voiceKG2.wav"); err == nil {
			//
		} else {
			pcmToWavKG("/tmp/"+strconv.Itoa(numBot)+"voiceKG.pcm", "/tmp/"+strconv.Itoa(numBot)+"voiceKG2.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumpedKG2")
		}
	}
	if voiceTimer == 3 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voiceKG3.wav"); err == nil {
			//
		} else {
			pcmToWavKG("/tmp/"+strconv.Itoa(numBot)+"voiceKG.pcm", "/tmp/"+strconv.Itoa(numBot)+"voiceKG3.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumpedKG3")
		}
	}
	if voiceTimer == 4 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voiceKG4.wav"); err == nil {
			//
		} else {
			pcmToWavKG("/tmp/"+strconv.Itoa(numBot)+"voiceKG.pcm", "/tmp/"+strconv.Itoa(numBot)+"voiceKG4.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumpedKG4")
		}
	}
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	var voiceTimer int = 0
	var transcription1 string = ""
	var transcription2 string = ""
	var transcription3 string = ""
	var transcription4 string = ""
	var successMatch bool = false
	var transcribedText string
	var die bool = false
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		log.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLoggingKG = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLoggingKG = true
		} else {
			debugLoggingKG = false
		}
	}
	matchedKG = 0
	botNumKG = botNumKG + 1
	if debugLoggingKG == true {
		log.Println("Stream " + strconv.Itoa(botNumKG) + " opened.")
	}
	data := []byte{}
	data = append(data, req.FirstReq.InputAudio...)
	stream := opus.OggStream{}
	go func() {
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			voiceTimer = voiceTimer + 1
			time.Sleep(time.Second * 1)
		}
	}()
	go func() {
		var processOne bool = false
		var processTwo bool = false
		var processThree bool = false
		var processFour bool = false
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			if processOne == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "dumpedKG1"); err == nil {
					processOne = true
					process1 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG1.wav")
					process1out, err := process1.Output()
					if err != nil {
						//
					}
					transcription1 = strings.TrimSpace(string(process1out))
					log.Println("1: " + transcription1)
				}
			}
			if processTwo == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "dumpedKG2"); err == nil {
					processTwo = true
					process2 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG2.wav")
					process2out, err := process2.Output()
					if err != nil {
						//
					}
					transcription2 = strings.TrimSpace(string(process2out))
					log.Println("2: " + transcription2)
				}
			}
			if processThree == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "dumpedKG3"); err == nil {
					processThree = true
					process3 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG3.wav")
					process3out, err := process3.Output()
					if err != nil {
						//
					}
					transcription3 = strings.TrimSpace(string(process3out))
					log.Println("3: " + transcription3)
				}
			}
			if processFour == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "dumpedKG4"); err == nil {
					processFour = true
					process4 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG4.wav")
					process4out, err := process4.Output()
					if err != nil {
						//
					}
					transcription4 = strings.TrimSpace(string(process4out))
					log.Println("4: " + transcription4)
					successMatch = true
				}
			}
		}
	}()
	for {
		chunk, err := req.Stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if transcription2 != "" {
			if transcription1 == transcription2 {
				log.Println("Speech stopped, 2: " + transcription1)
				transcribedText = transcription1
				die = true
				break
			} else if transcription2 != "" {
				if transcription2 == transcription3 {
					log.Println("Speech stopped, 3: " + transcription2)
					transcribedText = transcription2
					die = true
					break
				} else if transcription3 != "" {
					if transcription3 == transcription4 {
						log.Println("Speech stopped, 4: " + transcription3)
						transcribedText = transcription3
						die = true
						break
					} else if transcription4 != "" {
						if transcription3 == transcription4 {
							log.Println("Speech stopped, 4: " + transcription4)
							transcribedText = transcription4
							die = true
							break
						} else {
							log.Println("Speech stopped, 4 (nm): " + transcription4)
							transcribedText = transcription4
							die = true
							break
						}
					}
				}
			}
		}
		if transcription4 == "" && successMatch == true {
			transcribedText = ""
			break
		}
		data = append(data, chunk.InputAudio...)
		go bytesToIntKG(stream, data, botNumKG, voiceTimer, die)
	}
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG4").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped4").Run()
	NoResultSpoken = "This is a placeholder! You said: " + transcribedText
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  NoResultSpoken,
	}
	botNumKG = botNumKG - 1
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
