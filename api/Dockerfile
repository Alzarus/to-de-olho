FROM node:18-slim AS build
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:18-slim AS production
WORKDIR /app

RUN apt-get update && apt-get install -y curl && apt-get clean

COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules

CMD ["node", "dist/main"]
