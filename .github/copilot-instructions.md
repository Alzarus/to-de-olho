# GitHub Copilot Instructions for Tô De Olho

## 📦 Contexto do Projeto

Este repositório contém o projeto **Tô De Olho**, uma arquitetura modular baseada em containers Docker que realiza coleta, processamento e disponibilização de dados públicos municipais, com foco em transparência, fiscalização cidadã e acessibilidade de informações. O ecossistema se baseia em múltiplos serviços (API, crawlers, workers) que se comunicam por fila de mensagens (RabbitMQ), utilizando banco de dados PostgreSQL e estrutura escalável.

## 🚀 Objetivo das Instruções

Estas instruções ajudam o GitHub Copilot a gerar sugestões mais relevantes e consistentes com a arquitetura do projeto, promovendo boas práticas, reuso de código e modularidade no desenvolvimento dos crawlers, workers e API.

---

## 📂 Estrutura Esperada de Código

```
.
├── api/
├── crawlers/
│   ├── packages/
│   │   ├── contractDataJob/
│   │   ├── councilorDataJob/
│   │   ├── frequencyDataJob/
│   │   ├── generalProductivityDataJob/
│   │   ├── propositionDataJob/
│   │   ├── propositionProductivityDataJob/
│   │   ├── travelExpensesDataJob/
│   │   ├── tseDataJob/ (validação eleitores tse/liberação login blog)
│   └── run-crawlers.sh
├── json-processor/
├── validate-user-worker/
├── postgresql/
│   └── create-database.sql
└── docker-compose.prod.yml
```

---

## 🔁 Comunicação entre Serviços

- A comunicação entre `crawlers`, `validate-user-worker` e `json-processor` ocorre via **RabbitMQ**.
- Os crawlers devem utilizar o módulo `broker.js` (localizado em `crawlers/broker.js`) para publicar mensagens.
- As credenciais padrão para acesso ao painel do RabbitMQ (para depuração) são `to-de-olho:olho-de-to`.
- As filas seguem o padrão de nomeação:  
  - `json-processor-queue` para arquivos `.json` gerados
  - `validate-user-queue` para verificações de autenticação

### Exemplo de Publicação em Fila

```js
// Exemplo conceitual de envio de mensagem para json-processor-queue usando crawlers/broker.js
const broker = require('../broker'); // Supondo que está em um script dentro de packages/contractDataJob
const jobName = 'contractDataJob'; // Nome do Job atual
const fileName = 'contract_20250308_214530.json'; // Nome do arquivo gerado
const filePath = `/shared_data/${jobName}/${fileName}`; // Caminho dinâmico
const message = { filePath: filePath, jobName: jobName };
broker.publishMessage('json-processor-queue', JSON.stringify(message));
console.log(`Mensagem enviada para json-processor-queue: ${filePath}`);
```

---

## 📤 Padrão de Saída JSON dos Crawlers

Crawlers devem gerar arquivos `.json` no diretório `/shared_data/<JobName>/`, com nome no formato:

```
<nome-do-arquivo>_YYYYMMDD_HHmmss.json
```

Exemplo:

```
contract_20250308_214530.json
```

---

## 📁 Diretórios Compartilhados

- Os arquivos `.json` gerados devem estar disponíveis no volume compartilhado `shared_data`, montado nos containers `crawlers` e `json-processor`. Este volume é definido no arquivo `docker-compose.prod.yml`.

---

## 📚 Logs e Auditoria

- Os logs de execução devem ser claros e com timestamp no formato brasileiro.
- Arquivos de log devem ser armazenados no diretório `crawlers/logs/`.
  ```
  08/03/2025 21:45:30 - Download concluído: /shared_data/contract/contract_20250308_214530.json
  ```

---

## ✅ Boas Práticas

- Cada `Job` deve estar dentro de `packages/<JobName>`, com seus próprios scripts JS.
- Incluir tratamento de erro robusto.
- Nomear diretórios e arquivos com consistência.
- Ao usar Puppeteer ou Playwright, configurar corretamente o path do Chrome.
- Usar variáveis de ambiente para URLs e paths quando possível.

---

## 🧠 Sugestões Copilot

Quando sugerir novo código, Copilot deve:

- Criar funções reutilizáveis para downloads, criação de pastas e geração de nomes de arquivos, preferencialmente localizadas em `crawlers/utils/`.
- Usar `path.join(__dirname, ...)` para caminhos de arquivos.
- Seguir o padrão de `writeLog`, `getFormattedPath`, `checkAndCreateFolder`.

---

## 📎 Exemplo de Geração de JSON

```js
const fs = require('fs');
const path = require('path');

const jobName = 'contractDataJob'; // Exemplo de nome do Job
const baseFileName = 'contract'; // Base para o nome do arquivo, pode ser derivado do jobName

const now = new Date();
const year = now.getFullYear();
const month = (now.getMonth() + 1).toString().padStart(2, '0');
const day = now.getDate().toString().padStart(2, '0');
const hours = now.getHours().toString().padStart(2, '0');
const minutes = now.getMinutes().toString().padStart(2, '0');
const seconds = now.getSeconds().toString().padStart(2, '0');

const timestamp = `${year}${month}${day}_${hours}${minutes}${seconds}`;
const filename = `${baseFileName}_${timestamp}.json`;
// jsonData é o objeto com os dados a serem salvos
// const jsonData = { /* ... seus dados aqui ... */ }; 

const outputPath = path.join("/shared_data", jobName, filename);

// Certifique-se de que o diretório de output para o jobName existe
// fs.mkdirSync(path.join("/shared_data", jobName), { recursive: true }); // Descomente se necessário criar o diretório

fs.writeFileSync(outputPath, JSON.stringify(jsonData, null, 2));
console.log(`Arquivo JSON gerado em: ${outputPath}`);
```