#!/usr/bin/env bash

CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build -o target/rdb-linux-amd64 ./
go build -o target/rdb-darwin ./