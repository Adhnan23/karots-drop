# karots-drop — by Karots

Share text and files ephemerally via CLI or Web UI. Data is stored in-memory with a configurable TTL and a 6-digit code. No database, no persistence, no accounts.

[![Go](https://github.com/Adhnan23/karots-drop/actions/workflows/ci.yml/badge.svg)](https://github.com/Adhnan23/karots-drop/actions/workflows/ci.yml)

## Features

- **Single static binary** — zero runtime dependencies, no database, no JS framework
- **CLI-first** — works on headless servers, SSH sessions, and pipelines
- **Embedded Web UI** — serves a dark-themed SPA out of the box, no separate frontend
- **Optional AES-256-GCM encryption** — client-side, zero-knowledge (key never reaches the server)
- **QR codes** — terminal ASCII or server-side 256×256 PNG
- **Optional token auth** — lock the API down with `--token`
- **Configurable TTL** — set item expiry with `--ttl` (default 20m, e.g. `5m`, `1h`)
- **Rate limiting** — per-IP sliding window with `X-RateLimit-*` response headers
- **Delete on retrieve** — one-time read mode for ephemeral delivery
- **Localhost-only binding** — `--bind localhost` restricts to loopback interface
- **Clipboard integration** — `xclip`/`xsel` with watch mode
- **Client-requested TTL** — `karots-drop send --ttl 5m` asks the server for a custom TTL
- **Compact output** — `karots-drop send --compact` prints only the code (scripting-friendly)
- **Health endpoint** — `GET /api/health` for monitoring and Docker HEALTHCHECK
- **Copy-to-clipboard** — Web UI buttons to copy code and URL
- **Bash completion** — `karots-drop completion` outputs a shell completion script
- **Graceful shutdown** — drains connections on SIGINT/SIGTERM

## Quick start

```bash
# Build
make build

# Start the server
./bin/karots-drop serve
#   karots-drop listening on :8080
#   Web UI: http://localhost:8080
#   API:    http://localhost:8080/api/

# In another terminal, send text
./bin/karots-drop send "hello world"
#   Code: 482910
#   URL:  http://localhost:8080/api/get/482910
#   TTL:  1200 seconds

# Retrieve by code
./bin/karots-drop get 482910
#   hello world

# Or open http://localhost:8080 in a browser
```

---

## CLI reference

### serve

Start the HTTP server with the API, embedded Web UI, and optional middleware.

```bash
./bin/karots-drop serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:8080` | Listen address and port |
| `--bind` | `""` | Bind to interface (`localhost` restricts to 127.0.0.1) |
| `--token` | `""` | API token for `X-Auth-Token` header (empty = disabled) |
| `--rate-limit` | `60` | Max requests per minute per IP (`0` = unlimited) |
| `--max-items` | `500` | Max items in the in-memory store (`0` = unlimited) |
| `--delete-on-retrieve` | `false` | Delete items immediately after they are retrieved |
| `--ttl` | `20m` | Item TTL (e.g. `5m`, `30m`, `1h`) |

On SIGINT/SIGTERM the server drains in-flight connections for up to 10 seconds before exiting.

See the **REST API** section below for details on how token auth, rate limiting, TTL, and delete-on-retrieve behave at the HTTP level.

---

### send

Upload text, a file, or piped data to a running server.

```bash
# Send text as argument
./bin/karots-drop send "message here"

# Send text via pipe
echo "hello" | ./bin/karots-drop send

# Send a file
./bin/karots-drop send --file photo.jpg

# Encrypt before upload
./bin/karots-drop send --encrypt "secret data"

# Request custom TTL (capped at server's --ttl)
./bin/karots-drop send --ttl 5m "short-lived data"

# Compact output — prints only the code
code=$(./bin/karots-drop send --compact "data")

# Generate QR code after upload
./bin/karots-drop send --qr --file document.pdf

# JSON output (for scripts)
./bin/karots-drop send --json "data"

# Target a remote server
./bin/karots-drop send --server https://drop.example.com "hello"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--file` | `""` | Path to a file to upload |
| `--server` | `http://localhost:8080` | Server URL |
| `--encrypt` | `false` | Encrypt data with AES-256-GCM before upload |
| `--qr` | `false` | Display a QR code in the terminal |
| `--json` | `false` | Output as JSON |
| `--ttl` | `""` | Request custom TTL (e.g. `5m`, `1h`); capped at server max |
| `--compact` | `false` | Output only the code (silent on stderr) |

If `--file` is provided, the file is uploaded with its original filename. Otherwise the positional argument or piped stdin is sent as text.

When `--encrypt` is used, a random 32-byte AES key is generated client-side. The ciphertext is sent to the server; the key is printed locally and is **never stored on the server**. You need it to decrypt later.

---

### get

Retrieve data by its 6-digit code.

```bash
# Get data (prints to stdout)
./bin/karots-drop get 482910

# Get and decrypt
./bin/karots-drop get 482910 --key "a1b2c3...base64..."

# Get metadata as JSON
./bin/karots-drop get 482910 --json

# From a remote server
./bin/karots-drop get 482910 --server https://drop.example.com
```

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | `http://localhost:8080` | Server URL |
| `--json` | `false` | Output metadata as JSON |
| `--key` | `""` | Base64-encoded AES decryption key |

When the server returns `X-Encrypted: true` but no `--key` is given, the raw ciphertext is printed to stdout and a warning is shown on stderr.

---

### clip

Read the system clipboard and upload its contents. Linux only (requires `xclip` or `xsel`).

```bash
sudo apt install xclip   # Debian/Ubuntu
```

```bash
# Upload clipboard content
./bin/karots-drop clip

# Encrypt clipboard content
./bin/karots-drop clip --encrypt

# Watch mode — poll clipboard every 2 seconds and auto-upload changes
./bin/karots-drop clip --watch
```

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | `http://localhost:8080` | Server URL |
| `--encrypt` | `false` | Encrypt before upload |
| `--qr` | `false` | Show QR code |
| `--json` | `false` | JSON output |
| `--watch` | `false` | Watch clipboard for changes, auto-upload |

In watch mode the clipboard is polled every 2 seconds. Only changed content (compared to the previous read) triggers an upload. Press Ctrl+C to stop.

---

### version

Print the binary version.

```bash
./bin/karots-drop version
# or
./bin/karots-drop --version
```

### health

Check if a running server is healthy. Exits 0 on success, 1 on failure. Used by the Docker HEALTHCHECK.

```bash
./bin/karots-drop health
./bin/karots-drop health --server http://localhost:8080
```

### completion

Print a bash completion script. Source it to enable tab completion for commands and flags.

```bash
source <(./bin/karots-drop completion)
```

---

## REST API

All endpoints are served by `karots-drop serve`. The API is available at `/api/` and the Web UI at `/`.

### Middleware

Requests pass through the following middleware chain before reaching a handler:

1. **Logging** — logs `METHOD /path STATUS BYTES DURATION` to stderr for every request
2. **CORS** — sets `Access-Control-Allow-Origin: *`, allows `GET`, `POST`, `OPTIONS`, headers `Content-Type` and `X-Auth-Token`. Preflight `OPTIONS` returns `204 No Content`.
3. **Token auth** — if `--token` is set, every API request must carry `X-Auth-Token: <token>`. Requests to `/`, `/api/health`, and `/static/*` are exempt.
4. **Rate limiter** — if `--rate-limit` is set (default 60), each IP is limited to that many requests per minute. Every response includes `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers. Exceeded requests receive `429 Too Many Requests` with a `Retry-After: 60` header.

### `POST /api/store`

Upload text or a file. Accepts `multipart/form-data` or raw body.

**multipart form fields**

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | Text content to share. Ignored if `file` is present. |
| `file` | file | File upload. Overrides `text` when present. |
| `encrypted` | string | Set to `"true"` to mark data as encrypted (informational). |
| `ttl` | string | Request custom TTL (e.g. `5m`, `1h`); capped at the server's `--ttl` max. |

**raw body**

Send raw bytes with a `Content-Type` header.

```bash
curl -X POST -d "hello world" http://localhost:8080/api/store
curl -X POST -F "text=hello"   http://localhost:8080/api/store
curl -X POST -F "file=@photo.jpg" http://localhost:8080/api/store
curl -X POST -F "text=hello" -F "ttl=5m" http://localhost:8080/api/store
```

**Response** `201 Created`

```json
{
  "code": "482910",
  "ttl": 1200,
  "url": "http://localhost:8080/api/get/482910",
  "filename": "photo.jpg",
  "encrypted": false
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | string | 6-digit retrieval code |
| `ttl` | number | Time-to-live in seconds (reflects the actual TTL used) |
| `url` | string | Full retrieval URL |
| `filename` | string | Original filename (only for file uploads) |
| `encrypted` | bool | Whether the item is marked as encrypted |

**Errors**

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error":"Empty body"}` | Body is empty |
| `400` | `{"error":"Upload too large (max 20MB)"}` | Payload exceeds 20 MB |
| `400` | `{"error":"Data exceeds 20MB limit"}` | Parsed data exceeds 20 MB |
| `400` | `{"error":"No text or file provided"}` | Multipart form with neither `text` nor `file` |
| `401` | `{"error":"Unauthorized"}` | `--token` is set and `X-Auth-Token` is missing or wrong |
| `429` | `{"error":"Rate limit exceeded"}` | IP exceeded the rate limit |
| `503` | `{"error":"Store is full, try again later"}` | Store has reached `--max-items` capacity |

---

### `GET /api/get/{code}`

Retrieve stored data by its 6-digit code.

```bash
curl http://localhost:8080/api/get/482910
```

**Response** `200 OK`

Returns the raw stored bytes with the original `Content-Type` and optional `Content-Disposition` for file downloads.

| Header | When set |
|--------|----------|
| `Content-Type` | Original content type from upload |
| `Content-Disposition` | `attachment; filename="..."` for file uploads |
| `X-Encrypted` | `"true"` if the item was uploaded with `encrypted=true` |

When `--delete-on-retrieve` is enabled on the server, the item is deleted immediately after being read. A second request for the same code returns `404`.

**Errors**

| Status | Body |
|--------|------|
| `400` | `{"error":"Missing code"}` |
| `404` | `{"error":"Not found or expired"}` |

Expired items are deleted on access and return `404`.

---

### `GET /api/qr/{code}`

Generate a 256×256 PNG QR code encoding the retrieval URL for the given code.

```bash
curl -o qr.png http://localhost:8080/api/qr/482910
```

**Response** `200 OK` — `Content-Type: image/png`

**Errors**

| Status | Body |
|--------|------|
| `400` | `{"error":"Missing code"}` |
| `404` | `{"error":"Not found or expired"}` |

---

### `GET /api/health`

Health check endpoint for monitoring and Docker HEALTHCHECK.

```bash
curl http://localhost:8080/api/health
```

**Response** `200 OK`

```json
{"status":"ok"}
```

This endpoint is exempt from token authentication. Always returns 200 when the server is running.

---

### `GET /`

Serves the embedded Web UI (single-page application).

```bash
curl http://localhost:8080/
```

Returns `index.html` with linked `/style.css` and `/script.js`. Requires a browser for full functionality.

If `--token` is set on the server, open the Web UI with `?token=xxx` in the URL:

```
http://localhost:8080/?token=your-api-token
```

The token is sent as `X-Auth-Token` on all API calls made by the UI.

---

## Web UI

Single-page application embedded via `//go:embed`. No framework, vanilla JavaScript. Features:

- **Text upload** — textarea with upload button
- **File upload** — drag-and-drop zone + file picker
- **Retrieve** — 6-digit code input with inline text display or file download
- **QR code modal** — shown after every upload, generated server-side at `/api/qr/{code}`
- **Copy-to-clipboard** — buttons to copy the code and URL after upload, plus in the QR modal
- **Dynamic TTL display** — shows actual expiry time from the server response
- **Dark theme** — GitHub-inspired color palette
- **Mobile responsive** — 600px breakpoint, full-width layout on small screens

---

## Security model

| Property | Detail |
|----------|--------|
| Data at rest | In-memory only, no persistence to disk |
| Data in transit | No built-in TLS — deploy behind Caddy/nginx for HTTPS |
| Encryption at rest | Optional AES-256-GCM, client-side key (zero-knowledge) |
| Access control | Optional `--token` flag (`X-Auth-Token` header). Default: no auth. |
| Upload limit | 20 MB per item, enforced at the HTTP layer |
| TTL | Configurable via `--ttl` (default 20 min), enforced server-side |
| Expiry cleanup | Background goroutine every 60 seconds |
| Restart | All data lost on restart (by design) |
| Rate limiting | Per-IP sliding window (default 60 req/min), exposes `X-RateLimit-*` headers |
| Store capacity | Configurable max items (default 500, returns 503 when full) |
| Delete on retrieve | Optional one-time read mode |
| Localhost-only | `--bind localhost` restricts to 127.0.0.1 |
| Collision resistance | `crypto/rand` digits, up to 100 retries |

---

## Build from source

```bash
# Build for current platform (version from git tag)
make build
#   → bin/karots-drop

# Override version
make build VERSION=1.0.0

# Cross-compile
make build-linux      # linux/amd64   → bin/karots-drop-linux-amd64
make build-arm64      # linux/arm64   → bin/karots-drop-linux-arm64
make build-darwin     # darwin/amd64  → bin/karots-drop-darwin-amd64
make build-windows    # windows/amd64 → bin/karots-drop-windows-amd64.exe

# Run tests
make test

# Clean build artifacts
make clean
```

All builds are static (`CGO_ENABLED=0`) and stripped (`-s -w`). The version is baked into the binary via `-X main.version`. Requires Go 1.26.

---

## Deployment

### Direct

```bash
./bin/karots-drop serve --addr :8080 --token secret --delete-on-retrieve --ttl 10m --bind localhost
```

### systemd service

```
[Unit]
Description=karots-drop ephemeral sharing
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/karots-drop serve --addr :8080 --ttl 30m
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Docker

The repository includes a multi-stage `Dockerfile` that produces a ~6 MB scratch-based image with built-in HEALTHCHECK:

```bash
docker build -t karots-drop .
docker run -d -p 8080:8080 karots-drop
```

With custom flags:

```bash
docker run -d -p 8080:8080 karots-drop serve \
  --addr :8080 \
  --token "$TOKEN" \
  --rate-limit 30 \
  --max-items 200 \
  --ttl 10m \
  --delete-on-retrieve
```

The HEALTHCHECK runs every 30 seconds via the embedded `health` command:
```
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD ["/karots-drop", "health"]
```

### GitHub Container Registry (CI)

On tagged releases (`v*`), the CI pipeline automatically:

1. Runs `go vet` and all tests
2. Builds static binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `windows/amd64`
3. Builds and pushes a Docker image to `ghcr.io`
4. Creates a GitHub Release with all binaries attached

The pipeline is defined in `.github/workflows/ci.yml`.

### Railway / Fly.io

Deploy as a single-process HTTP service. Use the `Dockerfile` — no buildpack needed.

---

## Project structure

```
karots-drop/
├── cmd/
│   └── karots-drop/          # Binary entry point
│       ├── main.go            # Subcommand dispatch + version
│       ├── serve.go           # serve command (+ graceful shutdown, --bind)
│       ├── send.go            # send command (+ --ttl, --compact)
│       ├── get.go             # get command
│       ├── clip.go            # clip command (+ --watch)
│       ├── health.go          # health check command
│       ├── completion.go      # bash completion script
│       └── usage.go           # Help text
├── internal/
│   ├── store/                 # In-memory storage with TTL cleanup
│   ├── crypt/                 # AES-256-GCM encrypt/decrypt
│   ├── qr/                    # Terminal and PNG QR generation
│   ├── api/                   # HTTP server, middleware, route handlers
│   ├── clip/                  # Clipboard reader (xclip/xsel)
│   └── webui/                 # Embedded static assets
│       └── static/
│           ├── index.html
│           ├── style.css
│           └── script.js
├── bin/                       # Build output (gitignored)
├── Makefile                   # Static cross-compile targets + version ldflags
├── Dockerfile                 # Multi-stage scratch image + HEALTHCHECK
├── .github/workflows/ci.yml   # Test → build → docker → release
├── go.mod / go.sum
└── README.md
```

---

## FAQ

**Why not use a database?** By design: no persistence across restarts means no cleanup of stale data, no files to manage, no backup concerns. The TTL makes persistence unnecessary.

**Is the clipboard command available on macOS/Windows?** The `clip` command uses `xclip`/`xsel` via `os/exec` (Linux only). macOS and Windows clipboard support requires CGO and is not included in this static build.

**Can I run this behind a reverse proxy?** Yes. karots-drop doesn't handle TLS itself — run Caddy, nginx, or Traefik in front for HTTPS and domain binding.

**How are codes generated?** `crypto/rand` picks 6 decimal digits. If there's a collision (extremely rare in practice), the server retries up to 100 times.

**What happens to my data after the TTL expires?** The cleanup goroutine deletes expired entries every 60 seconds. Expired entries are also checked on read.

**Can I set a custom TTL?** Yes — configure the server default with `--ttl` (e.g. `--ttl 5m`). Clients can also request a shorter TTL per upload with `karots-drop send --ttl 5m` or via the `ttl` form field in the API. The server caps the client's request at its own `--ttl` max.

**What happens when the store is full?** `POST /api/store` returns `503 Service Unavailable` with `{"error":"Store is full, try again later"}`. Increase capacity with `--max-items` or wait for entries to expire.

**How does delete-on-retrieve work?** When `--delete-on-retrieve` is enabled, the item is removed from the store immediately after reading. A second GET for the same code returns `404`. This ensures one-time delivery.

**How does token auth work?** Set `--token some-secret` on the server. API clients must include `X-Auth-Token: some-secret` in every request. The Web UI reads the token from the URL query parameter: `http://localhost:8080/?token=some-secret`. The root page (`/`), health endpoint (`/api/health`), and static assets (`/static/*`) are exempt from auth checks.

**What are the rate limit headers?** Every response includes `X-RateLimit-Limit` (max requests/min), `X-RateLimit-Remaining` (requests left), and `X-RateLimit-Reset` (unix timestamp when the window resets). When exceeded, the server returns `429 Too Many Requests` with a `Retry-After` header.

**How do I use bash completion?** Run `source <(karots-drop completion)` to enable tab completion for all commands and flags.

**Is there TLS support?** Not built-in. Deploy behind Caddy, nginx, or use a Cloudflare Tunnel for HTTPS termination.

**How do I get just the code for scripting?** Use `karots-drop send --compact "data"` — it prints only the 6-digit code to stdout with no other output.

---

## License

MIT — © 2025 Adhnan (Karots)
