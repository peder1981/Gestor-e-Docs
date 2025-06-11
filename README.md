# Gestor-e-Docs

[English version below](#english)

## Português (Pt-BR)

### Visão Geral
Plataforma web para gerenciamento completo de documentos eletrônicos no formato Markdown, com funcionalidades de upload, armazenamento seguro, consulta, conversão e controle de acesso.

[![Licença MIT](https://img.shields.io/badge/Licença-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://docs.docker.com/compose/)
[![Go](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18.2+-blue.svg)](https://reactjs.org/)

### Resumo das Correções e Melhorias (Junho/2025)
Este projeto passou por uma extensa fase de depuração e refatoração para estabilizar o ambiente e corrigir o fluxo de autenticação. As principais melhorias incluem:

- **Criação de Usuário Admin Confiável**: A lógica de criação do usuário administrador foi integrada diretamente à inicialização do `identity-service`. Isso resolveu um problema crítico de persistência de dados no Docker, eliminando a necessidade de scripts externos (`create-admin-user.go`) e simplificando o Dockerfile.

- **Correção do Fluxo de Autenticação**: O principal bug que impedia o login foi resolvido. O problema era multifacetado:
  - **Backend**: O endpoint `/me` foi corrigido para retornar o objeto de usuário completo, em vez de apenas uma mensagem de sucesso, que era o que o frontend esperava.
  - **Frontend**: O cliente de API (`axios`) foi configurado com `withCredentials: true` para garantir que os cookies de autenticação (HttpOnly) fossem enviados em todas as requisições.
  - **Frontend**: A lógica no `AuthContext` foi refinada para lidar corretamente com a resposta da API, atualizando o estado de autenticação e permitindo o redirecionamento para a página principal após o login.

- **Estabilização do Build**: Foram resolvidos múltiplos erros de build no backend (conflitos de módulos Go, caminhos de importação incorretos) e no frontend (dependências ausentes, erros de lint, importações erradas).

- **Segurança**: O modelo de usuário no backend foi ajustado para nunca expor o hash da senha em respostas JSON (`json:"-"`), e a configuração de cookies HttpOnly foi validada para proteger contra ataques XSS.

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
