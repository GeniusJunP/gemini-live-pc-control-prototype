package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
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

		case "activate_window":
			var success bool
			if pid, ok := args["pid"].(float64); ok {
				log.Printf("[Activate] Window by PID: %d", int(pid))
				success = activateWindowByPID(int(pid))
			} else if title, ok := args["title"].(string); ok {
				log.Printf("[Activate] Window by title: %s", title)
				success = activateWindowByTitle(title)
			}

			if success {
				functionResponses = append(functionResponses, map[string]any{
					"id":       id,
					"name":     name,
					"response": map[string]any{"status": "success"},
				})
			} else {
				functionResponses = append(functionResponses, map[string]any{
					"id":       id,
					"name":     name,
					"response": map[string]any{"status": "error", "error": "Failed to activate window"},
				})
			}

		case "list_windows":
			windows, err := GetAllWindows()
			if err != nil {
				log.Printf("[List Windows] Error: %v", err)
				functionResponses = append(functionResponses, map[string]any{
					"id":       id,
					"name":     name,
					"response": map[string]any{"status": "error", "error": err.Error()},
				})
			} else {
				windowList := make([]map[string]any, len(windows))
				for i, w := range windows {
					windowList[i] = map[string]any{
						"pid":    w.PID,
						"title":  w.Title,
						"x":      w.X,
						"y":      w.Y,
						"width":  w.Width,
						"height": w.Height,
					}
				}
				log.Printf("[List Windows] Found %d windows", len(windows))
				functionResponses = append(functionResponses, map[string]any{
					"id":       id,
					"name":     name,
					"response": map[string]any{"status": "success", "windows": windowList},
				})
			}

		case "capture_screen":
			var img image.Image
			var err error

			x, hasX := args["x"].(float64)
			y, hasY := args["y"].(float64)
			width, hasWidth := args["width"].(float64)
			height, hasHeight := args["height"].(float64)

			if hasX && hasY && hasWidth && hasHeight {
				log.Printf("[Capture Screen] Region: (%d, %d) %dx%d", int(x), int(y), int(width), int(height))
				img, err = robotgo.CaptureImg(int(x), int(y), int(width), int(height))
			} else {
				log.Printf("[Capture Screen] Full screen")
				img, err = robotgo.CaptureImg()
			}

			if err != nil {
				log.Printf("[Capture Screen] Error: %v", err)
				functionResponses = append(functionResponses, map[string]any{
					"id":       id,
					"name":     name,
					"response": map[string]any{"status": "error", "error": err.Error()},
				})
			} else {
				bounds := img.Bounds()
				maxSize := 768
				if bounds.Dx() > maxSize || bounds.Dy() > maxSize {
					log.Printf("[Capture Screen] Resizing from %dx%d to %dx%d", bounds.Dx(), bounds.Dy(), maxSize, maxSize)
					thumb := imaging.Resize(img, maxSize, maxSize, imaging.Lanczos)
					img = thumb
				}

				var buf bytes.Buffer
				err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
				if err != nil {
					log.Printf("[Capture Screen] JPEG encode error: %v", err)
					functionResponses = append(functionResponses, map[string]any{
						"id":       id,
						"name":     name,
						"response": map[string]any{"status": "error", "error": err.Error()},
					})
				} else {
					imageData := base64.StdEncoding.EncodeToString(buf.Bytes())
					log.Printf("[Capture Screen] Success, size: %d bytes", len(imageData))
					if len(imageData) > 10*1024*1024 {
						log.Printf("[Capture Screen] Warning: Image too large (%d bytes), skipping", len(imageData))
						functionResponses = append(functionResponses, map[string]any{
							"id":       id,
							"name":     name,
							"response": map[string]any{"status": "error", "error": "image too large"},
						})
					} else {
						functionResponses = append(functionResponses, map[string]any{
							"id":       id,
							"name":     name,
							"response": map[string]any{"status": "success", "image": imageData, "format": "jpeg"},
						})
					}
				}
			}

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
		timestamp := time.Now().Format("15:04:05")
		ui.UpdateBufferStatus(bufferSize, true, timestamp)
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
			log.Printf("[WebSocket] Read error: %v", err)
			return
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
