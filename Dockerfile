FROM golang:1.11.1-alpine as builder
WORKDIR /go/src/github.com/fankserver/torchapi-hive-system
COPY . .
RUN apk add --no-cache pcre-dev alpine-sdk \
    && go get ./... \
    && go build -o app .

FROM alpine:latest
RUN adduser -D -u 1000 hive
USER hive

# Add app
COPY --from=builder /go/src/github.com/fankserver/torchapi-hive-system/app /app

# This container will be executable
ENTRYPOINT ["/app"]
