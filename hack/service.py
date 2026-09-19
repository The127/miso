#!/usr/bin/env python3
# A service beyond the builder that answers miso to every connection, on an
# IPv6 address the harness gives its own network, where QEMU reaches it as
# it would reach the internet.
# Usage: service.py <address> <port>
import socket
import sys

server = socket.create_server((sys.argv[1], int(sys.argv[2])), family=socket.AF_INET6)
while True:
    connection, _ = server.accept()
    connection.sendall(b"miso\n")
    connection.close()
