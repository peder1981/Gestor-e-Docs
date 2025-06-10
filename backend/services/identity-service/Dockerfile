# Etapa de build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Limpar cache de módulos Go explicitamente
RUN go clean -modcache

# Copiar arquivos de dependências e baixar dependências
COPY go.mod go.sum* ./
RUN go mod download

# Copiar o código fonte
COPY . .

# Garantir que go.mod e go.sum estejam corretos ANTES do build,
# baseado em todas as importações no código copiado:
RUN go mod tidy

# Construir a aplicação
# CGO_ENABLED=0 para um binário estático, GOOS=linux para cross-compilation (se necessário, mas alpine é linux)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o identity-service .

# Etapa final - imagem leve
FROM alpine:latest

WORKDIR /root/

# Copiar o binário construído da etapa anterior
COPY --from=builder /app/identity-service .

# Copiar quaisquer outros assets necessários (ex: templates, configs), se houver
# COPY --from=builder /app/templates ./templates

# Expor a porta que a aplicação vai rodar
# A porta real é definida pela variável de ambiente SERVICE_PORT ou padrão 8085 no código
# O docker-compose mapeia a porta do host para esta porta do container
EXPOSE 8085

# Comando para rodar a aplicação
CMD ["./identity-service"]
