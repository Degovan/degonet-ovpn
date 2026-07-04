#!/bin/bash
set -e

export DB_FILE="${DB_FILE:-/etc/openvpn/data/users.sqlite}"
export CCD_DIR="${CCD_DIR:-/etc/openvpn/ccds}"
export DEFAULT_NETMASK="${DEFAULT_NETMASK:-255.255.255.0}"
export API_PORT="${API_PORT:-8080}"

cleanup() {
    kill "$API_PID" 2>/dev/null || true
    kill "$OVPN_PID" 2>/dev/null || true
}
trap cleanup SIGTERM SIGINT

iptables -t nat -A POSTROUTING -s 10.10.0.0/24 -d 11.11.0.0/24 -j MASQUERADE

/etc/openvpn/auth serve &
API_PID=$!

openvpn --config /etc/openvpn/server.conf &
OVPN_PID=$!

wait -n $API_PID $OVPN_PID
cleanup
exit 1
