# Honeyshell

An SSH honeypot written entirely in Go, using `crypto/ssh` as its base.

Currently, it allows connections on the server and collects failed login attempts, meaning all usernames and passwords. Also, you can decrease the permissions on the process by flipping to a different user (ie 'nobody').

## Building

To compile the program, simply run `go build` in your terminal.

## Running

The default port is 22 and no user is mandatory, but the `-key` is, so make sure you set one.

```
Usage of ./honeyshell:
  -banner string
    	The banner for the SSH server (default "OpenSSH_7.4p1 Raspbian-10+deb9u3")
  -key string
    	The RSA key to use
  -port int
    	The port the deamon should run on (default 22)
  -user string
    	Set the permissions to a certain user (ie 'nobody')
```

Example usage:

``` sh
sudo ./honeyshell -user nobody -port 2222 -key ~/.ssh/key_for_honeypot_rsa
```

The output should look something like this:

```
2022/12/27 11:30:09 Starting on port 2222
2022/12/27 11:30:09 Changing permissions to user 'nobody'
2022/12/27 11:30:12 127.0.0.1:54476 test key:442f78a53b6188a6b18a225c86aeb9c77592add0d714d26f8d84c9e4f9f59a77 (ssh-rsa)
2022/12/27 11:30:12 127.0.0.1:54476 test key:dfde57ac7251135922968b933fd28944e384a504515ba7ce27cd925224af4657 (ssh-rsa)
2022/12/27 11:30:12 127.0.0.1:54476 test key:c09d0feddad874b7a2cd82bdc4b846632fa47a7113a17cf8c846876a5f4eaf4a (ssh-rsa)
2022/12/27 11:30:12 127.0.0.1:54476 test key:589adac413c8d42c9166dd0e56d2da5cdeaf4abd731f001b3fa40b3c6232383b (ssh-rsa)
2022/12/27 11:30:12 Error during handshake ssh: disconnect, reason 2: too many authentication failures
2022/12/27 11:30:19 127.0.0.1:54478 test pass:test
2022/12/27 11:30:20 127.0.0.1:54478 test pass:admin
2022/12/27 11:30:26 127.0.0.1:54478 test pass:password123
```
