FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

RUN apk add --no-cache ca-certificates && \
    chmod +x /app/main

EXPOSE 3000

CMD ["/app/main"]