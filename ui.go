package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rivo/tview"
)

type AudioUI struct {
	app           *tview.Application
	inputBefore   *tview.TextView
	inputAfter    *tview.TextView
	outputLevel   *tview.TextView
	bufferStatus  *tview.TextView
	transcription *tview.TextView
	mu            sync.Mutex
	transcriptBuf string
}

func NewAudioUI() *AudioUI {
	app := tview.NewApplication()

	inputBefore := tview.NewTextView()
	inputBefore.SetDynamicColors(true)

	inputAfter := tview.NewTextView()
	inputAfter.SetDynamicColors(true)

	outputLevel := tview.NewTextView()
	outputLevel.SetDynamicColors(true)

	bufferStatus := tview.NewTextView()
	bufferStatus.SetDynamicColors(true)
	bufferStatus.SetScrollable(false)

	transcription := tview.NewTextView()
	transcription.SetDynamicColors(true)
	transcription.SetScrollable(true)

	return &AudioUI{
		app:           app,
		inputBefore:   inputBefore,
		inputAfter:    inputAfter,
		outputLevel:   outputLevel,
		bufferStatus:  bufferStatus,
		transcription: transcription,
	}
}

func (ui *AudioUI) Start() {
	go func() {
		if err := ui.app.SetRoot(ui.createLayout(), true).Run(); err != nil {
			panic(err)
		}
	}()
}

func (ui *AudioUI) createLayout() *tview.Grid {
	grid := tview.NewGrid().
		SetRows(
			1, // Audio Input (Before) ラベル
			2, // inputBefore (RMS+バー)
			1, // Audio Input (After) ラベル
			2, // inputAfter (RMS+バー)
			1, // Audio Output ラベル
			3, // outputLevel (RMS, Samples, Buffer)
			1, // Buffer Status ラベル
			1, // bufferStatus（1行）
			1, // Transcription ラベル
			0, // transcription（残り全体）
		).
		SetColumns(0).
		AddItem(tview.NewTextView().SetText("[yellow]Audio Input (Before)[white]"), 0, 0, 1, 1, 0, 0, false).
		AddItem(ui.inputBefore, 1, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("[yellow]Audio Input (After)[white]"), 2, 0, 1, 1, 0, 0, false).
		AddItem(ui.inputAfter, 3, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("[yellow]Audio Output[white]"), 4, 0, 1, 1, 0, 0, false).
		AddItem(ui.outputLevel, 5, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("[yellow]Buffer Status[white]"), 6, 0, 1, 1, 0, 0, false).
		AddItem(ui.bufferStatus, 7, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("[yellow]Transcription[white]"), 8, 0, 1, 1, 0, 0, false).
		AddItem(ui.transcription, 9, 0, 1, 1, 0, 0, false)
	return grid
}

func (ui *AudioUI) UpdateInputBefore(rms float64, bar string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		ui.inputBefore.SetText(fmt.Sprintf("RMS: %.1f\n%s", rms, bar))
	})
}

func (ui *AudioUI) UpdateInputAfter(rms float64, bar string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		ui.inputAfter.SetText(fmt.Sprintf("RMS: %.1f\n%s", rms, bar))
	})
}

func (ui *AudioUI) UpdateOutput(rms float64, samples int, bufferRemaining int) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		ui.outputLevel.SetText(fmt.Sprintf("RMS: %.1f\nSamples: %d\nBuffer: %d", rms, samples, bufferRemaining))
	})
}

func (ui *AudioUI) UpdateBufferStatus(bufferSize int, cleared bool, timestamp string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		if cleared {
			ui.bufferStatus.SetText(fmt.Sprintf("[red]Cleared: %d samples @ %s[white]", bufferSize, timestamp))
		} else {
			ui.bufferStatus.SetText(fmt.Sprintf("Buffer: %d samples", bufferSize))
		}
	})
}

func (ui *AudioUI) UpdateTranscription(input, output string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		if input != "" {
			for _, line := range strings.Split(input, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					ui.transcriptBuf += fmt.Sprintf("[yellow]Input:[white] %s\n", trimmed)
				}
			}
		}
		if output != "" {
			for _, line := range strings.Split(output, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					ui.transcriptBuf += fmt.Sprintf("[green]Output:[white] %s\n", trimmed)
				}
			}
		}
		ui.truncateTranscript(500)
		ui.transcription.SetText(ui.transcriptBuf)
		ui.transcription.ScrollToEnd()
	})
}

func (ui *AudioUI) AddLogMessage(message string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		// 改行で分割し、空白行を除去して1行ずつ追加
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				ui.transcriptBuf += fmt.Sprintf("[blue]Log:[white] %s\n", trimmed)
			}
		}
		ui.transcription.SetText(ui.transcriptBuf)
		ui.transcription.ScrollToEnd()
	})
}

// truncateTranscript keeps the last maxLines lines in transcriptBuf.
func (ui *AudioUI) truncateTranscript(maxLines int) {
	lines := strings.Split(ui.transcriptBuf, "\n")
	// Remove any trailing empty element caused by trailing newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) <= maxLines {
		return
	}
	start := len(lines) - maxLines
	kept := lines[start:]
	ui.transcriptBuf = strings.Join(kept, "\n") + "\n"
}

// uiLogWriter implements io.Writer and routes log output into the UI.
type uiLogWriter struct {
	ui  *AudioUI
	buf string
	mu  sync.Mutex
}

func (w *uiLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	s := w.buf + string(p)
	parts := strings.Split(s, "\n")
	// The last element may be a partial line; keep it in buffer.
	for i := 0; i < len(parts)-1; i++ {
		trimmed := strings.TrimSpace(parts[i])
		if trimmed == "" {
			continue
		}
		w.ui.mu.Lock()
		w.ui.transcriptBuf += fmt.Sprintf("[blue]Log:[white] %s\n", trimmed)
		w.ui.truncateTranscript(500)
		w.ui.mu.Unlock()

		w.ui.app.QueueUpdateDraw(func() {
			w.ui.transcription.SetText(w.ui.transcriptBuf)
			w.ui.transcription.ScrollToEnd()
		})
	}

	// Save partial remainder (no newline yet) for next write.
	w.buf = parts[len(parts)-1]
	return len(p), nil
}

func (ui *AudioUI) Stop() {
	ui.app.Stop()
}
