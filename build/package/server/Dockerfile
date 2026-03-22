FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY ../../../go.mod ../go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY .. .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o server ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/server ./

CMD ["./server"]