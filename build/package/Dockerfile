FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/main.go

FROM debian:bookworm-slim

COPY --from=builder /app/server .

CMD ["./app"]