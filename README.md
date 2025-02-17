## Запуск

```sh
docker compose up --build -d
```

или

```sh
make compose-up
```

## Тесты

**Unit-тесты:**

```sh
go test -v -race ./internal/...
```

**Итнеграционнае тесты:**

```sh
make compose-up-integration-test
```
