FROM golang:1.24.2 AS builder

WORKDIR /app

# Copiar os arquivos de módulo primeiro para aproveitar o cache do Docker
COPY go.mod go.sum ./

# Baixar as dependências
RUN go mod download

# Copiar todo o código do serviço
COPY . ./

# Garantir que o go.sum está perfeitamente sincronizado com o código
RUN go mod tidy

# Compilar o binário principal do serviço
RUN CGO_ENABLED=0 GOOS=linux go build -o document-service .

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
