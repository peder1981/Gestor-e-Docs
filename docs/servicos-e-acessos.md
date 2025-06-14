# Documentação de Serviços e Acessos - Gestor-e-Docs

## Importante
⚠️ Este documento contém informações sensíveis. Mantenha-o seguro e compartilhe apenas com pessoal autorizado.

## Serviços da Aplicação

### Frontend (React)
- **URL de Acesso**: https://localhost:443
- **Credenciais de Administrador**:
  - Email: admin@example.com
  - Senha: password123
- **Observações**: 
  - Interface web principal do sistema
  - Hot-reloading habilitado para desenvolvimento
  - Servido via Nginx com SSL

### Serviço de Identidade (Identity Service)
- **URL da API**: https://localhost:8085
- **Responsabilidades**:
  - Autenticação de usuários
  - Gerenciamento de sessões
  - Emissão de tokens JWT
- **Variáveis de Ambiente**:
  - JWT_SECRET_KEY: seuSuperSegredoMuitoComplexoAqui
  - MONGO_URI: mongodb://mongo_db:27017/gestor_e_docs
  - SERVICE_PORT: 8085
  - GIN_MODE: debug

### Serviço de Documentos (Document Service)
- **URL da API**: https://localhost:8185
- **Responsabilidades**:
  - Gerenciamento de documentos
  - Armazenamento no MinIO
  - Metadados no MongoDB
- **Variáveis de Ambiente**:
  - JWT_SECRET_KEY: seuSuperSegredoMuitoComplexoAqui
  - MONGO_URI: mongodb://mongo_db:27017/gestor_e_docs
  - MINIO_ENDPOINT: minioserver:9000
  - MINIO_ACCESS_KEY: minioadmin
  - MINIO_SECRET_KEY: minioadmin
  - MINIO_BUCKET_NAME: documents
  - PORT: 8185
  - GIN_MODE: debug

### Serviço de Conversão (Conversion Service)
- **URL da API**: https://localhost:8285
- **Responsabilidades**:
  - Conversão de formatos de documentos
  - Integração com Gotenberg
- **Variáveis de Ambiente**:
  - JWT_SECRET_KEY: seuSuperSegredoMuitoComplexoAqui
  - PORT: 8285
  - GIN_MODE: debug
  - GOTENBERG_API_URL: http://gotenberg:3000

## Serviços de Armazenamento

### MongoDB (Banco de Dados)
- **URL**: mongodb://localhost:27185
- **Credenciais**: Autenticação não configurada (ambiente de desenvolvimento)
- **Base de Dados**: gestor_e_docs
- **Volumes**: /data/db (persistente)
- **Observações**: Armazena metadados dos documentos e dados do sistema

### MinIO (Armazenamento de Objetos)
- **URL API**: https://localhost:9085
- **URL Console**: https://localhost:9185
- **Credenciais**:
  - Usuário: minioadmin
  - Senha: minioadmin
- **Bucket Principal**: documents
- **Volumes**: /data (persistente)
- **Observações**: Armazena os arquivos físicos dos documentos

## Serviços de Monitoramento

### Elasticsearch
- **URL**: http://localhost:9285
- **Configurações**:
  - Modo: single-node
  - Memória: 512MB (min) - 512MB (max)
  - Segurança: desabilitada para desenvolvimento
- **Volumes**: /usr/share/elasticsearch/data (persistente)

### Kibana (Visualização de Logs)
- **URL**: http://localhost:5685
- **Observações**: 
  - Interface web para Elasticsearch
  - Visualização e análise de logs

### Grafana (Monitoramento)
- **URL**: http://localhost:3385
- **Credenciais**:
  - Usuário: admin
  - Senha: gestor_e_docs_admin
- **Volumes**: /var/lib/grafana (persistente)
- **Observações**: Dashboard de métricas e monitoramento

### Prometheus (Métricas)
- **URL**: http://localhost:9385
- **Volumes**: 
  - /etc/prometheus (configuração)
  - /prometheus (dados)
- **Integrações**:
  - Node Exporter (métricas do sistema)
  - Nginx Exporter (métricas do Nginx)

### Fluentd (Agregador de Logs)
- **URL**: http://localhost:24285
- **Responsabilidades**:
  - Coleta de logs dos containers
  - Envio para Elasticsearch
- **Volumes**:
  - /fluentd/etc (configuração)
  - /var/lib/docker/containers (logs dos containers)

## Portas Utilizadas

### Portas Externas (Acessíveis fora do container)
- 443: Frontend (HTTPS)
- 8085: API Identity Service
- 8185: API Document Service
- 8285: API Conversion Service
- 9085: MinIO API
- 9185: MinIO Console Web
- 27185: MongoDB (Acesso direto ao banco)
- 9285: Elasticsearch (9200 interno)
- 5685: Kibana (5601 interno)
- 24285: Fluentd (24224 interno)
- 9385: Prometheus (9090 interno)
- 3385: Grafana (3000 interno)

### Portas Internas (Apenas na rede do container)
- 80: Frontend App
- 27017: MongoDB
- 9000: MinIO API
- 9001: MinIO Console
- 9100: Node Exporter
- 9113: Nginx Exporter
- 3000: Gotenberg

**Observação**: Todas as portas externas seguem o padrão com final 85 para padronização do ambiente.

## Observações de Segurança

### Credenciais e Senhas
1. **Alteração Obrigatória para Produção**:
   - Frontend Admin: `admin@example.com`/`password123`
   - MinIO Admin: `minioadmin`/`minioadmin`
   - Grafana Admin: `admin`/`gestor_e_docs_admin`
   - JWT_SECRET_KEY: `seuSuperSegredoMuitoComplexoAqui`

2. **Política de Senhas**:
   - Comprimento mínimo: 12 caracteres
   - Combinação de maiúsculas, minúsculas, números e símbolos
   - Rotação a cada 90 dias
   - Sem reutilização das últimas 5 senhas

### Configurações por Serviço

1. **Frontend e APIs**:
   - Habilitar HTTPS em produção
   - Configurar CORS adequadamente
   - Implementar rate limiting
   - Adicionar proteção contra DDoS
   - Desabilitar modo debug

2. **Banco de Dados (MongoDB)**:
   - Habilitar autenticação
   - Criar usuários com privilégios mínimos
   - Restringir acesso por IP
   - Habilitar SSL/TLS
   - Configurar backup automático

3. **MinIO**:
   - Habilitar SSL/TLS
   - Configurar políticas de bucket
   - Implementar retenção de objetos
   - Habilitar criptografia em repouso
   - Configurar backup regular

4. **Elasticsearch e Kibana**:
   - Habilitar X-Pack Security
   - Configurar autenticação
   - Implementar SSL/TLS
   - Restringir acesso por IP
   - Configurar retenção de logs

5. **Prometheus e Grafana**:
   - Habilitar autenticação
   - Configurar SSL/TLS
   - Restringir acesso por IP
   - Implementar alertas de segurança

### Rede e Firewalls

1. **Exposição de Serviços**:
   - Limitar acesso externo apenas às portas necessárias
   - Usar VPN para acesso administrativo
   - Implementar firewall em nível de aplicação (WAF)

2. **Regras de Firewall**:
   ```
   # Exemplo de regras (adaptar conforme necessidade)
   - 443: Permitir todos (HTTPS)
   - 8085-8285: Permitir apenas IPs internos (APIs)
   - 27185: Bloquear acesso externo (MongoDB)
   - 9x85: Permitir apenas IPs administrativos
   ```

### Monitoramento e Logs

1. **Monitoramento de Segurança**:
   - Configurar alertas para tentativas de login mal sucedidas
   - Monitorar padrões suspeitos de tráfego
   - Acompanhar uso anômalo de recursos
   - Monitorar alterações em arquivos críticos

2. **Gestão de Logs**:
   - Centralizar logs no Elasticsearch
   - Implementar retenção adequada
   - Configurar alertas para eventos críticos
   - Manter logs de auditoria

### Procedimentos de Manutenção

1. **Backups**:
   - MongoDB: backup diário com retenção de 30 dias
   - MinIO: backup semanal com retenção de 90 dias
   - Elasticsearch: snapshot diário
   - Testar restauração mensalmente

2. **Atualizações**:
   - Manter todos os serviços atualizados
   - Testar atualizações em ambiente de homologação
   - Manter registro de versões e mudanças

### Procedimentos de Emergência

1. **Em caso de incidente**:
   - Isolar serviços comprometidos
   - Coletar logs para análise
   - Restaurar último backup válido
   - Documentar ocorrência

2. **Contatos de Emergência**:
   - Definir equipe de resposta
   - Manter lista de contatos atualizada
   - Estabelecer protocolo de comunicação
