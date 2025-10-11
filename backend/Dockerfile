# Build stage
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação (entrypoint em cmd/server)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ingestor ./cmd/ingestor && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler ./cmd/scheduler

# Production stage
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN adduser -D -s /bin/sh appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binários da aplicação
COPY --from=builder /app/server ./server
COPY --from=builder /app/ingestor ./ingestor
COPY --from=builder /app/scheduler ./scheduler

# Copiar scripts de inicialização
COPY scripts/smart-init.sh ./smart-init.sh
RUN chmod +x ./smart-init.sh

# Instalar curl para health checks
RUN apk add --no-cache curl

# Mudar ownership para usuário não-root
RUN chown -R appuser:appuser /app

# Mudar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Comando para executar a aplicação
CMD ["./server"]
