#!/usr/bin/env python3
"""Deterministic loopback-free TCP/UDP fixture used only inside VM namespaces."""
import select
import socket
import sys
import time

HOST = "198.19.0.1"
TCP_PORT = 18080
UDP_PORT = 18081
PAYLOAD = b"LUNA_VM_P0"
ACK = b"ACK_LUNA_VM_P0"

def serve():
    tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    udp = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    tcp.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    tcp.bind((HOST, TCP_PORT))
    tcp.listen(8)
    udp.bind((HOST, UDP_PORT))
    print("VM_NETPROBE_READY", flush=True)
    while True:
        ready,_,_ = select.select([tcp,udp], [], [], 2)
        for server in ready:
            if server is tcp:
                c,_ = tcp.accept()
                c.settimeout(1)
                try:
                    if c.recv(1024) == PAYLOAD:
                        c.sendall(ACK)
                except (OSError,TimeoutError):
                    pass
                finally:
                    c.close()
            else:
                data,addr = udp.recvfrom(1024)
                if data == PAYLOAD:
                    udp.sendto(ACK,addr)

def tcp_check():
    try:
        with socket.create_connection((HOST,TCP_PORT),timeout=0.7) as c:
            c.settimeout(0.7)
            c.sendall(PAYLOAD)
            return c.recv(64) == ACK
    except (OSError, TimeoutError):
        return False

def udp_check():
    try:
        with socket.socket(socket.AF_INET,socket.SOCK_DGRAM) as c:
            c.settimeout(0.7)
            c.sendto(PAYLOAD,(HOST,UDP_PORT))
            return c.recv(64) == ACK
    except (OSError, TimeoutError):
        return False

def client(expected):
    tcp_ok = tcp_check()
    udp_ok = udp_check()
    should_work = expected == "accept"
    if tcp_ok != should_work or udp_ok != should_work:
        raise RuntimeError(f"TCP/UDP unexpected verdict tcp={tcp_ok}, udp={udp_ok} expected={expected}")
    print("VM_TCP_UDP_" + expected.upper() + "_PASS")

if __name__ == "__main__":
    if len(sys.argv)==2 and sys.argv[1]=="server":
        serve()
    elif len(sys.argv)==3 and sys.argv[1]=="client" and sys.argv[2] in ("accept","drop"):
        client(sys.argv[2])
    else:
        raise SystemExit("usage: guest-netprobe.py server | client accept|drop")
