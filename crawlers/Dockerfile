# Etapa de build
FROM node:18-alpine AS build
WORKDIR /app/crawlers

# Copia apenas arquivos necessários para instalação das dependências
COPY ./crawlers/package.json ./crawlers/package-lock.json ./

# Instalando dependências sem cache desnecessário
RUN npm ci --prefer-offline --no-audit --no-fund && \
    rm -rf /root/.npm

# Copia restante do código
COPY ./crawlers /app/crawlers

# Etapa de produção
FROM mcr.microsoft.com/playwright:v1.34.0-focal AS production
WORKDIR /app/crawlers

# Instalando dependências do sistema em um único RUN para melhor aproveitamento de cache
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget gnupg libnss3 libatk-bridge2.0-0 libxcomposite1 \
    libxdamage1 libxrandr2 x11-utils libgbm-dev libasound2 jq \
    build-essential ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Instalando Google Chrome para Puppeteer e configurando chave de repositório
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | gpg --dearmor > /usr/share/keyrings/google-chrome.gpg && \
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
    apt-get update && apt-get install -y google-chrome-stable && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Criando cache local do Playwright para evitar re-downloads contínuos
RUN mkdir -p /ms-playwright && npx playwright install --with-deps chromium

# Copiando o código e as dependências da etapa de build
COPY --from=build /app/crawlers /app/crawlers

# Garantindo que o lerna esteja disponível no ambiente de produção
RUN npm install -g lerna

# Definindo o caminho correto do Chromium
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
ENV PLAYWRIGHT_CHROMIUM_PATH=/ms-playwright/chromium-1064/chrome-linux/chrome

# Criando link simbólico para Chromium para compatibilidade
RUN ln -sf /ms-playwright/chromium-1064/chrome-linux/chrome /usr/bin/chromium

# Garantindo que o script tem permissão de execução
RUN chmod +x /app/crawlers/run-crawlers.sh

# CMD para iniciar o serviço de crawlers
CMD ["bash", "/app/crawlers/run-crawlers.sh"]
