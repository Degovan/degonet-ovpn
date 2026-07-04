#!/bin/bash
set -e

export DB_FILE="${DB_FILE:-/etc/openvpn/data/users.sqlite}"
export CCD_DIR="${CCD_DIR:-/etc/openvpn/ccds}"
export DEFAULT_NETMASK="${DEFAULT_NETMASK:-255.255.255.0}"

iptables -t nat -A POSTROUTING -s 10.10.0.0/24 -d 11.11.0.0/24 -j MASQUERADE
exec openvpn --config /etc/openvpn/server.conf
