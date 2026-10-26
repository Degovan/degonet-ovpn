# DegoNet OpenVPN Server Configuration

## Setup

Generate certificate authority:

```bash
bash setup-certs.sh
```

Build authentication binary (the web UI is embedded, so build it first):

```bash
cd auth-src/web
npm ci && npm run build
cd ..
CGO_ENABLED=0 go build -ldflags="-s -w" -o auth .
cd ..
```

Skipping the web build still compiles; `/` then just says the UI is not built.

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

## Web UI

`auth serve` also serves a web dashboard at `/` (with Docker: `http://<server>:8080/`, port set by `API_HOST_PORT`). Log in with `API_KEY`; the browser gets a 12-hour `HttpOnly` session cookie.

- List, search and filter users by online/offline status
- Add (next free IP in `VPN_SUBNET` is suggested), edit IP/netmask/password, delete and bulk delete
- Live status from the OpenVPN management interface: public IP, connected time, traffic, and kick

The frontend is Preact + Vite in `auth-src/web`. For development, run `./auth serve` (port 8080) and then:

```bash
cd auth-src/web
npm run dev
```

Vite proxies `/api` to `http://localhost:8080` (override with `API_TARGET`).

## HTTP API

Start API server:

```bash
API_KEY=your-secret-key ./auth serve
API_KEY=your-secret-key ./auth serve --port 9090
```

All endpoints except `/api/auth`, `/api/login` and `/api/logout` require the `X-API-Key` header (or a web UI session cookie).

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

**PUT /api/users/{username}** — Update password, IP and/or netmask (omitted fields stay unchanged)

```bash
curl -X PUT -H "X-API-Key: your-secret-key" -H "Content-Type: application/json" \
  -d '{"ip":"10.8.0.11","password":"baru"}' \
  http://localhost/api/users/budi
```

**GET /api/status** — Connected clients from the OpenVPN management interface

```bash
curl -H "X-API-Key: your-secret-key" http://localhost/api/status
```

**POST /api/users/{username}/kick** — Disconnect a connected client

```bash
curl -X POST -H "X-API-Key: your-secret-key" http://localhost/api/users/budi/kick
```

**GET /api/config** — VPN subnet and default netmask (used by the web UI)

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

Or use the interactive manager, which sets the API secret in `.env`, creates the `degonet-network` network if missing, and wraps the common compose/user actions:

```bash
./manage.sh            # interactive menu
./manage.sh secret     # generate or enter API_KEY, then recreate the container
./manage.sh rebuild    # rebuild image and start
./manage.sh help       # all actions (status, up, restart, logs, health, users, user:add, user:delete, ...)
```

`API_KEY` is read from `.env` in the project root (next to `docker-compose.yml`); without it the container falls back to `changeme`. After changing it the container must be recreated (`./manage.sh up`), a plain restart keeps the old value.

Manage users inside container:

```bash
docker exec degonet-openvpn /etc/openvpn/auth add <username> <ip> [password] [netmask]
docker exec degonet-openvpn /etc/openvpn/auth list
docker exec degonet-openvpn /etc/openvpn/auth delete <username> [username...] [-y]
docker exec -it degonet-openvpn /etc/openvpn/auth delete
```

Running `auth delete` without a username opens an interactive checklist: type to search (username or IP), `space`/`tab` to select, `ctrl+a` to select all visible, `enter` to continue, `esc` to cancel. Deleting multiple users asks for confirmation; pass `-y` to skip it.

## Configuration

Environment variables (set via `.env` in `auth-src/` or Docker environment):

| Variable | Default | Description |
|---|---|---|
| `DB_FILE` | `/etc/openvpn/data/users.sqlite` | SQLite database path |
| `CCD_DIR` | `/etc/openvpn/ccds` | Client config directory |
| `DEFAULT_NETMASK` | `255.255.255.0` | Default netmask for new users |
| `API_KEY` | (required for serve) | API key for HTTP API authentication |
| `API_PORT` | `80` | HTTP API server port |
| `API_HOST_PORT` | `8080` | Host port the API/web UI is published on (docker compose only) |
| `MGMT_ADDR` | `localhost:7505` | OpenVPN management interface, used for online status and kick |
| `VPN_SUBNET` | `11.11.0.0/24` | VPN subnet, used by the web UI to suggest free IPs |
