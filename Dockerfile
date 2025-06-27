FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY config/config.yaml config/config.yaml
EXPOSE 8080
CMD ["./app"] 