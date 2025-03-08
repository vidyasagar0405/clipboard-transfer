#!/usr/bin/env python3

import socket
import argparse
import sys
import base64
import os
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend

# AES Key (Must be the same on both client and server)
SECRET_KEY = os.getenv("TRANSFER_SECRET_KEY", "").encode()
if len(SECRET_KEY) not in [16, 24, 32]:
    raise ValueError("Error: SECRET_KEY must be 16, 24, or 32 bytes long.")

def encrypt_message(message, key):
    """Encrypt the message using AES."""
    iv = os.urandom(16)  # Generate a random IV
    cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
    encryptor = cipher.encryptor()

    # Padding to ensure block size compatibility
    padded_message = message + ' ' * (16 - len(message) % 16)

    encrypted = encryptor.update(padded_message.encode()) + encryptor.finalize()
    return base64.b64encode(iv + encrypted).decode()  # Encode to Base64

def decrypt_message(encrypted_message, key):
    """Decrypt the received encrypted message."""
    data = base64.b64decode(encrypted_message)
    iv, encrypted = data[:16], data[16:]

    cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
    decryptor = cipher.decryptor()

    decrypted = decryptor.update(encrypted) + decryptor.finalize()
    return decrypted.decode().strip()  # Remove padding

def run_server(host, port):
    """Start a secure TCP server to receive encrypted text data."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        try:
            s.bind((host, port))
        except Exception as e:
            print(f"Bind failed: {e}")
            sys.exit(1)
        s.listen(1)
        print(f"Secure server listening on {host}:{port} ...")
        conn, addr = s.accept()
        with conn:
            print(f"Connected by {addr}")
            data = b""
            while True:
                packet = conn.recv(1024)
                if not packet:
                    break
                data += packet

            decrypted_message = decrypt_message(data.decode(), SECRET_KEY)
            print("Decrypted message received:")
            print(decrypted_message)

def run_client(host, port, message):
    """Connect to the secure TCP server and send encrypted text data."""
    encrypted_message = encrypt_message(message, SECRET_KEY)

    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        try:
            s.connect((host, port))
        except Exception as e:
            print(f"Connection failed: {e}")
            sys.exit(1)
        s.sendall(encrypted_message.encode())
        print("Encrypted message sent successfully!")

def main():
    parser = argparse.ArgumentParser(description="Send encrypted text data over TCP.")
    parser.add_argument("--mode", choices=["server", "client"], required=True,
                        help="Mode: 'server' to receive data, 'client' to send data.")
    parser.add_argument("--host", default="0.0.0.0",
                        help="Host to bind (server) or to connect to (client). For client mode, provide the server's IP address.")
    parser.add_argument("--port", type=int, default=5000,
                        help="Port number to bind/connect (default is 5000).")
    parser.add_argument("--message", default="Hello securely!",
                        help="The message to send (only for client mode).")
    args = parser.parse_args()

    if args.mode == "server":
        run_server(args.host, args.port)
    else:
        run_client(args.host, args.port, args.message)

if __name__ == "__main__":
    main()

