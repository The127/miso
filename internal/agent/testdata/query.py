# Asks a nameserver for a name and prints the address it answers, and
# nothing at all when no answer comes.
# Usage: query.py <nameserver> <name> <A|AAAA> [udp|tcp]
import socket
import struct
import sys

nameserver = sys.argv[1]
kind = 28 if sys.argv[3] == "AAAA" else 1
stream = len(sys.argv) > 4 and sys.argv[4] == "tcp"
labels = b"".join(bytes([len(p)]) + p.encode() for p in sys.argv[2].split("."))
query = struct.pack("!HHHHHH", 1, 0x0100, 1, 0, 0, 0) + labels + b"\0" + struct.pack("!HH", kind, 1)

family = socket.AF_INET6 if ":" in nameserver else socket.AF_INET
asking = socket.socket(family, socket.SOCK_STREAM if stream else socket.SOCK_DGRAM)
asking.settimeout(5)
try:
    if stream:
        # over TCP a message comes after its length, and goes back the same
        asking.connect((nameserver, 53))
        asking.sendall(struct.pack("!H", len(query)) + query)
        answer = asking.recv(514)[2:]
    else:
        asking.sendto(query, (nameserver, 53))
        answer = asking.recv(512)
except OSError:
    # the builder refused it, or nothing came back
    sys.exit(0)

if struct.unpack("!H", answer[6:8])[0]:
    record = socket.AF_INET6 if kind == 28 else socket.AF_INET
    size = 16 if kind == 28 else 4
    print(socket.inet_ntop(record, answer[-size:]))
