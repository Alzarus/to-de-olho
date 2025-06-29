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
- **Validate User Worker** 🔐 → Serviço para validação de usuários no TSE (verifica se são eleitores válidos).
- **Front-end (Futuro Blog)** 📰 → Interface para exibição de dados coletados.

---

## 🔄 **Fluxo de Dados**

1️⃣ **Os Crawlers** coletam arquivos JSON e armazenam em `/shared_data/<JobName>/`.  
2️⃣ **O Broker (RabbitMQ)** recebe mensagens dos crawlers e as distribui através da fila `json-processor-queue`.  
3️⃣ **O JSON Processor** escuta a fila e processa os arquivos JSON localizados no volume compartilhado.  
4️⃣ **A API** recebe os dados processados e armazena no **PostgreSQL**.  
5️⃣ **O Front-end (futuro blog)** exibe os dados de forma acessível.  
6️⃣ **O Validate User Worker** recebe solicitações, aciona o crawler `tseDataJob` para verificar se o usuário é um eleitor válido no TSE, e libera o acesso ao blog.

---

## 📂 Estrutura de Diretórios

```
/to-de-olho
│── /api                        # API de dados públicos
│   ├── /cmd                    # Ponto de entrada da API
│   ├── /configs                # Configurações de banco de dados
│   ├── /controllers            # Controladores para cada tipo de dado
│   ├── /models                 # Estruturas de dados
│   ├── /repositories           # Camada de acesso ao banco de dados
│   └── /routes                 # Definições de rotas da API
│── /crawlers                   # Scripts de coleta de dados
│   ├── broker.js               # Módulo para comunicação com o RabbitMQ
│   ├── run-crawlers.sh         # Script para execução de crawlers
│   ├── /logs                   # Logs de execução dos crawlers
│   ├── /packages
│   │   ├── contractDataJob      # Crawler de contratos públicos
│   │   ├── councilorDataJob     # Crawler de vereadores e dados pessoais
│   │   ├── frequencyDataJob     # Crawler de frequência parlamentar
│   │   ├── generalProductivityDataJob # Crawler de dados de produtividade geral
│   │   ├── propositionDataJob   # Crawler de proposições legislativas
│   │   ├── propositionProductivityDataJob # Crawler de produtividade de proposições
│   │   ├── travelExpensesDataJob # Crawler de despesas de viagem
│   │   └── tseDataJob           # Crawler de dados do TSE para validação de eleitores
│   └── /utils                  # Utilitários compartilhados entre crawlers
│── /json-processor              # Processamento de JSONs gerados pelos crawlers
│   ├── /api                    # Interface para comunicação com a API
│   ├── /broker                 # Comunicação com RabbitMQ
│   ├── /logs                   # Logs do processador
│   ├── /processing             # Lógica de processamento de dados
│   └── /utils                  # Utilitários compartilhados
│── /postgresql                  # Banco de dados PostgreSQL
│   └── create-database.sql     # Script de criação do banco de dados
│── /validate-user-worker        # Serviço de validação de usuários
│── /docs                        # Documentação adicional do projeto
│── /shared_data                 # 📌 Armazena os JSONs baixados pelos crawlers (volume compartilhado)
│── docker-compose.prod.yml      # Arquivo de orquestração dos serviços
└── README.md                    # Documentação do projeto
```

---

## 🏗️ **Componentes Principais**

### 🔹 **1. Crawlers (Coleta de Dados)**

Os crawlers utilizam **Puppeteer** e **Playwright** para navegar e extrair os dados diretamente das páginas públicas, salvando-os em arquivos JSON.

- Cada crawler (localizado em `crawlers/packages/<JobName>`) armazena os arquivos JSON dentro de `/shared_data/<JobName>/`
- Os arquivos são salvos no formato `<nome-do-arquivo>_YYYYMMDD_HHmmss.json` para auditoria
- Os logs de execução são armazenados em `crawlers/logs/` com timestamp no formato brasileiro

📌 **Exemplo de execução manual:**

```bash
docker exec -it to-de-olho-crawlers-1 npm run start-contract
```

---

### 🔹 **2. RabbitMQ (Broker de Mensagens)**

O RabbitMQ é o intermediador entre os crawlers e o processador de JSONs, além de gerenciar a autenticação de usuários.

- Cada vez que um JSON é baixado por um crawler, um evento é enviado para a fila `json-processor-queue` via `broker.js`
- Validações de autenticação são processadas através da fila `validate-user-queue`
- O serviço `validate-user-worker` despacha requisições para o `tseDataJob` verificar dados de eleitores no TSE
- Credenciais padrão para acesso ao painel: `to-de-olho:olho-de-to`

📌 **Verificando as filas no RabbitMQ:**

```bash
curl -u to-de-olho:olho-de-to -X GET http://broker:15672/api/queues
```

---

### 🔹 **3. JSON Processor (Processamento)**

O `json-processor` recebe os arquivos e faz a limpeza e organização dos dados.

- Escuta a fila `json-processor-queue` e recebe mensagens com o caminho do arquivo e nome do job
- Processa os arquivos localizados no volume compartilhado `/shared_data/<JobName>/`
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

### 🔹 **6. Validate User Worker (Validação de Usuários)**

O `validate-user-worker` é responsável por validar se os usuários do blog são eleitores reais, verificando seus dados no Tribunal Superior Eleitoral (TSE).

- Recebe requisições de validação via fila `validate-user-queue`
- Aciona o crawler `tseDataJob` para obter dados do eleitor no TSE
- Verifica se o usuário é um eleitor válido a partir do título de eleitor/CPF, data de nascimento e nome da mãe
- Libera o acesso ao blog após validação bem-sucedida

📌 **Exemplo de execução do crawler TSE manualmente:**

```bash
docker exec -it to-de-olho-crawlers-1 bash -c "npm_config_tituloCpf=\"01245678910\" npm_config_dataNascimento=\"01/01/2000\" npm_config_nomeMae=\"Nome da Mãe\" lerna run start --scope tse-data-job"
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

## 💻 **Guia para Desenvolvedores**

### Criando um novo Crawler

1. Crie um novo diretório em `crawlers/packages/<NomeDoJob>`
2. Implemente o script principal em `src/<NomeDoJob>.js`
3. Use o módulo `broker.js` para publicar mensagens na fila do RabbitMQ
4. Siga o padrão de nomenclatura para arquivos JSON: `<nome-do-arquivo>_YYYYMMDD_HHmmss.json`
5. Salve os arquivos no caminho `/shared_data/<NomeDoJob>/`

### Exemplo de código para comunicação com o broker:

```js
const broker = require("../../../broker");
const jobName = "meuNovoJob";
const fileName = `dados_${new Date().toISOString().replace(/[:.]/g, "")}.json`;
const filePath = `/shared_data/${jobName}/${fileName}`;

// Após salvar o arquivo JSON
const message = { filePath, jobName };
broker.publishMessage("json-processor-queue", JSON.stringify(message));
console.log(`Mensagem enviada para json-processor-queue: ${filePath}`);
```

## 🚀 **Conclusão**

O **Tô de Olho** está evoluindo para se tornar uma ferramenta robusta de **transparência política**.  
Com uma arquitetura distribuída baseada em microserviços, conseguimos garantir escalabilidade, flexibilidade e confiabilidade no processamento dos dados.

Se tiver sugestões ou quiser contribuir, entre em contato! 💡
