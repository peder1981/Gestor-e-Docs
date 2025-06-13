#!/bin/sh
set -e

# Remover qualquer nginx.conf em conf.d
rm -f /etc/nginx/conf.d/nginx.conf

# Garantir que o nginx.conf principal existe
if [ ! -f /etc/nginx/nginx.conf ]; then
    echo "Erro: /etc/nginx/nginx.conf não encontrado!"
    exit 1
fi

echo "Configuração do nginx verificada e corrigida."
