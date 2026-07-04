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
cd ..
```

- Manage users via CLI (run from the project root after build)

```bash
./auth.phar add <username> <ip> [password] [netmask]
./auth.phar list
./auth.phar delete <username>
```

Example:

```bash
./auth.phar add budi 10.8.0.10 rahasia 255.255.255.0
./auth.phar list
./auth.phar delete budi
```

Notes for the `add` command:

- Username must be unique in the database.
- IP address must be unique in the database.
- If password is not provided, it defaults to the username.
- Successful output is shown in table format and the password value is masked as `****`.

- Run OpenVPN Server

```bash
sudo openvpn server.conf
```
