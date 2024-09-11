FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk update --no-cache

WORKDIR /build
COPY src/go.mod .
COPY src/go.sum .

RUN go mod download

COPY src/ .

RUN go build -ldflags="-s -w" -o /app/service

RUN go build -o service main.go


FROM alpine:3.19.1

WORKDIR /app

EXPOSE 8080

COPY --from=builder /app/service /app/service
CMD ["./service"]