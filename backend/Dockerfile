FROM golang:1.25.0-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server \
    && mkdir -p /app/storage

FROM scratch

COPY --from=builder /app/main /app/main
COPY --from=builder /app/storage /app/storage

# USER 1000:1000

EXPOSE 8080

CMD ["/app/main"]