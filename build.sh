#!/bin/bash

echo "golint validations..."
golint .
golint ./controller/...
golint ./handler/...
golint ./model/...

echo "go clean ..."
go clean -i ./...

echo "go get ..."
go get ./...

echo "go build ..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock_bin .

echo "go test and coverage..."
go test -coverprofile=cover_controller.out ./controller
go test -coverprofile=cover_handler.out ./handler

go tool cover -html=cover_controller.out -o cover_controller.html
go tool cover -html=cover_handler.out -o cover_handler.html

echo "Done!"