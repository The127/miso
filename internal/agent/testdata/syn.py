# Sends a TCP SYN as a frame of its own to the gateway's MAC, past the IP
# stack of the run, and prints reached when the answer is a SYN-ACK, which
# QEMU sends only once it connected on the host.
# Usage: syn.py <gateway MAC> <own address> <target address> <port>
import socket
import struct
import sys
import time


def checksum(data):
    if len(data) % 2:
        data += b"\0"
    total = sum(struct.unpack(f"!{len(data) // 2}H", data))
    while total >> 16:
        total = (total & 0xFFFF) + (total >> 16)
    return ~total & 0xFFFF


def mac(text):
    return bytes.fromhex(text.strip().replace(":", ""))


gateway = mac(sys.argv[1])
source = socket.inet_aton(sys.argv[2])
target = socket.inet_aton(sys.argv[3])
port = int(sys.argv[4])
with open("/sys/class/net/eth0/address") as f:
    own = mac(f.read())

tcp = struct.pack("!HHIIBBHHH", 40000, port, 1, 0, 5 << 4, 0x02, 64240, 0, 0)
pseudo = source + target + struct.pack("!BBH", 0, socket.IPPROTO_TCP, len(tcp))
tcp = tcp[:16] + struct.pack("!H", checksum(pseudo + tcp)) + tcp[18:]
ip = struct.pack("!BBHHHBBH4s4s", 0x45, 0, 20 + len(tcp), 1, 0, 64, socket.IPPROTO_TCP, 0, source, target)
ip = ip[:10] + struct.pack("!H", checksum(ip)) + ip[12:]

frames = socket.socket(socket.AF_PACKET, socket.SOCK_RAW, socket.htons(0x0800))
frames.bind(("eth0", 0))
try:
    frames.send(gateway + own + b"\x08\x00" + ip + tcp)
except OSError:
    # the builder refused it, so it went nowhere
    sys.exit(0)
frames.settimeout(0.2)
end = time.time() + 3
while time.time() < end:
    try:
        packet = frames.recv(2048)[14:]
    except socket.timeout:
        continue
    header = (packet[0] & 0x0F) * 4
    if packet[9] == socket.IPPROTO_TCP and packet[12:16] == target and packet[header + 13] & 0x12 == 0x12:
        print("reached")
        break
