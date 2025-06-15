# Conversion Service

Microserviço responsável por converter documentos Markdown para múltiplos formatos (PDF, HTML, DOCX, LaTeX) no ecossistema Gestor-e-Docs.

## 📋 Visão Geral

O Conversion Service oferece conversão robusta e escalável de documentos Markdown, com suporte a processamento síncrono e assíncrono, validação de entrada, sistema de queue com workers, e monitoramento integrado.

## 🚀 Funcionalidades

### Formatos Suportados
- **PDF** - Via Gotenberg usando Chrome headless
- **HTML** - Geração de HTML responsivo e acessível  
- **DOCX** - Via LibreOffice através do Gotenberg
- **LaTeX** - Conversor nativo implementado em Go

### Modos de Operação

#### Conversão Síncrona (Resposta Imediata)
Ideal para documentos pequenos e casos onde o resultado é necessário imediatamente.

- `POST /markdown-to-pdf`
- `POST /markdown-to-html`
- `POST /markdown-to-docx`
- `POST /markdown-to-latex`

#### Conversão Assíncrona (Processamento em Background)
Ideal para documentos grandes ou quando performance é crítica.

- `POST /async/markdown-to-pdf`
- `POST /async/markdown-to-html`
- `POST /async/markdown-to-docx`
- `POST /async/markdown-to-latex`

### Gerenciamento de Jobs
- `GET /jobs/{jobId}/status` - Verificar status do job
- `GET /jobs/{jobId}/download` - Download do resultado
- `GET /jobs/stats` - Estatísticas da queue

### Endpoints Públicos
- `GET /formats` - Listar formatos suportados
- `GET /health` - Status de saúde do serviço

## 🏗️ Arquitetura

```
┌─────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   API Gateway   │───▶│ Validation       │───▶│ Conversion       │
│   (Gin Router)  │    │ Middleware       │    │ Handlers         │
└─────────────────┘    └──────────────────┘    └──────────────────┘
                                                          │
                       ┌──────────────────┐              │
                       │   Queue System   │◀─────────────┘
                       │   (3 Workers)    │
                       └──────────────────┘
                                │
                       ┌──────────────────┐
                       │ Gotenberg Client │
                       │ (PDF/HTML/DOCX)  │
                       └──────────────────┘
```

### Componentes Principais

#### 1. **Handlers** (`handlers/`)
- `conversion_handler.go` - Handlers síncronos
- `async_handlers.go` - Handlers assíncronos
- `models.go` - Estruturas de dados

#### 2. **Sistema de Queue** (`handlers/queue.go`)
- Processamento assíncrono com workers
- Gerenciamento de jobs em memória
- Cleanup automático de jobs antigos

#### 3. **Validação** (`handlers/validation.go`)
- Middleware de validação robusta
- Verificação de formato e tamanho
- Sanitização de entrada

#### 4. **Cliente Gotenberg** (`handlers/gotenberg_client.go`)
- Integração com API Gotenberg
- Conversões PDF, HTML e DOCX
- Conversor LaTeX nativo

## 🔧 Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `PORT` | Porta do serviço | `8083` |
| `GOTENBERG_API_URL` | URL da API Gotenberg | `http://gotenberg:3000` |
| `JWT_SECRET_KEY` | Chave secreta JWT | - |

### Dependências Externas

- **Gotenberg**: Serviço para conversão PDF/HTML/DOCX
- **Identity Service**: Para autenticação via JWT
- **Prometheus**: Para métricas (opcional)

## 📝 API Reference

### Estrutura de Requisição

```json
{
  "content": "# Título\n\nConteúdo em **Markdown**",
  "title": "nome-do-arquivo"
}
```

### Respostas

#### Conversão Síncrona
- **200 OK**: Arquivo binário ou texto (dependendo do formato)
- **400 Bad Request**: Erro de validação
- **500 Internal Server Error**: Erro na conversão

#### Conversão Assíncrona  
```json
{
  "success": true,
  "message": "Conversão para PDF iniciada",
  "job_id": "abc123def456"
}
```

#### Status do Job
```json
{
  "success": true,
  "job": {
    "id": "abc123def456",
    "type": "pdf",
    "status": "completed",
    "created_at": "2025-06-15T10:30:00Z",
    "completed_at": "2025-06-15T10:30:05Z"
  }
}
```

### Validações Aplicadas

- **Content**: Obrigatório, máximo 10MB, UTF-8 válido
- **Title**: Opcional, máximo 255 caracteres, sem caracteres especiais
- **Markdown**: Verificação de estrutura básica para documentos grandes

## 🔄 Queue System

### Funcionamento
- **Workers**: 3 workers concorrentes por padrão
- **Buffer**: Queue com buffer de 100 jobs
- **Timeout**: 5 segundos para adicionar job à queue
- **Cleanup**: Jobs removidos automaticamente após 24h

### Estados dos Jobs
- `pending` - Aguardando processamento
- `processing` - Em processamento  
- `completed` - Concluído com sucesso
- `failed` - Falha no processamento

## 📊 Monitoramento

### Métricas Prometheus
- Contador de conversões por formato
- Tempo de processamento
- Taxa de sucesso/falha
- Tamanho da queue

### Logs
- Logs estruturados com níveis apropriados
- Rastreamento de jobs por ID
- Métricas de performance

## 🧪 Testando o Serviço

### Conversão Síncrona
```bash
curl -X POST http://localhost:8083/api/v1/markdown-to-pdf \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "# Teste\n\nConteúdo de **teste**",
    "title": "documento-teste"
  }' \
  --output documento-teste.pdf
```

### Conversão Assíncrona
```bash
# Iniciar conversão
curl -X POST http://localhost:8083/api/v1/async/markdown-to-pdf \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "# Documento Grande\n\nConteúdo extenso...",
    "title": "documento-grande"
  }'

# Verificar status
curl -X GET http://localhost:8083/api/v1/jobs/{jobId}/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Download do resultado
curl -X GET http://localhost:8083/api/v1/jobs/{jobId}/download \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --output resultado.pdf
```

## 🚀 Deploy e Execução

### Docker Compose
O serviço é executado automaticamente como parte do stack do Gestor-e-Docs:

```bash
docker compose up -d --build
```

### Standalone
```bash
cd backend/services/conversion-service
go mod tidy
go build -o conversion-service .
./conversion-service
```

## 🛠️ Desenvolvimento

### Estrutura do Projeto
```
conversion-service/
├── main.go                    # Ponto de entrada
├── go.mod                     # Dependências Go
├── Dockerfile                 # Container definition
├── handlers/
│   ├── conversion_handler.go  # Handlers síncronos
│   ├── async_handlers.go      # Handlers assíncronos
│   ├── gotenberg_client.go    # Cliente Gotenberg
│   ├── queue.go               # Sistema de queue
│   ├── validation.go          # Middleware de validação
│   ├── models.go              # Estruturas de dados
│   └── auth_middleware.go     # Middleware de autenticação
└── metrics/
    └── prometheus.go          # Métricas Prometheus
```

### Adicionando Novos Formatos
1. Implementar método no `gotenberg_client.go`
2. Criar handler em `conversion_handler.go` 
3. Adicionar handler assíncrono em `async_handlers.go`
4. Atualizar `ListSupportedFormats()`
5. Adicionar rotas em `main.go`

## 🐛 Troubleshooting

### Problemas Comuns

**Erro de conexão com Gotenberg**
- Verificar se o serviço Gotenberg está rodando
- Validar a variável `GOTENBERG_API_URL`

**Jobs ficam em estado "pending"**
- Verificar se os workers estão iniciados
- Checar logs para erros de processamento

**Erro de autenticação**
- Validar token JWT
- Verificar configuração do `JWT_SECRET_KEY`

### Logs Úteis
```bash
# Logs do conversion-service
docker logs gestor-e-docs-conversion-service-1

# Logs do Gotenberg
docker logs gestor-e-docs-gotenberg-1
```

## 📜 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](../../../LICENSE) para detalhes.
