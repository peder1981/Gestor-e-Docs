FROM nginx:1.25

# Criar diretórios necessários
RUN mkdir -p /var/log/nginx /etc/nginx/certs

# Copiar script de inicialização personalizado
COPY ./docker-entrypoint.d/40-fix-nginx-conf.sh /docker-entrypoint.d/
RUN chmod +x /docker-entrypoint.d/40-fix-nginx-conf.sh

# Copiar configurações
COPY ./conf/nginx.conf /etc/nginx/nginx.conf
COPY ./conf/default.conf /etc/nginx/conf.d/default.conf

# Copiar certificados SSL
COPY ./certs/nginx.crt /etc/nginx/certs/nginx.crt
COPY ./certs/nginx.key /etc/nginx/certs/nginx.key
