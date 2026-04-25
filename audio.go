package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"log"
	"math"
	"sync"

	"github.com/gordonklaus/portaudio"
)

var (
	audioBuffer    []int16
	audioMutex     sync.Mutex
	audioAvailable sync.Cond
)

func init() {
	audioAvailable.L = &audioMutex
}

func initAudio() error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}
	return nil
}

func terminateAudio() {
	portaudio.Terminate()
}

func startOutputStream(ui *AudioUI) (*portaudio.Stream, error) {
	outStream, err := portaudio.OpenDefaultStream(0, 1, 24000, 2048, func(out []int16) {
		audioMutex.Lock()
		n := copy(out, audioBuffer)
		bufferSize := len(audioBuffer)
		audioBuffer = audioBuffer[n:]
		audioMutex.Unlock()
		for i := n; i < len(out); i++ {
			out[i] = 0
		}
		if bufferSize > 0 && ui != nil {
			rms := calculateRMS(out[:n])
			ui.UpdateOutput(rms, n, bufferSize-n)
		}
	})
	if err != nil {
		return nil, err
	}
	if err := outStream.Start(); err != nil {
		outStream.Close()
		return nil, err
	}
	log.Println("[Audio Output] Stream started (24kHz, 2048 frames)")
	return outStream, nil
}

func startInputStream(sendAudio func([]int16)) (*portaudio.Stream, error) {
	inStream, err := portaudio.OpenDefaultStream(1, 0, 48000, 480, func(in []int16) {
		bufCopy := make([]int16, len(in))
		copy(bufCopy, in)
		sendAudio(bufCopy)
	})
	if err != nil {
		return nil, err
	}
	if err := inStream.Start(); err != nil {
		inStream.Close()
		return nil, err
	}
	log.Println("[Audio Input] Stream started (48kHz, 10ms chunks)")
	return inStream, nil
}

func encodeAudio(data []int16) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, data)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func decodeAudio(dataStr string) []int16 {
	decoded, _ := base64.StdEncoding.DecodeString(dataStr)
	newAudio := make([]int16, len(decoded)/2)
	buf := bytes.NewReader(decoded)
	for i := range newAudio {
		binary.Read(buf, binary.LittleEndian, &newAudio[i])
	}
	for i := range newAudio {
		newAudio[i] = int16(float64(newAudio[i]) * 0.5)
	}
	return newAudio
}

func addToAudioBuffer(samples []int16) {
	audioMutex.Lock()
	audioBuffer = append(audioBuffer, samples...)
	audioMutex.Unlock()
}

func clearAudioBuffer() {
	audioMutex.Lock()
	audioBuffer = nil
	audioMutex.Unlock()
}

func calculateRMS(samples []int16) float64 {
	sum := 0.0
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}
	return math.Sqrt(sum / float64(len(samples)))
}
