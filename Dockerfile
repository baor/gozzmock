# Build stage
FROM golang:1.8.3 as builder

MAINTAINER Travix

COPY ./ /go/src/gozzmock

WORKDIR /go/src/gozzmock

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock_bin .

# Run stage
FROM scratch

COPY ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/gozzmock/gozzmock_bin .

EXPOSE 8080

CMD ["./gozzmock_bin"]
