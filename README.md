# DegoNet OpenVPN Server Configuration

## Setup
- Setup certificate folder
```bash
make-cadir certs
```

- Generate certificate authority
```bash
cd certs
./easyrsa init-pki
./easyrsa gen-dh
./easyrsa build-ca nopass
./easyrsa build-server-full server nopass
cd ..
```

- Compile authentication file
```bash
cd auth
cp .env.example .env
composer install --no-dev
composer run build
cd .. && chmod +x auth.phar
```

- Run OpenVPN Server
```bash
sudo openvpn server.conf
```
