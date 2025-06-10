# Gestor-e-Docs

[English version below](#english)

## Português (Pt-BR)

### Visão Geral
Plataforma web para gerenciamento completo de documentos eletrônicos no formato Markdown, com funcionalidades de upload, armazenamento seguro, consulta, conversão e controle de acesso.

### Tecnologias
- **Backend**: GoLang (Arquitetura de Microserviços)
- **Frontend**: React
- **Banco de Dados**: MongoDB
- **Armazenamento de Arquivos**: MinIO
- **Autenticação**: JWT com cookies HttpOnly
- **Comunicação API**: REST
- **Conversão de Documentos**: Pandoc

### Arquitetura
O sistema é composto por serviços independentes que se comunicam via API REST:

- **identity-service**: Responsável pela autenticação e autorização de usuários
- **document-service** (planejado): Gerenciamento de documentos
- **conversion-service** (planejado): Conversão entre formatos de documento
- **frontend-app**: Interface web para usuários

### Sistema de Autenticação
O Gestor-e-Docs utiliza um sistema moderno e seguro de autenticação:

- **Autenticação baseada em JWT** utilizando cookies HttpOnly
- **Proteção contra XSS** por não expor tokens ao JavaScript no navegador
- **Refresh tokens** para renovação automática de sessões
- **CORS** configurado para proteger contra acessos não autorizados

### Configuração do Ambiente de Desenvolvimento

#### Pré-requisitos
- Docker
- Docker Compose

#### Como Executar
1. Clone o repositório:
   ```bash
   git clone https://github.com/seu-usuario/gestor-e-docs.git
   cd Gestor-e-Docs
   ```

2. Inicie os serviços com Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Acesse a aplicação:
   - Frontend: http://localhost:3085
   - API de identidade: http://localhost:8085

### Estrutura de Diretórios
```
/
├── backend/
│   └── services/
│       └── identity-service/   # Serviço de autenticação em Go
├── frontend/
│   └── web-app/               # Aplicação React
└── docker-compose.yml         # Configuração dos serviços Docker
```

### Variáveis de Ambiente

#### Identity Service
- `MONGO_URI`: URI de conexão com o MongoDB (padrão: "mongodb://mongo_db:27017")
- `MONGO_DB_NAME`: Nome do banco de dados (padrão: "gestor_e_docs")
- `JWT_SECRET_KEY`: Chave secreta para assinar tokens JWT

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

### Architecture
The system consists of independent services that communicate via REST API:

- **identity-service**: Responsible for user authentication and authorization
- **document-service** (planned): Document management
- **conversion-service** (planned): Conversion between document formats
- **frontend-app**: Web interface for users

### Authentication System
Gestor-e-Docs uses a modern and secure authentication system:

- **JWT-based authentication** using HttpOnly cookies
- **XSS protection** by not exposing tokens to JavaScript in the browser
- **Refresh tokens** for automatic session renewal
- **CORS** configured to protect against unauthorized access

### Development Environment Setup

#### Prerequisites
- Docker
- Docker Compose

#### How to Run
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/gestor-e-docs.git
   cd Gestor-e-Docs
   ```

2. Start services with Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Access the application:
   - Frontend: http://localhost:3085
   - Identity API: http://localhost:8085

### Directory Structure
```
/
├── backend/
│   └── services/
│       └── identity-service/   # Go authentication service
├── frontend/
│   └── web-app/               # React application
└── docker-compose.yml         # Docker services configuration
```

### Environment Variables

#### Identity Service
- `MONGO_URI`: MongoDB connection URI (default: "mongodb://mongo_db:27017")
- `MONGO_DB_NAME`: Database name (default: "gestor_e_docs")
- `JWT_SECRET_KEY`: Secret key to sign JWT tokens
