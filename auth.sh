#!/bin/bash

# File credentials untuk menyimpan username dan password
CREDENTIALS_FILE="credentials.txt"

if [ ! -f "$CREDENTIALS_FILE" ]; then
    echo "Credentials file not found."
    exit 1
fi

# Ambil username dan password dari variabel environment
USER="$username"
PASS="$password"

# Verifikasi username dan password
if grep -qx "$USER $PASS$" "$CREDENTIALS_FILE"; then
    exit 0  # Sukses
else
    exit 1  # Gagal
fi

