#!/usr/bin/env python3
# The nameserver of the harness's own network, on the loopback of the host
# QEMU runs on, where QEMU's built-in forwarder looks for it. It answers
# every name with one address of each family, from the documentation range,
# over UDP and over TCP, which a resolver falls back to for long answers.
import selectors
import socket
import struct

RECORDS = {
    1: socket.inet_pton(socket.AF_INET, "192.0.2.53"),
    28: socket.inet_pton(socket.AF_INET6, "2001:db8::53"),
}


def answer(query):
    end = 12
    while query[end]:
        end += query[end] + 1
    end += 1
    kind = struct.unpack("!H", query[end:end + 2])[0]
    question = query[12:end + 4]
    record = RECORDS.get(kind)
    if record is None:
        return query[:2] + struct.pack("!HHHHH", 0x8180, 1, 0, 0, 0) + question

    return (query[:2] + struct.pack("!HHHHH", 0x8180, 1, 1, 0, 0) + question
            + struct.pack("!HHHIH", 0xC00C, kind, 1, 60, len(record)) + record)


waiting = selectors.DefaultSelector()
for address, family in (("127.0.0.1", socket.AF_INET), ("::1", socket.AF_INET6)):
    datagrams = socket.socket(family, socket.SOCK_DGRAM)
    datagrams.bind((address, 53))
    waiting.register(datagrams, selectors.EVENT_READ, datagrams)
    stream = socket.create_server((address, 53), family=family)
    waiting.register(stream, selectors.EVENT_READ, stream)

while True:
    for ready, _ in waiting.select():
        server = ready.data
        if server.type == socket.SOCK_DGRAM:
            query, sender = server.recvfrom(512)
            server.sendto(answer(query), sender)
            continue

        # over TCP a message comes after its length, and goes back the same
        connection, _ = server.accept()
        size = struct.unpack("!H", connection.recv(2))[0]
        reply = answer(connection.recv(size))
        connection.sendall(struct.pack("!H", len(reply)) + reply)
        connection.close()
