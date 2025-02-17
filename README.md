![workflow](https://github.com/demig00d/shop-api/actions/workflows/ci.yml/badge.svg)
![coverage](https://raw.githubusercontent.com/demig00d/shop-api/badges/.badges/master/coverage.svg)

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

**Интеграционные тесты:**

```sh
make compose-up-integration-test
```
