# Gestor-e-Docs

[English version below](#english)

## Português (Pt-BR)

### Visão Geral
Plataforma web para gerenciamento completo de documentos eletrônicos no formato Markdown, com funcionalidades de upload, armazenamento seguro, consulta, conversão e controle de acesso.

[![Licença MIT](https://img.shields.io/badge/Licença-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://docs.docker.com/compose/)
[![Go](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18.2+-blue.svg)](https://reactjs.org/)

### Resumo das Melhorias (Junho/2025)
Este projeto passou por uma extensa fase de aprimoramento e estabilização. As principais melhorias incluem:

#### ✅ Correções Gerais (Q1 2025)
- **Criação de Usuário Admin Confiável**: A lógica de criação do usuário administrador foi integrada diretamente à inicialização do `identity-service`, eliminando a necessidade de scripts externos.

- **Correção do Fluxo de Autenticação**: O principal bug que impedia o login foi resolvido, corrigindo o endpoint `/me` e configurando o cliente de API (`axios`) adequadamente.

- **Estabilização do Build**: Foram resolvidos múltiplos erros de build no backend e no frontend.

- **Segurança Básica**: O modelo de usuário no backend foi ajustado para nunca expor o hash da senha em respostas JSON, e a configuração de cookies HttpOnly foi implementada.

#### ✅ Serviço de Conversão (Q1 2025)
- **Conversion Service**: Implementação completa do serviço de conversão de documentos Markdown para PDF, HTML, DOCX e LaTeX.
- **Processamento Assíncrono**: Implementação de queue para processamento assíncrono de conversões.

#### ✅ Melhorias de Segurança (Q2 2025)
- **Rate Limiting**: Implementação de limitação de taxa para proteger contra ataques de força bruta (5 tentativas/min nas rotas de autenticação, 60 req/min global).
- **Log de Auditoria**: Sistema completo de registro de ações dos usuários, armazenando detalhes como endereço IP, ação realizada, status da resposta e timestamp.
- **Autenticação em Duas Etapas (2FA)**: Implementação de 2FA para maior segurança das contas de usuário.

#### ✅ Consolidação de Banco de Dados (Q2 2025)
- **Unificação do MongoDB**: Migração de todas as coleções para um único banco de dados `gestor_e_docs` (anteriormente dividido entre `gestor_e_docs` e `gestor_docs`).
- **Correção de Bugs de Persistência**: Resolução de problemas com persistência de IDs e outros dados no MongoDB.
- **Otimização da Comunicação entre Serviços**: Correção de problemas de comunicação entre serviços, especialmente na integração com Nginx.

### Tecnologias
- **Backend**: GoLang (Arquitetura de Microserviços)
- **Frontend**: React
- **Banco de Dados**: MongoDB
- **Armazenamento de Arquivos**: MinIO
- **Autenticação**: JWT com cookies HttpOnly
- **Comunicação API**: REST
- **Service Discovery**: DNS interno do Docker
- **SSL/TLS**: Nginx como proxy reverso seguro
- **Monitoramento**: Prometheus + Grafana
- **Centralização de Logs**: ELK Stack (Elasticsearch, Kibana) + Fluentd

### Arquitetura
O sistema é composto por serviços independentes que se comunicam via API REST:

- **identity-service**: Responsável pela autenticação e autorização de usuários
- **document-service**: Gerenciamento de documentos, incluindo upload, armazenamento, categorização e controle de versões
- **conversion-service**: Conversão de documentos Markdown para PDF, HTML, DOCX e LaTeX com processamento síncrono e assíncrono
- **frontend-app**: Interface web para usuários
- **nginx**: Proxy reverso para TLS/SSL e acesso seguro a todos os serviços
- **prometheus + grafana**: Coleta, armazenamento e visualização de métricas
- **fluentd + elasticsearch + kibana**: Coleta, armazenamento e visualização de logs

### Conversion Service
O **Conversion Service** oferece conversão robusta de documentos Markdown para múltiplos formatos:

#### Formatos Suportados
- **PDF**: Conversão via Gotenberg usando Chrome headless
- **HTML**: Geração de HTML responsivo e acessível
- **DOCX**: Conversão via LibreOffice através do Gotenberg
- **LaTeX**: Conversor nativo implementado em Go

#### Modos de Operação

**Conversão Síncrona** (resposta imediata):
- `POST /api/v1/convert/markdown-to-pdf`
- `POST /api/v1/convert/markdown-to-html`
- `POST /api/v1/convert/markdown-to-docx`
- `POST /api/v1/convert/markdown-to-latex`

**Conversão Assíncrona** (processamento em background):
- `POST /api/v1/convert/async/markdown-to-pdf`
- `POST /api/v1/convert/async/markdown-to-html`
- `POST /api/v1/convert/async/markdown-to-docx`
- `POST /api/v1/convert/async/markdown-to-latex`

#### Gerenciamento de Jobs Assíncronos
- `GET /api/v1/convert/jobs/{jobId}/status` - Verificar status do job
- `GET /api/v1/convert/jobs/{jobId}/download` - Download do resultado
- `GET /api/v1/convert/jobs/stats` - Estatísticas da queue de processamento

#### Funcionalidades Avançadas
- **Sistema de Queue**: Processamento assíncrono com workers concorrentes
- **Validação Robusta**: Middleware que valida formato, tamanho e estrutura do Markdown
- **Monitoramento**: Métricas Prometheus integradas
- **Rate Limiting**: Proteção contra sobrecarga do sistema
- **Cleanup Automático**: Remoção de jobs antigos após 24 horas

#### Exemplo de Uso
```bash
# Conversão síncrona para PDF
curl -X POST https://localhost/api/v1/convert/markdown-to-pdf \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"content": "# Meu Documento\n\nConteúdo em **Markdown**", "title": "exemplo"}'

# Conversão assíncrona
curl -X POST https://localhost/api/v1/convert/async/markdown-to-pdf \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"content": "# Documento Grande\n\nConteúdo extenso...", "title": "documento-grande"}'

# Verificar status do job assíncrono
curl -X GET https://localhost/api/v1/convert/jobs/{jobId}/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Configuração do Ambiente

#### Pré-requisitos
- Docker
- Docker Compose

#### Como Executar
1. Clone o repositório:
   ```bash
   git clone https://github.com/peder1981/Gestor-e-Docs.git
   cd Gestor-e-Docs
   ```

2. Inicie os serviços com Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Acesse a aplicação:
   - **Frontend**: [https://localhost](https://localhost)
   - **Credenciais de Admin**: `admin@example.com` / `password123`

   > **Nota:** Os certificados SSL são autoassinados. Pode ser necessário aceitar o risco de segurança no seu navegador na primeira vez.

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
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Faça commit das alterações (`git commit -m 'Adiciona nova feature'`)
4. Faça o push para a branch (`git push origin feature/nova-feature`)
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

### Summary of Fixes and Improvements (June 2025)
This project underwent an extensive debugging and refactoring phase to stabilize the environment and fix the authentication flow. Key improvements include:

- **Reliable Admin User Creation**: The admin user creation logic was integrated directly into the `identity-service` startup. This resolved a critical data persistence issue in Docker, eliminating the need for external scripts (`create-admin-user.go`) and simplifying the Dockerfile.

- **Authentication Flow Fix**: The main bug preventing login has been resolved. The issue was multifaceted:
  - **Backend**: The `/me` endpoint was fixed to return the full user object instead of just a success message, which the frontend expected.
  - **Frontend**: The API client (`axios`) was configured with `withCredentials: true` to ensure authentication cookies (HttpOnly) were sent with every request.
  - **Frontend**: The logic in `AuthContext` was refined to correctly handle the API response, update the authentication state, and allow redirection to the main page after login.

- **Build Stabilization**: Multiple build errors were resolved in the backend (Go module conflicts, incorrect import paths) and frontend (missing dependencies, linting errors, wrong imports).

- **Security**: The user model in the backend was adjusted to never expose the password hash in JSON responses (`json:"-"`), and the HttpOnly cookie configuration was validated to protect against XSS attacks.

### Technologies
- **Backend**: GoLang (Microservices Architecture)
- **Frontend**: React
- **Database**: MongoDB
- **File Storage**: MinIO
- **Authentication**: JWT with HttpOnly cookies
- **API Communication**: REST
- **Service Discovery**: Docker's internal DNS
- **SSL/TLS**: Nginx as a secure reverse proxy
- **Monitoring**: Prometheus + Grafana
- **Log Centralization**: ELK Stack (Elasticsearch, Kibana) + Fluentd

### Environment Setup

#### Prerequisites
- Docker
- Docker Compose

#### How to Run
1. Clone the repository:
   ```bash
   git clone https://github.com/peder1981/Gestor-e-Docs.git
   cd Gestor-e-Docs
   ```

2. Start the services with Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Access the application:
   - **Frontend**: [https://localhost](https://localhost)
   - **Admin Credentials**: `admin@example.com` / `password123`

   > **Note:** The SSL certificates are self-signed. You may need to accept the security risk in your browser on the first visit.

### Contributing

Contributions are welcome! To contribute to this project:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -m 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Open a Pull Request

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
