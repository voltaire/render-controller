FROM golang:1.15-alpine AS builder
ADD . /src/
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bin/server

FROM gcr.io/distroless/static
COPY --from=builder /bin/server /bin/server

ENTRYPOINT ["/bin/server"]