# --- build stage ---
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscription-service ./cmd/app

# --- runtime ---
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/subscription-service /app/subscription-service
COPY configs/.env.example /app/configs/.env
COPY docs /app/docs
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/subscription-service"]