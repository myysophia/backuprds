# syntax=docker/dockerfile:1

# Step 1: build golang binary
FROM golang:1.22 as builder
WORKDIR /opt/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-w" -o backuprds

# Step 2: copy binary from step1
FROM alpine:latest
ENV PATH /usr/local/bin:$PATH
ENV LANG C.UTF-8

ENV TZ=Asia/Shanghai

RUN apk update && apk upgrade \
    && apk add ca-certificates\
    && update-ca-certificates \
    && apk --no-cache add openssl wget \
        && apk add --no-cache bash tzdata curl \
        && set -ex \
    && mkdir -p /usr/bin \
    && mkdir -p /usr/sbin \
    && mkdir -p /data/backuprds/

COPY --from=builder /opt/app/backuprds /usr/bin/backuprds
ADD start.sh /data/backuprds/
WORKDIR /data/backuprds/

