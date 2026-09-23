# Metodologia do ranking — Tô De Olho

Documento versionado da metodologia do ranking de senadores. O mesmo conteúdo
está publicado em [todeolho.org/metodologia](https://todeolho.org/metodologia).
Toda mudança de regra entra aqui com data e motivo, na seção
[Histórico de versões](#histórico-de-versões).

**Versão vigente: v2**

---

## Nota final

Cada critério é normalizado de 0 a 100 antes da ponderação:

```
Nota = Produtividade × 0,35 + Presença × 0,25 + Economia da cota × 0,20 + Comissões × 0,20
```

**Período (recorte):** a partir da posse da legislatura em curso. Na 57ª
legislatura, **01/02/2023**. O recorte acompanha a legislatura: a partir de
01/02/2027 passa a ser a 58ª.

## 1. Produtividade legislativa (35%)

Avança conforme a proposição anda no processo legislativo:

| estágio alcançado | pontos |
|---|---|
| Apresentada | 1 |
| Em comissão | 2 |
| Aprovada em comissão | 4 |
| Aprovada no Plenário | 8 |
| Transformada em lei | 16 |

Multiplicador por tipo: PEC ×3; PLP ×2; PL ×1; RQS e MOC ×0,5; REQ ×0,1.

**Só o primeiro autor pontua.** Coautorias aparecem na ficha do senador como
contagem e não somam pontos. É a regra do *Legislative Effectiveness Score*
(Volden e Wiseman, 2014), base do índice: conta o *sponsor*, não o
*cosponsor*. Uma PEC exige ao menos 27 assinaturas. Se cada assinatura
pontuasse, a mesma PEC renderia os mesmos pontos a 27 senadores.

- A posição de cada autor vem da API do Senado: primeiro do texto de autoria
  da listagem de matérias e, quando o texto não basta, do campo
  `autoriaIniciativa[].ordem` do detalhe do processo. Numa amostra de 2.344
  matérias, as duas fontes concordaram em 100% das 55.168 posições
  comparadas.
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
| presente | Sim, Não, Abstenção, Votou (votação secreta), P-NRV (presente, não registrou voto), Presidente (art. 51 RISF), P-OD, Obstrução | conta como presença |
| ausência justificada | LS (licença saúde), LP (licença particular), LAP, MIS (missão oficial), LC, LG, REP, LAN | sai da conta |
| ausência | AP ("atividade parlamentar"), NCom (não compareceu) | conta como falta |
| não conta | NA (dispositivo não citado) | sai da conta |

```
Presença (usada no ranking) = presentes / (registros − NA − ausências justificadas) × 100
Presença bruta              = presentes / (registros − NA) × 100
```

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

O teto varia por UF (valores de março de 2025: de R$ 36.582 em DF, GO e TO a
R$ 52.798 no AM). Gasto igual ou acima do teto dá nota zero.

## 4. Participação em comissões (20%)

Titular: 2 pontos; suplente: 1 ponto; participação ativa no período: +1.
Normalizado pela maior pontuação da Casa (`pontos / maior × 100`). A API do
Senado não informa quem preside a comissão, então a presidência não pontua à
parte.

Em revisão (auditoria, item 6): hoje entram todos os colegiados que a API
devolve, inclusive frentes parlamentares e grupos de amizade.

## Dados insuficientes

Um senador sem nenhum registro de votação que conte no período não tem
presença a medir. Ele **fica fora da ordenação** e aparece à parte, como
"dados insuficientes". Ficar em último lugar seria uma afirmação que o dado
não sustenta. Produtividade zero continua sendo nota zero: ela é um dado real
(nenhuma matéria de autoria principal).

## Fontes

- Senado Federal, dados abertos legislativos: `legis.senado.leg.br/dadosabertos`
  (votações nominais por período, matérias por autor, detalhe do processo)
- Senado Federal, dados abertos administrativos: CEAPS
- Portal da Transparência: emendas parlamentares

---

## Histórico de versões

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

Efeito na presença (mandato, 423 votações): mediana de 94,3, com 7 de 81
senadores em 100% (antes, 58 de 81).

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
