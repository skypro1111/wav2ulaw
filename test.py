import subprocess
import tempfile
import os
import time
import soundfile as sf
from play_ulaw import play_ulaw

# Anti-aliasing filter types
AA_SIMPLE = 0      # Simple low-pass filter
AA_BUTTERWORTH = 1 # Butterworth filter (flattest frequency response)
AA_BESSEL = 2      # Bessel filter (best signal shape preservation)
AA_CHEBYSHEV = 3   # Chebyshev Type I filter (steepest roll-off)

def get_wav_info(wav_bytes):
    """
    Get WAV file information from bytes
    
    Returns:
    --------
    tuple: (sample_rate, channels, bit_depth)
    """
    with tempfile.NamedTemporaryFile(suffix='.wav', delete=False) as temp_wav:
        temp_wav.write(wav_bytes)
        temp_wav_path = temp_wav.name
    
    try:
        info = sf.info(temp_wav_path)
        return info.samplerate, info.channels, info.subtype
    finally:
        os.unlink(temp_wav_path)

# Optimal default parameters based on testing
DEFAULT_CONFIG = {
    'input_sample_rate': 0,     # Auto-detect
    'force_mono': True,         # Convert to mono by default
    'low_pass': 3400,          # Telephone bandwidth
    'high_pass': 200,          # Remove noise
    'normalize': 0.95,         # Prevent clipping
    'compression_ratio': 1.5,   # Light compression
    'compression_threshold': 0.5,
    'window_size': 12,         # Good balance
    'anti_aliasing_ratio': 0.95,  # Good quality
    'anti_aliasing_type': AA_BUTTERWORTH,  # Best frequency response
    'filter_order': 4,         # Good quality
    'chebyshev_ripple': 0.1    # Minimal ripple
}

def benchmark_conversion(wav_bytes, **kwargs):
    """
    Measures WAV to u-law conversion time with given parameters
    
    Returns:
    --------
    tuple: (ulaw_bytes, elapsed_time_sec, params)
    """
    start_time = time.time()
    ulaw_data = wav_to_ulaw(wav_bytes, **kwargs)
    elapsed_time = time.time() - start_time  # in seconds
    return ulaw_data, elapsed_time, kwargs

def wav_to_ulaw(wav_bytes, input_sample_rate=0, force_mono=True,
                low_pass=3400, high_pass=200, normalize=0.95,
                compression_ratio=1.5, compression_threshold=0.5,
                window_size=64, anti_aliasing_ratio=0.95,
                anti_aliasing_type=AA_SIMPLE, filter_order=2,
                chebyshev_ripple=0.1):
    """
    Convert WAV bytes to u-law format with advanced audio processing
    
    Parameters:
    -----------
    wav_bytes : bytes
        Input WAV file bytes
    input_sample_rate : int
        Input sample rate (Hz). Set to 0 for auto-detect
    force_mono : bool
        Convert multi-channel audio to mono
    low_pass : float
        Low-pass filter cutoff frequency (Hz). Recommended 3400 Hz (telephone bandwidth)
    high_pass : float
        High-pass filter cutoff frequency (Hz). Recommended 200 Hz
    normalize : float
        Normalization level (0.0-1.0). Recommended 0.95
    compression_ratio : float
        Compression ratio. 1.0 = no compression. Recommended 1.5
    compression_threshold : float
        Compression threshold (0.0-1.0). Recommended 0.5
    window_size : int
        Resampling window size. Recommended 64
    anti_aliasing_ratio : float
        Anti-aliasing filter ratio (0.0-1.0). Recommended 0.95
    anti_aliasing_type : int
        Filter type (AA_SIMPLE=0, AA_BUTTERWORTH=1, AA_BESSEL=2, AA_CHEBYSHEV=3)
    filter_order : int
        Filter order (2-6). Recommended 2
    chebyshev_ripple : float
        Chebyshev filter ripple (dB). Recommended 0.1
    """
    
    # Create a temporary file for WAV input
    with tempfile.NamedTemporaryFile(suffix='.wav', delete=False) as temp_wav:
        temp_wav.write(wav_bytes)
        temp_wav_path = temp_wav.name

    # Create a temporary file for normalized PCM WAV
    temp_pcm = tempfile.NamedTemporaryFile(suffix='.wav', delete=False)
    temp_pcm_path = temp_pcm.name
    temp_pcm.close()

    # Create a temporary file for u-law output
    temp_ulaw = tempfile.NamedTemporaryFile(suffix='.ulaw', delete=False)
    temp_ulaw_path = temp_ulaw.name
    temp_ulaw.close()
    
    try:
        # Get WAV info if needed
        if input_sample_rate == 0:
            detected_rate, _, _ = get_wav_info(wav_bytes)
            input_sample_rate = detected_rate

        # Read and normalize the audio data
        data, samplerate = sf.read(temp_wav_path)
        
        # Convert to mono if needed
        if force_mono and len(data.shape) > 1:
            data = data.mean(axis=1)
        
        # Normalize float to int16 range
        data = data * 32767
        data = data.astype('int16')
        
        # Write as PCM WAV
        sf.write(temp_pcm_path, data, samplerate, subtype='PCM_16')
        
        # Convert WAV to u-law using the Go program
        cmd = [
            './wav2ulaw',
            '--input', temp_pcm_path,
            '--output', temp_ulaw_path,
            '--mode', 'wav2ulaw',
            '--sample-rate', str(input_sample_rate),
            '--low-pass', str(low_pass),
            '--high-pass', str(high_pass),
            '--normalize', str(normalize),
            '--compress-ratio', str(compression_ratio),
            '--compress-threshold', str(compression_threshold),
            '--window-size', str(window_size),
            '--anti-aliasing-ratio', str(anti_aliasing_ratio),
            '--anti-aliasing-type', str(anti_aliasing_type),
            '--filter-order', str(filter_order)
        ]
        
        # Add Chebyshev ripple parameter if using Chebyshev filter
        if anti_aliasing_type == AA_CHEBYSHEV:
            cmd.extend(['--chebyshev-ripple', str(chebyshev_ripple)])
            
        subprocess.run(cmd, check=True)
        
        # Read the output u-law data
        with open(temp_ulaw_path, 'rb') as f:
            ulaw_data = f.read()
            
        return ulaw_data
        
    finally:
        # Clean up temporary files
        os.unlink(temp_wav_path)
        os.unlink(temp_pcm_path)
        os.unlink(temp_ulaw_path)

def format_params(params):
    """Formats parameters for output"""
    return f"""
    Parameters:
    - Input: rate={params.get('input_sample_rate', 'auto')}Hz, mono={params.get('force_mono', True)}
    - Filters: LPF={params['low_pass']}Hz, HPF={params['high_pass']}Hz
    - Volume: normalize={params['normalize']}, ratio={params['compression_ratio']}, threshold={params['compression_threshold']}
    - Resampling: window={params['window_size']}, aa_ratio={params['anti_aliasing_ratio']}
    - Filter: type={params['anti_aliasing_type']}, order={params['filter_order']}{', ripple=' + str(params['chebyshev_ripple']) if params['anti_aliasing_type'] == AA_CHEBYSHEV else ''}
    """

# Test parameter sets
TEST_CONFIGS = [
    {
        'name': 'Default parameters (optimized for TTS)',
        'params': DEFAULT_CONFIG.copy()
    },
    {
        'name': 'Fast mode',
        'params': {
            'input_sample_rate': 0,  # Auto-detect
            'force_mono': True,
            'low_pass': 3400,
            'high_pass': 200,
            'normalize': 0.95,
            'compression_ratio': 1.2,
            'compression_threshold': 0.5,
            'window_size': 8,        # Faster processing
            'anti_aliasing_ratio': 0.8,  # Faster
            'anti_aliasing_type': AA_SIMPLE,  # Fastest
            'filter_order': 2,
            'chebyshev_ripple': 0.1
        }
    },
    {
        'name': 'High quality (Bessel)',
        'params': {
            'input_sample_rate': 0,  # Auto-detect
            'force_mono': True,
            'low_pass': 3400,
            'high_pass': 200,
            'normalize': 0.98,
            'compression_ratio': 1.3,
            'compression_threshold': 0.6,
            'window_size': 16,       # Better quality
            'anti_aliasing_ratio': 0.98,
            'anti_aliasing_type': AA_BESSEL,  # Better shape
            'filter_order': 4,
            'chebyshev_ripple': 0.1
        }
    },
    {
        'name': 'Maximum quality (Butterworth)',
        'params': {
            'input_sample_rate': 0,  # Auto-detect
            'force_mono': True,
            'low_pass': 3400,
            'high_pass': 200,
            'normalize': 0.99,
            'compression_ratio': 1.2,
            'compression_threshold': 0.7,
            'window_size': 24,       # Best quality
            'anti_aliasing_ratio': 0.99,
            'anti_aliasing_type': AA_BUTTERWORTH,  # Best response
            'filter_order': 6,
            'chebyshev_ripple': 0.1
        }
    }
]

if __name__ == "__main__":
    # Read input WAV file
    with open('input.wav', 'rb') as f:
        wav_data = f.read()

    # Get WAV info
    sample_rate, channels, bit_depth = get_wav_info(wav_data)
    print("\n=== Input file info ===")
    print(f"Sample rate: {sample_rate} Hz")
    print(f"Channels: {channels}")
    print(f"Bit depth: {bit_depth} bits")
    print(f"File size: {len(wav_data)/1024:.1f} KB")

    print("\n=== Testing conversion speed ===")

    results = []
    for config in TEST_CONFIGS:
        print(f"\nTest: {config['name']}")
        ulaw_data, elapsed_time, params = benchmark_conversion(wav_data, **config['params'])
        
        # Update params with detected sample rate if auto-detect was used
        if params['input_sample_rate'] == 0:
            params['input_sample_rate'] = sample_rate
            
        results.append({
            'name': config['name'],
            'time': elapsed_time,
            'size': len(ulaw_data),
            'params': params
        })
        print(f"Conversion time: {elapsed_time:.3f} sec")
        print(f"Result size: {len(ulaw_data)/1024:.1f} KB")
        print(format_params(params))

    print("\n=== Results comparison ===")
    base_time = results[0]['time']
    for r in results:
        speedup = base_time / r['time'] if r['time'] > 0 else 0
        print(f"\n{r['name']}:")
        print(f"Time: {r['time']:.3f} sec ({speedup:.2f}x relative to baseline)")
        print(f"Size: {r['size']/1024:.1f} KB")
        print(f"Compression ratio: {len(wav_data)/r['size']:.1f}x")

    print("\n=== Playing results ===")
    for r in results:
        print(f"\n{r['name']}:")
        play_ulaw(ulaw_data, sample_rate=44100, window_size=r['params']['window_size'])