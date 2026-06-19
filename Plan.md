# 🧭 karots-drop — Project Plan

## 1. Goal

A single Go binary that provides:

* ephemeral text/file transfer (≤20MB)
* 6-digit retrieval codes
* 20-minute TTL auto-expiry
* CLI-first usage (works on headless Linux servers)
* optional Web UI (embedded static frontend)
* optional encryption + QR sharing + clipboard automation

---

# 🧱 2. Core System Architecture

### Components inside one binary:

* HTTP API server (core transport layer)
* CLI interface (primary user interaction)
* In-memory store (no database)
* TTL cleanup worker
* Optional Web UI (embedded assets)
* Optional crypto module (AES-GCM)
* QR generator module
* Clipboard integration module

---

# 📦 3. Functional Modules

## A. Storage System

* In-memory map keyed by 6-digit code
* Stores:

  * bytes (text/file)
  * filename (if file)
  * timestamp
  * expiry time
  * encryption metadata (optional)

Constraints:

* max size: 20MB per item
* auto-delete after 20 minutes
* periodic cleanup goroutine

---

## B. API Layer

Endpoints:

* `POST /api/store`
* `GET /api/get/{code}`
* `GET /` (web UI optional)

Responsibilities:

* accept multipart file/text
* return 6-digit code
* return metadata (TTL, URL, QR reference)

---

## C. CLI Layer (Primary Interface)

Commands:

* `send [text|file]`
* `get <code>`
* `serve`
* `clip` (clipboard upload)
* optional flags:

  * `--encrypt`
  * `--qr`
  * `--json`
  * `--server <url>`

Supports piping:

* stdin input support
* command chaining from shell tools

---

## D. Web UI (Optional Layer)

* Embedded static HTML/CSS/JS
* Features:

  * text upload
  * file upload (drag & drop)
  * code input retrieval
  * QR display after upload
* Can be disabled in headless deployments

---

## E. QR Code System

* Generate QR from:

  * retrieval URL OR
  * direct API endpoint
* Output modes:

  * terminal ASCII QR (CLI)
  * PNG image (web UI)
* Used for fast mobile → PC transfer

---

## F. Encryption System (Optional Mode)

Modes:

* off (default)
* per-item AES-GCM encryption

Design:

* random key per item
* nonce-based encryption
* store encrypted payload in memory only
* optional client-side decrypt mode for advanced use

---

## G. Clipboard Automation

CLI mode:

* read system clipboard
* auto-upload content
* return code + QR + URL

Optional:

* watch mode (auto-upload on clipboard change)

---

# 🌐 4. Deployment Model

Supports:

* local machine
* LAN server
* SSH/VPS headless server
* Railway / Fly.io / Docker container

Key constraint:

* in-memory storage → no persistence across restarts

---

# 🔐 5. Security Model

* 20MB max upload limit
* 20-minute TTL enforced server-side
* optional encryption per item
* optional localhost-only binding mode
* no authentication by default (optional token mode later)

---

# 🔁 6. User Flow Design

## Send flow:

1. CLI/web uploads data
2. server stores in memory
3. generates 6-digit code
4. generates share URL
5. optionally generates QR
6. returns response

---

## Receive flow:

1. user enters code or scans QR
2. request `/api/get/{code}`
3. server returns payload
4. item deleted if expired or optionally after fetch

---

# 📱 7. Output Formats

CLI output modes:

* human-readable (default)
* JSON mode (automation-friendly)
* QR mode (terminal or image)
* compact code-only mode

---

# ⚙️ 8. Runtime Behavior

* single binary
* embedded web assets
* no external dependencies at runtime
* background cleanup worker
* stateless restart behavior

---

# 🚀 9. Build Targets

* Linux (primary server use)
* Windows (desktop CLI)
* macOS (developer use)
* ARM64 (optional for servers/Raspberry Pi)

---

# 📈 10. Optional Future Extensions (not required in MVP)

* peer-to-peer LAN mode (no server)
* WebSocket live transfer updates
* password-protected shares
* device pairing trust model
* persistent mode (SQLite optional, explicitly not in current plan)
* multi-file bundles
