wsl sudo go build --trimpath --buildmode=plugin -o ./modules/backend.so

docker -D compose --progress=plain up