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

genmock:
	go generate ./internal/http/handlers/mocks

genapi:
	redocly bundle api/services.swagger.json --output api/gened.swagger.json

lint:
	golangci-lint run

test:
	go test ./...

hook:
	pip install pre-commit
	pre-commit install --hook-type pre-push

check_building:
	cp -n configs/env.example configs/.env || true
	mkdir -p deployments/pgdata
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d --build
	docker compose -f deployments/docker-compose.yml --env-file configs/.env down

check_fmt:
	@go fmt ./...
	@if ! git diff --exit-code --quiet; then \
		git add -u; \
		echo "✨ Код автоматически отформатирован и добавлен в коммит!"; \
	fi