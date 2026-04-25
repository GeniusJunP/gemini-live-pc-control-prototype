package main

import (
	"log"

	"github.com/CoyAce/apm"
)

type AudioProcessor struct {
	apm *apm.Processor
}

func NewAudioProcessor() (*AudioProcessor, error) {
	config := apm.Config{
		CaptureChannels:       1,
		RenderChannels:        1,
		HighPassFilterEnabled: true,
		EchoCancellation: apm.EchoCancellationConfig{
			Enabled: false,
		},
		NoiseSuppression: apm.NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: apm.NsLevelVeryHigh,
		},
	}

	processor, err := apm.New(config)
	if err != nil {
		return nil, err
	}

	processor.SetStreamAnalogLevel(128)

	log.Println("[AudioProcessor] Initialized with VeryHigh noise suppression")

	return &AudioProcessor{
		apm: processor,
	}, nil
}

func (ap *AudioProcessor) Process(samples []int16) []int16 {
	processed := make([]int16, len(samples))
	copy(processed, samples)

	frameSize := apm.GetNumSamplesPerFrame()
	for i := 0; i < len(processed); i += frameSize {
		end := i + frameSize
		if end > len(processed) {
			end = len(processed)
		}

		frame := processed[i:end]
		if err := ap.apm.ProcessCaptureInt16(frame); err != nil {
			log.Printf("[AudioProcessor] Error processing frame: %v", err)
		}
	}

	downsampled := downsample(processed, 3)
	return downsampled
}

func downsample(samples []int16, factor int) []int16 {
	if factor <= 1 {
		return samples
	}

	result := make([]int16, 0, len(samples)/factor)
	for i := 0; i < len(samples); i += factor {
		result = append(result, samples[i])
	}
	return result
}

func (ap *AudioProcessor) Close() {
	if ap.apm != nil {
		ap.apm.Close()
	}
}
