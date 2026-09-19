#!/usr/bin/env python3
# A service on the host's own loopback, which a run must never reach. It
# prints the port it got and answers loopback to every connection.
import socket

server = socket.create_server(("127.0.0.1", 0))
print(server.getsockname()[1], flush=True)
while True:
    connection, _ = server.accept()
    connection.sendall(b"loopback\n")
    connection.close()
