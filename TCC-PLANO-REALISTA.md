# 🎯 PLANO REALISTA - TCC "Tô De Olho" (Ago/25 → Fev/26)

## 🚨 **FOCO ABSOLUTO: MVP que FUNCIONA > Arquitetura Perfeita**

### 📊 **Cronograma Simplificado (6 meses)**

| Mês | Foco Principal | Entregas |
|-----|----------------|----------|
| **Set/25** | Backend Core | API Deputados + Despesas funcionando |
| **Out/25** | Frontend Base | Dashboard básico com dados reais |
| **Nov/25** | Recursos Essenciais | Busca, filtros, gráficos simples |
| **Dez/25** | UX/Polimento | Interface bonita, responsiva |
| **Jan/26** | TCC Writing | Documentação, relatório final |
| **Fev/26** | Apresentação | Demo funcionando + defesa |

---

## 🎯 **MVP REDUZIDO - O QUE REALMENTE IMPORTA**

### ✅ **Core Features (OBRIGATÓRIAS)**
1. **Listar Deputados** com foto, partido, estado
2. **Gastos por Deputado** com gráficos simples
3. **Busca e Filtros** por nome, estado, partido
4. **Dashboard de Transparência** com métricas básicas
5. **Responsivo** para mobile/desktop

### 🚀 **Nice-to-Have (SE DER TEMPO)**
1. **Comparar Deputados** (2-3 lado a lado)
2. **Alertas de Gastos** excessivos
3. **Sistema de Avaliação** simples (likes/stars)
4. **Export de Dados** (PDF/Excel)

### ❌ **CORTAR AGORA (Deixar para V2)**
- ❌ Gamificação completa
- ❌ Fórum de discussões  
- ❌ IA Generativa (Gemini)
- ❌ Microsserviços (usar monolito)
- ❌ Plebiscitos digitais
- ❌ Sistema de usuários complexo

---

## 🛠️ **STACK SIMPLIFICADA (Sem Over-Engineering)**

### **Backend Simples (Go)**
```
Monolito Go + Gin + GORM
├── main.go
├── handlers/          # REST APIs
├── models/           # Structs dos dados  
├── services/         # Business logic
├── database/         # PostgreSQL
└── utils/           # Helpers
```

### **Frontend Direto (Next.js)**
```
Next.js 15 + TypeScript + Tailwind
├── pages/api/        # API routes (se precisar)
├── components/       # Componentes reutilizáveis
├── pages/           # Páginas da aplicação
├── hooks/           # Custom hooks
└── utils/           # Helpers
```

### **Banco Simplificado**
```sql
-- 3 tabelas principais apenas
deputados (id, nome, partido, uf, foto_url...)
despesas (id, deputado_id, valor, tipo, data...)
partidos (id, sigla, nome)
```

---

## 📈 **ESTRATÉGIA DE SUCESSO PARA TCC**

### 🎯 **1. Começar pelo Backend (Setembro)**
```bash
# Criar estrutura mínima viável
mkdir backend
cd backend
go mod init to-de-olho
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres

# 3 endpoints principais:
GET /deputados              # Lista todos
GET /deputados/:id          # Detalhes + gastos  
GET /deputados/:id/despesas # Gastos paginados
```

### 🎨 **2. Frontend Funcional (Outubro)**
```bash
# Next.js com Tailwind
npx create-next-app@latest frontend --typescript --tailwind
npm install recharts          # Gráficos simples
npm install @headlessui/react # Componentes prontos
npm install lucide-react      # Ícones modernos

# 4 páginas principais:
/                    # Home com stats gerais
/deputados          # Lista paginada
/deputados/[id]     # Perfil do deputado
/transparencia      # Dashboard geral
```

### 📊 **3. Dados Reais da Câmara (Novembro)**
```go
// ETL simples para popular dados
func SyncDeputados() {
    // Consumir API oficial: https://dadosabertos.camara.leg.br/api/v2/deputados
    // Salvar no PostgreSQL
    // Rodar 1x por dia via cron
}

func SyncDespesas() {
    // Para cada deputado, buscar despesas do ano atual
    // Focar apenas em 2024-2025 (dados recentes)
}
```

---

## 🎓 **FOCO NO TCC - Não na Startup**

### ✅ **O que Professores/Banca VÃO VALORIZAR:**
1. **Funciona de verdade** ✅
2. **Código limpo e organizado** ✅
3. **Dados reais e atuais** ✅
4. **Interface bonita e responsiva** ✅
5. **Documentação clara** ✅
6. **Testes básicos** ✅

### ❌ **O que NÃO vai somar pontos:**
- ❌ Over-engineering (microsserviços, DDD complexo)
- ❌ Features que não funcionam 100%
- ❌ Muitas tecnologias sem necessidade
- ❌ Gamificação incompleta

---

## 📝 **ESTRUTURA DO TCC (Paralelo ao Código)**

### **Capítulos Sugeridos:**
1. **Introdução** - Transparência política no Brasil
2. **Fundamentação Teórica** - E-gov, transparência, participação
3. **Metodologia** - Análise de requisitos, escolha da stack
4. **Desenvolvimento** - Backend, frontend, integração
5. **Resultados** - Screenshots, métricas, testes
6. **Conclusão** - Impacto, limitações, trabalhos futuros

### **Métricas para Mostrar Valor:**
- Nº de deputados cadastrados (513)
- Volume de despesas processadas (+100k registros)
- Performance (tempo de resposta <500ms)
- Responsividade (Mobile + Desktop)

---

## 🚀 **PRÓXIMOS PASSOS IMEDIATOS**

### **Esta Semana (Ago/25):**
1. ✅ Escolher: Monolito Go OU Next.js Full-Stack
2. ✅ Configurar banco PostgreSQL local
3. ✅ Criar 1º endpoint `/deputados` funcionando
4. ✅ Testar com dados da API da Câmara

### **Semana que vem:**
1. Frontend básico consumindo API
2. Layout responsivo com Tailwind
3. Deploy simples (Vercel + Railway/Supabase)

---

## 💡 **DICAS DE OURO PARA TCC**

### 🎯 **Mindset Correto:**
- **"Done is better than perfect"**
- **Funcionalidade > Arquitetura**
- **Simplicidade > Complexidade**
- **Demo working > 1000 features broken**

### 📱 **Demonstração Final (Fev/26):**
- Site responsivo funcionando
- Dados reais atualizados
- 3-4 funcionalidades bem feitas
- Código no GitHub
- Documentation clara

---

> **🎯 LEMBRE-SE**: Seu objetivo é **FORMAR** e **IMPRESSIONAR** a banca, não criar o próximo unicórnio. Foque no essencial, execute bem, e entregue no prazo! 🚀
