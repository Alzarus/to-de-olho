# Etapa de build
FROM golang:1.23 AS builder
WORKDIR /app

# Copia arquivos necessários primeiro para otimizar cache
COPY go.mod go.sum ./
RUN go mod download

# Copia código-fonte
COPY . .

# Define variáveis de ambiente para build otimizado
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Compila o binário
RUN go build -o main .

# Etapa final: usar 'scratch' para imagem ultraleve
FROM scratch
WORKDIR /app

# Copia certificados para conexões HTTPS funcionarem
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copia o binário compilado
COPY --from=builder /app/main .

# Expor a porta, se necessário
EXPOSE 3000

# Define entrada do container
ENTRYPOINT ["/app/main"]
