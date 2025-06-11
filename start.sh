#!/bin/bash

# Script de inicialização para Gestor-e-Docs
# Autor: Equipe Gestor-e-Docs
# Data: 10/06/2025
# Versão: 1.1

# Cores para melhor visualização
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Iniciando Gestor-e-Docs ===${NC}"
echo -e "${YELLOW}Verificando pré-requisitos...${NC}"

# Limpar variáveis de ambiente do Docker que possam causar conflitos
unset DOCKER_HOST
unset DOCKER_TLS_VERIFY
unset DOCKER_CERT_PATH

# Verificar se o Docker está instalado
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker não está instalado! Por favor, instale o Docker antes de continuar.${NC}"
    echo "Visite: https://docs.docker.com/get-docker/"
    exit 1
fi

# Verificar se o Docker está rodando (usando opção simplificada)
echo "Verificando status do Docker..."
if ! docker ps > /dev/null 2>&1; then
    echo -e "${RED}Não foi possível conectar ao daemon do Docker!${NC}"
    echo "Possíveis soluções:"
    echo "1. Verifique se o serviço Docker está em execução:"
    echo "   - Em sistemas Linux, execute: sudo systemctl start docker"
    echo "   - No Mac ou Windows, inicie o aplicativo Docker Desktop"
    echo "2. Verifique se seu usuário tem permissão para acessar o Docker:"
    echo "   - Execute: sudo usermod -aG docker $USER"
    echo "   - Depois faça logout e login novamente ou execute: newgrp docker"
    exit 1
fi

# Determinar qual comando do Docker Compose usar
DOCKER_COMPOSE=""
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
    echo "Usando docker-compose standalone"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
    echo "Usando docker compose plugin"
else
    echo -e "${RED}Docker Compose não está instalado! Por favor, instale o Docker Compose antes de continuar.${NC}"
    echo "Visite: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}Todos os pré-requisitos estão satisfeitos!${NC}"

# Criar um arquivo .env se não existir
if [ ! -f .env ]; then
    echo "Criando arquivo .env com configurações padrão..."
    cat > .env << EOF
# Configurações do MongoDB
MONGO_URI=mongodb://mongo_db:27017/gestor_e_docs
MONGO_DB_NAME=gestor_e_docs

# Configurações de segurança
JWT_SECRET_KEY=seuSuperSegredoMuitoComplexoAqui

# Configurações do MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
MINIO_ENDPOINT=minio_server:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET_NAME=documents

# Portas dos serviços - Todas terminadas em 85 conforme padronização
IDENTITY_SERVICE_PORT=8085
DOCUMENT_SERVICE_PORT=8185
FRONTEND_PORT=3085
GRAFANA_PORT=3385
KIBANA_PORT=5685
ELASTICSEARCH_PORT=9285
PROMETHEUS_PORT=9385
MINIO_API_PORT=9085
MINIO_CONSOLE_PORT=9185
FLUENTD_PORT=24285
MONGO_PORT=27185

# Configurações de monitoramento
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=gestor_e_docs_admin
EOF
    echo -e "${GREEN}Arquivo .env criado com sucesso!${NC}"
    echo -e "${YELLOW}ATENÇÃO: Em ambiente de produção, altere os valores de senha/chaves no arquivo .env${NC}"
fi

# Executar o docker-compose com comando adequado
echo -e "${YELLOW}Iniciando os serviços com $DOCKER_COMPOSE...${NC}"

# Tentar executar o comando, lidando com erros comuns
if ! $DOCKER_COMPOSE up -d --build; then
    echo -e "${RED}Falha ao iniciar os serviços com $DOCKER_COMPOSE!${NC}"
    
    # Verificação alternativa para Docker em execução
    if [ "$DOCKER_COMPOSE" = "docker-compose" ]; then
        echo "Tentando método alternativo com 'docker compose'..."
        if docker compose up -d --build; then
            echo -e "${GREEN}Serviços iniciados com sucesso usando 'docker compose'!${NC}"
            DOCKER_COMPOSE="docker compose"
        else
            echo -e "${RED}Todos os métodos falharam. Verifique a instalação do Docker e Docker Compose.${NC}"
            echo "Dica: Execute 'docker info' para verificar a conexão com o daemon do Docker"
            exit 1
        fi
    else
        exit 1
    fi
fi

# Verificar se os containers estão rodando
echo -e "${YELLOW}Verificando status dos serviços...${NC}"
sleep 5 # Dar tempo para os containers iniciarem

# Use docker diretamente em vez de docker-compose para verificar containers
CONTAINER_COUNT=$(docker ps --filter "name=identity_service|document_service|frontend_app|mongo_db|minio_server" --format "{{.Names}}" | wc -l)

if [ "$CONTAINER_COUNT" -gt 0 ]; then
    echo -e "${GREEN}Serviços iniciados com sucesso! Containers em execução: $CONTAINER_COUNT${NC}"
    
    # Exibir URLs de acesso
    echo -e "\n${GREEN}=== Acesso ao Sistema ===${NC}"
    echo -e "\n${YELLOW}Acesso Seguro (HTTPS):${NC}"
    echo -e "Frontend: ${GREEN}https://localhost${NC}"
    echo -e "API Identity Service: ${GREEN}https://localhost:8085${NC}"
    echo -e "API Document Service: ${GREEN}https://localhost:8185${NC}"
    echo -e "MinIO Console: ${GREEN}https://localhost:9185${NC}"
    echo -e "MinIO API: ${GREEN}https://localhost:9085${NC}"
    
    echo -e "\n${YELLOW}Acesso Direto (HTTP, sem proxy/SSL):${NC}"
    echo -e "MongoDB: ${GREEN}localhost:27185${NC}"
    
    echo -e "\n${YELLOW}Monitoramento e Logging:${NC}"
    echo -e "Grafana: ${GREEN}http://localhost:3385${NC}  (admin / gestor_e_docs_admin)"
    echo -e "Kibana: ${GREEN}http://localhost:5685${NC}"
    echo -e "Elasticsearch: ${GREEN}http://localhost:9285${NC}"
    echo -e "Prometheus: ${GREEN}http://localhost:9385${NC}"
    
    echo -e "\n${YELLOW}Credenciais iniciais:${NC}"
    echo -e "MinIO: minioadmin / minioadmin"
    echo -e "Grafana: admin / gestor_e_docs_admin"
    
    echo -e "\n${YELLOW}IMPORTANTE: Os certificados SSL são autoassinados.${NC}"
    echo -e "Navegadores podem exibir alertas de segurança. Para uso em produção,"
    echo -e "substitua por certificados válidos emitidos por uma CA confiável."
    
    echo -e "\n${YELLOW}Para parar os serviços, execute: docker compose down${NC}"
else
    echo -e "${RED}Erro ao iniciar os serviços! Verifique os logs:${NC}"
    docker logs identity_service 2>/dev/null || echo "Container identity_service não encontrado"
    docker logs document_service 2>/dev/null || echo "Container document_service não encontrado"
fi

echo -e "\n${GREEN}=== Fim do Script ===${NC}"
