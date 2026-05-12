FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# модуль в корне проекта
COPY go.mod go.sum ./
RUN go mod download

# копируем весь проект (или только нужное)
COPY . .

# собираем из каталога backend/cmd/server
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./cmd/server
RUN mkdir -p /app/storage

FROM scratch

COPY --from=builder /app/main /app/main
COPY --from=builder /app/storage /app/storage

EXPOSE 8080
CMD ["/app/main"]