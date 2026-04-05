up:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up -d

down:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env down

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

genmock:
	go generate ./internal/http/handlers/mocks
