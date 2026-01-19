# Build Stage
FROM golang:1.21-alpine AS builder

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

# Runtime Stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copiar binario do stage anterior
COPY --from=builder /server /server

# Expor porta
EXPOSE 8080

# Usuario nao-root para seguranca
USER nonroot:nonroot

# Comando de entrada
ENTRYPOINT ["/server"]
