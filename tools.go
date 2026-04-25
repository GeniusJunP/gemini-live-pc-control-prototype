package main

import (
	"fmt"
	"runtime"

	"github.com/go-vgo/robotgo"
)

func getTools() []map[string]any {
	return []map[string]any{
		{
			"functionDeclarations": []map[string]any{
				{
					"name":        "execute_keybind",
					"description": "任意のキーボードショートカットを実行します。修飾キー（cmd/ctrl/shift/alt/option）とメインキーを組み合わせ可能。例: ['cmd', 'c'], ['ctrl', 'shift', 'n'], ['enter'], ['escape'], ['space']。必ず発火させるメインキーを配列の最後に配置してください。対応キー: a-z, 0-9, f1-f12, enter, return, escape, tab, space, up, down, left, right, delete, backspace, cmd, ctrl, shift, alt, option",
					"parameters": map[string]any{
						"type": "OBJECT",
						"properties": map[string]any{
							"keys": map[string]any{
								"type":        "ARRAY",
								"items":       map[string]any{"type": "STRING"},
								"description": "押下するキーのリスト",
							},
						},
						"required": []string{"keys"},
					},
				},
				{
					"name":        "type_text",
					"description": "テキストを入力します。カーソル位置に文字を挿入します。特殊キーは execute_keybind を使用してください。制御文字（改行など）は含めません。",
					"parameters": map[string]any{
						"type": "OBJECT",
						"properties": map[string]any{
							"text": map[string]any{
								"type":        "STRING",
								"description": "挿入するテキスト",
							},
						},
						"required": []string{"text"},
					},
				},
				{
					"name":        "activate_window",
					"description": "特定のウィンドウをアクティブにします。PIDまたはウィンドウタイトルを指定できます。タイトルは部分一致で検索します。",
					"parameters": map[string]any{
						"type": "OBJECT",
						"properties": map[string]any{
							"pid": map[string]any{
								"type":        "INTEGER",
								"description": "アクティブにするウィンドウのPID（pidとtitleのどちらかを指定）",
							},
							"title": map[string]any{
								"type":        "STRING",
								"description": "アクティブにするウィンドウのタイトル（部分一致、pidとtitleのどちらかを指定）",
							},
						},
					},
				},
				{
					"name":        "list_windows",
					"description": "全てのウィンドウの情報を取得します。PID、タイトル、位置、サイズを含みます。",
					"parameters": map[string]any{
						"type":       "OBJECT",
						"properties": map[string]any{},
					},
				},
				{
					"name":        "capture_screen",
					"description": "スクリーンショットを撮ります。特定のウィンドウの領域を指定することもできます。",
					"parameters": map[string]any{
						"type": "OBJECT",
						"properties": map[string]any{
							"x": map[string]any{
								"type":        "INTEGER",
								"description": "キャプチャ領域のX座標（省略時は全画面）",
							},
							"y": map[string]any{
								"type":        "INTEGER",
								"description": "キャプチャ領域のY座標（省略時は全画面）",
							},
							"width": map[string]any{
								"type":        "INTEGER",
								"description": "キャプチャ領域の幅（省略時は全画面）",
							},
							"height": map[string]any{
								"type":        "INTEGER",
								"description": "キャプチャ領域の高さ（省略時は全画面）",
							},
						},
					},
				},
			},
		},
	}
}

func getSystemPrompt() string {
	return `あなたはユーザーのPCで動く音声アシスタントです。短い相槌や返事を行ってください。
ユーザーは現在 ` + runtime.GOOS + ` を使用しています。
画面操作を求められたら execute_keybind ツールを使用してキーボードショートカットを実行してください。`
}

func getActiveWindow() string {
	pid := robotgo.GetPid()
	title := robotgo.GetTitle(pid)
	x, y, w, h := robotgo.GetBounds(pid)

	return fmt.Sprintf("Active Window:\n  Title: %s\n  PID: %d\n  Position: (%d, %d)\n  Size: %dx%d", title, pid, x, y, w, h)
}

func activateWindowByPID(pid int) bool {
	err := robotgo.ActivePid(pid)
	return err == nil
}

func activateWindowByTitle(title string) bool {
	pids, err := robotgo.FindIds(title)
	if err != nil || len(pids) == 0 {
		return false
	}
	err = robotgo.ActivePid(pids[0])
	return err == nil
}
