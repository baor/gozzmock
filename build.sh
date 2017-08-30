#!/bin/bash


echo "go clean ..."
go clean -i ./...

echo "go get ..."
go get ./...

echo "go build ..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock_bin .

echo "go test ..."
go test ./...

#set -e

#echo "Retrieving/updating vendor packages..."

#gvt update google.golang.org/api/pubsub/v1 || true

echo "Done!"