# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/shop ./cmd/shop

# Stage 2: Create a minimal image
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /bin/shop /app/shop
COPY --from=builder /app/swagger /app/swagger

EXPOSE 8080

CMD ["/app/shop"]
