#!/bin/bash
set -e

CERTS_DIR="./certs"

mkdir -p "$CERTS_DIR"

curl -Lo easy-rsa.tgz https://github.com/OpenVPN/easy-rsa/releases/download/v3.2.2/EasyRSA-3.2.2.tgz
tar xzf easy-rsa.tgz
mv EasyRSA-3.2.2 certs/easy-rsa
rm easy-rsa.tgz

cd certs/easy-rsa

./easyrsa init-pki

./easyrsa gen-dh

echo "yes" | ./easyrsa build-ca nopass

echo "yes" | ./easyrsa build-server-full degonet nopass

cp pki/ca.crt pki/issued/degonet.crt pki/private/degonet.key pki/private/ca.key pki/dh.pem ..

cd .. && rm -rf easy-rsa

echo "✅ Certificates created successfully"
