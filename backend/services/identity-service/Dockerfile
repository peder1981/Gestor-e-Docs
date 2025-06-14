# Etapa 1: Builder
FROM golang:1.24.2 AS builder

WORKDIR /app

# Copiar os arquivos de módulo primeiro para aproveitar o cache do Docker
COPY go.mod go.sum ./

# Baixar as dependências
RUN go mod download

# Copiar todo o código backend
COPY backend/ ./

# Garantir que o go.sum está perfeitamente sincronizado com o código.
# Isso resolve os erros de "missing go.sum entry".
RUN go mod tidy

# Compilar o binário principal do serviço
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/identity-service ./services/identity-service/main.go

# Voltar para o WORKDIR /app se precisar fazer mais algo no root do builder (opcional aqui)
WORKDIR /app

# Etapa 2: Imagem final leve
FROM alpine:latest

WORKDIR /app/

# Copiar os binários da etapa de build
COPY --from=builder /app/identity-service .

ENV PORT=8085
ENV JWT_SECRET_KEY=your-secret-key
ENV MONGODB_URI=mongodb://mongo_db:27017
ENV MONGODB_DATABASE=gestor_e_docs
COPY backend/services/identity-service/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expor a porta do serviço
EXPOSE 8085

# Definir o entrypoint para executar o script de inicialização
ENTRYPOINT ["./entrypoint.sh"]
