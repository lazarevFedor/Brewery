# Brewery

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26.1-blue.svg)](#prerequisites)

Microbrewery management service that allows you to work with categories, beers, reviews, enum classes and enum values.

## Features

- category hierarchy
- beers CRUD and listing by category
- beer features and reviews
- enum classes and enum values

## Tech stack

**Backend:** Go, gin

**QA**: Go, minimock, testify, testing

**Data**: PostgreSQL, pgx, goose migrations

**Infrastructure**: Docker, Nginx, GitHub Actions

**SRE/Observability**: Prometheus, Grafana, OpenAPI spec

## Prerequisites

- Docker & Docker Compose
- Go 1.26.1 (for local development)
- Make (optional, but recommended)

## Configuration

Create `configs/.env` from the example file:

```bash
cp configs/env.example configs/.env
```

Main variables:

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `SERVER_PORT`
- `GF_PORT`, `GF_PASSWORD`
- `PROMETHEUS_PORT`, `NODE_EXPORTER_PORT`

## Quick start

1. Start the application:

```bash
make buildup
```

2. Fill the database with test data:

```bash
make filldb
```

3. Try the API:

Get all categories:
```bash
curl http://localhost:80/api/categories
```

Get all beers:
```bash
curl http://localhost:80/api/beers
```

Create a new category:
```bash
curl -X POST http://localhost:80/api/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Pilsner","parent_id":null}'
```

For full API details, see [`api/gened.swagger.json`](api/gened.swagger.json).

**Note:** For local development without Docker, use `http://localhost:$SERVER_PORT` instead of `http://localhost:80`.

## Run application

```bash
make buildup
```

or, without rebuild:

```bash
make up
```

The API will be available through Nginx at `http://localhost:80/api`.

By default, the server listens on the port from `SERVER_PORT` (see `configs/env.example`).

## Migrations

Database migrations are stored in `migrator/sql` and are embedded into the binary.
Use the helper scripts to run them against PostgreSQL.

Apply all migrations and fill the database with test data:
```bash
make filldb
```

## Testing

Run all tests:
```bash
go test ./...
```

Generate mocks:
```bash
make genmock
```

## API documentation

- OpenAPI spec: [`api/gened.swagger.json`](api/gened.swagger.json)
- If you regenerate the spec, use:

```bash
make genapi
```

## Monitoring

**Prometheus:** http://localhost:9090
- Metrics endpoint: `http://localhost:80/metrics` (via Nginx)

**Grafana:** http://localhost:3000
- Default credentials: `admin` / (see `GF_PASSWORD` in `.env`)
- Datasource: Prometheus (pre-configured)

## Repository structure

- `.github/` — GitHub Actions workflows
- `api/` — OpenAPI spec
- `build/` — Docker images
- `cmd/` — application entry point
- `configs/` — configuration files and examples
- `deployments/` — Docker Compose, Nginx, Prometheus and Grafana configs
- `internal/` — handlers, use cases, repositories, entities
- `migrator/sql/` — DB migrations
- `pkg/` — shared packages, e.g. for postgres connection and logging
- `scripts/` — DB fill and clean helpers

## Releases

### v1.1.0

**New features:**
- Enum classes and enum values management (create, read, update, delete)
- Extended beer features (beer_features) operations
- Extended reviews (reviews) operations

**Fixes:**
- Various bug fixes and stability improvements

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
