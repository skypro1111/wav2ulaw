// Copyright (c) 2024 skypro1111@gmail.com
// All rights reserved.

package wav2ulaw

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/zaf/g711"
	"io"
	"math"
	"os"
)

// AntiAliasingType defines the type of anti-aliasing filter to use
type AntiAliasingType int

const (
	AASimple AntiAliasingType = iota  // Simple lowpass filter
	AAButterworth                      // Butterworth filter
	AABessel                          // Bessel filter
	AAChebyshev                       // Chebyshev Type I filter
)

// AudioConfig contains configuration for audio processing
type AudioConfig struct {
	// Input sample rate (Hz). If not specified, will be detected from WAV file
	InputSampleRate int
	// Force mono conversion before processing
	ForceMono bool
	// Low-pass filter cutoff frequency (Hz)
	LowPassCutoff float64
	// High-pass filter cutoff frequency (Hz)
	HighPassCutoff float64
	// Normalize audio to this peak level (-1.0 to 1.0)
	NormalizePeak float64
	// Compression ratio (1.0 means no compression)
	CompressionRatio float64
	// Compression threshold (-1.0 to 1.0)
	CompressionThreshold float64
	// Resampling window size (larger = better quality but slower)
	ResamplingWindowSize int
	// Anti-aliasing filter cutoff ratio (0.0 to 1.0, relative to Nyquist frequency)
	AntiAliasingCutoffRatio float64
	// Anti-aliasing filter type
	AntiAliasingType AntiAliasingType
	// Filter order for Butterworth/Bessel/Chebyshev
	FilterOrder int
	// Ripple in dB for Chebyshev filter
	ChebyshevRipple float64
}

// DefaultAudioConfig returns default audio configuration
func DefaultAudioConfig() *AudioConfig {
	return &AudioConfig{
		InputSampleRate:        0,      // Auto-detect from WAV
		ForceMono:             true,    // Convert to mono by default
		LowPassCutoff:         3400,    // Telephone bandwidth
		HighPassCutoff:        200,     // Soft low-frequency cutoff
		NormalizePeak:         0.95,    // 95% of maximum amplitude
		CompressionRatio:      1.5,     // Light compression
		CompressionThreshold:  0.5,     // Start compression at 50% of maximum amplitude
		ResamplingWindowSize:  64,      // Larger window for better quality
		AntiAliasingCutoffRatio: 0.95,  // Soft anti-aliasing
		AntiAliasingType:      AASimple, // Simple filter for stability
		FilterOrder:           2,       // Low order for stability
		ChebyshevRipple:      0.1,     // Minimal ripple
	}
}

// applyButterworthFilter applies a Butterworth low-pass filter
func applyButterworthFilter(samples []int16, sampleRate, cutoffFreq float64, order int) []int16 {
	// Normalize frequency
	wc := 2.0 * math.Pi * cutoffFreq / sampleRate
	
	// Create filter
	result := make([]int16, len(samples))
	copy(result, samples)
	
	// Calculate filter coefficients
	alpha := math.Tan(wc / 2.0)
	cosw := math.Cos(wc)
	
	a0 := 1.0 + alpha
	b0 := (1.0 - cosw) / 2.0
	b1 := 1.0 - cosw
	b2 := (1.0 - cosw) / 2.0
	a1 := -2.0 * cosw
	a2 := 1.0 - alpha
	
	// Normalize coefficients
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0
	
	// Apply filter
	x1, x2 := 0.0, 0.0
	y1, y2 := 0.0, 0.0
	
	for i := range result {
		x := float64(result[i]) / 32767.0 // Normalize to [-1, 1]
		
		// Direct form II
		y := b0*x + b1*x1 + b2*x2 - a1*y1 - a2*y2
		
		// Update delays
		x2, x1 = x1, x
		y2, y1 = y1, y
		
		// Scale back to int16
		result[i] = int16(math.Max(-32768, math.Min(32767, y*32767.0)))
	}
	
	return result
}

// applyBesselFilter applies a Bessel low-pass filter
func applyBesselFilter(samples []int16, sampleRate, cutoffFreq float64, order int) []int16 {
	// Normalize frequency
	wc := 2.0 * math.Pi * cutoffFreq / sampleRate
	
	// Create filter
	result := make([]int16, len(samples))
	copy(result, samples)
	
	// Calculate filter coefficients (3rd order Bessel approximation)
	alpha := math.Tan(wc / 2.0)
	alphaSq := alpha * alpha
	alphaCu := alphaSq * alpha
	
	a0 := 1.0 + 2.15*alpha + 3.15*alphaSq + 1.0*alphaCu
	b0 := 1.0 * alphaCu
	b1 := 3.0 * alphaCu
	b2 := 3.0 * alphaCu
	b3 := 1.0 * alphaCu
	a1 := (-2.0 + 2.15*alpha - 3.15*alphaSq + 3.0*alphaCu) 
	a2 := (1.0 - 2.15*alpha + 3.15*alphaSq - 3.0*alphaCu)
	a3 := (-0.0 + 0.0*alpha - 0.0*alphaSq + 1.0*alphaCu)
	
	// Normalize coefficients
	b0 /= a0
	b1 /= a0
	b2 /= a0
	b3 /= a0
	a1 /= a0
	a2 /= a0
	a3 /= a0
	
	// Apply filter
	x1, x2, x3 := 0.0, 0.0, 0.0
	y1, y2, y3 := 0.0, 0.0, 0.0
	
	for i := range result {
		x := float64(result[i]) / 32767.0 // Normalize to [-1, 1]
		
		// Direct form II
		y := b0*x + b1*x1 + b2*x2 + b3*x3 - a1*y1 - a2*y2 - a3*y3
		
		// Update delays
		x3, x2, x1 = x2, x1, x
		y3, y2, y1 = y2, y1, y
		
		// Scale back to int16
		result[i] = int16(math.Max(-32768, math.Min(32767, y*32767.0)))
	}
	
	return result
}

// applyChebyshevFilter applies a Chebyshev Type I low-pass filter
func applyChebyshevFilter(samples []int16, sampleRate, cutoffFreq, rippleDb float64, order int) []int16 {
	// Normalize frequency
	wc := 2.0 * math.Pi * cutoffFreq / sampleRate
	
	// Create filter
	result := make([]int16, len(samples))
	copy(result, samples)
	
	// Calculate filter coefficients (2nd order Chebyshev approximation)
	epsilon := math.Sqrt(math.Pow(10, rippleDb/10) - 1)
	v0 := math.Asinh(1/epsilon) / 2.0
	
	alpha := math.Tan(wc / 2.0)
	alphaSq := alpha * alpha
	
	sinh := math.Sinh(v0)
	
	// Calculate coefficients
	a0 := 1.0 + 2.0*alpha*sinh + alphaSq
	b0 := alphaSq
	b1 := 2.0 * alphaSq
	b2 := alphaSq
	a1 := 2.0 * (alphaSq - 1.0)
	a2 := 1.0 - 2.0*alpha*sinh + alphaSq
	
	// Normalize coefficients
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0
	
	// Apply filter
	x1, x2 := 0.0, 0.0
	y1, y2 := 0.0, 0.0
	
	for i := range result {
		x := float64(result[i]) / 32767.0 // Normalize to [-1, 1]
		
		// Direct form II
		y := b0*x + b1*x1 + b2*x2 - a1*y1 - a2*y2
		
		// Update delays
		x2, x1 = x1, x
		y2, y1 = y1, y
		
		// Scale back to int16
		result[i] = int16(math.Max(-32768, math.Min(32767, y*32767.0)))
	}
	
	return result
}

// applyAntiAliasingFilter applies the selected anti-aliasing filter
func applyAntiAliasingFilter(samples []int16, sampleRate, targetRate float64, config *AudioConfig) []int16 {
	// Nyquist frequency of target sample rate
	nyquistFreq := targetRate / 2.0
	
	// Use configured cutoff ratio
	cutoffFreq := nyquistFreq * config.AntiAliasingCutoffRatio
	
	// If source sample rate is lower than target, no need for anti-aliasing
	if sampleRate <= targetRate {
		return samples
	}
	
	// Apply selected filter type
	switch config.AntiAliasingType {
	case AAButterworth:
		return applyButterworthFilter(samples, sampleRate, cutoffFreq, config.FilterOrder)
	case AABessel:
		return applyBesselFilter(samples, sampleRate, cutoffFreq, config.FilterOrder)
	case AAChebyshev:
		return applyChebyshevFilter(samples, sampleRate, cutoffFreq, config.ChebyshevRipple, config.FilterOrder)
	default: // AASimple
		return applyLowPassFilter(samples, sampleRate, cutoffFreq)
	}
}

// resamplePCM16 resamples 16-bit PCM audio to a new sample rate using windowed sinc interpolation
func resamplePCM16(input []int16, inputRate, outputRate float64, windowSize int) []int16 {
	ratio := outputRate / inputRate
	outputLen := int(float64(len(input)) * ratio)
	output := make([]int16, outputLen)

	// Pre-calculate window coefficients
	window := make([]float64, windowSize*2+1)
	for i := range window {
		// Blackman window
		x := float64(i) / float64(len(window)-1)
		window[i] = 0.42 - 0.5*math.Cos(2*math.Pi*x) + 0.08*math.Cos(4*math.Pi*x)
	}

	for i := range output {
		pos := float64(i) / ratio
		idx := int(pos)
		
		// Calculate sinc interpolation
		sum := 0.0
		weightSum := 0.0

		for j := -windowSize; j <= windowSize; j++ {
			inputIdx := idx + j
			if inputIdx < 0 || inputIdx >= len(input) {
				continue
			}

			// Calculate sinc value
			x := math.Pi * (pos - float64(inputIdx))
			var sinc float64
			if x == 0 {
				sinc = 1.0
			} else {
				sinc = math.Sin(x) / x
			}

			// Apply window function
			weight := window[j+windowSize] * sinc
			sum += float64(input[inputIdx]) * weight
			weightSum += weight
		}

		// Normalize and convert back to int16
		if weightSum > 0 {
			sum /= weightSum
		}
		output[i] = int16(math.Round(sum))
	}

	return output
}

// applyHighPassFilter applies a simple high-pass filter to the samples
func applyHighPassFilter(samples []int16, sampleRate float64, cutoffFreq float64) []int16 {
	// Calculate RC constant for the filter
	rc := 1.0 / (2.0 * math.Pi * cutoffFreq)
	dt := 1.0 / sampleRate
	alpha := rc / (rc + dt)

	filtered := make([]int16, len(samples))
	filtered[0] = samples[0]
	var prevInput float64
	var prevOutput float64

	for i := 1; i < len(samples); i++ {
		input := float64(samples[i])
		// High pass filter formula: y[i] = alpha * (y[i-1] + x[i] - x[i-1])
		output := alpha * (prevOutput + input - prevInput)
		filtered[i] = int16(math.Round(output))
		prevInput = input
		prevOutput = output
	}

	return filtered
}

// applyLowPassFilter applies a simple low-pass filter to the samples
func applyLowPassFilter(samples []int16, sampleRate float64, cutoffFreq float64) []int16 {
	// Calculate RC constant for the filter
	rc := 1.0 / (2.0 * math.Pi * cutoffFreq)
	dt := 1.0 / sampleRate
	alpha := dt / (rc + dt)

	filtered := make([]int16, len(samples))
	filtered[0] = samples[0]

	for i := 1; i < len(samples); i++ {
		// Low pass filter formula: y[i] = y[i-1] + alpha * (x[i] - y[i-1])
		float_sample := float64(filtered[i-1]) + alpha*float64(samples[i]-filtered[i-1])
		filtered[i] = int16(math.Round(float_sample))
	}

	return filtered
}

// normalizeAudio normalizes audio to the specified peak level
func normalizeAudio(samples []int16, peakLevel float64) []int16 {
	// Find current peak
	maxAbs := float64(0)
	for _, sample := range samples {
		abs := math.Abs(float64(sample))
		if abs > maxAbs {
			maxAbs = abs
		}
	}

	// Calculate scaling factor
	scale := (peakLevel * 32767.0) / maxAbs

	// Apply normalization
	normalized := make([]int16, len(samples))
	for i, sample := range samples {
		normalized[i] = int16(math.Round(float64(sample) * scale))
	}

	return normalized
}

// applyCompression applies dynamic range compression
func applyCompression(samples []int16, ratio, threshold float64) []int16 {
	compressed := make([]int16, len(samples))
	thresholdAbs := threshold * 32767.0

	for i, sample := range samples {
		sampleFloat := float64(sample)
		sampleAbs := math.Abs(sampleFloat)

		if sampleAbs > thresholdAbs {
			// Apply compression above threshold
			excess := sampleAbs - thresholdAbs
			compressed[i] = int16(math.Round(math.Copysign(
				thresholdAbs + (excess/ratio),
				sampleFloat,
			)))
		} else {
			compressed[i] = sample
		}
	}

	return compressed
}

// ConvertWavBytesToUlaw converts WAV file bytes to u-law encoded bytes
func ConvertWavBytesToUlaw(wavBytes []byte, config *AudioConfig) ([]byte, error) {
	if config == nil {
		config = DefaultAudioConfig()
	}

	// Create a decoder
	reader := bytes.NewReader(wavBytes)
	decoder := wav.NewDecoder(reader)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file")
	}

	// Read audio format
	format := decoder.Format()
	if format == nil {
		return nil, fmt.Errorf("error reading WAV format")
	}

	// Read audio data
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("error reading WAV data: %v", err)
	}

	// Get actual input sample rate
	inputSampleRate := config.InputSampleRate
	if inputSampleRate == 0 {
		inputSampleRate = format.SampleRate
	}

	// Convert samples to int16 and handle mono conversion if needed
	var samples []int16
	if config.ForceMono && format.NumChannels > 1 {
		// Average all channels to mono
		samples = make([]int16, len(buf.Data)/format.NumChannels)
		for i := 0; i < len(samples); i++ {
			sum := 0
			for ch := 0; ch < format.NumChannels; ch++ {
				idx := i*format.NumChannels + ch
				if idx < len(buf.Data) {
					sum += buf.Data[idx]
				}
			}
			avg := sum / format.NumChannels
			if buf.SourceBitDepth == 8 {
				samples[i] = int16((avg + 128) << 8)
			} else {
				samples[i] = int16(avg)
			}
		}
	} else {
		// Convert to int16 without channel mixing
		samples = make([]int16, len(buf.Data))
		for i, sample := range buf.Data {
			if buf.SourceBitDepth == 8 {
				samples[i] = int16((sample + 128) << 8)
			} else {
				samples[i] = int16(sample)
			}
		}
	}

	// Apply audio processing on original sample rate
	if config.HighPassCutoff > 0 {
		samples = applyHighPassFilter(samples, float64(inputSampleRate), config.HighPassCutoff)
	}

	if config.LowPassCutoff > 0 {
		samples = applyLowPassFilter(samples, float64(inputSampleRate), config.LowPassCutoff)
	}

	// Apply anti-aliasing filter before resampling
	samples = applyAntiAliasingFilter(samples, float64(inputSampleRate), 8000, config)

	// Resample to 8kHz
	if inputSampleRate != 8000 {
		samples = resamplePCM16(samples, float64(inputSampleRate), 8000, config.ResamplingWindowSize)
	}

	// Apply volume processing after resampling
	if config.CompressionRatio > 1.0 {
		samples = applyCompression(samples, config.CompressionRatio, config.CompressionThreshold)
	}

	if config.NormalizePeak > 0 {
		samples = normalizeAudio(samples, config.NormalizePeak)
	}

	// Convert samples to bytes for g711
	pcmBytes := make([]byte, len(samples)*2)
	for i, sample := range samples {
		binary.LittleEndian.PutUint16(pcmBytes[i*2:], uint16(sample))
	}

	// Convert to u-law
	ulawData := g711.EncodeUlaw(pcmBytes)
	return ulawData, nil
}

// ConvertUlawBytesToWav converts u-law encoded bytes back to WAV file bytes
func ConvertUlawBytesToWav(ulawBytes []byte, sampleRate uint32, windowSize int) ([]byte, error) {
	// Convert u-law to PCM
	pcmData := g711.DecodeUlaw(ulawBytes)

	// Convert bytes to int16 samples
	samples := make([]int16, len(pcmData)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2:]))
	}

	// Resample if needed
	if sampleRate != 8000 {
		samples = resamplePCM16(samples, 8000, float64(sampleRate), windowSize)
	}

	// Create temporary file for WAV encoder
	tmpFile, err := os.CreateTemp("", "wav_*.wav")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create WAV encoder
	enc := wav.NewEncoder(tmpFile, int(sampleRate), 16, 1, 1)

	// Convert samples to PCM buffer
	audioBuf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate: int(sampleRate),
		},
		Data:           make([]int, len(samples)),
		SourceBitDepth: 16,
	}

	// Convert samples to int
	for i, sample := range samples {
		audioBuf.Data[i] = int(sample)
	}

	// Write audio data
	if err := enc.Write(audioBuf); err != nil {
		return nil, fmt.Errorf("error writing WAV data: %v", err)
	}

	// Close encoder
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("error closing WAV encoder: %v", err)
	}

	// Read the temporary file
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking temporary file: %v", err)
	}

	wavData, err := io.ReadAll(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("error reading temporary file: %v", err)
	}

	return wavData, nil
} 