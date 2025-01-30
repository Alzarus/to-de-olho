# Etapa de build
FROM golang:1.23-alpine AS builder

# Instala apenas as dependências necessárias para a compilação
RUN apk add --no-cache git

WORKDIR /app

# Copia e baixa dependências antes do código-fonte para otimizar cache
COPY go.mod go.sum ./
RUN go mod download

# Copia código-fonte
COPY . .

# Compila o binário de forma otimizada
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o validate-user-worker .

# Etapa final: usa scratch para um container minimalista
FROM scratch

WORKDIR /app

# Copia certificados para permitir conexões HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copia o binário compilado
COPY --from=builder /app/validate-user-worker .

# Define o ponto de entrada
ENTRYPOINT ["/app/validate-user-worker"]
