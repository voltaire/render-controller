FROM golang:1.15.5-alpine3.12 AS builder
ADD . /src/
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o /bin/main ./cmd/render-edge

FROM gcr.io/distroless/static
EXPOSE 8080
COPY --from=builder /bin/main /bin/main

ENTRYPOINT ["/bin/main"]