import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const criterios = [
  {
    nome: "Produtividade Legislativa",
    peso: "35%",
    descricao:
      "Mede a capacidade do senador de criar e aprovar leis. Quanto mais longe o projeto avança, mais pontos ele ganha.",
    formula: "Nota = ln(1 + Pontos do Senador) / ln(1 + Maior Pontuador) x 100",
    detalhes: [
      "Apresentado: 1 ponto",
      "Em discussão nas comissões: 2 pontos",
      "Aprovado em comissão: 4 pontos",
      "Aprovado no Plenário: 8 pontos",
      "Virou lei: 16 pontos",
      "Só o primeiro autor pontua, e só como senador: coautorias aparecem na ficha, sem pontos",
      "Matérias de quando o senador era deputado não pontuam",
      "Vetos não pontuam: são ato do Presidente sobre lei já aprovada",
      "Matérias de autoria institucional (Mesa, comissão, partido) não pontuam para ninguém",
    ],
    extras: [
      "PECs (mudanças na Constituição): peso x3",
      "PLPs (Lei Complementar): peso x2",
      "Projetos de lei, de decreto legislativo e de resolução (PL, PLS, PDL, PDS, PRS, PRN): peso x1",
      "Requerimentos ao Plenário, moções e propostas de fiscalização (RQS, MOC, PFS): peso x0,5",
      "Requerimentos de comissão e indicações (REQ, INS, RDH e outros): peso x0,1",
      "Emendas da Câmara, ofícios, mensagens, petições e outros documentos que não são autoria do senador: não pontuam",
      "Ajuste logarítmico impede que quantidade supere qualidade",
    ],
  },
  {
    nome: "Presença em Votações",
    peso: "25%",
    descricao:
      "Mede se o senador comparece quando o Senado vota. Cada votação nominal do Plenário conta separadamente.",
    formula:
      "Nota = Presenças / (Votações no período − Licenças de saúde e missões oficiais) x 100",
    detalhes: [
      "Conta como presença: Sim, Não, Abstenção, voto secreto, presente sem registrar voto e presidência da sessão",
      "Licença de saúde (e demais licenças legais) e missões oficiais saem da conta: não penalizam nem ajudam",
      "Licença particular (LP) conta como falta: o TCC só justifica licença médica e missão oficial",
      "\"Atividade parlamentar\" (AP) conta como falta: é uma justificativa declarada pelo próprio senador, sem verificação",
      "\"Não compareceu\" conta como falta",
      "Obstrução fica fora da conta: não é presença nem falta",
      "Considera as votações desde a posse da legislatura atual (01/02/2023). Nos 6 primeiros meses de uma nova legislatura, o ranking continua mostrando a anterior, já encerrada",
      "A ficha mostra também a presença bruta, sem descontar licenças",
    ],
  },
  {
    nome: "Economia da Cota Parlamentar (CEAPS)",
    peso: "20%",
    descricao:
      "Avalia quanto o senador economiza da sua cota mensal de gastos. Quanto menos gastar, melhor a nota.",
    formula: "Nota = (1 − Gasto no período / Teto no período) x 100",
    detalhes: [
      "Teto no período = teto mensal do estado x meses em exercício",
      "Quem assumiu depois tem teto menor: a nota mede economia, não a data da posse",
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
    formula: "Nota = (Pontos do Senador / Maior Pontuador) x 100",
    detalhes: [
      "Membro titular (com direito a voto): 2 pontos",
      "Suplente (substituto eventual): 1 ponto",
      "Cada comissão conta uma vez no período, pelo papel mais alto",
      "Frentes parlamentares, grupos de amizade e conselhos de honrarias não contam",
      "A API do Senado não informa quem preside a comissão; presidência não pontua à parte",
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

      {/* Sem dados */}
      <Card className="mt-12">
        <CardHeader>
          <CardTitle>Período e dados insuficientes</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">
            Todos os critérios usam o mesmo período: desde a posse da
            legislatura (01/02/2023) no ranking do mandato, ou o ano escolhido
            no ranking anual.
          </p>
          <p className="text-sm text-muted-foreground">
            Fica fora da ordenação, como &quot;dados insuficientes&quot;, quem
            esteve menos de 6 meses em exercício no período ou não tem nenhum
            registro de votação: o último lugar seria uma afirmação que o dado
            não sustenta. Produtividade zero continua sendo nota zero, porque é
            um dado real.
          </p>
        </CardContent>
      </Card>

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
                Despesas da cota parlamentar (CEAPS) e estrutura de
                gabinete (quantidade de servidores por local e vínculo,
                auxílio-moradia e imóvel funcional).
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
          <div className="mt-6 border-t pt-6">
            <h4 className="font-semibold text-foreground">
              Nome popular e descrição das matérias
            </h4>
            <p className="mt-1 text-sm text-muted-foreground">
              O nome popular de uma matéria (como &ldquo;Marco Legal da
              Inteligência Artificial&rdquo; para o PL 2338/2023) é o apelido
              registrado pelo próprio Senado. Quando o Senado não registra
              apelido, usamos uma lista curta de nomes que aparecem em páginas
              oficiais (Agência Senado, Câmara ou gov.br), marcada como
              &ldquo;nome popular&rdquo; e com o link da fonte. A descrição
              breve é a explicação da ementa publicada pelo Senado ou, na falta
              dela, a própria ementa, e os temas são a classificação do Senado.
              Nenhum texto é gerado por inteligência artificial. A
              identificação oficial (&ldquo;PL 2338/2023&rdquo;) fica sempre
              visível.
            </p>
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
              <strong>Gabinete em números agregados</strong> -- por
              privacidade, mostramos só quantos servidores há por local e
              vínculo, sem nomes nem salários; o detalhe está na página de
              transparência do Senado. Quem ocupa cargo na Mesa Diretora
              (como a Presidência) tem parte da equipe fora do gabinete.
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
            <li>
              A versão completa, com fórmulas e a tabela de códigos de voto,
              está em{" "}
              <a
                href="https://github.com/Alzarus/to-de-olho/blob/master/METODOLOGIA.md"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                METODOLOGIA.md
              </a>
              .
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
