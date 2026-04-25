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
	outputLog     *tview.TextView
	mu            sync.Mutex
	transcriptBuf string
}

func NewAudioUI() *AudioUI {
	app := tview.NewApplication()

	inputBefore := tview.NewTextView()
	inputBefore.SetTitle("Audio Input (Before)")
	inputBefore.SetBorder(true)
	inputBefore.SetDynamicColors(true)

	inputAfter := tview.NewTextView()
	inputAfter.SetTitle("Audio Input (After)")
	inputAfter.SetBorder(true)
	inputAfter.SetDynamicColors(true)

	outputLevel := tview.NewTextView()
	outputLevel.SetTitle("Audio Output")
	outputLevel.SetBorder(true)
	outputLevel.SetDynamicColors(true)

	bufferStatus := tview.NewTextView()
	bufferStatus.SetTitle("Buffer Status")
	bufferStatus.SetBorder(true)
	bufferStatus.SetDynamicColors(true)

	transcription := tview.NewTextView()
	transcription.SetTitle("Transcription")
	transcription.SetBorder(true)
	transcription.SetDynamicColors(true)

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
	return tview.NewGrid().
		SetRows(3, 3, 3, 3, 0).
		SetColumns(0, 0, 0).
		AddItem(ui.inputBefore, 0, 0, 1, 3, 0, 0, true).
		AddItem(ui.inputAfter, 1, 0, 1, 3, 0, 0, false).
		AddItem(ui.outputLevel, 2, 0, 1, 3, 0, 0, false).
		AddItem(ui.bufferStatus, 3, 0, 1, 3, 0, 0, false).
		AddItem(ui.transcription, 4, 0, 1, 3, 0, 0, false)
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

func (ui *AudioUI) UpdateBufferStatus(bufferSize int, cleared bool) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.app.QueueUpdateDraw(func() {
		if cleared {
			ui.bufferStatus.SetText(fmt.Sprintf("[red]Buffer cleared (size: %d samples)[white]", bufferSize))
		} else {
			ui.bufferStatus.SetText(fmt.Sprintf("Buffer size: %d samples", bufferSize))
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
		ui.transcription.Clear()
		ui.transcription.SetText("\n" + ui.transcriptBuf)
		ui.transcription.ScrollToEnd()
	})
}

func (ui *AudioUI) Stop() {
	ui.app.Stop()
}
