# Stage 1: build
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app ./cmd/main.go

# Stage 2: runtime
FROM alpine:latest

RUN apk add --no-cache tzdata \
 && cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime \
 && echo "Europe/Moscow" > /etc/timezone \
 && apk del tzdata

WORKDIR /app

COPY --from=builder /app/app .
COPY config/config.dev.yaml ./config/config.dev.yaml

EXPOSE 8080
CMD ["./app"]