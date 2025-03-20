# Etapa de build
FROM node:18-alpine AS build
WORKDIR /app/crawlers

# Adicionando package.json primeiro para usar cache eficiente do Docker
COPY ./crawlers/package.json ./crawlers/package-lock.json ./

# Instalando dependências com cache do npm (ACELERA a build!)
RUN --mount=type=cache,target=/root/.npm \
    npm install --no-audit --no-fund --legacy-peer-deps

# Instalando Lerna globalmente
RUN npm install -g lerna

# Copiando restante do código após instalação das dependências
COPY ./crawlers /app/crawlers

# Etapa de produção
FROM mcr.microsoft.com/playwright:v1.34.0-focal AS production
WORKDIR /app/crawlers

# Instalando dependências do sistema em um único RUN (EVITA múltiplas camadas desnecessárias)
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget gnupg libnss3 libatk-bridge2.0-0 libxcomposite1 \
    libxdamage1 libxrandr2 x11-utils libgbm-dev libasound2 jq \
    build-essential ca-certificates \
    fonts-liberation \
    xdg-utils \
    libu2f-udev \
    libvulkan1 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Adicionando repositório oficial do Google Chrome e instalando a versão estável
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
    apt-get update && apt-get install -y --no-install-recommends google-chrome-stable

# Copiando o código da etapa de build para a imagem de produção
COPY --from=build /app/crawlers /app/crawlers

# Instalando `amqplib` para suporte ao RabbitMQ
RUN npm install amqplib --omit=dev

# Definindo o caminho correto do Chrome para Puppeteer e Playwright
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome
ENV PLAYWRIGHT_BROWSERS_PATH=/usr/bin/google-chrome
ENV PLAYWRIGHT_CHROMIUM_PATH=/usr/bin/google-chrome

# Instalando Lerna globalmente na produção
RUN npm install -g lerna

# Permissão de execução para o script principal dos crawlers
RUN chmod +x /app/crawlers/run-crawlers.sh

# Comando de execução padrão
CMD ["bash", "/app/crawlers/run-crawlers.sh"]
