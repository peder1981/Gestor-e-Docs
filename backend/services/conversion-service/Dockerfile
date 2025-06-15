# Build stage
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

# Compilar o serviço
RUN CGO_ENABLED=0 GOOS=linux go build -o conversion-service .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copiar o binário compilado
COPY --from=builder /app/conversion-service .

# Expor a porta do serviço
EXPOSE 8285

# Executar o serviço
CMD ["./conversion-service"]
