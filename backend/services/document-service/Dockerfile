FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar ferramentas necessárias
RUN apk add --no-cache git gcc musl-dev ca-certificates

# Configurar para ignorar verificações de integridade
ENV GOSUMDB=off
ENV GO111MODULE=on
ENV GOPROXY=direct

# Copiar todo o código-fonte (importante: copie antes de manipular dependências)
COPY . .

# Gerar go.mod do zero (ignorando possíveis problemas existentes)
RUN rm -f go.mod go.sum
RUN go mod init gestor-e-docs/document-service

# Forçar adição das dependências necessárias com versões específicas
RUN go get github.com/golang-jwt/jwt/v5@v5.0.0
RUN go get github.com/gin-gonic/gin@v1.9.1
RUN go get github.com/gin-contrib/cors@v1.4.0
RUN go get github.com/minio/minio-go/v7@v7.0.61
RUN go get go.mongodb.org/mongo-driver/mongo@v1.12.0
RUN go get github.com/google/uuid@v1.3.0

# Desativar a verificação de módulos completamente
RUN go env -w GOFLAGS="-mod=mod"

# Compilar com flags para ignorar erros menos graves
RUN CGO_ENABLED=0 GOOS=linux go build -o document-service . || \
    # Se falhar, tente um build mais agressivo ignorando mais erros
    CGO_ENABLED=0 GOOS=linux go build -tags=nomsgpack -o document-service .

# Imagem final
FROM alpine:latest

WORKDIR /app

# Instalar CA certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Copiar o binário compilado
COPY --from=builder /app/document-service .

# Expor a porta da aplicação
EXPOSE 8185

# Definir comando de inicialização
CMD ["./document-service"]
