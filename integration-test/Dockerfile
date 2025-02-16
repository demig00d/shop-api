# Кэширование зависимостей
FROM golang:1.23-alpine AS modules
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Запуск интеграционных тестов
FROM golang:1.23-alpine
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]

