# Honeyshell

An SSH honeypot based on the [libssh](https://www.libssh.org/) library written entirely in Go.

Currently, it allows connections on the server and collects failed login attempts, meaning all usernames and passwords. Also, you can decrease the permissions on the process by flipping to a different user (ie 'nobody').

## Building

There are three dependencies, which are the following:

### 1. [libssh](https://www.libssh.org/)

Download the source from the website and build with the following script:

``` sh
cd libssh-<version>/
mkdir build && cd build
cmake \
    -DWITH_STATIC_LIB=ON \
    -DWITH_GSSAPI=OFF \
    -DWITH_GCRYPT=ON \
    -DWITH_SERVER=ON \
    ..
make
sudo make install
```

### 2. `libgcrypt` and `libgpg-error`

Install them from your package manager:

``` sh
sudo apt install libgcrypt20-dev libgpg-error-dev
```

### Compile

To compile the program, simply run `./build.sh` in your terminal.

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
sudo ./honeyshell -user nobody -port 2222 -key ~/.ssh/id_rsa
```

The output should look something like this:

```
Starting on port 2222
Changing permissions to user 'nobody'
192.168.0.5 35034 connection request
192.168.0.5 35034 client connected with SSH-2.0-OpenSSH_7.6p1 Ubuntu-4
192.168.0.5 admin password
192.168.0.5 admin password123
192.168.0.5 connection terminated
192.168.0.5 35036 connection request
192.168.0.5 35036 client connected with SSH-2.0-OpenSSH_7.6p1 Ubuntu-4
192.168.0.5 admin test
192.168.0.5 admin test123
192.168.0.5 connection terminated
192.168.0.200 45250 connection request
192.168.0.200 45250 client connected with SSH-2.0-OpenSSH_7.4p1 Raspbian-10+deb9u3
192.168.0.200 admin password
192.168.0.200 admin mindy
192.168.0.200 admin testytest
192.168.0.200 connection terminated
```
