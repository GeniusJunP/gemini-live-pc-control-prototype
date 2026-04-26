package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	UI *AudioUI
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY 環境変数を設定してください")
	}

	debugMode := os.Getenv("DEBUG_MODE") == "true"

	if err := initAudio(); err != nil {
		log.Fatalf("PortAudio init error: %v", err)
	}
	defer terminateAudio()

	conn, err := connectWebSocket(apiKey)
	if err != nil {
		log.Fatalf("WebSocket connection error: %v", err)
	}
	defer conn.Close()
	// Note: route this message into the TUI log after UI is started.

	// Narrow the visualizer width so small RMS changes are more visible.
	visualizer := NewAudioVisualizer(10, 5)
	// Adjust sensitivity to match observed RMS (~3.5k). Tweak as needed.
	visualizer.SetMaxRMS(2000.0)
	UI = NewAudioUI()
	UI.Start()

	// Route standard logger output into the TUI to avoid terminal mixing.
	log.SetOutput(&uiLogWriter{ui: UI})
	UI.AddLogMessage("[Info] Gemini Live API に接続しました")
	defer UI.Stop()

	if debugMode {
		UI.AddLogMessage("[Mode] Debug mode: Manual ActivityStart/ActivityEnd")
	} else {
		UI.AddLogMessage("[Mode] Normal mode: Automatic VAD")
	}

	if err := sendSetupMessage(conn, debugMode); err != nil {
		log.Fatalf("Setup send error: %v", err)
	}

	processor, err := NewAudioProcessor()
	if err != nil {
		log.Fatalf("Audio processor error: %v", err)
	}
	defer processor.Close()

	outStream, err := startOutputStream(UI)
	if err != nil {
		log.Fatalf("Audio out stream error: %v", err)
	}
	defer outStream.Close()

	inStream, err := startInputStream(func(data []int16) {
		go func(d []int16) {
			rmsBefore := visualizer.CalculateRMS(d)
			processed := processor.Process(d)
			rmsAfter := visualizer.CalculateRMS(processed)

			UI.UpdateInputBefore(rmsBefore, visualizer.Visualize(d, "Before"))
			UI.UpdateInputAfter(rmsAfter, visualizer.Visualize(processed, "After"))

			if err := sendAudioMessage(conn, processed); err != nil {
				UI.AddLogMessage(fmt.Sprintf("[Audio Input] WebSocket write error: %v", err))
			}
		}(data)
	})
	if err != nil {
		log.Fatalf("Audio in stream error: %v", err)
	}
	defer inStream.Close()

	go startTextInputLoop(conn, debugMode)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				windowInfo := getActiveWindow()
				if err := sendTextMessage(conn, windowInfo); err != nil {
					UI.AddLogMessage(fmt.Sprintf("[Context] WebSocket write error: %v", err))
				} else {
					UI.AddLogMessage("[Context] Sent window info")
				}
			}
		}
	}()

	UI.AddLogMessage("[Info] マイクに向かって話しかけてください")

	startReceiveLoop(conn, UI)
}
