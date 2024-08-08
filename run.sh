#!/bin/fish
go build --trimpath --buildmode=plugin -o ./modules/backend.so

docker -D compose --progress=plain up -d
docker -D compose --progress=plain restart