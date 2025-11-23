FROM golang:1.23-alpine AS builder

WORKDIR /app

# Модули
COPY go.mod go.sum ./
RUN go mod download

# Остальной код
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/avito-pr-reviewer ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/avito-pr-reviewer .

# Порт сервиса
ENV APP_PORT=8080

EXPOSE 8080

CMD ["./avito-pr-reviewer"]
