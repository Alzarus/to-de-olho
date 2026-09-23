# Build Stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Instalar dependencias do sistema necessarias para build (se houver)
# RUN apk add --no-cache git

# Copiar arquivos de dependencia
COPY go.mod go.sum ./
RUN go mod download

# Copiar codigo fonte
COPY . .

# Compilar binario estatico otimizado
# CGO_ENABLED=0 garante que nao dependa de libc dinamica
# -ldflags="-s -w" remove simbolos de debug para reduzir tamanho
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /force_sync cmd/force_sync_all/main.go

# Runtime Stage
# Tag explicita (nao :latest). "nonroot" ja traz o usuario usado abaixo.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copiar binario do stage anterior
COPY --from=builder /server /server
COPY --from=builder /force_sync /force_sync

# Expor porta
EXPOSE 8080

# Usuario nao-root para seguranca
USER nonroot:nonroot

# Sem shell na imagem: o proprio binario checa GET /health (cmd/api/main.go).
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD ["/server", "healthcheck"]

# Comando de entrada
ENTRYPOINT ["/server"]
