#!/bin/bash

if [ $# -gt 0 ] && [ "$1" == "termux" ]; then
	export CGO_LDFLAGS="-L/system/lib64 -L/data/data/com.termux/files/lib -L/data/data/com.termux/files/usr/lib -lssh -lgcrypt -lgpg-error -lz"
	export CGO_CFLAGS="-I/usr/include -I/usr/local/include -I/data/data/com.termux/files/usr/include -I/data/data/com.termux/files/usr/local/include"
else
	export CGO_LDFLAGS="-L/usr/lib -L/usr/local/lib -lssh -lgcrypt -lgpg-error -lz"
	export CGO_CFLAGS="-I/usr/include -I/usr/local/include"
fi

echo $CGO_LDFLAGS
echo $CGO_CFLAGS

go build .

#go build -ldflags "-linkmode external -extldflags -static" -o honeyshell-c .
