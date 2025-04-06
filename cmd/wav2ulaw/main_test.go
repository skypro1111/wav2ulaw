package main

import (
	"os"
	"runtime/pprof"
	"testing"
	"wav2ulaw"
)

func BenchmarkWavToUlaw(b *testing.B) {
	// Read test WAV file
	wavData, err := os.ReadFile("../../test_files/test.wav")
	if err != nil {
		b.Fatal(err)
	}

	// Create CPU profile
	f, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	err = pprof.StartCPUProfile(f)
	if err != nil {
		b.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Create memory profile
	memf, err := os.Create("mem.prof")
	if err != nil {
		b.Fatal(err)
	}
	defer memf.Close()

	config := wav2ulaw.DefaultAudioConfig()
	
	b.ResetTimer()
	
	// Run benchmark
	for i := 0; i < b.N; i++ {
		_, err := wav2ulaw.ConvertWavBytesToUlaw(wavData, config)
		if err != nil {
			b.Fatal(err)
		}
	}

	err = pprof.WriteHeapProfile(memf)
	if err != nil {
		b.Fatal(err)
	}
} 