import subprocess
import tempfile
import os

def play_ulaw(ulaw_bytes, sample_rate=8000, window_size=32):
    """Play u-law audio bytes on macOS"""
    # Create a temporary file for u-law data
    with tempfile.NamedTemporaryFile(suffix='.ulaw', delete=False) as temp_ulaw:
        temp_ulaw.write(ulaw_bytes)
        temp_ulaw_path = temp_ulaw.name

    # Create a temporary file for WAV output
    temp_wav = tempfile.NamedTemporaryFile(suffix='.wav', delete=False)
    temp_wav_path = temp_wav.name
    temp_wav.close()

    try:
        # Convert u-law to WAV using our wav2ulaw utility with new interface
        subprocess.run([
            './wav2ulaw',
            '--input', temp_ulaw_path,
            '--output', temp_wav_path,
            '--mode', 'ulaw2wav',
            '--sample-rate', str(sample_rate),
            '--window-size', str(window_size)
        ], check=True)

        # Play the WAV file using sox (play) with explicit sample rate
        # If sox is not installed, you can install it with: brew install sox
        subprocess.run([
            'play', 
            '-r', str(sample_rate),  # input sample rate
            '-b', '16',              # bit depth
            '-c', '1',               # number of channels (mono)
            '-e', 'signed',          # signed PCM
            temp_wav_path
        ], check=True)
    except FileNotFoundError:
        # If sox is not found, try using afplay
        print("Warning: sox not found. Using afplay instead. For better playback, install sox: brew install sox")
        subprocess.run(['afplay', temp_wav_path], check=True)
    finally:
        # Clean up the temporary files
        os.unlink(temp_wav_path)
        os.unlink(temp_ulaw_path)

# Example usage:
if __name__ == "__main__":
    # Read u-law file
    with open('audio.ulaw', 'rb') as f:
        ulaw_data = f.read()
    
    # Play audio with improved resampling quality
    play_ulaw(ulaw_data, sample_rate=44100, window_size=32) 