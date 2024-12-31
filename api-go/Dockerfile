FROM golang:1.23 AS builder
WORKDIR /app

COPY . .

WORKDIR /app/cmd

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /app/cmd/main .

EXPOSE 3000

CMD ["./main"]
