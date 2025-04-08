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

## Features

- High-quality audio processing pipeline:
  - Support for float and PCM WAV formats
  - Configurable anti-aliasing filters (Simple, Butterworth, Bessel, Chebyshev)
  - Telephone bandwidth optimization (200-3400 Hz)
  - Automatic volume normalization
  - Optional dynamic range compression
  - High-quality resampling with precomputed tables
  - Multi-channel to mono conversion
  - Support for various input sample rates (8kHz-48kHz)
- Fast Go implementation with Python bindings
- Simple command-line interface
- Easy integration with Python TTS systems

## Python Usage Example

```python
import subprocess
import tempfile
import os
import soundfile as sf
from play_ulaw import play_ulaw

# Anti-aliasing filter types
AA_SIMPLE = 0      # Simple low-pass filter
AA_BUTTERWORTH = 1 # Butterworth filter (flattest frequency response)
AA_BESSEL = 2      # Bessel filter (best signal shape preservation)
AA_CHEBYSHEV = 3   # Chebyshev Type I filter (steepest roll-off)

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

# Optional: Play the result
play_ulaw(ulaw_data, sample_rate=44100, window_size=12)
```

## Optimization Profiles

The library comes with several pre-configured optimization profiles:

1. **Default parameters (optimized for TTS)**:
   - Low-pass filter: 3400 Hz (telephone bandwidth)
   - High-pass filter: 200 Hz (remove noise)
   - Normalization: 0.95 (prevent clipping)
   - Compression ratio: 1.5 (light compression)
   - Window size: 12 (good balance)
   - Filter type: Butterworth (best frequency response)
   - Filter order: 4 (good quality)

2. **Fast mode**:
   - Window size: 8 (faster processing)
   - Filter type: Simple (fastest)
   - Filter order: 2
   - Compression ratio: 1.2

3. **Maximum quality**:
   - Window size: 24 (best quality)
   - Filter type: Butterworth (best response)
   - Filter order: 6
   - Normalization: 0.99

## Performance

The tool is highly optimized for both quality and speed:
- Uses optimized DSP algorithms with precomputed tables:
  - Cached sinc function values for resampling
  - Fast linear interpolation for table lookups
- Typical processing time: ~200ms for 30 seconds of audio
- Memory efficient: ~7MB peak memory usage

### Performance Comparison

| Version | Processing Time (30s audio) | Memory Usage | Allocations |
|---------|---------------------------|--------------|-------------|
| v1.0.0  | ~400ms                   | 7.15 MB      | 70 allocs   |
| v1.1.0  | ~170ms                   | 7.11 MB      | 54 allocs   |

## Common Issues and Solutions

1. **Distorted Audio**: If you experience audio distortion, try:
   - Specifying the correct input sample rate
   - Enabling mono conversion for multi-channel audio
   - Reducing the normalization level

2. **Performance Issues**: If processing is too slow:
   - Use a smaller window size (8-12)
   - Use the Simple anti-aliasing filter
   - Use the Fast mode profile

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

skypro1111@gmail.com 