package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

var wsMutex sync.Mutex

func connectWebSocket(apiKey string) (*websocket.Conn, error) {
	url := "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=" + apiKey
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func createSetupMessage(debugMode bool) map[string]any {
	return map[string]any{
		"setup": map[string]any{
			"model": "models/gemini-3.1-flash-live-preview",
			"generationConfig": map[string]any{
				"responseModalities": []string{"AUDIO"},
				"speechConfig": map[string]any{
					"voiceConfig": map[string]any{
						"prebuiltVoiceConfig": map[string]any{
							"voiceName": "Aoede",
						},
					},
				},
			},
			"systemInstruction": map[string]any{
				"parts": []map[string]any{{"text": getSystemPrompt()}},
			},
			"tools": getTools(),
			"realtimeInputConfig": map[string]any{
				"automaticActivityDetection": map[string]any{
					"disabled": debugMode,
				},
			},
		},
	}
}

func sendSetupMessage(conn *websocket.Conn, debugMode bool) error {
	setupMsg := createSetupMessage(debugMode)
	wsMutex.Lock()
	err := conn.WriteJSON(setupMsg)
	wsMutex.Unlock()
	return err
}

func sendAudioMessage(conn *websocket.Conn, data []int16) error {
	encoded := encodeAudio(data)
	msg := map[string]any{
		"realtimeInput": map[string]any{
			"audio": map[string]any{
				"mimeType": "audio/pcm;rate=16000",
				"data":     encoded,
			},
		},
	}
	wsMutex.Lock()
	err := conn.WriteJSON(msg)
	wsMutex.Unlock()
	return err
}

func sendTextMessage(conn *websocket.Conn, text string) error {
	msg := map[string]any{
		"realtimeInput": map[string]any{
			"text": text,
		},
	}
	wsMutex.Lock()
	err := conn.WriteJSON(msg)
	wsMutex.Unlock()
	return err
}

func sendActivityStart(conn *websocket.Conn) error {
	msg := map[string]any{
		"realtimeInput": map[string]any{
			"activityStart": map[string]any{},
		},
	}
	wsMutex.Lock()
	err := conn.WriteJSON(msg)
	wsMutex.Unlock()
	return err
}

func sendActivityEnd(conn *websocket.Conn) error {
	msg := map[string]any{
		"realtimeInput": map[string]any{
			"activityEnd": map[string]any{},
		},
	}
	wsMutex.Lock()
	err := conn.WriteJSON(msg)
	wsMutex.Unlock()
	return err
}

func sendToolResponse(conn *websocket.Conn, responses []map[string]any) error {
	toolResp := map[string]any{
		"toolResponse": map[string]any{
			"functionResponses": responses,
		},
	}
	wsMutex.Lock()
	err := conn.WriteJSON(toolResp)
	wsMutex.Unlock()
	return err
}

func handleToolCall(functionCalls []any) []map[string]any {
	var functionResponses []map[string]any

	for _, fcAny := range functionCalls {
		fc := fcAny.(map[string]any)
		id := fc["id"].(string)
		name := fc["name"].(string)
		args := fc["args"].(map[string]any)

		log.Printf("[Tool Call] %s: %v", name, args)

		switch name {
		case "execute_keybind":
			keysAny := args["keys"].([]any)
			var keys []string
			for _, k := range keysAny {
				keys = append(keys, k.(string))
			}

			log.Printf("[Execute] Keybind: %v", keys)

			if len(keys) > 0 {
				mainKey := keys[len(keys)-1]
				var mods []interface{}
				for i := 0; i < len(keys)-1; i++ {
					mods = append(mods, keys[i])
				}
				robotgo.KeyTap(mainKey, mods...)
			}

			functionResponses = append(functionResponses, map[string]any{
				"id":       id,
				"name":     name,
				"response": map[string]any{"status": "success"},
			})

		case "type_text":
			text := args["text"].(string)
			log.Printf("[Type] Text: %s", text)
			robotgo.TypeStr(text)

			functionResponses = append(functionResponses, map[string]any{
				"id":       id,
				"name":     name,
				"response": map[string]any{"status": "success"},
			})

		default:
			log.Printf("[Tool] Unknown tool: %s", name)
			functionResponses = append(functionResponses, map[string]any{
				"id":       id,
				"name":     name,
				"response": map[string]any{"status": "error", "error": "Unknown tool"},
			})
		}
	}

	return functionResponses
}

func handleServerContent(serverContent map[string]any, ui *AudioUI) {
	var inputText, outputText string

	if inputTranscription, ok := serverContent["inputTranscription"].(map[string]any); ok {
		if text, ok := inputTranscription["text"].(string); ok {
			inputText = text
		}
	}

	if outputTranscription, ok := serverContent["outputTranscription"].(map[string]any); ok {
		if text, ok := outputTranscription["text"].(string); ok {
			outputText = text
		}
	}

	if inputText != "" || outputText != "" {
		ui.UpdateTranscription(inputText, outputText)
	}

	if interrupted, ok := serverContent["interrupted"].(bool); ok && interrupted {
		audioMutex.Lock()
		bufferSize := len(audioBuffer)
		audioBuffer = nil
		audioMutex.Unlock()
		ui.UpdateBufferStatus(bufferSize, true)
		log.Printf("[Audio Output] Buffer cleared due to interruption (buffer size: %d samples)", bufferSize)
	}

	if modelTurn, ok := serverContent["modelTurn"].(map[string]any); ok {
		if parts, ok := modelTurn["parts"].([]any); ok {
			for _, partAny := range parts {
				part := partAny.(map[string]any)
				if inlineData, ok := part["inlineData"].(map[string]any); ok {
					if dataStr, ok := inlineData["data"].(string); ok {
						samples := decodeAudio(dataStr)
						rms := calculateRMS(samples)
						addToAudioBuffer(samples)
						ui.UpdateOutput(rms, len(samples), len(audioBuffer))
					}
				}
			}
		}
	}
}

func startReceiveLoop(conn *websocket.Conn, ui *AudioUI) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("[WebSocket] Read error: %v", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(message, &resp); err != nil {
			log.Printf("[WebSocket] JSON unmarshal error: %v", err)
			continue
		}

		if _, ok := resp["setupComplete"]; ok {
			log.Println("[WebSocket] Setup complete")
		}

		if toolCall, ok := resp["toolCall"].(map[string]any); ok {
			if functionCalls, ok := toolCall["functionCalls"].([]any); ok {
				responses := handleToolCall(functionCalls)
				sendToolResponse(conn, responses)
			}
		}

		if serverContent, ok := resp["serverContent"].(map[string]any); ok {
			handleServerContent(serverContent, ui)
		}
	}
}

func startTextInputLoop(conn *websocket.Conn, debugMode bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if debugMode {
			if input == "start" {
				if err := sendActivityStart(conn); err != nil {
					log.Printf("[Activity] WebSocket write error: %v", err)
				} else {
					log.Println("[Activity] Sent: ActivityStart")
				}
				continue
			}

			if input == "end" {
				if err := sendActivityEnd(conn); err != nil {
					log.Printf("[Activity] WebSocket write error: %v", err)
				} else {
					log.Println("[Activity] Sent: ActivityEnd")
				}
				continue
			}
		}

		if err := sendTextMessage(conn, input); err != nil {
			log.Printf("[Text Input] WebSocket write error: %v", err)
		} else {
			log.Printf("[Text Input] Sent: %s", input)
		}
	}
}
