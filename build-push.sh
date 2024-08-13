#!/bin/fish
go build -v -x --trimpath --buildmode=plugin -o ./modules/backend.so
scp modules/backend.so root@172.233.149.134:~/graven/modules