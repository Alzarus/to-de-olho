# Modificação no Dockerfile dos crawlers
FROM node:18-alpine AS build
WORKDIR /app/crawlers

# Copie apenas os arquivos package.json/lock antes das dependências
COPY ./crawlers/package*.json ./
COPY ./crawlers/lerna.json ./
COPY ./crawlers/packages/*/package.json ./packages/

# Crie diretórios para cada pacote e configure npm para maximizar velocidade
RUN for dir in packages/*/; do mkdir -p "${dir}"; done && \
    npm config set registry https://registry.npmjs.org/ && \
    npm config set fetch-retries 5 && \
    npm config set fetch-retry-factor 10 && \
    npm config set fetch-retry-mintimeout 60000 && \
    npm config set fetch-retry-maxtimeout 300000

# Instale as dependências com cache otimizado e flags de velocidade
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit --no-fund --loglevel=error --no-progress

# Agora copie o restante do código
COPY ./crawlers /app/crawlers

# Etapa de produção - Use uma imagem que já tenha o Chrome instalado
FROM mcr.microsoft.com/playwright:v1.34.0-focal AS production
WORKDIR /app/crawlers

# Instalando apenas dependências adicionais necessárias
RUN apt-get update && apt-get install -y --no-install-recommends \
    jq \
    ca-certificates \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Identificar o caminho correto do Chromium na imagem
RUN which chromium || which google-chrome || which chromium-browser || echo "Chromium não encontrado" && \
    ls -la /ms-playwright/ && \
    find /ms-playwright -name "chrome*" -o -name "chromium*"

# Copiando o código da etapa de build para a imagem de produção
COPY --from=build /app/crawlers /app/crawlers

# Instalando apenas amqplib para suporte ao RabbitMQ
RUN npm install -g lerna && \
    npm install amqplib --omit=dev

# Definindo o caminho correto do Chrome para Puppeteer e Playwright
ENV PUPPETEER_EXECUTABLE_PATH=/ms-playwright/chromium-1064/chrome-linux/chrome
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
ENV PLAYWRIGHT_CHROMIUM_PATH=/ms-playwright/chromium-1064/chrome-linux/chrome

# Criando um link simbólico para compatibilidade
RUN mkdir -p /usr/bin && \
    ln -sf /ms-playwright/chromium-1064/chrome-linux/chrome /usr/bin/google-chrome-stable

# Atribuindo permissões de execução
RUN chmod +x /ms-playwright/chromium-1064/chrome-linux/chrome && \
    chmod +x /usr/bin/google-chrome-stable && \
    chmod +x /app/crawlers/run-crawlers.sh

# Criar diretórios de log e dados compartilhados
RUN mkdir -p /logs /shared_data

# Comando de execução padrão
CMD ["bash", "/app/crawlers/run-crawlers.sh"]
