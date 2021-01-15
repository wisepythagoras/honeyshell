#!/bin/bash

export CGO_LDFLAGS="-L/usr/lib -L/usr/local/lib -lssh -lgcrypt -lgpg-error -lz"
export CGO_CFLAGS="-I/usr/include -I/usr/local/include"

go build .

#go build -ldflags "-linkmode external -extldflags -static" -o honeyshell-c .
