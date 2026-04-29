# Guia de Deploy no Servidor GSORT

Este documento contém o passo a passo para implantar o projeto **Tô De Olho** no servidor do laboratório GSORT. A infraestrutura utiliza Docker Compose e inclui o túnel do Cloudflare para roteamento de domínio público (`todeolho.org`) sem necessidade de abrir portas no firewall.

## 1. Pré-requisitos na Máquina do GSORT

Certifique-se de que a máquina possui instalados:
- **Git**
- **Docker** e **Docker Compose** (V2)
- Acesso à internet para baixar as imagens e conectar ao Cloudflare.

## 2. Clonando o Repositório

No terminal do servidor, execute:

```bash
# Clone o repositório
git clone https://github.com/SEU_USUARIO/todeolho.git
cd todeolho

# Acesse a branch específica de infraestrutura
git checkout infra/docker-gsort
```

## 3. Configuração do Ambiente (`.env.gsort`)

O arquivo `.env.gsort` já contém a estrutura necessária. Certifique-se de que os seguintes tokens estão válidos e atualizados:

1. **TRANSPARENCIA_API_KEY**: Chave do Portal da Transparência (obrigatória para as emendas parlamentares).
2. **CLOUDFLARE_TUNNEL_TOKEN**: Token do túnel gerado na dashboard do Cloudflare Zero Trust. Este token garante que o tráfego do domínio chegue ao container `web` rodando no GSORT.

> [!CAUTION]
> Caso o token do Cloudflare tenha sido revogado ou expirado, será necessário gerar um novo no painel da Cloudflare e atualizar o arquivo `.env.gsort`.

## 4. Iniciando os Containers

Com o `.env.gsort` configurado, inicie a infraestrutura:

```bash
# O comando construirá as imagens (frontend e backend) e iniciará o banco de dados e o túnel
docker compose --env-file .env.gsort up -d --build
```

### O que vai subir:
- `todeolho-db`: PostgreSQL 15 com os dados persistidos no volume local.
- `todeolho-api`: Backend em Go rodando na porta 8080.
- `todeolho-web`: Frontend em Next.js rodando na porta 3000.
- `todeolho-tunnel`: Cliente do Cloudflared, que conectará a porta 3000 ao domínio `todeolho.org`.

## 5. Verificando a Saúde dos Serviços

Para conferir se todos os serviços estão no ar e sem erros:

```bash
# Verificar status dos containers
docker compose ps

# Acompanhar os logs (exemplo: backend)
docker compose logs -f api

# Acompanhar os logs do túnel Cloudflare
docker compose logs -f cloudflared
```

## 6. Sincronização Inicial de Dados (Backfill)

Como o banco de dados estará zerado na primeira inicialização, será necessário rodar a rotina de sincronização (backfill) para buscar os senadores, emendas e votações:

```bash
# Executando um curl interno no container do backend para iniciar a rotina de sync manual (ou aguarde o cron diário)
docker exec todeolho-api curl -X POST http://localhost:8080/api/v1/sync/full
```

## 7. Manutenção

Para parar a aplicação ou aplicar atualizações:

```bash
# Parar os serviços
docker compose down

# Atualizar o código e reiniciar
git pull origin infra/docker-gsort
docker compose --env-file .env.gsort up -d --build
```
