up:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up

buildup:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up --build

down:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env down

logs:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env logs server

psqlup:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d postgres
  
filldb:
	go run scripts/fill/db_filler.go

cleandb:
	go run scripts/clean/db_cleaner.go

genjson:
	easyjson -all internal/entities/beer.go
	easyjson -all internal/entities/category.go
	easyjson -all internal/entities/review.go
	easyjson -all internal/entities/enum_class.go
	easyjson -all internal/entities/enum_value.go
	easyjson -all internal/entities/aggregate.go
	easyjson -all internal/entities/parameter.go

genmock:
	go generate ./internal/http/handlers/mocks

genapi:
	redocly bundle api/services.swagger.json --output api/gened.swagger.json

lint:
	golangci-lint run

test:
	go test ./...

hook:
	pre-commit install --hook-type pre-commit --hook-type pre-push

check_building:
	cp -n configs/env.example configs/.env || true
	mkdir -p deployments/pgdata
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d --build
	docker compose -f deployments/docker-compose.yml --env-file configs/.env down

check_fmt:
	@go fmt ./...
	@git diff --exit-code --quiet


check_mod:
	@go mod tidy
	@if ! git diff --exit-code --quiet go.mod go.sum; then \
		git add go.mod go.sum; \
		git commit --amend --no-edit; \
		echo "Зависимости обновлены и добавлены в коммит!"; \
	fi