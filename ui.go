package main

import (
	"fmt"
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
		SetRows(1, 1, 1, 1, 1, 1, 1, 0, 1, 0).
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
			ui.transcriptBuf += fmt.Sprintf("[yellow]Input:[white] %s\n", input)
		}
		if output != "" {
			ui.transcriptBuf += fmt.Sprintf("[green]Output:[white] %s\n", output)
		}
		ui.transcription.SetText(ui.transcriptBuf)
		ui.transcription.ScrollToEnd()
	})
}

func (ui *AudioUI) AddLogMessage(message string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		ui.transcriptBuf += fmt.Sprintf("[blue]Log:[white] %s\n", message)
		ui.transcription.SetText(ui.transcriptBuf)
		ui.transcription.ScrollToEnd()
	})
}

func (ui *AudioUI) Stop() {
	ui.app.Stop()
}
