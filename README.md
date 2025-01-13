# DegoNet OpenVPN Server Configuration

## Setup
- Setup certificate folder
```bash
make-cadir certs
```

- Generate certificate authority
```bash
./easyrsa init-pki
./easyrsa gen-dh
./easyrsa build-ca nopass
./easyrsa build-server-full server nopass
```

- Run OpenVPN Server
```bash
sudo openvpn server.conf
```
