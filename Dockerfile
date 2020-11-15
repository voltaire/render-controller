FROM golang:1.15.5-alpine3.12 AS builder
ADD . /src/
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bin/server

FROM alpine:3.12 AS unzipper
ADD https://github.com/linode/docker-machine-driver-linode/releases/download/v0.1.8/docker-machine-driver-linode_linux-amd64.zip /root/docker-machine-driver-linode.zip
RUN apk add --no-cache unzip
WORKDIR /root
RUN unzip /root/docker-machine-driver-linode.zip

FROM gcr.io/distroless/static
EXPOSE 8080
COPY --from=unzipper /root/docker-machine-driver-linode /usr/bin/docker-machine-driver-linode
COPY --from=builder /bin/server /bin/server

ENTRYPOINT ["/bin/server"]