# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copiar arquivos necessários
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY main.go .
COPY handlers/ handlers/
COPY metrics/ metrics/

# Compilar o serviço
RUN CGO_ENABLED=0 GOOS=linux go build -o conversion-service

# Final stage
FROM alpine:latest

WORKDIR /app

# Copiar o binário compilado
COPY --from=builder /app/conversion-service .

# Expor a porta do serviço
EXPOSE 8285

# Executar o serviço
CMD ["./conversion-service"]
