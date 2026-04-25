# Audio Control

Gemini Live API を使った音声アシスタントのプロトタイプ

## 機能

マイクからの音声入力（48kHz PCM16）を Gemini Live API に送信
Gemini からの音声応答（24kHz PCM16）をスピーカーで再生
ツール呼び出し（execute_keybind, type_text）によるキーボード操作実行
30秒ごとにアクティブなウィンドウ情報を自動送信
音声前処理（ノイズ除去、VAD）の実装
リアルタイムUIによる音声波形の可視化

## インストール

```bash
go mod download
```

## 使い方

1 `.env` ファイルを作成して API キーを設定

```env
GEMINI_API_KEY=your_api_key_here
DEBUG_MODE=false
```

2 ビルド

```bash
go build -o audio-control
```

3 実行

```bash
./audio-control
```

## 依存関係

- github.com/go-vgo/robotgo
- github.com/gordonklaus/portaudio
- github.com/gorilla/websocket
- github.com/CoyAce/apm
- github.com/rivo/tview
- github.com/joho/godotenv

## 動作環境

- macOS
- Windows
