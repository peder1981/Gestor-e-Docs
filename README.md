# Gestor-e-Docs

[English version below](#english)

## Português (Pt-BR)

### Visão Geral
Plataforma web para gerenciamento completo de documentos eletrônicos no formato Markdown, com funcionalidades de upload, armazenamento seguro, consulta, conversão e controle de acesso.

[![Licença MIT](https://img.shields.io/badge/Licença-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://docs.docker.com/compose/)
[![Go](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18.2+-blue.svg)](https://reactjs.org/)

### Tecnologias
- **Backend**: GoLang (Arquitetura de Microserviços)
- **Frontend**: React
- **Banco de Dados**: MongoDB
- **Armazenamento de Arquivos**: MinIO
- **Autenticação**: JWT com cookies HttpOnly
- **Comunicação API**: REST
- **Conversão de Documentos**: Pandoc
- **Service Discovery**: DNS interno do Docker
- **SSL/TLS**: Nginx como proxy reverso seguro
- **Monitoramento**: Prometheus + Grafana
- **Centralização de Logs**: ELK Stack (Elasticsearch, Kibana) + Fluentd

### Arquitetura
O sistema é composto por serviços independentes que se comunicam via API REST:

- **identity-service**: Responsável pela autenticação e autorização de usuários
- **document-service**: Gerenciamento de documentos, incluindo upload, armazenamento, categorização e controle de versões
- **conversion-service** (planejado): Conversão entre formatos de documento
- **frontend-app**: Interface web para usuários
- **nginx**: Proxy reverso para TLS/SSL e acesso seguro a todos os serviços
- **prometheus + grafana**: Coleta, armazenamento e visualização de métricas
- **fluentd + elasticsearch + kibana**: Coleta, armazenamento e visualização de logs

### Sistema de Autenticação
O Gestor-e-Docs utiliza um sistema moderno e seguro de autenticação:

- **Autenticação baseada em JWT** utilizando cookies HttpOnly
- **Proteção contra XSS** por não expor tokens ao JavaScript no navegador
- **Refresh tokens** para renovação automática de sessões
- **CORS** configurado para proteger contra acessos não autorizados

### Configuração do Ambiente

#### Pré-requisitos
- Docker
- Docker Compose

#### Como Executar
1. Clone o repositório:
   ```bash
   git clone https://github.com/seu-usuario/gestor-e-docs.git
   cd Gestor-e-Docs
   ```

2. Use o script de inicialização (recomendado):
   ```bash
   chmod +x start.sh
   ./start.sh
   ```
   
   Ou inicie os serviços diretamente com Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Acesse a aplicação:

   **Acesso seguro (HTTPS):**
   - Frontend: https://localhost
   - API de identidade: https://localhost:8085
   - API de documentos: https://localhost:8185
   - Console MinIO: https://localhost:9185 (credenciais: minioadmin/minioadmin)
   - API MinIO: https://localhost:9085
   
   **Monitoramento e Logging:**
   - Grafana: http://localhost:3385 (credenciais: admin/gestor_e_docs_admin)
   - Kibana: http://localhost:5685
   - Elasticsearch: http://localhost:9285
   - Prometheus: http://localhost:9385
   
   **Acesso direto:**
   - MongoDB: localhost:27185
   
   > **Nota:** Os certificados SSL são autoassinados. Para uso em produção, 
   > substitua por certificados válidos emitidos por uma CA confiável.

### Estrutura de Diretórios
```
/
│── backend/
│   └── services/
│       │── identity-service/   # Serviço de autenticação em Go
│       └── document-service/   # Serviço de gerenciamento de documentos em Go
│── frontend/
│   └── web-app/              # Aplicação React
│── docker-compose.yml        # Configuração dos serviços Docker
│── nginx/
│   │── conf/                 # Configuração do proxy reverso
│   └── logs/                 # Logs do Nginx
│── certs/                     # Certificados SSL
│── prometheus/                # Configuração do Prometheus
│── grafana/
│   └── provisioning/          # Dashboards e datasources
│── fluentd/
│   │── conf/                 # Configuração do Fluentd
│   └── Dockerfile             # Build personalizado do Fluentd
└── start.sh                  # Script de inicialização automatizada
```

### Variáveis de Ambiente

#### Identity Service
- `MONGO_URI`: URI de conexão com o MongoDB (padrão: "mongodb://mongo_db:27017")
- `MONGO_DB_NAME`: Nome do banco de dados (padrão: "gestor_e_docs")
- `JWT_SECRET_KEY`: Chave secreta para assinar tokens JWT

#### Document Service
- `MONGO_URI`: Mesma URI de conexão com o MongoDB
- `MONGO_DB_NAME`: Mesmo nome de banco de dados
- `JWT_SECRET_KEY`: Mesma chave secreta para validação de tokens
- `MINIO_ENDPOINT`: Endpoint do MinIO (padrão: "minio_server:9000")
- `MINIO_ACCESS_KEY`: Chave de acesso do MinIO (padrão: "minioadmin")
- `MINIO_SECRET_KEY`: Chave secreta do MinIO (padrão: "minioadmin")
- `MINIO_BUCKET_NAME`: Nome do bucket para armazenamento de documentos (padrão: "documents")
- `PORT`: Porta para o serviço (padrão: "8185")

#### Monitoramento e Logging
- `GRAFANA_ADMIN_USER`: Usuário administrador do Grafana (padrão: "admin")
- `GRAFANA_ADMIN_PASSWORD`: Senha do administrador do Grafana (padrão: "gestor_e_docs_admin")
- `GRAFANA_PORT`: Porta de acesso ao Grafana (padrão: "3385")
- `KIBANA_PORT`: Porta de acesso ao Kibana (padrão: "5685")
- `ELASTICSEARCH_PORT`: Porta de acesso ao Elasticsearch (padrão: "9285")
- `PROMETHEUS_PORT`: Porta de acesso ao Prometheus (padrão: "9385")
- `FLUENTD_PORT`: Porta para envio de logs ao Fluentd (padrão: "24285")

### Troubleshooting

#### Problemas Comuns

1. **Erro `Not supported URL scheme http+docker` ao executar scripts**
   - **Causa**: Variáveis de ambiente conflitantes do Docker
   - **Solução**: Use o script `start.sh` que limpa variáveis como `DOCKER_HOST` antes da execução

2. **Erro de conexão com MinIO**
   - **Causa**: Problemas de resolução de hostname dentro dos containers
   - **Solução**: Verifique o valor de `MINIO_ENDPOINT` no `.env` e certifique-se que os containers estão na mesma rede Docker

3. **Erro 401 (Unauthorized) ao tentar fazer login**
   - **Causa**: Credenciais inválidas ou problemas na configuração de cookies
   - **Solução**: Verifique as configurações de CORS e cookies no Nginx e no identity-service
   - **Dica**: Use o endpoint de registro para criar um novo usuário: `curl -k -X POST -H "Content-Type: application/json" -d '{"name": "Admin", "email": "admin@example.com", "password": "senha123", "role": "admin"}' https://localhost/api/v1/identity/register`

4. **Erro 502 Bad Gateway no frontend**
   - **Causa**: Serviços backend não estão rodando ou acessíveis
   - **Solução**: Verifique o status dos containers com `docker compose ps` e os logs com `docker compose logs identity-service`

5. **Timeout ao baixar dependências Go no build**
   - **Causa**: Problemas de DNS ou proxy em algumas redes
   - **Solução**: Recrie as imagens com o parâmetro `--no-cache` ou ajuste os Dockerfiles para usar mirrors alternativos

### Contribuição

Contribuições são bem-vindas! Para contribuir com este projeto:

1. Faça um fork do repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Faça commit das alterações (`git commit -m 'Adiciona nova funcionalidade'`)
4. Envie para o branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

#### Padrões de Código

- **Go**: Siga as convenções de formatação do Go. Use `go fmt` antes de cada commit
- **React**: Mantenha componentes pequenos e reutilizáveis
- **Testes**: Adicione testes para funcionalidades novas ou modificadas

#### Relatando Problemas

Use a seção de Issues do GitHub para relatar bugs ou sugerir melhorias. Forneça:
- Descrição clara do problema
- Passos para reproduzir
- Comportamento esperado vs. atual
- Logs relevantes ou screenshots

### Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

<a name="english"></a>
## English

### Overview
Web platform for complete management of electronic documents in Markdown format, with features for uploading, secure storage, querying, conversion, and access control.

### Technologies
- **Backend**: GoLang (Microservices Architecture)
- **Frontend**: React
- **Database**: MongoDB
- **File Storage**: MinIO
- **Authentication**: JWT with HttpOnly cookies
- **API Communication**: REST
- **Document Conversion**: Pandoc
- **Service Discovery**: Docker internal DNS
- **SSL/TLS**: Nginx as secure reverse proxy
- **Monitoring**: Prometheus + Grafana
- **Log Management**: ELK Stack (Elasticsearch, Kibana) + Fluentd

### Architecture
The system consists of independent services that communicate via REST API:

- **identity-service**: Responsible for user authentication and authorization
- **document-service**: Document management, including upload, storage, categorization, and version control
- **conversion-service** (planned): Conversion between document formats
- **frontend-app**: Web interface for users
- **nginx**: Reverse proxy for TLS/SSL and secure access to all services
- **prometheus + grafana**: Collection, storage, and visualization of metrics
- **fluentd + elasticsearch + kibana**: Collection, storage, and visualization of logs

### Authentication System
Gestor-e-Docs uses a modern and secure authentication system:

- **JWT-based authentication** using HttpOnly cookies
- **XSS protection** by not exposing tokens to JavaScript in the browser
- **Refresh tokens** for automatic session renewal
- **CORS** configured to protect against unauthorized access

### Environment Setup

#### Prerequisites
- Docker
- Docker Compose

#### How to Run
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/gestor-e-docs.git
   cd Gestor-e-Docs
   ```

2. Use the startup script (recommended):
   ```bash
   chmod +x start.sh
   ./start.sh
   ```
   
   Or start services directly with Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Access the application:

   **Secure access (HTTPS):**
   - Frontend: https://localhost
   - Identity API: https://localhost:8085
   - Document API: https://localhost:8185
   - MinIO Console: https://localhost:9185 (credentials: minioadmin/minioadmin)
   - MinIO API: https://localhost:9085
   
   **Monitoring and Logging:**
   - Grafana: http://localhost:3385 (credentials: admin/gestor_e_docs_admin)
   - Kibana: http://localhost:5685
   - Elasticsearch: http://localhost:9285
   - Prometheus: http://localhost:9385
   
   **Direct access:**
   - MongoDB: localhost:27185
   
   > **Note:** SSL certificates are self-signed. For production use,
   > replace with valid certificates issued by a trusted CA.

### Directory Structure
```
/
├── backend/
│   └── services/
│       ├── identity-service/   # Go authentication service
│       └── document-service/   # Go document management service
├── frontend/
│   └── web-app/               # React application
├── docker-compose.yml         # Docker services configuration
├── nginx/
│   ├── conf/                  # Reverse proxy configuration
│   └── logs/                  # Nginx logs
├── certs/                      # SSL certificates
├── prometheus/                 # Prometheus configuration
├── grafana/
│   └── provisioning/           # Dashboards and datasources
├── fluentd/
│   ├── conf/                  # Fluentd configuration
│   └── Dockerfile              # Custom Fluentd build
└── start.sh                   # Automated startup script
```

### Environment Variables

#### Identity Service
- `MONGO_URI`: MongoDB connection URI (default: "mongodb://mongo_db:27017")
- `MONGO_DB_NAME`: Database name (default: "gestor_e_docs")
- `JWT_SECRET_KEY`: Secret key to sign JWT tokens

#### Document Service
- `MONGO_URI`: Same MongoDB connection URI
- `MONGO_DB_NAME`: Same database name
- `JWT_SECRET_KEY`: Same secret key for token validation
- `MINIO_ENDPOINT`: MinIO endpoint (default: "minio_server:9000")
- `MINIO_ACCESS_KEY`: MinIO access key (default: "minioadmin")
- `MINIO_SECRET_KEY`: MinIO secret key (default: "minioadmin")
- `MINIO_BUCKET_NAME`: Bucket name for document storage (default: "documents")
- `PORT`: Service port (default: "8185")

#### Monitoring and Logging
- `GRAFANA_ADMIN_USER`: Grafana admin user (default: "admin")
- `GRAFANA_ADMIN_PASSWORD`: Grafana admin password (default: "gestor_e_docs_admin")
- `GRAFANA_PORT`: Grafana access port (default: "3385")
- `KIBANA_PORT`: Kibana access port (default: "5685")
- `ELASTICSEARCH_PORT`: Elasticsearch access port (default: "9285")
- `PROMETHEUS_PORT`: Prometheus access port (default: "9385")
- `FLUENTD_PORT`: Fluentd log forwarding port (default: "24285")

### Troubleshooting

#### Common Issues

1. **Error `Not supported URL scheme http+docker` when running scripts**
   - **Cause**: Conflicting Docker environment variables
   - **Solution**: Use the `start.sh` script which clears variables like `DOCKER_HOST` before execution

2. **Connection error with MinIO**
   - **Cause**: Hostname resolution issues inside containers
   - **Solution**: Check the `MINIO_ENDPOINT` value in `.env` and ensure containers are in the same Docker network

3. **Error 401 (Unauthorized) when trying to log in**
   - **Cause**: Invalid credentials or cookie configuration issues
   - **Solution**: Check CORS and cookie settings in Nginx and identity-service
   - **Tip**: Use the registration endpoint to create a new user: `curl -k -X POST -H "Content-Type: application/json" -d '{"name": "Admin", "email": "admin@example.com", "password": "password123", "role": "admin"}' https://localhost/api/v1/identity/register`

4. **Error 502 Bad Gateway in frontend**
   - **Cause**: Backend services are not running or accessible
   - **Solution**: Check container status with `docker compose ps` and logs with `docker compose logs identity-service`

5. **Timeout when downloading Go dependencies during build**
   - **Cause**: DNS or proxy issues on some networks
   - **Solution**: Rebuild images with `--no-cache` parameter or adjust Dockerfiles to use alternative mirrors

### Contributing

Contributions are welcome! To contribute to this project:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Open a Pull Request

#### Code Standards

- **Go**: Follow Go formatting conventions. Use `go fmt` before each commit
- **React**: Keep components small and reusable
- **Tests**: Add tests for new or modified functionality

#### Reporting Issues

Use the GitHub Issues section to report bugs or suggest improvements. Please provide:
- Clear description of the issue
- Steps to reproduce
- Expected vs. actual behavior
- Relevant logs or screenshots

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
