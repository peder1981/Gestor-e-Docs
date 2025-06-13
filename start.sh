#!/bin/bash

# Script de inicialização para Gestor-e-Docs
# Autor: Equipe Gestor-e-Docs
# Data: 10/06/2025
# Versão: 1.2

# Versões mínimas requeridas
MIN_DOCKER_VERSION="20.10.0"
MIN_COMPOSE_VERSION="1.29.0"

# Cores para melhor visualização
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para comparar versões
version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1";
}

# Função para verificar versão do Docker
check_docker_version() {
    local docker_version=$(docker version --format '{{.Server.Version}}' 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Erro ao obter versão do Docker${NC}"
        return 1
    fi
    if version_gt "$MIN_DOCKER_VERSION" "$docker_version"; then
        echo -e "${RED}Versão do Docker ($docker_version) é menor que a mínima requerida ($MIN_DOCKER_VERSION)${NC}"
        return 1
    fi
    echo -e "${GREEN}Versão do Docker ($docker_version) OK${NC}"
    return 0
}

# Função para verificar versão do Docker Compose
check_compose_version() {
    local compose_version
    if command -v docker-compose &> /dev/null; then
        compose_version=$(docker-compose version --short)
    else
        compose_version=$(docker compose version --short)
    fi
    if [ $? -ne 0 ]; then
        echo -e "${RED}Erro ao obter versão do Docker Compose${NC}"
        return 1
    fi
    if version_gt "$MIN_COMPOSE_VERSION" "$compose_version"; then
        echo -e "${RED}Versão do Docker Compose ($compose_version) é menor que a mínima requerida ($MIN_COMPOSE_VERSION)${NC}"
        return 1
    fi
    echo -e "${GREEN}Versão do Docker Compose ($compose_version) OK${NC}"
    return 0
}

# Função para verificar e remover contêineres conflitantes
check_and_remove_conflicts() {
    local services=("$@")
    local conflicts_found=0
    
    for service in "${services[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${service}$"; then
            echo -e "${YELLOW}Contêiner conflitante encontrado: $service${NC}"
            echo -e "${YELLOW}Removendo contêiner antigo...${NC}"
            if ! docker rm -f "$service" > /dev/null 2>&1; then
                echo -e "${RED}Erro ao remover contêiner $service${NC}"
                conflicts_found=1
            fi
        fi
    done
    
    return $conflicts_found
}

# Função para verificar se uma porta está disponível
check_port_available() {
    local port=$1
    if ! command -v nc &> /dev/null; then
        echo -e "${YELLOW}netcat não está instalado. Pulando verificação de porta $port.${NC}"
        return 0
    fi
    
    if nc -z localhost $port 2>/dev/null; then
        echo -e "${RED}Porta $port já está em uso${NC}"
        return 1
    fi
    return 0
}

# Função para verificar todas as portas necessárias
check_required_ports() {
    local -a ports=(
        "${IDENTITY_SERVICE_PORT:-8085}"
        "${DOCUMENT_SERVICE_PORT:-8185}"
        "${GRAFANA_PORT:-3385}"
        "${KIBANA_PORT:-5685}"
        "${ELASTICSEARCH_PORT:-9285}"
        "${PROMETHEUS_PORT:-9385}"
        "${MINIO_API_PORT:-9085}"
        "${MINIO_CONSOLE_PORT:-9185}"
        "${FLUENTD_PORT:-24285}"
        "${MONGO_PORT:-27185}"
    )
    
    local port_conflicts=0
    for port in "${ports[@]}"; do
        echo -e "${YELLOW}Verificando porta $port...${NC}"
        if ! check_port_available "$port"; then
            port_conflicts=1
        fi
    done
    
    if [ $port_conflicts -eq 1 ]; then
        echo -e "${RED}Uma ou mais portas estão em uso. Por favor, libere as portas ou altere no arquivo .env${NC}"
        return 1
    fi
    
    echo -e "${GREEN}Todas as portas estão disponíveis${NC}"
    return 0
}

# Função para fazer backup dos dados
backup_data() {
    local backup_dir="./backups"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$backup_dir/backup_$timestamp.tar.gz"
    
    # Criar diretório de backup se não existir
    if [ ! -d "$backup_dir" ]; then
        mkdir -p "$backup_dir"
    fi
    
    echo -e "${YELLOW}Iniciando backup dos dados...${NC}"
    
    # Verificar se existem volumes para backup
    if ! docker volume ls --format "{{.Name}}" | grep -q "gestor-e-docs"; then
        echo -e "${YELLOW}Nenhum volume de dados encontrado para backup${NC}"
        return 0
    fi
    
    # Criar container temporário para backup
    echo -e "${YELLOW}Criando backup dos volumes Docker...${NC}"
    docker run --rm \
        -v gestor-e-docs_mongodb_data:/mongodb:ro \
        -v gestor-e-docs_minio_data:/minio:ro \
        -v "$backup_dir:/backup" \
        alpine tar czf "/backup/backup_$timestamp.tar.gz" /mongodb /minio 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Backup criado com sucesso: $backup_file${NC}"
        # Manter apenas os 5 backups mais recentes
        ls -t "$backup_dir"/*.tar.gz 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null
        return 0
    else
        echo -e "${RED}Erro ao criar backup${NC}"
        return 1
    fi
}

# Função para tratamento de erros
handle_error() {
    local error_msg=$1
    local error_code=${2:-1}
    local show_logs=${3:-false}
    
    echo -e "\n${RED}ERRO: $error_msg${NC}"
    
    # Sugestões de solução baseadas no tipo de erro
    case $error_code in
        1)
            echo -e "${YELLOW}Sugestões:${NC}"
            echo "1. Verifique se todos os serviços estão parados: $DOCKER_COMPOSE down"
            echo "2. Verifique se há conflitos de porta: netstat -tulpn | grep LISTEN"
            echo "3. Verifique os logs do Docker: docker logs [container_id]"
            ;;
        2)
            echo -e "${YELLOW}Sugestões:${NC}"
            echo "1. Verifique a conexão com a internet"
            echo "2. Verifique se o registro Docker está acessível: docker pull hello-world"
            echo "3. Limpe o cache do Docker: docker system prune -a"
            ;;
        3)
            echo -e "${YELLOW}Sugestões:${NC}"
            echo "1. Verifique o espaço em disco: df -h"
            echo "2. Limpe volumes não utilizados: docker volume prune"
            echo "3. Verifique as permissões dos diretórios"
            ;;
    esac
    
    if [ "$show_logs" = true ] && [ ! -z "$DOCKER_COMPOSE" ]; then
        echo -e "\n${YELLOW}Exibindo logs dos serviços para diagnóstico:${NC}"
        $DOCKER_COMPOSE logs
    fi
    
    exit $error_code
}

# Função para validar certificados SSL
check_ssl_certificates() {
    local cert_dir="./certs"
    local missing_certs=0
    
    if [ ! -d "$cert_dir" ]; then
        echo -e "${YELLOW}Diretório de certificados não encontrado. Criando...${NC}"
        mkdir -p "$cert_dir"
    fi
    
    # Lista de certificados necessários
    local -a required_certs=(
        "nginx.crt"
        "nginx.key"
        "minio.crt"
        "minio.key"
    )
    
    for cert in "${required_certs[@]}"; do
        if [ ! -f "$cert_dir/$cert" ]; then
            echo -e "${YELLOW}Certificado $cert não encontrado${NC}"
            missing_certs=1
        else
            # Verificar validade do certificado (apenas para arquivos .crt)
            if [[ "$cert" == *.crt ]] && ! openssl x509 -in "$cert_dir/$cert" -noout -checkend 0 2>/dev/null; then
                echo -e "${RED}Certificado $cert está expirado ou inválido${NC}"
                missing_certs=1
            fi
        fi
    done
    
    if [ $missing_certs -eq 1 ]; then
        echo -e "${YELLOW}Gerando novos certificados autoassinados...${NC}"
        
        # Gerar certificado para Nginx
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$cert_dir/nginx.key" -out "$cert_dir/nginx.crt" \
            -subj "/C=BR/ST=SP/L=Sao Paulo/O=Gestor-e-Docs/CN=localhost" 2>/dev/null
        
        # Gerar certificado para MinIO
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$cert_dir/minio.key" -out "$cert_dir/minio.crt" \
            -subj "/C=BR/ST=SP/L=Sao Paulo/O=Gestor-e-Docs/CN=minio_server" 2>/dev/null
        
        echo -e "${GREEN}Novos certificados gerados com sucesso${NC}"
    else
        echo -e "${GREEN}Certificados SSL válidos encontrados${NC}"
    fi
    
    return 0
}

echo -e "${GREEN}=== Iniciando Gestor-e-Docs ===${NC}"
echo -e "${YELLOW}Verificando pré-requisitos...${NC}"

# Verificar versões do Docker e Docker Compose
check_docker_version || exit 1
check_compose_version || exit 1

# Verificar portas disponíveis
check_required_ports || exit 1

# Verificar certificados SSL
check_ssl_certificates || handle_error "Falha na validação dos certificados SSL" 1

# Realizar backup se solicitado
if [ "${PERFORM_BACKUP:-false}" = true ]; then
    backup_data || handle_error "Falha ao realizar backup dos dados" 3
fi

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
echo -e "${YELLOW}Verificando contêineres conflitantes...${NC}"
check_and_remove_conflicts "${ESSENTIAL_SERVICES[@]}" || {
    echo -e "${RED}Falha ao remover contêineres conflitantes${NC}"
    exit 1
}

echo -e "${YELLOW}Garantindo que a imagem do frontend esteja atualizada...${NC}"
$DOCKER_COMPOSE build frontend-app

echo -e "${YELLOW}Iniciando os serviços com $DOCKER_COMPOSE...${NC}"

# Tentar executar o comando, lidando com erros comuns
if ! $DOCKER_COMPOSE up -d --build; then
    if [ "$DOCKER_COMPOSE" = "docker-compose" ]; then
        echo "Tentando método alternativo com 'docker compose'..."
        if docker compose up -d --build; then
            echo -e "${GREEN}Serviços iniciados com sucesso usando 'docker compose'!${NC}"
            DOCKER_COMPOSE="docker compose"
        else
            handle_error "Falha ao iniciar serviços com docker compose" 2 true
        fi
    else
        handle_error "Falha ao iniciar serviços com $DOCKER_COMPOSE" 2 true
    fi
fi

# Verificar se os containers principais estão rodando
echo -e "${YELLOW}Verificando status dos serviços...${NC}"
sleep 8 # Dar um tempo maior para os containers iniciarem

# Lista de serviços essenciais para verificar
ESSENTIAL_SERVICES=("nginx_proxy" "frontend_app" "identity_service" "document_service" "mongo_db" "minio_server")
FAILED_SERVICES=()

# Espera um pouco para garantir que o docker compose ps tenha a informação mais recente
sleep 2

for service in "${ESSENTIAL_SERVICES[@]}"; do
    # Verifica se o container está com o status 'running'
    status=$($DOCKER_COMPOSE ps --status=running | grep "$service")
    if [ -z "$status" ]; then
        FAILED_SERVICES+=("$service")
    fi
done

if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
    echo -e "${GREEN}Serviços essenciais iniciados com sucesso!${NC}"
    
    # Exibir URLs de acesso
    echo -e "\n${GREEN}=== Acesso ao Sistema ===${NC}"
    echo -e "\n${YELLOW}Acesso Principal (via Proxy Reverso com HTTPS):${NC}"
    echo -e "Interface Gráfica (Frontend): ${GREEN}https://localhost${NC}"
    echo -e "MinIO Console: ${GREEN}https://localhost:9185${NC}"
    
    echo -e "\n${YELLOW}Acesso direto aos serviços (para depuração):${NC}"
    echo -e "API Identity (HTTPS direto): ${GREEN}https://localhost:8085${NC}"
    echo -e "API Document (HTTPS direto): ${GREEN}https://localhost:8185${NC}"
    echo -e "MinIO API (HTTPS direto): ${GREEN}https://localhost:9085${NC}"
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
    echo -e "Seu navegador exibirá um alerta de segurança. Você pode aceitá-lo para continuar."
    
    echo -e "\n${YELLOW}Para parar os serviços, execute: $DOCKER_COMPOSE down${NC}"
else
    error_msg="Os seguintes serviços essenciais falharam ao iniciar:\n"
    for service in "${FAILED_SERVICES[@]}"; do
        error_msg+="- $service\n"
    done
    handle_error "$error_msg" 1 true
fi

echo -e "\n${GREEN}=== Fim do Script ===${NC}"
