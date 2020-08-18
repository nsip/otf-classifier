###########################
# INSTRUCTIONS
############################
# BUILD: docker build -t nsip/otf-classifier:develop .
# TEST: docker run -it -p1576:1576 nsip/otf-classifier:develop .
# RUN: docker run -d -p1576:1576 nsip/otf-classifie:developr
#
###########################
# DOCUMENTATION
############################

###########################
# STEP 0 Get them certificates
############################
FROM alpine:latest as certs
RUN apk --no-cache add ca-certificates

############################
# STEP 1 build executable binary (go.mod version)
############################
FROM golang:1.14-stretch as builder
RUN mkdir -p /build
WORKDIR /build
COPY . .
WORKDIR cmd/otf-classifier
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/server

############################
# STEP 2 build a small image
############################
FROM debian:stretch
COPY --from=builder /go/bin/server /go/bin/server
# NOTE - make sure it is the last build that still copies the files
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/cmd/otf-classifier/curricula/* /data/curricula/
WORKDIR /data
CMD ["/go/bin/server"]




