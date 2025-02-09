# DegoNet OpenVPN Server Configuration

## Setup

- Setup certificate folder

```bash
mkdir certs
```

- Generate certificate authority

```bash
bash setup-certs.sh
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
