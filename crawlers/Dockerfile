# Etapa de build
FROM node:18-slim AS build
WORKDIR /app/crawlers

# Copiando apenas os arquivos necessários para dependências
COPY ./crawlers/package.json ./crawlers/package-lock.json ./

# Instalando dependências, incluindo lerna globalmente
RUN npm ci --prefer-offline --cache /app/.npm && \
    npm install -g lerna && \
    rm -rf /app/.npm

# Copiando o restante do código
COPY ./crawlers /app/crawlers

# Etapa de produção
FROM mcr.microsoft.com/playwright:v1.34.0-focal AS production
WORKDIR /app/crawlers

# Instalando dependências do sistema para cron, scripts e Google Chrome/Chromium
RUN apt-get update && \
    apt-get install -y curl cron wget gnupg libnss3 libatk-bridge2.0-0 libxcomposite1 libxdamage1 libxrandr2 x11-utils libgbm-dev libasound2 && \
    wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | gpg --dearmor > /usr/share/keyrings/google-chrome.gpg && \
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
    apt-get update && \
    apt-get install -y google-chrome-stable && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copiando o código e as dependências do build
COPY --from=build /app/crawlers /app/crawlers

# Garantindo que o lerna esteja disponível no ambiente de produção
RUN npm install -g lerna

# Garantindo que o script tem permissão de execução
RUN chmod +x /app/crawlers/run-crawlers.sh

# CMD para iniciar o serviço de crawlers
CMD ["bash", "/app/crawlers/run-crawlers.sh"]
