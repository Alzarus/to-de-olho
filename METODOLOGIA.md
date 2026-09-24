# Metodologia do ranking — Tô De Olho

Documento versionado da metodologia do ranking de senadores. O mesmo conteúdo
está publicado em [todeolho.org/metodologia](https://todeolho.org/metodologia).
Toda mudança de regra entra aqui com data e motivo, na seção
[Histórico de versões](#histórico-de-versões).

**Versão vigente: v2.2**

---

## Nota final

Cada critério é normalizado de 0 a 100 antes da ponderação:

```
Nota = Produtividade × 0,35 + Presença × 0,25 + Economia da cota × 0,20 + Comissões × 0,20
```

**Período (recorte):** a partir da posse da legislatura em curso. Na 57ª
legislatura, **01/02/2023**. O recorte acompanha a legislatura, com uma
transição: nos 6 primeiros meses de uma legislatura nova, todos teriam menos
que o tempo mínimo em exercício (abaixo). Nesse intervalo o ranking do mandato
mostra a **legislatura anterior, encerrada**, com os senadores que ocupavam a
cadeira no último dia dela. Na 58ª: de 01/02/2027 a 31/07/2027 o ranking
mostra a 57ª inteira (01/02/2023 a 31/01/2027); a partir de 01/08/2027, a 58ª.

**Todos os critérios usam o mesmo período.** No ranking do mandato, ele vai do
início do recorte até hoje. No ranking de um ano, vai do início do ano (ou do
recorte, o que vier depois) até o fim do ano (ou hoje). Assim, janeiro de 2023
pertence à legislatura anterior e não entra em nenhum critério.

**Tempo mínimo em exercício: 6 meses no período.** Quem ocupou a cadeira por
menos tempo (suplente recém-empossado, senador que voltou de ministério) fica
fora da ordenação, como "dados insuficientes". O tempo em exercício vem dos
períodos oficiais de exercício de cada senador (`/senador/{codigo}/mandatos`).

## 1. Produtividade legislativa (35%)

Avança conforme a proposição anda no processo legislativo:

| estágio alcançado | pontos |
|---|---|
| Apresentada | 1 |
| Em comissão | 2 |
| Aprovada em comissão | 4 |
| Aprovada no Plenário | 8 |
| Transformada em lei | 16 |

Multiplicador por tipo (v2.2). Critério: força jurídica do instrumento e
dificuldade de aprová-lo; vale zero o que não é iniciativa legislativa do
próprio senador.

| peso | siglas | por quê |
|---|---|---|
| ×3 | PEC | Muda a Constituição: 3/5 em dois turnos |
| ×2 | PLP | Lei complementar: maioria absoluta |
| ×1 | PL, PLS, PDL, PDS, PRS, PRN | Normas com efeito próprio. PLS e PDS são as siglas antigas de PL e PDL; o PDL susta atos do Executivo; o PRS trata de competências exclusivas do Senado |
| ×0,5 | RQS, MOC, PFS | Requerimento ao Plenário, moção e proposta de fiscalização e controle |
| ×0,1 | REQ, INS e requerimentos de comissão (RDH, RQN, RAS, RCE, RQJ, RMA, RCT, RQE, RQI, RDR, RRA, RRE, RFF, RTG, RQR, R.S, R.C) | Pedidos de audiência, informação ou homenagem; indicação é sugestão sem efeito vinculante |
| ×0 | ECD, SCD, OFS, MSG, OFN, PET, DEN, CON, SIN, DIV, ATS, PCE, PRM | Não são autoria legislativa do senador. ECD e SCD são emenda e substitutivo da Câmara à mesma matéria (contariam em dobro) |

Sigla que não está na tabela não pontua e é registrada no log da carga até
ter peso decidido.

**Só o primeiro autor pontua, e só como senador.** Coautorias aparecem na ficha
do senador como contagem e não somam pontos. É a regra do *Legislative Effectiveness Score*
(Volden e Wiseman, 2014), base do índice: conta o *sponsor*, não o
*cosponsor*. Uma PEC exige ao menos 27 assinaturas. Se cada assinatura
pontuasse, a mesma PEC renderia os mesmos pontos a 27 senadores.

- A posição de cada autor vem da API do Senado: primeiro do texto de autoria
  da listagem de matérias e, quando o texto não basta, do campo
  `autoriaIniciativa[].ordem` do detalhe do processo. Numa amostra de 2.344
  matérias, as duas fontes concordaram em 100% das 55.168 posições
  comparadas.
- Conta a autoria como senador, inclusive na função de líder ou de presidente
  do Senado. Matérias apresentadas quando o parlamentar era **deputado** não
  pontuam: não são produção do mandato de senador.
- **Vetos não pontuam.** O veto é ato do Presidente da República sobre uma
  matéria já aprovada. A API o lista na autoria do parlamentar que propôs a
  matéria original, com estágio de "transformada em lei"; contá-lo pontuaria
  a mesma lei duas vezes.
- Matérias de autoria institucional (Comissão Diretora, comissão, partido,
  liderança) não têm primeiro autor individual e não pontuam para ninguém.
- Só entram no mandato as matérias apresentadas a partir do início do recorte.

A pontuação do senador é normalizada em escala logarítmica em relação à maior
pontuação da Casa: `nota = ln(1 + pontos) / ln(1 + maior) × 100`.

## 2. Presença em votações (25%)

Base: todas as votações nominais do Plenário no período, com o registro de
cada senador que ocupava a cadeira naquele dia. Quem assumiu no meio do
mandato é avaliado só pelas votações do período em que esteve em exercício.

Cada registro da API cai numa de quatro classes:

| classe | códigos da API | efeito |
|---|---|---|
| presente | Sim, Não, Abstenção, Votou (votação secreta), P-NRV (presente, não registrou voto), Presidente (art. 51 RISF) | conta como presença |
| ausência justificada | LS (licença saúde), LAP, MIS (missão oficial), LC, LG, REP, LAN | sai da conta |
| ausência | AP ("atividade parlamentar"), LP (licença particular), NCom (não compareceu) | conta como falta |
| não conta | NA (dispositivo não citado), Obstrução, P-OD | sai da conta |

```
Presença (usada no ranking) = presentes / (registros − NA − ausências justificadas) × 100
Presença bruta              = presentes / (registros − NA) × 100
```

- **Licença particular (LP) conta como falta.** O TCC só aceita como
  justificada a licença médica e a missão oficial; a licença particular é
  escolha do próprio senador, como a AP.
- **Obstrução fica fora da conta.** Pelo TCC, não conta como presença; também
  não é falta, por ser uma estratégia regimental. Nenhuma ocorrência no
  mandato até a v2.1.
- **AP ("atividade parlamentar") conta como falta.** É uma justificativa
  declarada pelo próprio senador, sem verificação independente. É também a
  ausência mais frequente no mandato: 2.243 registros, contra 153 de "não
  compareceu".
- A ficha de cada senador mostra as duas taxas, as ausências por AP, as
  licenças e missões e os "não compareceu".
- Código de voto que não esteja na tabela fica fora da conta até ser
  classificado. Ele gera um alerta no log e não vira falta nem presença.

## 3. Economia da cota parlamentar — CEAPS (20%)

```
Economia = (1 − gasto do senador / teto da CEAPS no período) × 100
```

O teto do período é o **teto mensal da UF × meses em exercício no período**
(valores de março de 2025: de R$ 36.582 em DF, GO e TO a R$ 52.798 no AM).
Gasto igual ou acima do teto dá nota zero. As despesas entram pelo mês de
competência informado pelo Senado.

O teto é proporcional ao tempo de cada senador. Um teto único desde fevereiro
de 2023 para todos daria economia perto de 100 a quem acabou de tomar posse, e
o critério mediria a data da posse, não a economia.

## 4. Participação em comissões (20%)

Contam os colegiados legislativos em que o senador teve participação no
período: comissões permanentes e temporárias, subcomissões, CPIs, CPMIs e
comissões mistas.

- **Cada colegiado conta uma vez**, pelo papel mais alto no período: titular
  (ou membro nato) 2 pontos, suplente 1 ponto. Recondução não pontua de novo.
- **Não contam:** frentes parlamentares, grupos parlamentares (de amizade ou
  de relacionamento) e conselhos de comendas, diplomas e prêmios. A ficha os
  mostra como "fora da conta".
- A fórmula é a mesma no ranking do mandato e no anual.
- A API do Senado não informa quem preside o colegiado, então a presidência não
  pontua à parte.

Normalizado pela maior pontuação entre os senadores ordenados
(`pontos / maior × 100`).

## Dados insuficientes

O senador **fica fora da ordenação** e aparece à parte, com o motivo, quando:

- esteve **menos de 6 meses em exercício** no período; ou
- não tem nenhum registro de votação que conte no período (sem presença a
  medir).

O último lugar seria uma afirmação que o dado não sustenta. A escala de cada
critério (a "maior pontuação da Casa") é calculada só com os senadores
ordenados. Produtividade zero continua sendo nota zero: é um dado real
(nenhuma matéria de autoria principal).

## Fontes

- Senado Federal, dados abertos legislativos: `legis.senado.leg.br/dadosabertos`
  (votações nominais por período, matérias por autor, detalhe do processo)
- Senado Federal, dados abertos administrativos: CEAPS
- Portal da Transparência: emendas parlamentares
- Nome popular e descrição das matérias (só exibição, fora do cálculo): o
  apelido oficial do Senado (`apelido` de `/processo/{id}`); sem ele, uma
  curadoria curta do projeto, marcada como "nome popular" e com o link da página
  oficial que usa o nome (Agência Senado, senado.leg.br, camara.leg.br ou
  gov.br). A descrição é a `explicacaoEmenta` do Senado ou, na falta dela, a
  ementa; os temas são as `classificacoes` do processo. Nada é gerado por IA.

---

## Histórico de versões

### v2.2 — 23/09/2026: peso para todas as siglas de proposição

Até a v2.1 só PEC, PLP, RQS/MOC e REQ tinham peso definido. As outras ~40
siglas valiam ×1, como um projeto de lei, sem que isso estivesse decidido ou
publicado. Uma indicação (INS), que é só uma sugestão a outro Poder, valia o
mesmo que um PL; emendas da Câmara (ECD) pontuavam de novo a mesma matéria.
A v2.2 dá peso a todas as siglas (tabela na seção 1). PEC ×3 e PLP ×2, os
pesos que o TCC define, não mudam.

Efeito medido antes da troca (ranking do mandato, 78 senadores): 26 mudam de
posição, 0,5 posição em média; só dois mudam 4 ou mais (Mara Gabrilli 30º →
37º e Jaime Bagattoli 35º → 39º, ambos com muitas indicações). No top 10, só
o 10º e o 11º trocam de lugar.

### v2.1 — 23/09/2026: presença mais próxima do TCC e virada de legislatura

| problema na v2 | correção na v2.1 |
|---|---|
| Licença particular (LP) saía da conta, mas o TCC só justifica licença médica e missão oficial | LP conta como falta. São 120 registros no mandato, de 12 senadores. O maior efeito é de 82,4 para 74,3 na presença (41 LP) |
| Obstrução contava como presença; o TCC diz que não conta | Fica fora da conta (nem presença nem falta). Nenhuma ocorrência no mandato |
| Em 01/02/2027, com a 58ª legislatura, todos ficariam abaixo de 6 meses em exercício e o ranking sairia vazio até agosto | Nos 6 primeiros meses de uma legislatura, o ranking do mandato mostra a anterior, encerrada |

### v2 — 23/09/2026: correção do cálculo (auditoria interna)

Motivo: uma auditoria do próprio projeto encontrou divergências entre o que o
TCC descreve e o que o código calculava. A v2 corrige o código para a
metodologia publicada. As fórmulas do TCC não mudam.

| # | problema na v1 | correção na v2 |
|---|---|---|
| 1 | Cada matéria ficava com um único autor (o primeiro em ordem alfabética); as coautorias eram descartadas. Correlação de −0,33 entre ordem alfabética e produtividade | Uma linha por senador e matéria |
| 9 | Com as coautorias gravadas, cada assinatura de PEC pontuaria cheio | Só o primeiro autor pontua |
| 3 | Votações de uma mesma sessão colapsavam em uma. Das 423 votações do mandato, só 155 contavam | Cada votação nominal conta separadamente |
| 2 | A presença só enxergava Sim, Não, Abstenção e "não compareceu". "Atividade parlamentar" (AP), votação secreta e presidência sumiam da conta. Resultado: 58 de 81 senadores com 100% | Dicionário completo de códigos; AP conta como falta; licenças e missões saem da conta, como o TCC descreve |
| 4 | Falha ao buscar os votos de um senador virava presença 0 (caso Marcelo Castro: 81º lugar) | Nova rotina de carga, com nova tentativa quando a API falha; sem dado, o senador fica fora da ordenação |
| 6 | Frentes parlamentares, grupos de amizade e conselhos de honrarias contavam como comissão; cada recondução pontuava de novo; o bônus de "ativa" contava em dobro | Só colegiados legislativos, uma vez cada, titular 2 e suplente 1 |
| 7 | No ranking anual, a fórmula de comissões era outra (todas as participações ganhavam o bônus) | Mesma fórmula no ano e no mandato |
| 8 | Teto da cota igual para todos desde fev/2023: quem tomou posse em 2026 tinha economia perto de 100 | Teto proporcional aos meses em exercício; mínimo de 6 meses para entrar na ordenação |
| — | Vetos e matérias de quando o senador era deputado pontuavam (41 vetos = 586 pontos no mandato) | Não pontuam |
| — | Critérios com períodos diferentes: a cota incluía janeiro de 2023 (R$ 1,24 milhão, legislatura anterior), as comissões contavam participações encerradas naquele mês e o ranking anual de proposições usava o ano do número da matéria | Mesmo período para todas as fontes |
| — | Dados de origem incompletos: a cota perdia lançamentos distintos com mesmo fornecedor, dia e valor (R$ 3,75 milhões dos 81 senadores atuais); as comissões juntavam períodos diferentes numa linha (11% com fim antes do início) | Carga refeita: cada lançamento e cada período como na fonte |

Efeito na presença (mandato, 423 votações): mediana de 94,3; antes, 58 de 81
senadores tinham 100%.

O texto anterior do site também divergia do código, e foi alinhado:

- presença: dizia "justificativa de ausência não anula a falta", o oposto do
  TCC e do cálculo;
- produtividade: dizia `pontos / maior × 100`, mas o cálculo usa escala
  logarítmica;
- comissões: dizia "presidente 5, titular 3, suplente 1", mas o cálculo dá
  titular 2, suplente 1 e +1 por participação ativa.

### v1 — TCC (2026)

Metodologia descrita no Trabalho de Conclusão de Curso (IFBA): quatro
critérios, pesos 35/25/20/20, presença descontando licenças e missões
oficiais, produtividade por estágio de tramitação com multiplicador por tipo.
