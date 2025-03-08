import contextlib
import signal
import socket
import struct
import threading
import queue
from datetime import datetime
from time import sleep
import time
from typing import Any, Dict, List, Optional
import asyncio
import uuid

UUID_LENGTH = 16  # Standard UUID string length


class TCPLogServer:
    def __init__(
        self, host: str = "localhost", port: int = 9999, max_connections: int = 10
    ):
        self.host = host
        self.port = port
        self.max_connections = max_connections
        self.log_queue = asyncio.Queue()  # Use asyncio's Queue for async operations
        self.running = False
        self.server_socket = None
        self.clients = []

    async def start(self) -> bool:
        """Start the server with graceful error handling."""
        if self.running:
            print("Server is already running")
            return True

        try:
            # Create the server socket and set up for async operation
            self.server_socket = await asyncio.start_server(
                self._handle_client_connection, self.host, self.port
            )

            self.running = True
            print(f"TCP Log Server listening on {self.host}:{self.port}")

            # Start the main server loop to accept incoming connections
            await self._accept_connections()

            return True

        except Exception as e:
            print(f"Error starting server: {e}")
            # self._cleanup_resources()
            return False

    async def _accept_connections(self):
        """Accept client connections and manage client lifecycle."""
        async with self.server_socket:
            # The server will continuously listen for new connections
            await self.server_socket.serve_forever()

    async def _handle_client_connection(
        self, reader: asyncio.StreamReader, writer: asyncio.StreamWriter
    ):
        """Handle an incoming client connection asynchronously."""
        client_addr = writer.get_extra_info("peername")
        print(f"New connection from {client_addr}")

        try:
            while self.running:
                # 1. Read 4 bytes for the message length
                length_bytes = await self._read_exact_bytes(reader, 4)
                if not length_bytes:
                    print(f"Client {client_addr} disconnected while reading length")
                    return
                msg_length = struct.unpack("!I", length_bytes)[0]
                print(f"Message length: {msg_length}")

                # 2. Read the full message body based on the length
                message_body = await self._read_exact_bytes(reader, msg_length)
                if not message_body:
                    print(f"Client {client_addr} disconnected during message body")
                    return

                # 3. Extract the client_id and compressed_data from message_body
                client_id_bytes = message_body[:UUID_LENGTH]
                compressed_data = message_body[UUID_LENGTH:]
                client_id = str(uuid.UUID(bytes=client_id_bytes))

                print("Client ID:", client_id)
                print("Compressed Data:", compressed_data)

                # Queue the complete message
                log_data = {
                    "timestamp": int(datetime.now().timestamp()),
                    "client_id": client_id,
                    "message": message_body,  # includes both client_id and compressed_data
                    "process_id": client_id,
                }

                await self.log_queue.put(log_data)
                print(f"Received message from {client_addr}")

        except asyncio.CancelledError:
            print(f"Client {client_addr} connection cancelled")
        except Exception as e:
            print(f"Error handling client {client_addr}: {e}")
        finally:
            # Cleanup client on disconnection
            writer.close()
            await writer.wait_closed()

    async def _read_exact_bytes(self, reader, num_bytes: int) -> Optional[bytes]:
        """Helper function to read an exact number of bytes."""
        data = b""
        while len(data) < num_bytes:
            chunk = await reader.read(num_bytes - len(data))
            if not chunk:
                return None  # Connection closed
            data += chunk
        return data

    async def stop(self) -> bool:
        """Stop the server gracefully."""
        if not self.running:
            return True

        print("Stopping TCP Log Server...")
        self.running = False

        # Close the server socket
        if self.server_socket:
            self.server_socket.close()

        # Wait until all clients finish processing their current operations
        await asyncio.gather(*[client["task"] for client in self.clients])

        print("TCP Log Server shutdown complete")
        return True

    def get_log(self, timeout: float = 0.001) -> Optional[dict]:
        """Get the next log message from the queue."""
        try:
            return self.log_queue.get_nowait()
        except asyncio.QueueEmpty:
            return None

    async def drain_logs(self, timeout: float = 1.0) -> List[Dict[str, Any]]:
        """Drain and return all logs from the queue."""
        logs = []
        end_time = time.time() + timeout

        while time.time() < end_time:
            log = await self.get_log(timeout=0.1)
            if log:
                logs.append(log)

        return logs

    async def __aenter__(self):
        """Support for async context manager."""
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Support for async context manager."""
        await self.stop()
