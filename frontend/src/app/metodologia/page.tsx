import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const criterios = [
  {
    nome: "Produtividade Legislativa",
    peso: "35%",
    descricao:
      "Avalia a quantidade e qualidade das proposições apresentadas pelo senador, considerando o tipo de matéria e seu estágio de tramitação.",
    formula: "Score = (Proposições * PesoTipo * PesoEstágio) / MaxPontos",
    detalhes: [
      "PECs (Propostas de Emenda Constitucional): peso x3",
      "PLPs (Projetos de Lei Complementar): peso x2",
      "PLs (Projetos de Lei): peso x1",
      "Moções (RQS/MOC): peso x0.5",
      "Requerimentos (REQ): peso x0.1",
      "Bônus por estágio: Apresentado +1, Em Comissão +2, Aprovado Comissão +4, Aprovado Plenário +8, Transformado em Lei +16",
    ],
  },
  {
    nome: "Presença em Votações",
    peso: "25%",
    descricao:
      "Mede a participação do senador nas sessões deliberativas do Senado Federal, considerando votos registrados em plenário.",
    formula: "Score = (VotosRegistrados / TotalVotações) * 100",
    detalhes: [
      "Contabiliza todos os votos: Sim, Não, Abstenção",
      "Ausências são contabilizadas negativamente",
      "Período de análise: mandato atual",
      "Justificativas de ausência são desconsideradas",
    ],
  },
  {
    nome: "Economia CEAPS",
    peso: "20%",
    descricao:
      "Avalia o uso responsável da Cota para Exercício da Atividade Parlamentar, comparando o gasto real com o teto disponível por UF.",
    formula: "Score = (1 - (GastoReal / TetoCEAPS)) * 100",
    detalhes: [
      "Teto = Verba Indenizatória (R$ 15.000) + Verba Transporte Aéreo (varia por UF)",
      "AM: R$ 52.798/mês (maior teto) | DF/GO/TO: R$ 36.582/mês (menor teto)",
      "Média nacional: R$ 46.402/mês (referência março 2025, reajuste 12%)",
      "Cada UF tem seu teto específico baseado no custo aéreo até Brasília",
      "Quanto menor o gasto, maior o score",
    ],
  },
  {
    nome: "Participação em Comissões",
    peso: "20%",
    descricao:
      "Mede o engajamento do senador nas comissões temáticas, valorizando posições de liderança e titularidade.",
    formula:
      "Score = (Titularidades * 3 + Suplências * 1 + Presidências * 5) / MaxPontos",
    detalhes: [
      "Presidência de comissão: 5 pontos",
      "Titularidade: 3 pontos",
      "Suplência: 1 ponto",
      "Relatorias não estão contabilizadas nesta versão (planejado para versão futura)",
    ],
  },
];

export default function MetodologiaPage() {
  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-12">
        <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          Metodologia do Ranking
        </h1>
        <p className="mt-4 max-w-3xl text-lg text-muted-foreground">
          Entenda como calculamos o score de cada senador. Nossa metodologia é
          baseada em dados públicos e critérios objetivos, inspirada em
          literatura acadêmica sobre avaliação legislativa.
        </p>
      </div>

      {/* Formula Overview */}
      <Card className="mb-12">
        <CardHeader>
          <CardTitle>Fórmula Geral</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg bg-muted p-6 font-mono text-sm">
            <p className="text-foreground">
              <span className="text-primary font-bold">Score Total</span> =
            </p>
            <p className="mt-2 pl-4 text-muted-foreground">
              (Produtividade * <span className="text-primary">0.35</span>) +
            </p>
            <p className="pl-4 text-muted-foreground">
              (Presença * <span className="text-primary">0.25</span>) +
            </p>
            <p className="pl-4 text-muted-foreground">
              (Economia * <span className="text-primary">0.20</span>) +
            </p>
            <p className="pl-4 text-muted-foreground">
              (Comissões * <span className="text-primary">0.20</span>)
            </p>
          </div>
          <p className="mt-4 text-sm text-muted-foreground">
            Cada critério é normalizado para uma escala de 0 a 100 antes da
            aplicação dos pesos.
          </p>
        </CardContent>
      </Card>

      {/* Criteria Details */}
      <div className="space-y-8">
        <h2 className="text-2xl font-bold tracking-tight text-foreground">
          Detalhamento dos Critérios
        </h2>

        {criterios.map((criterio, index) => (
          <Card key={criterio.nome}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-3">
                  <span className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground">
                    {index + 1}
                  </span>
                  {criterio.nome}
                </CardTitle>
                <span className="rounded-full bg-primary/10 px-3 py-1 text-sm font-bold text-primary">
                  {criterio.peso}
                </span>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-muted-foreground">{criterio.descricao}</p>

              <div className="rounded-lg bg-muted p-4">
                <p className="font-mono text-sm text-foreground">
                  {criterio.formula}
                </p>
              </div>

              <div>
                <h4 className="mb-2 text-sm font-semibold text-foreground">
                  Detalhes:
                </h4>
                <ul className="list-inside list-disc space-y-1 text-sm text-muted-foreground">
                  {criterio.detalhes.map((detalhe, i) => (
                    <li key={i}>{detalhe}</li>
                  ))}
                </ul>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Data Sources */}
      <Card className="mt-12">
        <CardHeader>
          <CardTitle>Fontes de Dados</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-3">
            <div>
              <h4 className="font-semibold text-foreground">
                API Legislativa do Senado
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Dados legislativos, votações, comissões e informações dos
                senadores
              </p>
              <a
                href="https://legis.senado.leg.br/dadosabertos"
                target="_blank"
                rel="noopener noreferrer"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                legis.senado.leg.br/dadosabertos
              </a>
            </div>
            <div>
              <h4 className="font-semibold text-foreground">
                API Administrativa do Senado
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Dados de despesas CEAPS e informações administrativas
              </p>
              <a
                href="https://adm.senado.gov.br/adm-dadosabertos/swagger-ui"
                target="_blank"
                rel="noopener noreferrer"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                adm.senado.gov.br/adm-dadosabertos/swagger-ui
              </a>
            </div>
            <div>
              <h4 className="font-semibold text-foreground">
                Portal da Transparência
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Dados de contratos, convênios e despesas do Governo Federal
              </p>
              <a
                href="https://portaldatransparencia.gov.br"
                target="_blank"
                rel="noopener noreferrer"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                portaldatransparencia.gov.br
              </a>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Academic References */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>Fundamentação Teórica</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            A metodologia do ranking é inspirada no{" "}
            <strong>State Legislative Effectiveness Score (SLES)</strong>,
            desenvolvido por Volden e Wiseman (2014), adaptado para o contexto do
            Senado brasileiro. A abordagem combina indicadores quantitativos de
            produtividade legislativa com métricas de engajamento e
            responsabilidade fiscal.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
