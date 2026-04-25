package main

import (
	"fmt"
	"math"
	"strings"
)

type AudioVisualizer struct {
	width  int
	height int
}

func NewAudioVisualizer(width, height int) *AudioVisualizer {
	return &AudioVisualizer{
		width:  width,
		height: height,
	}
}

func (av *AudioVisualizer) Visualize(samples []int16, label string) string {
	if len(samples) == 0 {
		return fmt.Sprintf("%s: [no data]", label)
	}

	rms := av.CalculateRMS(samples)
	bar := av.createBar(rms)
	return fmt.Sprintf("%s: %s RMS: %.1f", label, bar, rms)
}

func (av *AudioVisualizer) CalculateRMS(samples []int16) float64 {
	sum := 0.0
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func (av *AudioVisualizer) createBar(rms float64) string {
	maxRMS := 32768.0
	normalized := rms / maxRMS
	if normalized > 1.0 {
		normalized = 1.0
	}

	barWidth := int(normalized * float64(av.width))
	if barWidth > av.width {
		barWidth = av.width
	}

	bar := strings.Repeat("█", barWidth)
	spaces := strings.Repeat(" ", av.width-barWidth)
	return "[" + bar + spaces + "]"
}

func (av *AudioVisualizer) VisualizeWaveform(samples []int16, label string) string {
	if len(samples) == 0 {
		return fmt.Sprintf("%s: [no data]", label)
	}

	height := av.height
	width := av.width

	step := len(samples) / width
	if step < 1 {
		step = 1
	}

	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", width)
	}

	for x := 0; x < width && x*step < len(samples); x++ {
		sample := samples[x*step]
		normalized := float64(sample) / 32768.0
		if normalized > 1.0 {
			normalized = 1.0
		}
		if normalized < -1.0 {
			normalized = -1.0
		}

		center := height / 2
		offset := int(normalized * float64(center))
		y := center - offset

		if y >= 0 && y < height {
			line := []rune(lines[y])
			if x < len(line) {
				line[x] = '█'
				lines[y] = string(line)
			}
		}
	}

	result := fmt.Sprintf("%s:\n", label)
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
