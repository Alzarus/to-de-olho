FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o validate-user-worker .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/validate-user-worker .

CMD ["./validate-user-worker"]
