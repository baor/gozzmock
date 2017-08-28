MAINTAINER Travix

# Build stage
FROM golang:1.8 as builder

COPY ../gozzmock /go/src/

WORKDIR /go/src/gozzmock

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gozzmock .

# Run stage
FROM alpine:3.6

WORKDIR /root/

COPY --from=builder /go/src/gozzmock .

EXPOSE 8080

CMD ["./gozzmock"]

#FROM scratch
#ADD ca-certificates.crt /etc/ssl/certs/
#ADD main /