###########################
# INSTRUCTIONS
############################
# BUILD
#   Confirm v0.1.0 tag has been commited to master, and checkout at that point
#	docker build -t nsip/otf-classifier:latest -t nsip/otf-classifier:v0.1.0 .
# TEST: docker run -it -v $PWD/test/data:/data -v $PWD/test/config.json:/config.json nsip/otf-classifier:develop .
# RUN: docker run -d nsip/otf-classifier:develop
#
# PUSH
#	Public:
#		docker push nsip/otf-classifier:v0.1.0
#		docker push nsip/otf-classifier:latest
#
#	Private:
#		docker tag nsip/otf-classifier:v0.1.0 the.hub.nsip.edu.au:3500/nsip/otf-classifier:v0.1.0
#		docker tag nsip/otf-classifier:latest the.hub.nsip.edu.au:3500/nsip/otf-classifier:latest
#		docker push the.hub.nsip.edu.au:3500/nsip/otf-classifier:v0.1.0
#		docker push the.hub.nsip.edu.au:3500/nsip/otf-classifier:latest
#
###########################
# DOCUMENTATION
############################

###########################
# STEP 0 Get them certificates
############################
# (note, step 2 is using alpine now)
# FROM alpine:latest as certs

############################
# STEP 1 build executable binary (go.mod version)
############################
FROM golang:1.15.0-alpine3.12 as builder
RUN apk --no-cache add ca-certificates
RUN apk update && apk add git
RUN apk add gcc g++
RUN mkdir -p /build
WORKDIR /build
COPY . .
WORKDIR cmd/otf-classifier
RUN go build -o /build/app

############################
# STEP 2 build a small image
############################
#FROM debian:stretch
FROM alpine
COPY --from=builder /build/app /app
# NOTE - make sure it is the last build that still copies the files
CMD ["./app"]
WORKDIR /
ENTRYPOINT ["./app"]
