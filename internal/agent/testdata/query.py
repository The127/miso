# Asks a nameserver for a name over UDP and prints the address it answers,
# and nothing at all when no answer comes.
# Usage: query.py <nameserver> <name> <A|AAAA>
import socket
import struct
import sys

nameserver = sys.argv[1]
kind = 28 if sys.argv[3] == "AAAA" else 1
labels = b"".join(bytes([len(p)]) + p.encode() for p in sys.argv[2].split("."))
query = struct.pack("!HHHHHH", 1, 0x0100, 1, 0, 0, 0) + labels + b"\0" + struct.pack("!HH", kind, 1)

asking = socket.socket(socket.AF_INET6 if ":" in nameserver else socket.AF_INET, socket.SOCK_DGRAM)
asking.settimeout(5)
try:
    asking.sendto(query, (nameserver, 53))
    answer = asking.recv(512)
except OSError:
    # the builder refused it, or nothing came back
    sys.exit(0)

if struct.unpack("!H", answer[6:8])[0]:
    record = socket.AF_INET6 if kind == 28 else socket.AF_INET
    size = 16 if kind == 28 else 4
    print(socket.inet_ntop(record, answer[-size:]))
