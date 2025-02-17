export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Показать help
	@awk 'BEGIN {FS = ":.*##"; printf "\nИспользование:\n  make \033[36m<цель>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Запустить docker-compose
	docker compose up --build -d
.PHONY: compose-up

compose-up-integration-test: ### Запустить docker-compose с интеграционным тестом
	docker compose -f docker-compose.integration.yml up --build --abort-on-container-exit --exit-code-from integration
.PHONY: compose-up-integration-test

compose-down: ### Остановить docker-compose
	docker compose down --remove-orphans
.PHONY: compose-down

linter-golangci: ### Проверить с помощью golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### Проверить с помощью hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

test: ### Запустить тесты
	go test -v -cover -race ./internal/...
.PHONY: test

mock: ### Запустить mockgen
	mockgen -source=./internal/db/db.go -destination=./internal/db/mocks/db_mock.go -package=mocks
	mockgen -source=./internal/usecase/transaction.go -destination=./internal/usercase/mocks/transaction_mock.go -package=mocks
	mockgen -source=./internal/usecase/item.go -destination=./internal/usercase/mocks/item_mock.go -package=mocks
	mockgen -source=./internal/usecase/user.go -destination=./internal/usercase/mocks/user_mock.go -package=mocks
.PHONY: mock


.PHONY: check-coverage
check-coverage: 
	go test ./internal/... -coverprofile=./cover_unit.out -covermode=atomic -coverpkg=./...
	go-test-coverage --config=./.testcoverage.yml

bin-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest
	GOBIN=$(LOCAL_BIN) go install github.com/vladopajic/go-test-coverage/v2@latest
