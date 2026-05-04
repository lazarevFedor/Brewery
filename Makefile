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
	cp .githooks/pre-push .git/hooks