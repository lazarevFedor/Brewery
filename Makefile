up:
	docker compose -f deployments/docker-compose.yml --env-file configs/.env up

down:
	docker compose -f ./docker/docker-compose.yml down -v

lint:
	golangci-lint run