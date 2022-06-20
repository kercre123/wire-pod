package wirepod

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"strconv"

	opus "github.com/digital-dream-labs/opus-go/opus"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func pcmToWav(pcmFile string, wavFile string) {
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
	audioBuf, err := newAudioIntBuffer(in)
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

func newAudioIntBuffer(r io.Reader) (*audio.IntBuffer, error) {
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

func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func bytesToInt(stream opus.OggStream, data []byte, numBot int, voiceTimer int, die bool) {
	if die == true {
		return
	}
	f, err := os.Create("/tmp/" + strconv.Itoa(numBot) + "voice.pcm")
	if err != nil {
		log.Println(err)
	}
	n, err := stream.Decode(data)
	f.Write(n)
	if voiceTimer == 1 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice1.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice1.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped1")
		}
	}
	if voiceTimer == 2 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice2.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice2.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped2")
		}
	}
	if voiceTimer == 3 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice3.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice3.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped3")
		}
	}
	if voiceTimer == 4 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice4.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice4.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped4")
		}
	}
	if err != nil {
		log.Println(err)
	}
}
