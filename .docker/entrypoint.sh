#!/bin/bash
set -e

iptables -t nat -A POSTROUTING -s 10.10.0.0/24 -d 11.11.0.0/24 -j MASQUERADE
exec openvpn --config /etc/openvpn/server.conf
