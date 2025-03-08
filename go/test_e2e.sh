#!/bin/bash
set -e

# Build the application
echo "Building application..."
go build -o secure-transfer

# Generate test file
echo "Creating test file..."
dd if=/dev/urandom of=test_file.bin bs=1K count=256

# Find available port
PORT=$(python -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')
echo "Using port: $PORT"

# Start server in background
echo "Starting server..."
./secure-transfer server --port "$PORT" --save received_file.bin &
SERVER_PID=$!

# Give server time to start
sleep 1

# Send file
echo "Sending file..."
./secure-transfer client --ip localhost --port "$PORT" --file test_file.bin

# Give time for completion
sleep 1

# Kill server
echo "Stopping server..."
kill $SERVER_PID || true

# Verify files match
echo "Verifying file integrity..."
if cmp -s test_file.bin received_file.bin; then
    echo "Success! Files match."
else
    echo "ERROR: Files don't match."
    exit 1
fi

# Test echo functionality
echo "Testing echo functionality..."
./secure-transfer echo --port "$PORT" &
ECHO_PID=$!

# Give server time to start
sleep 1

# Message to send
TEST_MESSAGE="This is a test message for the echo server functionality"

# Send message
echo "Sending message..."
./secure-transfer client --ip localhost --port "$PORT" --message "$TEST_MESSAGE"

# Kill echo server
echo "Stopping echo server..."
kill $ECHO_PID || true

# Clean up
echo "Cleaning up..."
rm -f test_file.bin received_file.bin secure-transfer

echo "All tests completed successfully!"
