# --- build stage ---
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0: both glebarez/sqlite and jackc/pgx are pure Go — no cgo needed
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /deckforge ./cmd/server

# --- run stage ---
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /deckforge .
EXPOSE 8080
ENTRYPOINT ["/app/deckforge"]
