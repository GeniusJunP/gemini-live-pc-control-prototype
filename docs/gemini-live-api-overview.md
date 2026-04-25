# Gemini Live API Overview

## Caution
`gemini-live-2.5-flash-preview-native-audio-09-2025` will be deprecated and removed on March 19, 2026. Migrate any workflows to `gemini-live-2.5-flash-native-audio`.

## Example use cases
- E-commerce and retail: Shopping assistants, support agents
- Gaming: Interactive NPCs, in-game help assistants, real-time translation
- Next-gen interfaces: Voice- and video-enabled experiences in robotics, smart glasses, vehicles
- Healthcare: Health companions for patient support and education
- Financial services: AI advisors for wealth management and investment guidance
- Education: AI mentors and learner companions

## Key features
- High audio quality: Natural, realistic-sounding speech across multiple languages
- Multilingual support: Converse in 24 supported languages
- Barge-in: Users can interrupt the model at any time for responsive interactions
- Affective dialog: Adapts response style and tone to match user's input expression
- Tool use: Integrates tools like function calling and Google Search for dynamic interactions
- Audio transcriptions: Provides text transcripts of both user input and model output
- Proactive audio: (Preview) Lets you control when the model responds and in what contexts

## Technical specifications

| Category | Details |
|----------|---------|
| Input modalities | Audio (raw 16-bit PCM audio, 16kHz, little-endian), images/video (JPEG 1FPS), text |
| Output modalities | Audio (raw 16-bit PCM audio, 24kHz, little-endian), text |
| Protocol | Stateful WebSocket connection (WSS) |

## Supported models

| Model ID | Availability | Use case | Key features |
|----------|--------------|----------|--------------|
| `gemini-live-2.5-flash-native-audio` | Generally available | **Recommended**. Low-latency voice agents. Supports seamless multilingual switching and emotional tone. | Native audio, Audio transcriptions, Voice activity detection, Affective dialog, Proactive audio, Tool use |
| `gemini-live-2.5-flash-preview-native-audio-09-2025` | Public preview | Cost-efficiency in real-time voice agents. | Native audio, Audio transcriptions, Voice activity detection, Affective dialog, Proactive audio, Tool use |

## Get started
- Gen AI SDK tutorial: Connect using the Gen AI SDK
- WebSocket tutorial: Connect using WebSockets
- ADK tutorial: Create an agent using Agent Development Kit (ADK) Streaming

## Partner integrations
- Daily
- LiveKit
- Twilio
- Voximplant
