package transfer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"secure-transfer/internal/crypto"
)

// Setup a mock server/client for testing
func setupTestServerClient(t *testing.T) (int, []byte) {
	// Find an available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Generate a key for testing
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i % 256)
	}

	return port, key
}

func TestSendReceiveIntegration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port, key := setupTestServerClient(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a test file
	testContent := []byte("This is a test file content for integration testing")
	testFile := "test_send.txt"
	receivedFile := "test_receive.txt"

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)
	defer os.Remove(receivedFile)

	// Start server in a goroutine with context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Track errors from goroutine
	var serverErr error

	go func() {
		defer wg.Done()

		// Setup server with a small timeout so test doesn't hang forever
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			serverErr = err
			return
		}
		defer listener.Close()

		// Accept with timeout
		listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second))

		// Wait for connection or cancellation
		connChan := make(chan net.Conn, 1)
		errChan := make(chan error, 1)

		go func() {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			}
			connChan <- conn
		}()

		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			serverErr = err
			return
		case conn := <-connChan:
			defer conn.Close()

			// Read file size
			sizeBuf := make([]byte, 8)
			_, err = io.ReadFull(conn, sizeBuf)
			if err != nil {
				serverErr = err
				return
			}

			// Read encrypted data
			var buf bytes.Buffer
			_, err = io.Copy(&buf, conn)
			if err != nil {
				serverErr = err
				return
			}

			// Decrypt
			decrypted, err := crypto.Decrypt(buf.Bytes(), key)
			if err != nil {
				serverErr = err
				return
			}

			// Save
			err = os.WriteFile(receivedFile, decrypted, 0644)
			if err != nil {
				serverErr = err
				return
			}
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Send the file
	err = SendFile("localhost", port, testFile, key, logger)
	if err != nil {
		t.Fatalf("Failed to send file: %v", err)
	}

	// Wait for server to finish
	wg.Wait()

	// Check server error
	if serverErr != nil {
		t.Fatalf("Server error: %v", serverErr)
	}

	// Verify received file
	received, err := os.ReadFile(receivedFile)
	if err != nil {
		t.Fatalf("Failed to read received file: %v", err)
	}

	if !bytes.Equal(received, testContent) {
		t.Errorf("Received content doesn't match original. Got %v, want %v",
			received, testContent)
	}
}
