up:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up --build 

down:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env down

logs:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env logs

lint:
	golangci-lint run

monitor:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d --build prometheus
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d --build node_exporter
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d --build grafana

genjson:
	easyjson -all internal/entities/beer.go
	easyjson -all internal/entities/category.go
	easyjson -all internal/entities/review.go

psqlup:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d postgres
  
genmock:
	go generate ./internal/http/handlers/mocks

filldb:
	go run script/db_filler.go
