# Build stage
FROM golang:1.9 as builder

MAINTAINER Travix

COPY ./ /go/src/gozzmock

WORKDIR /go/src/gozzmock

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock_bin .

# Run stage
FROM scratch

MAINTAINER Travix

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/gozzmock/gozzmock_bin .

EXPOSE 8080

ENTRYPOINT ["./gozzmock_bin"]
