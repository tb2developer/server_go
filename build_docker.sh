#!/bin/bash

#CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/server_go ./src/main.go

docker build -t server_go -f Dockerfile.scratch .