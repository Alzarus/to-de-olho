import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const criterios = [
  {
    nome: "Produtividade Legislativa",
    peso: "35%",
    descricao:
      "Mede a capacidade do senador de criar e aprovar leis. Quanto mais longe o projeto avança, mais pontos ele ganha.",
    formula: "Nota = (Pontos do Senador / Maior Pontuador) x 100",
    detalhes: [
      "Apresentado: 1 ponto",
      "Em discussão nas comissões: 2 pontos",
      "Aprovado em comissão: 4 pontos",
      "Aprovado no Plenário: 8 pontos",
      "Virou lei: 16 pontos",
    ],
    extras: [
      "PECs (mudanças na Constituição): peso x3",
      "PLPs (Lei Complementar): peso x2",
      "PLs (Lei Ordinária): peso x1",
      "Moções (RQS/MOC): peso x0,5",
      "Requerimentos (REQ): peso x0,1",
      "Ajuste logarítmico impede que quantidade supere qualidade",
    ],
  },
  {
    nome: "Presença em Votações",
    peso: "25%",
    descricao:
      "Mede se o senador comparece quando o Senado vota. Quem não aparece, não está cumprindo seu papel.",
    formula:
      "Nota = Votações em que participou / Total de votações disponíveis x 100",
    detalhes: [
      "Qualquer voto conta: Sim, Não ou Abstenção",
      "Justificativa de ausência não anula a falta",
      "Considera todo o mandato atual",
    ],
  },
  {
    nome: "Economia da Cota Parlamentar (CEAPS)",
    peso: "20%",
    descricao:
      "Avalia quanto o senador economiza da sua cota mensal de gastos. Quanto menos gastar, melhor a nota.",
    formula: "Nota = Quanto mais economizar, maior a pontuação",
    detalhes: [
      "Cada estado tem um teto diferente de gastos",
      "Maior teto: Amazonas (~R$ 52 mil/mês)",
      "Menor teto: DF/Goiás (~R$ 36 mil/mês)",
      "Gastar ou ultrapassar o teto zera a nota neste critério",
      "Gastar não é ruim -- mas economia é premiada",
    ],
  },
  {
    nome: "Participação em Comissões",
    peso: "20%",
    descricao:
      "Mede o envolvimento do senador nas comissões, onde projetos são discutidos antes de irem ao Plenário.",
    formula: "Nota = Baseada nos cargos e na quantidade de comissões",
    detalhes: [
      "Presidente de comissão: 5 pontos",
      "Membro Titular (com direito a voto): 3 pontos",
      "Suplente (substituto eventual): 1 ponto",
    ],
  },
];

export default function MetodologiaPage() {
  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-12">
        <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          Como funciona o Ranking
        </h1>
        <p className="mt-4 max-w-3xl text-lg text-muted-foreground">
          Explicamos de forma transparente como a nota de cada senador é
          calculada. Todos os dados são públicos e verificáveis.
        </p>
      </div>

      {/* How It Works Summary */}
      <Card className="mb-12">
        <CardHeader>
          <CardTitle>Como a nota é calculada?</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            A nota final combina quatro critérios com pesos diferentes, todos
            em escala de 0 a 100:
          </p>
          <div className="rounded-lg bg-muted p-6 font-mono text-sm">
            <p className="text-foreground">
              <span className="text-primary font-bold">Nota Final</span> =
            </p>
            <p className="mt-2 pl-4 text-muted-foreground">
              Produtividade Legislativa x{" "}
              <span className="text-primary font-semibold">0,35</span> +
            </p>
            <p className="pl-4 text-muted-foreground">
              Presença em Votações x{" "}
              <span className="text-primary font-semibold">0,25</span> +
            </p>
            <p className="pl-4 text-muted-foreground">
              Economia da Cota x{" "}
              <span className="text-primary font-semibold">0,20</span> +
            </p>
            <p className="pl-4 text-muted-foreground">
              Participação em Comissões x{" "}
              <span className="text-primary font-semibold">0,20</span>
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Criteria Details */}
      <div className="space-y-8">
        <h2 className="text-2xl font-bold tracking-tight text-foreground">
          Os quatro critérios
        </h2>

        {criterios.map((criterio, index) => (
          <Card key={criterio.nome}>
            <CardHeader>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <CardTitle className="flex items-center gap-3">
                  <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground">
                    {index + 1}
                  </span>
                  {criterio.nome}
                </CardTitle>
                <span className="w-fit rounded-full bg-primary/10 px-3 py-1 text-sm font-bold text-primary">
                  Peso: {criterio.peso}
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
                  Pontuação:
                </h4>
                <ul className="list-inside list-disc space-y-1 text-sm text-muted-foreground">
                  {criterio.detalhes.map((detalhe, i) => (
                    <li key={i}>{detalhe}</li>
                  ))}
                </ul>
              </div>

              {"extras" in criterio && criterio.extras && (
                <div>
                  <h4 className="mb-2 text-sm font-semibold text-foreground">
                    Peso por tipo de proposta:
                  </h4>
                  <ul className="list-inside list-disc space-y-1 text-sm text-muted-foreground">
                    {criterio.extras.map((extra, i) => (
                      <li key={i}>{extra}</li>
                    ))}
                  </ul>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Data Sources */}
      <Card className="mt-12">
        <CardHeader>
          <CardTitle>De onde vêm os dados?</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-6 text-sm text-muted-foreground">
            Todos os dados são públicos. Qualquer pessoa pode verificar
            nas fontes oficiais:
          </p>
          <div className="grid gap-6 sm:grid-cols-3">
            <div>
              <h4 className="font-semibold text-foreground">
                Senado (Legislativo)
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Projetos de lei, votações, comissões e mandatos.
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
                Senado (Administrativo)
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Despesas da cota parlamentar (CEAPS).
              </p>
              <a
                href="https://adm.senado.gov.br/adm-dadosabertos/swagger-ui"
                target="_blank"
                rel="noopener noreferrer"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                adm.senado.gov.br/adm-dadosabertos
              </a>
            </div>
            <div>
              <h4 className="font-semibold text-foreground">
                Portal da Transparência
              </h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Emendas parlamentares e contratos do Governo Federal.
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

      {/* Limitations */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>Limitações</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="list-inside list-disc space-y-2 text-sm text-muted-foreground">
            <li>
              <strong>Trabalho de bastidores</strong> (negociações,
              articulações) não é mensurável com dados públicos.
            </li>
            <li>
              <strong>Gastar menos não é ser melhor</strong> -- a cota tem peso
              moderado (20%) e varia por estado.
            </li>
            <li>
              <strong>Dados podem ter atraso</strong> -- as fontes oficiais
              nem sempre atualizam em tempo real.
            </li>
          </ul>
        </CardContent>
      </Card>

      {/* Academic References */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>Base científica</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="list-inside list-disc space-y-2 text-sm text-muted-foreground">
            <li>
              Baseado no{" "}
              <strong>
                Legislative Effectiveness Score
              </strong>{" "}
              de Volden e Wiseman (Universidade Vanderbilt, 2014).
            </li>
            <li>
              Índice usado internacionalmente para avaliar parlamentares de
              forma objetiva.
            </li>
            <li>
              Adaptado para o Senado brasileiro, considerando o sistema
              multipartidário e os dados disponíveis nas APIs do governo.
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
