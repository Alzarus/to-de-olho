# 🚀 Script de Start IMEDIATO - TCC "Tô De Olho"

## 🎯 **AÇÃO IMEDIATA: Próximas 2 Horas**

### **1. Setup Backend Mínimo (30 min)**

```powershell
# 1. Criar estrutura backend
mkdir backend
cd backend

# 2. Inicializar Go module
go mod init to-de-olho-backend

# 3. Instalar dependências básicas
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv

# 4. Voltar para raiz
cd ..
```

### **2. Setup Frontend Mínimo (30 min)**

```powershell
# 1. Criar Next.js app
npx create-next-app@latest frontend --typescript --tailwind --app --src-dir --import-alias "@/*"

# 2. Entrar no frontend
cd frontend

# 3. Instalar dependências essenciais
npm install lucide-react          # Ícones bonitos
npm install @headlessui/react     # Componentes prontos
npm install recharts              # Gráficos simples
npm install axios                 # Cliente HTTP

# 4. Voltar para raiz
cd ..
```

### **3. Testar Conexão API Câmara (30 min)**

```powershell
# Criar script de teste
mkdir scripts
```

### **4. Primeira Demo Funcionando (30 min)**

---

## 📁 **Estrutura Final Esperada:**

```
to-de-olho/
├── backend/
│   ├── main.go               # ✅ Server principal
│   ├── handlers/             # ✅ APIs REST
│   ├── models/               # ✅ Structs dos dados
│   ├── services/             # ✅ Business logic
│   └── go.mod                # ✅ Dependências
├── frontend/
│   ├── src/
│   │   ├── app/              # ✅ Pages Next.js 15
│   │   ├── components/       # ✅ Componentes React
│   │   └── lib/              # ✅ Utils e configs
│   ├── package.json          # ✅ Dependências
│   └── tailwind.config.js    # ✅ Styles
├── scripts/
│   ├── test-api.js           # ✅ Testar API Câmara
│   └── start-dev.ps1         # ✅ Comando único para subir tudo
└── docker-compose.yml        # ✅ PostgreSQL local
```

---

## 🔥 **CÓDIGO INICIAL - COPY & PASTE**

### **backend/main.go**
```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

type Deputado struct {
    ID      int    `json:"id"`
    Nome    string `json:"nome"`
    Partido string `json:"siglaPartido"`
    UF      string `json:"siglaUf"`
    Foto    string `json:"urlFoto"`
}

func main() {
    // Carregar .env
    godotenv.Load()
    
    // Criar router
    r := gin.Default()
    
    // CORS para desenvolvimento
    r.Use(cors.Default())
    
    // Rota de teste
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    
    // Rota principal - listar deputados
    r.GET("/api/deputados", func(c *gin.Context) {
        // TODO: Buscar da API da Câmara e salvar no banco
        // Por enquanto, dados mock
        deputados := []Deputado{
            {ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP", Foto: ""},
            {ID: 2, Nome: "Maria Santos", Partido: "PSDB", UF: "RJ", Foto: ""},
        }
        
        c.JSON(200, gin.H{
            "data": deputados,
            "total": len(deputados),
        })
    })
    
    // Buscar deputado específico
    r.GET("/api/deputados/:id", func(c *gin.Context) {
        id := c.Param("id")
        
        // TODO: Buscar no banco
        deputado := Deputado{
            ID: 1, 
            Nome: "João Silva", 
            Partido: "PT", 
            UF: "SP",
        }
        
        c.JSON(200, gin.H{"data": deputado})
    })
    
    // Iniciar servidor
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("🚀 Servidor rodando na porta %s", port)
    r.Run(":" + port)
}
```

### **frontend/src/app/page.tsx**
```tsx
'use client';

import { useState, useEffect } from 'react';
import { Search, Users, DollarSign, TrendingUp } from 'lucide-react';

interface Deputado {
  id: number;
  nome: string;
  siglaPartido: string;
  siglaUf: string;
  urlFoto?: string;
}

export default function Home() {
  const [deputados, setDeputados] = useState<Deputado[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Buscar deputados do backend
    fetch('http://localhost:8080/api/deputados')
      .then(res => res.json())
      .then(data => {
        setDeputados(data.data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Erro ao carregar deputados:', err);
        setLoading(false);
      });
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-blue-600 text-white p-6">
        <div className="max-w-6xl mx-auto">
          <h1 className="text-3xl font-bold">🏛️ Tô De Olho</h1>
          <p className="text-blue-100 mt-2">
            Transparência política da Câmara dos Deputados
          </p>
        </div>
      </header>

      {/* Stats Cards */}
      <div className="max-w-6xl mx-auto p-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center">
              <Users className="h-8 w-8 text-blue-600" />
              <div className="ml-4">
                <p className="text-sm text-gray-600">Deputados</p>
                <p className="text-2xl font-bold">513</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center">
              <DollarSign className="h-8 w-8 text-green-600" />
              <div className="ml-4">
                <p className="text-sm text-gray-600">Gastos 2025</p>
                <p className="text-2xl font-bold">R$ 1.2B</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center">
              <TrendingUp className="h-8 w-8 text-purple-600" />
              <div className="ml-4">
                <p className="text-sm text-gray-600">Transparência</p>
                <p className="text-2xl font-bold">100%</p>
              </div>
            </div>
          </div>
        </div>

        {/* Search Bar */}
        <div className="bg-white p-6 rounded-lg shadow mb-8">
          <div className="flex items-center">
            <Search className="h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder="Buscar deputado por nome, partido ou estado..."
              className="ml-3 flex-1 border-0 focus:ring-0 text-lg"
            />
            <button className="bg-blue-600 text-white px-6 py-2 rounded-lg ml-4">
              Buscar
            </button>
          </div>
        </div>

        {/* Deputados List */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-6 border-b">
            <h2 className="text-xl font-bold">Deputados Federais</h2>
          </div>
          
          <div className="p-6">
            {loading ? (
              <div className="text-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                <p className="mt-2 text-gray-600">Carregando deputados...</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {deputados.map((deputado) => (
                  <div key={deputado.id} className="border rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div className="flex items-center space-x-3">
                      <div className="w-12 h-12 bg-gray-200 rounded-full flex items-center justify-center">
                        <Users className="h-6 w-6 text-gray-400" />
                      </div>
                      <div>
                        <h3 className="font-semibold">{deputado.nome}</h3>
                        <p className="text-sm text-gray-600">
                          {deputado.siglaPartido} - {deputado.siglaUf}
                        </p>
                      </div>
                    </div>
                    <button className="mt-3 w-full bg-blue-50 text-blue-600 py-2 rounded-lg text-sm font-medium hover:bg-blue-100">
                      Ver Detalhes
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
```

### **scripts/test-api.js**
```javascript
// Script para testar API da Câmara
const https = require('https');

console.log('🧪 Testando API da Câmara dos Deputados...\n');

// Testar endpoint de deputados
const url = 'https://dadosabertos.camara.leg.br/api/v2/deputados?idLegislatura=57&ordem=ASC&ordenarPor=nome&itens=5';

https.get(url, (res) => {
    let data = '';
    
    res.on('data', (chunk) => {
        data += chunk;
    });
    
    res.on('end', () => {
        try {
            const json = JSON.parse(data);
            console.log('✅ API funcionando!');
            console.log(`📊 Encontrados ${json.dados.length} deputados`);
            console.log('\n📋 Primeiros deputados:');
            
            json.dados.slice(0, 3).forEach(dep => {
                console.log(`- ${dep.nome} (${dep.siglaPartido}/${dep.siglaUf})`);
            });
            
            console.log('\n🚀 Pronto para integrar!');
        } catch (err) {
            console.error('❌ Erro ao parsear JSON:', err);
        }
    });
}).on('error', (err) => {
    console.error('❌ Erro na requisição:', err);
});
```

### **scripts/start-dev.ps1**
```powershell
#!/usr/bin/env pwsh

Write-Host "🚀 Iniciando ambiente de desenvolvimento..." -ForegroundColor Green

# Iniciar backend
Write-Host "📡 Iniciando backend..." -ForegroundColor Blue
Start-Process -FilePath "powershell" -ArgumentList "-Command", "cd backend; go run main.go" -WindowStyle Normal

# Aguardar 3 segundos
Start-Sleep 3

# Iniciar frontend  
Write-Host "🎨 Iniciando frontend..." -ForegroundColor Blue
Start-Process -FilePath "powershell" -ArgumentList "-Command", "cd frontend; npm run dev" -WindowStyle Normal

Write-Host "✅ Ambiente iniciado!" -ForegroundColor Green
Write-Host "📱 Frontend: http://localhost:3000" -ForegroundColor Yellow
Write-Host "📡 Backend: http://localhost:8080" -ForegroundColor Yellow
Write-Host "🧪 Teste: http://localhost:8080/ping" -ForegroundColor Yellow
```

---

## ⚡ **EXECUTAR AGORA (15 minutos):**

```powershell
# 1. Setup backend
mkdir backend
cd backend
go mod init to-de-olho-backend
go get github.com/gin-gonic/gin github.com/gin-contrib/cors github.com/joho/godotenv
# [Criar main.go com código acima]
cd ..

# 2. Setup frontend
npx create-next-app@latest frontend --typescript --tailwind --app --src-dir --import-alias "@/*"
cd frontend
npm install lucide-react @headlessui/react
# [Substituir src/app/page.tsx com código acima]
cd ..

# 3. Testar API da Câmara
mkdir scripts
# [Criar test-api.js com código acima]
node scripts/test-api.js

# 4. Rodar aplicação
cd backend
go run main.go &
cd ../frontend  
npm run dev
```

## 🎯 **Resultado em 15 min:**
- ✅ Backend Go funcionando (localhost:8080)
- ✅ Frontend Next.js funcionando (localhost:3000)
- ✅ Interface bonita com Tailwind
- ✅ Consumindo dados (mock por enquanto)
- ✅ Base sólida para continuar

---

> **💡 PRÓXIMO PASSO**: Integrar API real da Câmara no backend e substituir dados mock!
