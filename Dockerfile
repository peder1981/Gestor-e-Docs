# Etapa de build
FROM node:18-alpine AS build

WORKDIR /app

# Copiar package.json e package-lock.json (ou yarn.lock)
COPY package*.json ./

# Instalar dependências
RUN npm install

# Copiar o restante dos arquivos da aplicação
COPY . .

# Construir a aplicação para produção
RUN npm run build

# Etapa de produção - servir com Nginx
FROM nginx:stable-alpine

# Copiar os arquivos construídos da etapa anterior para o diretório padrão do Nginx
COPY --from=build /app/build /usr/share/nginx/html

# Copiar a configuração personalizada do Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expor a porta 80 (Nginx escuta na porta 80 por padrão)
EXPOSE 80

# Comando para iniciar o Nginx
CMD ["nginx", "-g", "daemon off;"]
