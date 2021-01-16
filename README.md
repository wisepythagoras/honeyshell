# Honeyshell

An SSH honeypot based on the [libssh](https://www.libssh.org/) library written entirely in Go.

Currently, it allows connections on the server and collects failed login attempts, meaning all usernames and passwords. Also, you can decrease the permissions on the process by flipping to a different user (ie 'nobody').

## Building

There are three dependencies, which are the following:

### 1. `libgcrypt` and `libgpg-error`

Install them from your package manager:

``` sh
sudo apt install libgcrypt20-dev libgpg-error-dev
```

### 2. [libssh](https://www.libssh.org/)

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
127.0.0.1 60812 connection request
127.0.0.1 60812 client connected with SSH-2.0-OpenSSH_7.4p1 Raspbian-10+deb9u6
127.0.0.1 admin 216247142fed250e4c5bfdfe1af2262a8000f0581f1f6bd20509cd49d542a27a
127.0.0.1 admin 3bc98ba6299bcb10d0eb185be884a12ee35e5eb311edb77fe99f36e92fbba603
127.0.0.1 admin 1f671129e0ca2917e24809271109e20378f4f50de41bfa9f5b578056535f64e1
127.0.0.1 admin password123
127.0.0.1 admin passwordtest
127.0.0.1 connection terminated
```
