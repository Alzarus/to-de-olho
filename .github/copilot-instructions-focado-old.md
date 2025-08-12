# 🤖 GitHub Copilot - TCC "Tô De Olho" (MVP Focado)

## 🎯 Visão do Projeto (TCC - IFBA)

O **"Tô De Olho"** é uma plataforma de transparência política para **TCC** que democratiza o acesso aos dados da Câmara dos Deputados com foco em **SIMPLICIDADE e EFETIVIDADE**.

### 🚨 **PRIORIDADE ABSOLUTA: MVP que FUNCIONA**
1. **Listar Deputados** com dados reais
2. **Exibir Gastos** com gráficos simples  
3. **Interface Responsiva** e moderna
4. **Performance** adequada

### ❌ **EVITAR Over-Engineering:**
- Não usar microsserviços (MONOLITO é OK)
- Não implementar gamificação complexa
- Foco em funcionalidades que FUNCIONAM 100%

## 🛠️ Stack SIMPLIFICADA (TCC-Friendly)

```
Backend:     Go 1.23 + Gin + GORM (MONOLITO)
Frontend:    Next.js 15 + TypeScript + Tailwind CSS
Database:    PostgreSQL (simples, 3 tabelas principais)
Deploy:      Vercel (frontend) + Railway (backend)
Testing:     Básico (não precisa 80% coverage para TCC)
```

## 🏗️ Estrutura SIMPLES

### Backend Monolito
```go
// Estrutura SIMPLES para TCC
/backend/
├── main.go                     # Entry point único
├── handlers/                   # HTTP handlers  
│   ├── deputados.go           # CRUD deputados
│   └── despesas.go            # Gastos
├── models/                    # Structs GORM
│   ├── deputado.go           
│   └── despesa.go            
├── services/                  # Business logic
├── database/                  # DB connection
└── utils/                     # Helpers
```

### Frontend Direto
```
/frontend/
├── pages/                     # Next.js pages
│   ├── index.tsx             # Home
│   ├── deputados/            # Lista + detalhes
│   └── api/                  # API routes (se precisar)
├── components/               # Componentes reutilizáveis
├── hooks/                    # Custom hooks
└── utils/                    # Helpers
```

## 📋 MVP Requirements (FOCO TOTAL)

### ✅ **Core Features OBRIGATÓRIAS**
- [ ] **GET /deputados** - Lista paginada com filtros
- [ ] **GET /deputados/:id** - Perfil completo
- [ ] **GET /deputados/:id/despesas** - Gastos do deputado
- [ ] **Frontend responsivo** funcionando
- [ ] **Dados reais** da API da Câmara

### 🚀 **Nice-to-Have (SE DER TEMPO)**
- [ ] Comparar 2-3 deputados
- [ ] Dashboard com estatísticas
- [ ] Export de dados (PDF/Excel)
- [ ] Sistema de busca avançada
## � Padrões de Código SIMPLES (TCC)

### Naming Convention Go
```go
// ✅ Simples e direto
type Deputado struct {
    ID       uint   `gorm:"primaryKey"`
    Nome     string `json:"nome"`
    Partido  string `json:"partido"`
    UF       string `json:"uf"`
    FotoURL  string `json:"foto_url"`
}

// ✅ Handlers simples
func GetDeputados(c *gin.Context) {
    var deputados []Deputado
    db.Find(&deputados)
    c.JSON(200, deputados)
}

// ✅ Errors simples
var (
    ErrDeputadoNotFound = errors.New("deputado não encontrado")
    ErrInvalidInput     = errors.New("dados inválidos")
)
```

### Frontend Patterns
```tsx
// ✅ Componentes funcionais simples
interface DeputadoCardProps {
  deputado: Deputado;
}

export function DeputadoCard({ deputado }: DeputadoCardProps) {
  return (
    <div className="p-4 border rounded-lg">
      <h3 className="font-bold">{deputado.nome}</h3>
      <p className="text-gray-600">{deputado.partido} - {deputado.uf}</p>
    </div>
  );
}

// ✅ Hooks simples para API
export function useDeputados() {
  const [deputados, setDeputados] = useState<Deputado[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    fetch('/api/deputados')
      .then(res => res.json())
      .then(setDeputados)
      .finally(() => setLoading(false));
  }, []);
  
  return { deputados, loading };
}
```

## 📊 Dados da Câmara (API Oficial)

### API Base: `https://dadosabertos.camara.leg.br/api/v2/`

#### Endpoints Essenciais para TCC
```bash
# Deputados ativos (513 deputados)
GET /deputados?idLegislatura=57&ordem=ASC&ordenarPor=nome

# Dados específicos do deputado  
GET /deputados/{id}

# Despesas do deputado (últimos 6 meses)
GET /deputados/{id}/despesas?ano=2025&mes=8&ordem=DESC&ordenarPor=dataDocumento

# Partidos ativos
GET /partidos?idLegislatura=57&ordem=ASC&ordenarPor=sigla
```

#### Rate Limiting
- **Limite**: 100 requisições/minuto
- **Estratégia**: Cache simples + requisições em lote

## 🎯 Foco do TCC: FUNCIONALIDADE > ARQUITETURA

### ✅ **O que a Banca VAI valorizar:**
1. **Funciona 100%** - Sem bugs na apresentação
2. **Interface bonita** - Tailwind CSS bem usado
3. **Dados reais** - API da Câmara funcionando
4. **Código limpo** - Fácil de entender
5. **Documentação** - README claro

### ❌ **O que NÃO vai somar pontos:**
- Over-engineering sem necessidade
- Features incompletas ou bugadas
- Complexidade desnecessária
- Muitas tecnologias sem uso real

---

## 📚 Documentação Adicional

Para detalhes específicos do MVP:
- **API Reference**: `.github/docs/api-reference.md`  
- **Business Rules**: `.github/docs/business-rules.md`
- **Plano Realista**: `TCC-PLANO-REALISTA.md`

---

> **🎯 Objetivo TCC**: Criar plataforma funcional de transparência política que demonstre competência técnica e impacto social, priorizando qualidade sobre quantidade.
