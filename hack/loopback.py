#!/usr/bin/env python3
# A service on the host's own loopback, IPv4 and IPv6 on one port, which a
# run must never reach. It prints the port it got and answers loopback to
# every connection.
import selectors
import socket

ipv4 = socket.create_server(("127.0.0.1", 0))
port = ipv4.getsockname()[1]
ipv6 = socket.create_server(("::1", port), family=socket.AF_INET6)
print(port, flush=True)
waiting = selectors.DefaultSelector()
waiting.register(ipv4, selectors.EVENT_READ, ipv4)
waiting.register(ipv6, selectors.EVENT_READ, ipv6)
while True:
    for ready, _ in waiting.select():
        connection, _ = ready.data.accept()
        connection.sendall(b"loopback\n")
        connection.close()
