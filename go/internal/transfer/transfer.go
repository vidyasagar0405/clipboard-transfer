package transfer

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"secure-transfer/internal/clipboard"
	"secure-transfer/internal/crypto"
)

// SendFile sends a file over TCP
func SendFile(ip string, port int, filePath string, key []byte, logger *slog.Logger) error {
	logger.Info("Sending file", "ip", ip, "port", port, "file", filePath)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer conn.Close()

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	encryptedData, err := crypto.Encrypt(fileData, key)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	// Send file size first
	size := make([]byte, 8)
	copy(size, fmt.Sprintf("%08d", len(encryptedData)))
	_, err = conn.Write(size)
	if err != nil {
		return fmt.Errorf("error sending file size: %w", err)
	}

	// Send encrypted file
	_, err = conn.Write(encryptedData)
	if err != nil {
		return fmt.Errorf("error sending file data: %w", err)
	}

	logger.Info("File sent successfully!")
	return nil
}

// ReceiveFile receives a file over TCP
func ReceiveFile(port int, saveAs string, key []byte, logger *slog.Logger) error {
	logger.Info("Starting file receiver", "port", port, "saveAs", saveAs)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	defer listener.Close()

	logger.Info("Waiting for connection", "port", port)
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer conn.Close()

	logger.Info("Connection established", "from", conn.RemoteAddr())

	// Read file size first
	sizeBuf := make([]byte, 8)
	_, err = io.ReadFull(conn, sizeBuf)
	if err != nil {
		return fmt.Errorf("error reading file size: %w", err)
	}

	// Read encrypted file data
	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return fmt.Errorf("error receiving file: %w", err)
	}

	decryptedData, err := crypto.Decrypt(buf.Bytes(), key)
	if err != nil {
		return fmt.Errorf("decryption error: %w", err)
	}

	err = os.WriteFile(saveAs, decryptedData, 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %w", err)
	}

	// Try to copy the content to clipboard
	if len(decryptedData) < 1024*1024 { // Only copy if less than 1MB
		err = clipboard.CopyToClipboard(string(decryptedData))
		if err != nil {
			logger.Warn("Could not copy to clipboard", "error", err)
		} else {
			logger.Info("Copied file content to clipboard")
		}
	} else {
		logger.Info("File too large to copy to clipboard")
	}

	logger.Info("File received and saved", "filename", saveAs)
	return nil
}

// EchoResponse starts an echo server
func EchoResponse(port int, key []byte, logger *slog.Logger) error {
	logger.Info("Starting echo server", "port", port)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	defer listener.Close()

	logger.Info("Waiting for connection", "port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Connection error", "error", err)
			continue
		}

		go handleEchoConnection(conn, key, logger)
	}
}

// handleEchoConnection handles a single echo connection
func handleEchoConnection(conn net.Conn, key []byte, logger *slog.Logger) {
	defer conn.Close()
	logger.Info("Connection established", "from", conn.RemoteAddr())

	// Read file size first
	sizeBuf := make([]byte, 8)
	_, err := io.ReadFull(conn, sizeBuf)
	if err != nil {
		logger.Error("Error reading message size", "error", err)
		return
	}

	// Read encrypted data
	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		logger.Error("Error receiving data", "error", err)
		return
	}

	decryptedData, err := crypto.Decrypt(buf.Bytes(), key)
	if err != nil {
		logger.Error("Decryption error", "error", err)
		return
	}

	message := string(decryptedData)
	logger.Info("Received message", "length", len(message))

	// Copy to clipboard
	err = clipboard.CopyToClipboard(message)
	if err != nil {
		logger.Warn("Could not copy to clipboard", "error", err)
	} else {
		logger.Info("Copied message to clipboard")
	}

	// Send response back
	response := fmt.Sprintf("Received message (%d bytes)", len(message))
	encryptedResponse, err := crypto.Encrypt([]byte(response), key)
	if err != nil {
		logger.Error("Encryption error", "error", err)
		return
	}

	// Send response size
	respSize := make([]byte, 8)
	copy(respSize, fmt.Sprintf("%08d", len(encryptedResponse)))
	conn.Write(respSize)

	// Send encrypted response
	conn.Write(encryptedResponse)
	logger.Info("Sent response to client")
}

// SendMessage sends a message to the echo server
func SendMessage(ip string, port int, filePath string, message string, key []byte, logger *slog.Logger) error {
	logger.Info("Sending message", "ip", ip, "port", port)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer conn.Close()

	var messageData []byte
	if filePath != "" {
		// Read from file
		messageData, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
	} else {
		// Use provided message
		messageData = []byte(message)
	}

	encryptedData, err := crypto.Encrypt(messageData, key)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	// Send message size first
	size := make([]byte, 8)
	copy(size, fmt.Sprintf("%08d", len(encryptedData)))
	_, err = conn.Write(size)
	if err != nil {
		return fmt.Errorf("error sending message size: %w", err)
	}

	// Send encrypted message
	_, err = conn.Write(encryptedData)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	logger.Info("Message sent", "bytes", len(messageData))

	// Read response size
	respSizeBuf := make([]byte, 8)
	_, err = io.ReadFull(conn, respSizeBuf)
	if err != nil {
		return fmt.Errorf("error reading response size: %w", err)
	}

	// Read encrypted response
	var respBuf bytes.Buffer
	_, err = io.Copy(&respBuf, conn)
	if err != nil {
		return fmt.Errorf("error receiving response: %w", err)
	}

	decryptedResp, err := crypto.Decrypt(respBuf.Bytes(), key)
	if err != nil {
		return fmt.Errorf("response decryption error: %w", err)
	}

	logger.Info("Response received", "message", string(decryptedResp))
	return nil
}
