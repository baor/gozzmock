go clean -i ./...
go get ./...
go build ./...
go test ./...

#!/bin/bash

set -e

echo "Retrieving/updating vendor packages..."

gvt update google.golang.org/api/pubsub/v1 || true
gvt update golang.org/x/oauth2/ || true
gvt update github.com/rs/cors || true
gvt update github.com/julienschmidt/httprouter || true
gvt update google.golang.org/cloud || true

echo "Building application..."

CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s" -o frogger main.go

echo "Done!"