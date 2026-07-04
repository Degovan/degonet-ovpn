# DegoNet OpenVPN Server Configuration

## Setup

Generate certificate authority:

```bash
bash setup-certs.sh
```

Build authentication binary:

```bash
cd auth-src
CGO_ENABLED=0 go build -ldflags="-s -w" -o auth .
cd ..
```

## CLI Usage

Manage users via the auth binary:

```bash
./auth add <username> <ip> [password] [netmask]
./auth list
./auth delete <username>
```

Example:

```bash
./auth add budi 10.8.0.10 rahasia 255.255.255.0
./auth list
./auth delete budi
```

Notes for the `add` command:

- Username and IP must be unique in the database.
- If password is not provided, it defaults to the username.
- Successful output is shown in table format and the password value is masked as `****`.

## Run

Locally:

```bash
sudo openvpn server.conf
```

Docker:

```bash
docker compose up -d --build
```

Manage users inside container:

```bash
docker exec degonet-openvpn /etc/openvpn/auth add <username> <ip> [password] [netmask]
docker exec degonet-openvpn /etc/openvpn/auth list
docker exec degonet-openvpn /etc/openvpn/auth delete <username>
```

## Configuration

Environment variables (set via `.env` in `auth-src/` or Docker environment):

| Variable | Default | Description |
|---|---|---|
| `DB_FILE` | `/etc/openvpn/data/users.sqlite` | SQLite database path |
| `CCD_DIR` | `/etc/openvpn/ccds` | Client config directory |
| `DEFAULT_NETMASK` | `255.255.255.0` | Default netmask for new users |
