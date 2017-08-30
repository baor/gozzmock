#!/bin/bash

echo "golint validations..."
golint .

echo "go clean ..."
go clean -i ./...

echo "go get ..."
go get ./...

echo "go build ..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock_bin .

echo "go test and coverage..."
go test -coverprofile=cover_main.out .
go tool cover -html=cover_main.out -o cover_main.html

echo "Done!"