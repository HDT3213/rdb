#!/usr/bin/env bash

CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build -o target/rdb-linux-amd64 ./
CGO_ENABLED=0  GOOS=darwin GOARCH=amd64 go build -o target/rdb-darwin-amd64 ./
CGO_ENABLED=0  GOOS=darwin GOARCH=arm64 go build -o target/rdb-darwin-arm64 ./
CGO_ENABLED=0  GOOS=windows GOARCH=amd64 go build -o target/rdb-windows-amd64.exe ./