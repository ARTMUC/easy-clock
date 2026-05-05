FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/easy-clock ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/easy-clock .
COPY --from=builder /app/static ./static
RUN mkdir -p ./static/uploads
EXPOSE 8080
CMD ["./easy-clock"]
