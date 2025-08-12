# 📋 Business Rules - Regras de Negócio

## 🏛️ Domínio: Deputados

### Entidade Deputado
```go
type Deputado struct {
    ID              uuid.UUID    `json:"id"`              // UUID interno
    CamaraID        int          `json:"camara_id"`       // ID oficial da Câmara
    Nome            string       `json:"nome"`
    NomeCivil       string       `json:"nome_civil"`
    CPF             CPF          `json:"cpf"`             // Value Object
    Sexo            Sexo         `json:"sexo"`            // M, F
    DataNascimento  time.Time    `json:"data_nascimento"`
    Estado          Estado       `json:"estado"`          // Value Object
    Partido         Partido      `json:"partido"`         // Aggregate
    Status          Status       `json:"status"`          // ativo, inativo, licenciado
    Mandatos        []Mandato    `json:"mandatos"`
    Contatos        Contatos     `json:"contatos"`        // Value Object
    RedesSociais    RedesSociais `json:"redes_sociais"`   // Value Object
    UltimaSync      time.Time    `json:"ultima_sync"`     // Última sincronização com API Câmara
}
```

### Regras de Validação

#### BR-DEP-001: Idade Mínima
- **Regra**: Deputado deve ter no mínimo 21 anos na data da posse
- **Validação**: `data_nascimento + 21 anos <= data_posse`
- **Erro**: `ErrIdadeInsuficiente`

#### BR-DEP-002: CPF Único
- **Regra**: Não pode existir dois deputados com o mesmo CPF
- **Validação**: Constraint unique no banco de dados
- **Erro**: `ErrCPFJaExiste`

#### BR-DEP-003: Nome Obrigatório
- **Regra**: Nome não pode ser vazio e deve ter no mínimo 2 caracteres
- **Validação**: `len(nome) >= 2 && nome != ""`
- **Erro**: `ErrNomeInvalido`

#### BR-DEP-004: Estado Válido
- **Regra**: Estado deve ser uma UF válida do Brasil
- **Validação**: UF deve estar na lista de estados brasileiros
- **Erro**: `ErrEstadoInvalido`

### Regras de Negócio

#### BR-DEP-005: Mudança de Partido
- **Regra**: Deputado pode mudar de partido, mas deve ser registrado o histórico
- **Implementação**: Criar registro em `mudancas_partido` com data e motivo
- **Janela**: Mudança só é válida dentro da janela legal (março-abril)

```go
func (d *Deputado) MudarPartido(novoPartido Partido, motivo string, data time.Time) error {
    // Verificar janela de mudança
    if !d.isJanelaMudancaPartido(data) {
        return ErrForaJanelaMudanca
    }
    
    // Registrar mudança
    mudanca := MudancaPartido{
        DeputadoID:    d.ID,
        PartidoOrigin: d.Partido,
        PartidoDestino: novoPartido,
        Data:          data,
        Motivo:        motivo,
    }
    
    d.HistoricoPartidos = append(d.HistoricoPartidos, mudanca)
    d.Partido = novoPartido
    
    return nil
}
```

#### BR-DEP-006: Status do Mandato
- **Regra**: Status deve refletir a situação atual do deputado
- **Estados**: `ativo`, `licenciado`, `afastado`, `cassado`, `renunciou`
- **Transições válidas**:
  - `ativo` → `licenciado`, `afastado`, `cassado`, `renunciou`
  - `licenciado` → `ativo`, `cassado`, `renunciou`
  - `afastado` → `ativo`, `cassado`

## 💰 Domínio: Despesas

### Entidade Despesa
```go
type Despesa struct {
    ID           uuid.UUID  `json:"id"`
    DeputadoID   uuid.UUID  `json:"deputado_id"`
    Tipo         TipoDespesa `json:"tipo"`
    Descricao    string     `json:"descricao"`
    Valor        decimal.Decimal `json:"valor"`
    Data         time.Time  `json:"data"`
    Fornecedor   Fornecedor `json:"fornecedor"`
    Documento    Documento  `json:"documento"`
    Status       StatusDespesa `json:"status"`
}
```

### Regras de Validação

#### BR-DESP-001: Valor Positivo
- **Regra**: Valor da despesa deve ser positivo e maior que zero
- **Validação**: `valor > 0`
- **Erro**: `ErrValorInvalido`

#### BR-DESP-002: Data Válida
- **Regra**: Data da despesa não pode ser futura
- **Validação**: `data <= time.Now()`
- **Erro**: `ErrDataFutura`

#### BR-DESP-003: Tipo Válido
- **Regra**: Tipo deve estar na lista de tipos permitidos pela Câmara
- **Validação**: Verificar em tabela de referência
- **Erro**: `ErrTipoInvalido`

### Regras de Negócio

#### BR-DESP-004: Limite Mensal por Tipo
- **Regra**: Cada tipo de despesa tem um limite mensal específico
- **Limites** (valores 2024):
  - Passagem Aérea: sem limite específico
  - Hospedagem: R$ 8.000/mês
  - Alimentação: R$ 4.500/mês
  - Combustível: R$ 6.000/mês

```go
func (d *Despesa) ValidarLimiteMensal(deputadoID uuid.UUID, mes, ano int) error {
    limite := d.Tipo.LimiteMensal()
    if limite == 0 {
        return nil // Sem limite
    }
    
    gastoMes := d.calcularGastoMensal(deputadoID, d.Tipo, mes, ano)
    
    if gastoMes+d.Valor > limite {
        return &ErrLimiteExcedido{
            Tipo:       d.Tipo,
            Limite:     limite,
            GastoAtual: gastoMes,
            NovoValor:  d.Valor,
        }
    }
    
    return nil
}
```

#### BR-DESP-005: Detecção de Anomalias
- **Regra**: Sistema deve detectar gastos suspeitos automaticamente
- **Critérios**:
  - Valor 3x superior à média histórica do deputado
  - Múltiplas despesas no mesmo dia/fornecedor
  - Gasto em fim de semana/feriado (exceto viagens)

```go
func (d *Despesa) DetectarAnomalia(historico []Despesa) *Anomalia {
    media := calcularMedia(historico, d.Tipo)
    
    // Valor muito acima da média
    if d.Valor > media*3 {
        return &Anomalia{
            Tipo:       "VALOR_ELEVADO",
            Severidade: "ALTA",
            Descricao:  fmt.Sprintf("Valor %.2f é 3x superior à média %.2f", d.Valor, media),
        }
    }
    
    // Múltiplas despesas mesmo fornecedor/dia
    if d.contarDespesasMesmoDia(historico) > 3 {
        return &Anomalia{
            Tipo:       "MULTIPLAS_DESPESAS",
            Severidade: "MEDIA",
            Descricao:  "Múltiplas despesas no mesmo dia/fornecedor",
        }
    }
    
    return nil
}
```

## 📜 Domínio: Proposições

### Entidade Proposição
```go
type Proposicao struct {
    ID                uuid.UUID     `json:"id"`
    Numero            int           `json:"numero"`
    Ano               int           `json:"ano"`
    Tipo              TipoProposicao `json:"tipo"`
    Ementa            string        `json:"ementa"`
    EmentaDetalhada   string        `json:"ementa_detalhada"`
    AutorPrincipal    Deputado      `json:"autor_principal"`
    Coautores         []Deputado    `json:"coautores"`
    DataApresentacao  time.Time     `json:"data_apresentacao"`
    Status            StatusProposicao `json:"status"`
    Temas             []string      `json:"temas"`
    Tramitacao        []Tramitacao  `json:"tramitacao"`
}
```

### Regras de Validação

#### BR-PROP-001: Número Único por Ano/Tipo
- **Regra**: Não pode existir duas proposições com mesmo número, ano e tipo
- **Validação**: Constraint unique(numero, ano, tipo)
- **Erro**: `ErrProposicaoJaExiste`

#### BR-PROP-002: Ementa Obrigatória
- **Regra**: Ementa deve ter no mínimo 10 caracteres
- **Validação**: `len(ementa) >= 10`
- **Erro**: `ErrEmentaInvalida`

#### BR-PROP-003: Autor Principal Obrigatório
- **Regra**: Toda proposição deve ter um autor principal
- **Validação**: `autor_principal_id != nil`
- **Erro**: `ErrAutorObrigatorio`

### Regras de Negócio

#### BR-PROP-004: Tramitação Sequencial
- **Regra**: Tramitação deve seguir ordem cronológica
- **Implementação**: Cada nova tramitação deve ter data >= última tramitação

```go
func (p *Proposicao) AdicionarTramitacao(nova Tramitacao) error {
    if len(p.Tramitacao) > 0 {
        ultima := p.Tramitacao[len(p.Tramitacao)-1]
        if nova.Data.Before(ultima.Data) {
            return ErrDataTramitacaoInvalida
        }
    }
    
    p.Tramitacao = append(p.Tramitacao, nova)
    p.Status = nova.NovoStatus
    
    return nil
}
```

#### BR-PROP-005: Coautores Limitados
- **Regra**: Máximo de 10 coautores por proposição (exceto PEC)
- **PEC**: Mínimo 171 apoiadores
- **Validação**: Verificar tipo e quantidade

```go
func (p *Proposicao) ValidarCoautores() error {
    switch p.Tipo {
    case TipoPEC:
        if len(p.Coautores) < 171 {
            return ErrPECInsuficienteApoio
        }
    default:
        if len(p.Coautores) > 10 {
            return ErrMuitosCoautores
        }
    }
    
    return nil
}
```

## 🗳️ Domínio: Votações

### Entidade Votação
```go
type Votacao struct {
    ID           uuid.UUID     `json:"id"`
    Data         time.Time     `json:"data"`
    Orgao        Orgao         `json:"orgao"`
    ProposicaoID uuid.UUID     `json:"proposicao_id"`
    Objeto       string        `json:"objeto"`
    Resultado    ResultadoVotacao `json:"resultado"`
    Placar       Placar        `json:"placar"`
    Votos        []Voto        `json:"votos"`
}

type Voto struct {
    DeputadoID uuid.UUID   `json:"deputado_id"`
    TipoVoto   TipoVoto    `json:"tipo_voto"` // sim, nao, abstencao, ausente
    Data       time.Time   `json:"data"`
}
```

### Regras de Validação

#### BR-VOT-001: Quorum Mínimo
- **Regra**: Votações no plenário precisam de quorum mínimo
- **Plenário**: Mínimo 257 deputados (maioria absoluta)
- **Comissões**: Varia por comissão
- **Erro**: `ErrQuorumInsuficiente`

#### BR-VOT-002: Deputado Vota Uma Vez
- **Regra**: Cada deputado pode votar apenas uma vez por votação
- **Validação**: Constraint unique(votacao_id, deputado_id)
- **Erro**: `ErrDeputadoJaVotou`

### Regras de Negócio

#### BR-VOT-003: Cálculo do Resultado
- **Regra**: Resultado baseado no tipo de matéria e quorum
- **Maioria Simples**: mais votos "sim" que "não"
- **Maioria Absoluta**: mínimo 257 votos "sim"
- **2/3**: mínimo 342 votos "sim" (PEC)

```go
func (v *Votacao) CalcularResultado() ResultadoVotacao {
    placar := v.Placar
    quorumMinimo := v.Orgao.QuorumMinimo()
    
    // Verificar quorum
    if placar.TotalPresentes() < quorumMinimo {
        return ResultadoQuorumInsuficiente
    }
    
    // Verificar tipo de maioria necessária
    switch v.TipoMaioria() {
    case MaioriaSimples:
        if placar.Sim > placar.Nao {
            return ResultadoAprovado
        }
        return ResultadoRejeitado
        
    case MaioriaAbsoluta:
        if placar.Sim >= 257 {
            return ResultadoAprovado
        }
        return ResultadoRejeitado
        
    case DoisTercos:
        if placar.Sim >= 342 {
            return ResultadoAprovado
        }
        return ResultadoRejeitado
    }
    
    return ResultadoRejeitado
}
```

## 👥 Domínio: Usuários

### Entidade Usuário
```go
type Usuario struct {
    ID              uuid.UUID      `json:"id"`
    Nome            string         `json:"nome"`
    Email           string         `json:"email"`
    CPF             CPF            `json:"cpf"`
    DataNascimento  time.Time      `json:"data_nascimento"`
    EstadoUF        string         `json:"estado_uf"`
    Cidade          string         `json:"cidade"`
    Role            Role           `json:"role"`
    Verificado      bool           `json:"verificado"`
    Gamificacao     Gamificacao    `json:"gamificacao"`
    Configuracoes   Configuracoes  `json:"configuracoes"`
}
```

### Regras de Validação

#### BR-USR-001: Email Único
- **Regra**: Não pode existir dois usuários com mesmo email
- **Validação**: Constraint unique(email)
- **Erro**: `ErrEmailJaExiste`

#### BR-USR-002: Idade Mínima
- **Regra**: Usuário deve ter no mínimo 16 anos (idade para votar)
- **Validação**: `data_nascimento + 16 anos <= hoje`
- **Erro**: `ErrIdadeInsuficiente`

#### BR-USR-003: CPF Válido para Eleitor
- **Regra**: Usuários com role "eleitor" devem ter CPF válido
- **Validação**: Algoritmo de validação de CPF
- **Erro**: `ErrCPFInvalido`

### Regras de Negócio

#### BR-USR-004: Verificação de Eleitor
- **Regra**: Validação via API do TSE para role "eleitor"
- **Processo**: 
  1. Validar CPF no TSE
  2. Verificar situação eleitoral
  3. Confirmar domicílio eleitoral
- **Status**: `pendente`, `verificado`, `rejeitado`

```go
func (u *Usuario) VerificarEleitor(tseAPI TSEAPIClient) error {
    if u.Role != RoleEleitor {
        return nil
    }
    
    // Consultar TSE
    situacao, err := tseAPI.ConsultarSituacao(u.CPF)
    if err != nil {
        return err
    }
    
    if situacao.Regular && situacao.DomicilioEleitoral == u.EstadoUF {
        u.Verificado = true
        u.DataVerificacao = time.Now()
        return nil
    }
    
    return ErrVerificacaoTSEFalhou
}
```

## 🎮 Domínio: Gamificação

### Sistema de Pontos

#### BR-GAM-001: Pontuação por Atividade
- **Leitura de Proposição**: 5 pontos
- **Comentário no Fórum**: 10 pontos
- **Participação em Plebiscito**: 15 pontos
- **Fiscalização de Gastos**: 20 pontos
- **Denúncia Procedente**: 50 pontos

#### BR-GAM-002: Multiplicadores
- **Primeira ação do dia**: x2
- **Sequência de 7 dias**: x1.5
- **Ação em proposição popular**: x1.3

```go
func (g *Gamificacao) CalcularPontos(acao Acao, usuario Usuario) int {
    pontos := acao.PontosBase()
    
    // Multiplicador primeira ação do dia
    if g.isPrimeiraAcaoHoje(usuario.ID) {
        pontos *= 2
    }
    
    // Multiplicador sequência
    if g.getSequenciaDias(usuario.ID) >= 7 {
        pontos = int(float64(pontos) * 1.5)
    }
    
    return pontos
}
```

### Badges e Conquistas

#### BR-GAM-003: Conquistas Automáticas
- **Fiscal Ativo**: 10 verificações de gastos
- **Cidadão Engajado**: 50 comentários
- **Vigilante**: Denúncia procedente
- **Conhecedor**: Acertar 10 quiz seguidos

---

> **📝 Nota**: Estas regras são implementadas nas camadas de domínio e aplicação, garantindo consistência e integridade dos dados.
