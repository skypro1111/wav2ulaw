#!/bin/bash

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Please provide version number (e.g. v1.0.0)"
    exit 1
fi

# Create release directory
mkdir -p releases

# Build for different platforms
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

for PLATFORM in "${PLATFORMS[@]}"; do
    # Split platform into OS and ARCH
    IFS='/' read -r -a array <<< "$PLATFORM"
    OS="${array[0]}"
    ARCH="${array[1]}"
    
    echo "Building for $OS/$ARCH..."
    
    # Set environment variables for cross-compilation
    export GOOS=$OS
    export GOARCH=$ARCH
    
    # Set output binary name
    if [ "$OS" = "windows" ]; then
        OUTPUT="releases/wav2ulaw_${VERSION}_${OS}_${ARCH}.exe"
    else
        OUTPUT="releases/wav2ulaw_${VERSION}_${OS}_${ARCH}"
    fi
    
    # Build
    go build -o "$OUTPUT" ./cmd/wav2ulaw
    
    # Create zip archive
    if [ "$OS" = "windows" ]; then
        zip "releases/wav2ulaw_${VERSION}_${OS}_${ARCH}.zip" "$OUTPUT" README.md LICENSE
    else
        tar -czf "releases/wav2ulaw_${VERSION}_${OS}_${ARCH}.tar.gz" "$OUTPUT" README.md LICENSE
    fi
done

echo "Build complete! Release files are in the releases directory" 