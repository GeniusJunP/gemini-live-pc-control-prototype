package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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
	fmt.Println(" Gemini Live API に接続しました")

	if debugMode {
		log.Println("[Mode] Debug mode: Manual ActivityStart/ActivityEnd")
	} else {
		log.Println("[Mode] Normal mode: Automatic VAD")
	}

	if err := sendSetupMessage(conn, debugMode); err != nil {
		log.Fatalf("Setup send error: %v", err)
	}

	processor, err := NewAudioProcessor()
	if err != nil {
		log.Fatalf("Audio processor error: %v", err)
	}
	defer processor.Close()

	visualizer := NewAudioVisualizer(50, 10)
	ui := NewAudioUI()
	ui.Start()
	defer ui.Stop()

	outStream, err := startOutputStream(ui)
	if err != nil {
		log.Fatalf("Audio out stream error: %v", err)
	}
	defer outStream.Close()

	inStream, err := startInputStream(func(data []int16) {
		go func(d []int16) {
			rmsBefore := visualizer.CalculateRMS(d)
			processed := processor.Process(d)
			rmsAfter := visualizer.CalculateRMS(processed)

			ui.UpdateInputBefore(rmsBefore, visualizer.Visualize(d, "Before"))
			ui.UpdateInputAfter(rmsAfter, visualizer.Visualize(processed, "After"))

			if err := sendAudioMessage(conn, processed); err != nil {
				log.Printf("[Audio Input] WebSocket write error: %v", err)
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
					log.Printf("[Context] WebSocket write error: %v", err)
				} else {
					log.Printf("[Context] Sent window info")
				}
			}
		}
	}()

	fmt.Println(" マイクに向かって話しかけてください")

	startReceiveLoop(conn, ui)
}
