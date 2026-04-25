# Get started with Gemini Live API using WebSockets

This tutorial shows you how to connect to Gemini Live API by using WebSockets. In this tutorial, you build a real-time multimodal application with a vanilla JavaScript frontend and a Python server handling the authentication and proxying.

## Before you begin

Complete the following steps to set up your environment.

- Sign in to your Google Cloud account. If you're new to Google Cloud, create an account to evaluate how our products perform in real-world scenarios. New customers also get $300 in free credits to run, test, and deploy workloads.
- In the Google Cloud console, on the project selector page, select or create a Google Cloud project.
- Verify that billing is enabled for your Google Cloud project.
- Enable the Vertex AI API.
- Install the Google Cloud CLI.
- Initialize the gcloud CLI: `gcloud init`
- Install Git
- Install Python 3

## Clone the demo app

Clone the demo app repository and navigate to that directory:

```bash
git clone https://github.com/GoogleCloudPlatform/generative-ai.git &&
cd generative-ai/gemini/multimodal-live-api/native-audio-websocket-demo-apps/plain-js-demo-app
```

### Project structure

The application includes the following files:

```
/
├── server.py            # WebSocket proxy + HTTP server
├── requirements.txt     # Python dependencies
└── frontend/
    ├── index.html       # UI
    ├── geminilive.js    # Gemini API client
    ├── mediaUtils.js    # Audio/video streaming
    ├── tools.js         # Custom tool definitions
    └── script.js        # App logic
```

## Run the backend server

The backend (`server.py`) handles the authentication and acts as a WebSocket proxy between the client and Gemini Live API.

To run the backend server, run the following commands:

1. Install dependencies:
   ```
   pip3 install -r requirements.txt
   ```

2. Run the server:
   ```
   python3 server.py
   ```

The server starts on `http://localhost:8080`.

## Run the frontend

Open `frontend/index.html` in your browser to start the application.

## Key implementation details

### WebSocket connection

The WebSocket connection is established to the Gemini Live API endpoint:

```
wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=YOUR_API_KEY
```

### Setup message

The setup message configures the session:

```json
{
  "setup": {
    "model": "models/gemini-live-2.5-flash-native-audio",
    "generationConfig": {
      "responseModalities": ["AUDIO"],
      "speechConfig": {
        "voiceConfig": {
          "prebuiltVoiceConfig": {
            "voiceName": "Aoede"
          }
        }
      }
    },
    "systemInstruction": {
      "parts": [{"text": "You are a helpful assistant."}]
    }
  }
}
```

### Audio streaming

Audio is sent in chunks using the `realtimeInput` message:

```json
{
  "realtimeInput": {
    "audio": {
      "mimeType": "audio/pcm;rate=16000",
      "data": "BASE64_ENCODED_AUDIO_DATA"
    }
  }
}
```

### Tool calling

Tools can be defined in the setup message and called by the model:

```json
{
  "setup": {
    "tools": [
      {
        "functionDeclarations": [
          {
            "name": "my_function",
            "description": "Description of the function",
            "parameters": {
              "type": "OBJECT",
              "properties": {
                "param1": {
                  "type": "STRING",
                  "description": "Parameter description"
                }
              },
              "required": ["param1"]
            }
          }
        ]
      }
    ]
  }
}
```

### Handling interruptions

When the model detects an interruption, it sends an `interrupted: true` signal. The client should immediately clear its audio playback buffer upon receiving this signal.
