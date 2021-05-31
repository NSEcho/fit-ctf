# fit-ctf

This is the API used for FIT Coding Challenge CTF in Mostar 2021. It is used as the first part of the CTF where the user is expected to grab a first flag and a clue how to proceed further.

It is a relatively small API will authentication as well as using specific header as a method of authorization.

It consists of the following endpoints:

* /api - the main endpoint
* /api/v1 - specific version, there are no others
* /api/v1/list - list all keys present or specific with /api/v1/list/key
* /api/v1/login - expects username and password in JSON and return the `FIT-Token` header
* /api/v1/users - return the list of the users

The idea here is to use password spraying and get people more familiar with using `curl` or whatever they want to communicate with API.
