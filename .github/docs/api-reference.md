# 📊 API Reference - Tô De Olho

## 🌐 API Base URL
- **Development**: `http://localhost:8080/api/v1`
- **Staging**: `https://staging-api.to-de-olho.com/api/v1`  
- **Production**: `https://api.to-de-olho.com/api/v1`

## 🔐 Autenticação

### JWT Bearer Token
```http
Authorization: Bearer <jwt_token>
```

### Obter Token
```http
POST /auth/login
Content-Type: application/json

{
  "email": "usuario@example.com",
  "password": "senha123"
}
```

## 📋 Deputados API

### Listar Deputados
```http
GET /deputados?uf=SP&partido=PT&page=1&limit=20
```

**Query Parameters:**
- `uf` (string): Filtrar por estado (opcional)
- `partido` (string): Filtrar por sigla do partido (opcional)
- `page` (int): Página (default: 1)
- `limit` (int): Itens por página (default: 20, max: 100)
- `status` (string): ativo, inativo (default: ativo)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "nome": "João Silva",
      "nome_civil": "João da Silva Santos",
      "cpf": "123.456.789-01",
      "sexo": "M",
      "data_nascimento": "1970-05-15",
      "estado": {
        "uf": "SP",
        "nome": "São Paulo"
      },
      "partido": {
        "id": "uuid",
        "sigla": "PT",
        "nome": "Partido dos Trabalhadores"
      },
      "foto_url": "https://...",
      "email": "joao.silva@camara.leg.br",
      "status": "ativo",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "total": 513,
    "page": 1,
    "limit": 20,
    "total_pages": 26
  }
}
```

### Buscar Deputado por ID
```http
GET /deputados/{id}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "nome": "João Silva",
    "nome_civil": "João da Silva Santos",
    "cpf": "123.456.789-01",
    "sexo": "M",
    "data_nascimento": "1970-05-15",
    "escolaridade": "Superior",
    "profissao": "Advogado",
    "estado": {
      "uf": "SP",
      "nome": "São Paulo"
    },
    "partido": {
      "id": "uuid",
      "sigla": "PT",
      "nome": "Partido dos Trabalhadores"
    },
    "mandatos": [
      {
        "legislatura": 57,
        "data_inicio": "2023-02-01",
        "data_fim": "2027-01-31",
        "situacao": "ativo"
      }
    ],
    "contatos": {
      "email": "joao.silva@camara.leg.br",
      "telefone": "(11) 99999-9999",
      "gabinete": "123"
    },
    "redes_sociais": {
      "twitter": "@joaosilva",
      "instagram": "@joaosilva",
      "facebook": "joaosilva"
    },
    "estatisticas": {
      "total_proposicoes": 45,
      "total_votacoes": 234,
      "presenca_plenario": 85.5,
      "presenca_comissoes": 78.2,
      "gastos_ano_atual": 125000.50
    },
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Gastos do Deputado
```http
GET /deputados/{id}/despesas?ano=2024&mes=10&tipo=PASSAGEM_AEREA
```

**Query Parameters:**
- `ano` (int): Ano das despesas (opcional)
- `mes` (int): Mês das despesas (opcional)
- `tipo` (string): Tipo de despesa (opcional)
- `page`, `limit`: Paginação

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "tipo": "PASSAGEM_AEREA",
      "descricao": "Emissão Bilhete Aéreo",
      "valor": 1250.00,
      "data": "2024-10-15",
      "fornecedor": {
        "cnpj": "12.345.678/0001-90",
        "nome": "Companhia Aérea TAM"
      },
      "documento": {
        "numero": "12345",
        "url": "https://..."
      },
      "created_at": "2024-10-16T00:00:00Z"
    }
  ],
  "meta": {
    "total": 156,
    "total_valor": 125000.50,
    "page": 1,
    "limit": 20
  },
  "resumo": {
    "por_tipo": {
      "PASSAGEM_AEREA": 25000.00,
      "HOSPEDAGEM": 15000.00,
      "COMBUSTIVEL": 8000.00
    },
    "media_mensal": 10416.67
  }
}
```

## 🗳️ Proposições API

### Listar Proposições
```http
GET /proposicoes?tipo=PL&ano=2024&autor_id=uuid&tema=educacao
```

**Query Parameters:**
- `tipo` (string): Tipo da proposição (PL, PEC, etc.)
- `ano` (int): Ano de apresentação
- `autor_id` (uuid): ID do deputado autor
- `tema` (string): Tema da proposição
- `status` (string): Status de tramitação

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "numero": 1234,
      "ano": 2024,
      "tipo": "PL",
      "ementa": "Dispõe sobre...",
      "autor": {
        "id": "uuid",
        "nome": "João Silva",
        "partido": "PT",
        "uf": "SP"
      },
      "data_apresentacao": "2024-10-01",
      "status": "tramitando",
      "temas": ["educacao", "ensino_superior"],
      "ultima_acao": {
        "data": "2024-10-15",
        "descricao": "Aprovado na Comissão de Educação"
      },
      "created_at": "2024-10-01T00:00:00Z"
    }
  ],
  "meta": {
    "total": 1234,
    "page": 1,
    "limit": 20
  }
}
```

### Detalhes da Proposição
```http
GET /proposicoes/{id}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "numero": 1234,
    "ano": 2024,
    "tipo": "PL",
    "ementa": "Dispõe sobre...",
    "ementa_detalhada": "Texto completo da ementa...",
    "justificativa": "Texto da justificativa...",
    "autor_principal": {
      "id": "uuid",
      "nome": "João Silva",
      "partido": "PT",
      "uf": "SP"
    },
    "coautores": [
      {
        "id": "uuid",
        "nome": "Maria Santos",
        "partido": "PSDB",
        "uf": "RJ"
      }
    ],
    "data_apresentacao": "2024-10-01",
    "status": "tramitando",
    "temas": ["educacao", "ensino_superior"],
    "tramitacao": [
      {
        "data": "2024-10-01",
        "orgao": "Mesa Diretora",
        "acao": "Recebimento da proposição",
        "texto": "Proposição recebida e distribuída..."
      },
      {
        "data": "2024-10-15",
        "orgao": "Comissão de Educação",
        "acao": "Aprovado",
        "texto": "Aprovado na forma do substitutivo..."
      }
    ],
    "votacoes": [
      {
        "id": "uuid",
        "data": "2024-10-15",
        "orgao": "Comissão de Educação",
        "resultado": "aprovado",
        "placar": {
          "sim": 15,
          "nao": 3,
          "abstencao": 2
        }
      }
    ],
    "arquivos": [
      {
        "tipo": "texto_original",
        "nome": "PL1234-2024.pdf",
        "url": "https://..."
      }
    ],
    "created_at": "2024-10-01T00:00:00Z",
    "updated_at": "2024-10-15T00:00:00Z"
  }
}
```

## 🗳️ Votações API

### Listar Votações
```http
GET /votacoes?data_inicio=2024-10-01&data_fim=2024-10-31&orgao=PLENARIO
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "data": "2024-10-15T14:30:00Z",
      "orgao": "PLENARIO",
      "proposicao": {
        "id": "uuid",
        "numero": 1234,
        "tipo": "PL",
        "ano": 2024,
        "ementa": "Dispõe sobre..."
      },
      "objeto": "Aprovação do Projeto de Lei",
      "resultado": "aprovado",
      "placar": {
        "sim": 267,
        "nao": 145,
        "abstencao": 12,
        "ausente": 89
      },
      "created_at": "2024-10-15T14:30:00Z"
    }
  ]
}
```

### Votos da Votação
```http
GET /votacoes/{id}/votos
```

**Response:**
```json
{
  "data": [
    {
      "deputado": {
        "id": "uuid",
        "nome": "João Silva",
        "partido": "PT",
        "uf": "SP"
      },
      "voto": "sim",
      "data": "2024-10-15T14:30:00Z"
    }
  ],
  "meta": {
    "total_votos": 513,
    "placar": {
      "sim": 267,
      "nao": 145,
      "abstencao": 12,
      "ausente": 89
    }
  }
}
```

## 👥 Usuários API

### Registrar Usuário
```http
POST /usuarios/registro
Content-Type: application/json

{
  "nome": "Maria Silva",
  "email": "maria@example.com",
  "senha": "senha123",
  "cpf": "123.456.789-01",
  "data_nascimento": "1985-06-20",
  "estado_uf": "SP",
  "cidade": "São Paulo",
  "aceita_termos": true
}
```

### Perfil do Usuário
```http
GET /usuarios/perfil
Authorization: Bearer <token>
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "nome": "Maria Silva",
    "email": "maria@example.com",
    "avatar_url": "https://...",
    "estado_uf": "SP",
    "cidade": "São Paulo",
    "role": "eleitor",
    "verificado": true,
    "gamificacao": {
      "pontos": 1250,
      "nivel": "Cidadão Ativo",
      "badges": [
        {
          "id": "fiscal_ativo",
          "nome": "Fiscal Ativo",
          "descricao": "Acompanha gastos regularmente",
          "conquistado_em": "2024-09-15T00:00:00Z"
        }
      ],
      "ranking_nacional": 1523,
      "ranking_estadual": 234
    },
    "configuracoes": {
      "notificacoes_email": true,
      "notificacoes_push": false,
      "deputados_seguidos": ["uuid1", "uuid2"],
      "temas_interesse": ["educacao", "saude", "economia"]
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-10-15T00:00:00Z"
  }
}
```

## 💬 Fórum API

### Tópicos do Fórum
```http
GET /forum/topicos?categoria=geral&page=1&limit=20
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "titulo": "Discussão sobre a PL 1234/2024",
      "conteudo": "Gostaria de debater os impactos...",
      "categoria": "proposicoes",
      "autor": {
        "id": "uuid",
        "nome": "Maria Silva",
        "avatar_url": "https://...",
        "badges": ["eleitor_verificado"]
      },
      "proposicao_relacionada": {
        "id": "uuid",
        "numero": 1234,
        "tipo": "PL",
        "ano": 2024
      },
      "estatisticas": {
        "total_comentarios": 45,
        "total_likes": 23,
        "total_visualizacoes": 234
      },
      "ultimo_comentario": {
        "data": "2024-10-15T14:30:00Z",
        "autor": "João Santos"
      },
      "tags": ["educacao", "ensino_superior"],
      "fixado": false,
      "fechado": false,
      "created_at": "2024-10-10T00:00:00Z",
      "updated_at": "2024-10-15T14:30:00Z"
    }
  ],
  "meta": {
    "total": 156,
    "page": 1,
    "limit": 20
  }
}
```

### Comentários do Tópico
```http
GET /forum/topicos/{id}/comentarios?page=1&limit=20
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "conteudo": "Concordo com a proposta, mas...",
      "autor": {
        "id": "uuid",
        "nome": "João Santos",
        "avatar_url": "https://...",
        "badges": ["fiscal_ativo"]
      },
      "comentario_pai_id": null,
      "nivel_aninhamento": 0,
      "estatisticas": {
        "total_likes": 12,
        "total_respostas": 3
      },
      "editado": false,
      "moderado": false,
      "created_at": "2024-10-15T10:30:00Z",
      "updated_at": "2024-10-15T10:30:00Z"
    },
    {
      "id": "uuid",
      "conteudo": "Mas você considerou que...",
      "autor": {
        "id": "uuid",
        "nome": "Ana Costa",
        "avatar_url": "https://..."
      },
      "comentario_pai_id": "uuid_comentario_anterior",
      "nivel_aninhamento": 1,
      "estatisticas": {
        "total_likes": 5,
        "total_respostas": 0
      },
      "created_at": "2024-10-15T11:15:00Z"
    }
  ]
}
```

## 📊 Analytics API

### Estatísticas Gerais
```http
GET /analytics/dashboard
Authorization: Bearer <admin_token>
```

**Response:**
```json
{
  "data": {
    "usuarios": {
      "total": 15234,
      "ativos_mes": 8456,
      "novos_mes": 1234,
      "crescimento": 8.5
    },
    "deputados": {
      "total": 513,
      "ativos": 498,
      "com_perfil_verificado": 156
    },
    "proposicoes": {
      "total_ano": 2345,
      "tramitando": 1456,
      "aprovadas": 234,
      "rejeitadas": 89
    },
    "forum": {
      "total_topicos": 1234,
      "total_comentarios": 15678,
      "usuarios_ativos": 2345
    },
    "engajamento": {
      "pageviews_mes": 156789,
      "tempo_medio_sessao": "4m 32s",
      "bounce_rate": 45.6
    }
  }
}
```

## 🚨 Códigos de Erro

### HTTP Status Codes
- `200` - OK
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Unprocessable Entity
- `429` - Too Many Requests
- `500` - Internal Server Error

### Error Response Format
```json
{
  "error": {
    "code": "DEPUTADO_NOT_FOUND",
    "message": "Deputado não encontrado",
    "details": "Não foi possível encontrar um deputado com o ID fornecido",
    "timestamp": "2024-10-15T14:30:00Z",
    "trace_id": "uuid"
  }
}
```

### Common Error Codes
- `INVALID_INPUT` - Dados de entrada inválidos
- `UNAUTHORIZED` - Token de autenticação inválido
- `FORBIDDEN` - Acesso negado
- `RESOURCE_NOT_FOUND` - Recurso não encontrado
- `RATE_LIMIT_EXCEEDED` - Limite de requisições excedido
- `INTERNAL_ERROR` - Erro interno do servidor

## 🔒 Rate Limiting

### Limites por Endpoint
- **Público**: 100 requests/hora por IP
- **Autenticado**: 1000 requests/hora por usuário
- **Admin**: 5000 requests/hora

### Headers de Rate Limit
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1698854400
```

---

> **📝 Nota**: Esta documentação é atualizada automaticamente. Para versões específicas da API, consulte `/docs/v1` no ambiente correspondente.
