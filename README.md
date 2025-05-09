# 🕵️‍♂️ Tô de Olho - Arquitetura do Projeto

## 📖 Visão Geral
O **Tô de Olho** é um sistema automatizado para coletar, processar e armazenar informações públicas sobre atividades políticas, garantindo transparência e acesso democrático aos dados.

A arquitetura é composta por diferentes módulos que funcionam de forma distribuída, utilizando **Docker** e **RabbitMQ** para orquestração de tarefas e processamento assíncrono.

Créditos à Câmara Municipal de Salvador pela disponibilidade dos dados de transparência - https://www.cms.ba.gov.br

---

## 🏗️ Arquitetura do Sistema
O sistema segue uma abordagem baseada em **microserviços**, onde cada componente tem uma responsabilidade bem definida:

- **Crawlers** 🕷️ → Capturam dados de sites públicos e armazenam em JSONs.
- **RabbitMQ (Broker)** 📨 → Coordena a comunicação entre os módulos.
- **JSON Processor** 🏭 → Processa e organiza os arquivos baixados pelos crawlers e envia os dados limpos para a API.
- **API** 🌐 → Exposição de dados via REST para consumo público.
- **Banco de Dados (PostgreSQL)** 🗄️ → Armazena informações processadas.
- **Front-end (Futuro Blog)** 📰 → Interface para exibição de dados coletados.

---

## 🔄 **Fluxo de Dados**
1️⃣ **Os Crawlers** coletam arquivos JSON e armazenam em `/data/crawler_nome/`.  
2️⃣ **O Broker (RabbitMQ)** recebe mensagens dos crawlers e as distribui.  
3️⃣ **O JSON Processor** escuta a fila e processa os arquivos JSON.  
4️⃣ **A API** recebe os dados processados e armazena no **PostgreSQL**.  
5️⃣ **O Front-end (futuro blog)** exibe os dados de forma acessível.

---

## 📂 Estrutura de Diretórios
```
/to-de-olho
│── /api                        # API de dados públicos
│── /crawlers                   # Scripts de coleta de dados
│   ├── /packages
│   │   ├── contractDataJob      # Crawler de contratos públicos
│   │   ├── councilorDataJob     # Crawler de vereadores
│   │   ├── frequencyDataJob     # Crawler de frequência parlamentar
│   │   ├── ...                  # Outros crawlers
│── /json-processor              # Processamento de JSONs gerados pelos crawlers
│── /postgresql                  # Banco de dados PostgreSQL
│── /validate-user-worker        # Serviço de validação de usuários
│── /shared_data                 # 📌 Armazena os JSONs baixados pelos crawlers (volume compartilhado)
│── /docker-compose.prod.yml     # Arquivo de orquestração dos serviços
│── /README.md                   # Documentação do projeto
```

---

## 🏗️ **Componentes Principais**
### 🔹 **1. Crawlers (Coleta de Dados)**
Os crawlers utilizam **Puppeteer** e **Playwright** para navegar e extrair os dados diretamente das páginas públicas, salvando-os em arquivos JSON.

- Cada crawler armazena os arquivos JSON dentro de `/shared_data/<nome-do-crawler>/`
- Os arquivos são salvos com timestamps para auditoria.

📌 **Exemplo de execução manual:**
```bash
docker exec -it to-de-olho-crawlers-1 npm run start-contract
```

---

### 🔹 **2. RabbitMQ (Broker de Mensagens)**
O RabbitMQ é o intermediador entre os crawlers e o processador de JSONs.

- Cada vez que um JSON é baixado por um crawler, um evento é enviado para a fila `json-processor-queue`.

📌 **Verificando as filas no RabbitMQ:**
```bash
curl -u to-de-olho:olho-de-to -X GET http://broker:15672/api/queues
```

---

### 🔹 **3. JSON Processor (Processamento)**
O `json-processor` recebe os arquivos e faz a limpeza e organização dos dados.

- Ele escuta a fila `json-processor-queue`
- Processa os arquivos localizados em `/data/`
- Envia os dados limpos para a **API**

📌 **Executando o `json-processor` manualmente:**
```bash
docker exec -it to-de-olho-json-processor-1 /app/main
```

---

### 🔹 **4. API (Serviço de Dados)**
A API expõe os dados coletados e processados.

- Criada com **Golang**
- Permite consultas aos dados via **REST API**
- Se comunica diretamente com o **PostgreSQL**

📌 **Verificando se a API está rodando:**
```bash
curl http://localhost:3000/api/v1/health
```

---

### 🔹 **5. Banco de Dados (PostgreSQL)**
O PostgreSQL armazena todas as informações coletadas.

📌 **Acessando o banco manualmente:**
```bash
docker exec -it to-de-olho-db-1 psql -U prod_username -d to_de_olho_prod
```

---

## 🔧 **Configuração do Docker**
Todos os serviços são orquestrados via **Docker Compose**.

📌 **Subindo o ambiente de produção:**
```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

📌 **Parando e limpando containers órfãos:**
```bash
docker-compose -f docker-compose.prod.yml down --remove-orphans
```

📌 **Reiniciando um serviço específico:**
```bash
docker-compose -f docker-compose.prod.yml restart crawlers
```

---

## 📅 **Próximos Passos**
1️⃣ **Implementar o Front-end (Blog) para exibir os dados coletados.**  
2️⃣ **Otimizar a estrutura de mensagens no RabbitMQ para melhorar o processamento paralelo.**  
3️⃣ **Adicionar testes automatizados para garantir qualidade e estabilidade.**

---

## 🚀 **Conclusão**
O **Tô de Olho** está evoluindo para se tornar uma ferramenta robusta de **transparência política**.  
Com uma arquitetura distribuída baseada em microserviços, conseguimos garantir escalabilidade, flexibilidade e confiabilidade no processamento dos dados.

Se tiver sugestões ou quiser contribuir, entre em contato! 💡
