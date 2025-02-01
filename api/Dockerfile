# Etapa de build
FROM golang:1.23 AS builder
WORKDIR /app

# Copiando arquivos necessários para o build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Definição das flags de compilação para binário mais otimizado
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o main ./cmd

FROM alpine:latest
WORKDIR /root/

# Copia certificados para conexões HTTPS funcionarem
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copia o binário da aplicação
COPY --from=builder /app/main .

# Define porta exposta
EXPOSE 3000

# Define entrada do container
ENTRYPOINT ["./main"]
