// Copyright (c) 2024 skypro1111@gmail.com
// All rights reserved.

package main

import (
	"flag"
	"fmt"
	"wav2ulaw"
	"os"
)

func main() {
	// Define command line flags
	inputFile := flag.String("input", "", "Input file path")
	outputFile := flag.String("output", "", "Output file path")
	mode := flag.String("mode", "wav2ulaw", "Conversion mode: wav2ulaw or ulaw2wav")
	sampleRate := flag.Uint("sample-rate", 8000, "Sample rate for output WAV file (only for ulaw2wav mode)")
	lowPass := flag.Float64("low-pass", 3400, "Low-pass filter cutoff frequency in Hz")
	highPass := flag.Float64("high-pass", 300, "High-pass filter cutoff frequency in Hz")
	normalize := flag.Float64("normalize", 0.9, "Normalize audio to this peak level (0.0 to 1.0)")
	compressRatio := flag.Float64("compress-ratio", 2.0, "Compression ratio (1.0 means no compression)")
	compressThreshold := flag.Float64("compress-threshold", 0.5, "Compression threshold (0.0 to 1.0)")
	windowSize := flag.Int("window-size", 16, "Resampling window size (larger = better quality but slower)")
	antiAliasingRatio := flag.Float64("anti-aliasing-ratio", 0.9, "Anti-aliasing filter cutoff ratio (0.0 to 1.0)")
	antiAliasingType := flag.Int("anti-aliasing-type", int(wav2ulaw.AAButterworth), "Anti-aliasing filter type (0=Simple, 1=Butterworth, 2=Bessel, 3=Chebyshev)")
	filterOrder := flag.Int("filter-order", 4, "Filter order for Butterworth/Bessel/Chebyshev (2-6)")
	chebyshevRipple := flag.Float64("chebyshev-ripple", 0.5, "Ripple in dB for Chebyshev filter (0.1-3.0)")

	flag.Parse()

	// Validate input parameters
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Error: Input and output file paths are required")
		flag.Usage()
		os.Exit(1)
	}

	// Read input file
	inputData, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var outputData []byte

	// Process based on mode
	if *mode == "wav2ulaw" {
		config := &wav2ulaw.AudioConfig{
			LowPassCutoff:          *lowPass,
			HighPassCutoff:         *highPass,
			NormalizePeak:          *normalize,
			CompressionRatio:       *compressRatio,
			CompressionThreshold:   *compressThreshold,
			ResamplingWindowSize:   *windowSize,
			AntiAliasingCutoffRatio: *antiAliasingRatio,
			AntiAliasingType:       wav2ulaw.AntiAliasingType(*antiAliasingType),
			FilterOrder:            *filterOrder,
			ChebyshevRipple:       *chebyshevRipple,
		}

		outputData, err = wav2ulaw.ConvertWavBytesToUlaw(inputData, config)
		if err != nil {
			fmt.Printf("Error converting WAV to u-law: %v\n", err)
			os.Exit(1)
		}
	} else if *mode == "ulaw2wav" {
		outputData, err = wav2ulaw.ConvertUlawBytesToWav(inputData, uint32(*sampleRate), *windowSize)
		if err != nil {
			fmt.Printf("Error converting u-law to WAV: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Error: Invalid mode '%s'. Must be 'wav2ulaw' or 'ulaw2wav'\n", *mode)
		os.Exit(1)
	}

	// Write output file
	err = os.WriteFile(*outputFile, outputData, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully")
} 