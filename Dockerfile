FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY src  /app/
RUN go mod download
RUN go build -o service main.go

EXPOSE 8080


FROM alpine:3.19.1
COPY --from=builder /app/service /app/service
ENTRYPOINT ["/app/service"]
