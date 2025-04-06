# WAV to μ-law Converter

High-quality WAV to μ-law (u-law) converter optimized for Text-to-Speech (TTS) output. This tool is designed to efficiently convert WAV files (including float format) to telephone-quality μ-law format while maintaining the best possible sound quality.

## Requirements

### Go requirements
- Go 1.21 or higher
- The following Go packages (will be installed automatically):
  - github.com/go-audio/audio v1.0.0
  - github.com/go-audio/wav v1.1.0
  - github.com/zaf/g711 v1.4.0

### Python requirements
- Python 3.8 or higher
- Required packages:
  - soundfile
  - numpy (dependency of soundfile)

### System requirements (for audio playback)
- macOS: 
  - sox (recommended): `brew install sox`
  - or afplay (built-in)
- Linux:
  - sox: `sudo apt-get install sox` or equivalent
- Windows:
  - sox: Download from [Sox website](http://sox.sourceforge.net/)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/wav2ulaw.git
cd wav2ulaw
```

2. Install Go dependencies and build the binary:
```bash
# Install Go dependencies
go mod download

# Build the binary
go build -o wav2ulaw cmd/wav2ulaw/main.go
```

3. Install Python dependencies:
```bash
pip install soundfile numpy
```

4. (Optional) Install sox for better audio playback:
```bash
# macOS
brew install sox

# Ubuntu/Debian
sudo apt-get install sox

# Windows
# Download and install from http://sox.sourceforge.net/
```

## Development Setup

1. Install development tools:
```bash
# Install Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install Python tools
pip install black pylint
```

2. Set up pre-commit hooks (optional):
```bash
# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
go fmt ./...
golangci-lint run
black .
pylint *.py
EOF

chmod +x .git/hooks/pre-commit
```

## Building from source

### Debug build
```bash
go build -o wav2ulaw -gcflags=all="-N -l" cmd/wav2ulaw/main.go
```

### Release build
```bash
# Linux/macOS
go build -o wav2ulaw -ldflags="-s -w" cmd/wav2ulaw/main.go

# Windows
go build -o wav2ulaw.exe -ldflags="-s -w" cmd/wav2ulaw/main.go
```

### Cross-compilation

For Linux (from any platform):
```bash
GOOS=linux GOARCH=amd64 go build -o wav2ulaw-linux-amd64 -ldflags="-s -w" cmd/wav2ulaw/main.go
```

For macOS (from any platform):
```bash
GOOS=darwin GOARCH=amd64 go build -o wav2ulaw-macos-amd64 -ldflags="-s -w" cmd/wav2ulaw/main.go
```

For Windows (from any platform):
```bash
GOOS=windows GOARCH=amd64 go build -o wav2ulaw-windows-amd64.exe -ldflags="-s -w" cmd/wav2ulaw/main.go
```

## Testing

Run Go tests:
```bash
go test ./...
```

Run Python tests:
```bash
python test.py
```

## Purpose

This project addresses the common need to convert TTS-generated WAV files to telephone-compatible μ-law format. It's specifically optimized for:

- Converting TTS output (WAV, including float format) to telephone format
- Maintaining maximum possible voice quality
- Providing efficient processing speed
- Easy integration with Python TTS workflows
- Handling various input sample rates (8kHz-48kHz)
- Automatic mono conversion for multi-channel audio

## Features

- High-quality audio processing pipeline:
  - Support for float and PCM WAV formats
  - Configurable anti-aliasing filters (Simple, Butterworth, Bessel, Chebyshev)
  - Telephone bandwidth optimization (200-3400 Hz)
  - Automatic volume normalization
  - Optional dynamic range compression
  - High-quality resampling
  - Multi-channel to mono conversion
  - Support for various input sample rates
- Fast Go implementation with Python bindings
- Simple command-line interface
- Easy integration with Python TTS systems

## Python Usage Example

Here's a simple example of converting TTS output to μ-law format:

```python
from wav2ulaw import wav_to_ulaw, AA_BUTTERWORTH

# Read your TTS-generated WAV file
with open('tts_output.wav', 'rb') as f:
    wav_data = f.read()

# Convert to μ-law with optimal TTS settings
ulaw_data = wav_to_ulaw(
    wav_data,
    input_sample_rate=44100,  # Specify input sample rate if known
    force_mono=True,          # Convert to mono if needed
    low_pass=3400,           # Telephone bandwidth
    high_pass=200,           # Remove low frequency noise
    normalize=0.95,          # Normalize volume
    compression_ratio=1.5,    # Light compression for better audibility
    window_size=12,          # Resampling window size (smaller = faster)
    anti_aliasing_type=AA_BUTTERWORTH  # Better frequency response
)

# Save the result
with open('output.ulaw', 'wb') as f:
    f.write(ulaw_data)
```

## Advanced Configuration

The converter supports various audio processing parameters that can be tuned for your specific needs:

```python
from wav2ulaw import wav_to_ulaw, AA_BUTTERWORTH, AA_BESSEL, AA_CHEBYSHEV

# Example with all available parameters
ulaw_data = wav_to_ulaw(
    wav_bytes,
    input_sample_rate=44100,       # Input sample rate (Hz), 0 for auto-detect
    force_mono=True,               # Convert multi-channel to mono
    low_pass=3400,                 # Cutoff frequency for low-pass filter (Hz)
    high_pass=200,                 # Cutoff frequency for high-pass filter (Hz)
    normalize=0.95,                # Output volume normalization (0.0-1.0)
    compression_ratio=1.5,         # Dynamic range compression ratio
    compression_threshold=0.5,      # Compression threshold (0.0-1.0)
    window_size=12,                # Resampling window size (smaller = faster)
    anti_aliasing_ratio=0.95,      # Anti-aliasing filter cutoff ratio
    anti_aliasing_type=AA_BUTTERWORTH,  # Filter type
    filter_order=4,                # Filter order (2-6)
    chebyshev_ripple=0.1          # Ripple for Chebyshev filter (dB)
)
```

## Optimization Profiles

The library comes with several pre-configured optimization profiles:

1. **Default parameters (optimized for TTS)**:
   - Input sample rate: auto-detect
   - Force mono: true
   - Low-pass filter: 3400 Hz (telephone bandwidth)
   - High-pass filter: 200 Hz (remove noise)
   - Normalization: 0.95 (prevent clipping)
   - Compression ratio: 1.5 (light compression)
   - Compression threshold: 0.5
   - Window size: 12 (good balance)
   - Anti-aliasing ratio: 0.95
   - Filter type: Butterworth (best frequency response)
   - Filter order: 4 (good quality)

2. **Fast mode**:
   - Input sample rate: auto-detect
   - Force mono: true
   - Low-pass filter: 3400 Hz
   - High-pass filter: 200 Hz
   - Normalization: 0.95
   - Compression ratio: 1.2
   - Compression threshold: 0.5
   - Window size: 8 (faster processing)
   - Anti-aliasing ratio: 0.8 (faster)
   - Filter type: Simple (fastest)
   - Filter order: 2

3. **High quality (Bessel)**:
   - Input sample rate: auto-detect
   - Force mono: true
   - Low-pass filter: 3400 Hz
   - High-pass filter: 200 Hz
   - Normalization: 0.98
   - Compression ratio: 1.3
   - Compression threshold: 0.6
   - Window size: 16 (better quality)
   - Anti-aliasing ratio: 0.98
   - Filter type: Bessel (better shape)
   - Filter order: 4

4. **Maximum quality (Butterworth)**:
   - Input sample rate: auto-detect
   - Force mono: true
   - Low-pass filter: 3400 Hz
   - High-pass filter: 200 Hz
   - Normalization: 0.99
   - Compression ratio: 1.2
   - Compression threshold: 0.7
   - Window size: 24 (best quality)
   - Anti-aliasing ratio: 0.99
   - Filter type: Butterworth (best response)
   - Filter order: 6

## Performance

The tool is highly optimized for both quality and speed:
- Written in Go for efficient processing
- Uses optimized DSP algorithms with precomputed tables:
  - Cached sinc function values for resampling
  - Optimized window function calculations
  - Fast linear interpolation for table lookups
- Processes audio in chunks to minimize memory usage
- Supports float and PCM WAV formats
- Typical processing time: ~200ms for 30 seconds of audio (2.3x faster than v1.0.0)
- Memory efficient: ~7MB peak memory usage
- Automatic handling of various input formats and sample rates

### Performance Comparison

| Version | Processing Time (30s audio) | Memory Usage | Allocations |
|---------|---------------------------|--------------|-------------|
| v1.0.0  | ~400ms                   | 7.15 MB      | 70 allocs   |
| v1.1.0  | ~170ms                   | 7.11 MB      | 54 allocs   |

### Optimization Details

1. **Resampling Optimizations**:
   - Precomputed sinc function lookup tables
   - Cached tables for different window sizes
   - Linear interpolation for fast value lookups
   - Reduced memory allocations

2. **Filter Processing**:
   - Optimized filter coefficient calculations
   - Efficient sample processing loops
   - Improved numerical stability

3. **Memory Management**:
   - Reduced temporary allocations
   - Reuse of existing buffers
   - Optimized data structures

## Common Issues and Solutions

1. **Distorted Audio**: If you experience audio distortion, try:
   - Specifying the correct input sample rate
   - Enabling mono conversion for multi-channel audio
   - Adjusting the anti-aliasing filter settings
   - Reducing the normalization level

2. **Incorrect Duration**: If the output duration is wrong:
   - Make sure to specify the correct input sample rate
   - Check if the input WAV file is properly formatted
   - Try different resampling window sizes

3. **Performance Issues**: If processing is too slow:
   - Use a smaller window size (12 is recommended)
   - Use the Simple anti-aliasing filter
   - Reduce the anti-aliasing ratio
   - Use the Fast mode profile

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

skypro1111@gmail.com 