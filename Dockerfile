FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/main.go

FROM alpine:latest
# Устанавливаем timezone для московского времени
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo "Europe/Moscow" > /etc/timezone && \
    apk del tzdata

WORKDIR /app
COPY --from=builder /app/app .
COPY config/config.yaml config/config.yaml
EXPOSE 8080
CMD ["./app"] 