#!/bin/sh

# Este script é executado antes do início da aplicação principal.
# É usado para executar migrações de banco de dados ou outras tarefas de configuração.

# Iniciar o serviço principal
echo "[ENTRYPOINT] Iniciando o identity-service..."
exec /app/identity-service
