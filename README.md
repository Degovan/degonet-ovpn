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

## HTTP API

Start API server:

```bash
API_KEY=your-secret-key ./auth serve
API_KEY=your-secret-key ./auth serve --port 9090
```

All endpoints (except `/api/auth`) require `X-API-Key` header.

### Endpoints

**POST /api/auth** — Authenticate user (public)

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"username":"budi","password":"rahasia"}' \
  http://localhost/api/auth
```

**GET /api/users** — List all users

```bash
curl -H "X-API-Key: your-secret-key" http://localhost/api/users
```

**POST /api/users** — Add user

```bash
curl -X POST -H "X-API-Key: your-secret-key" -H "Content-Type: application/json" \
  -d '{"username":"budi","password":"rahasia","ip":"10.8.0.10","netmask":"255.255.255.0"}' \
  http://localhost/api/users
```

**DELETE /api/users/{username}** — Delete user

```bash
curl -X DELETE -H "X-API-Key: your-secret-key" http://localhost/api/users/budi
```

### Response Format

```json
{"success": true, "user": {"username": "budi", "ip": "10.8.0.10", "netmask": "255.255.255.0"}}
{"success": false, "error": "username already exists"}
```

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
| `API_KEY` | (required for serve) | API key for HTTP API authentication |
| `API_PORT` | `80` | HTTP API server port |
