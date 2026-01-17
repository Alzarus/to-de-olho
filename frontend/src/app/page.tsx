import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const features = [
  {
    title: "Ranking de Senadores",
    description:
      "Avaliação objetiva baseada em produtividade legislativa, presença em votações, economia de recursos e participação em comissões.",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-6 w-6"
        aria-hidden="true"
      >
        <line x1="18" x2="18" y1="20" y2="10" />
        <line x1="12" x2="12" y1="20" y2="4" />
        <line x1="6" x2="6" y1="20" y2="14" />
      </svg>
    ),
    href: "/ranking",
  },
  {
    title: "Dados Abertos",
    description:
      "Informações atualizadas diretamente das APIs oficiais do Senado Federal e Portal da Transparência.",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-6 w-6"
        aria-hidden="true"
      >
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M3 5v14a9 3 0 0 0 18 0V5" />
        <path d="M3 12a9 3 0 0 0 18 0" />
      </svg>
    ),
    href: "/metodologia",
  },
  {
    title: "Metodologia Transparente",
    description:
      "Critérios e pesos do ranking completamente documentados e baseados em literatura acadêmica.",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-6 w-6"
        aria-hidden="true"
      >
        <path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H19a1 1 0 0 1 1 1v18a1 1 0 0 1-1 1H6.5a1 1 0 0 1 0-5H20" />
      </svg>
    ),
    href: "/metodologia",
  },
];

const stats = [
  { value: "81", label: "Senadores Monitorados" },
  { value: "85.877", label: "Votações Registradas" },
  { value: "4", label: "Critérios de Avaliação" },
  { value: "100%", label: "Dados Públicos" },
];

export default function Home() {
  return (
    <div className="flex flex-col">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-b from-primary/5 via-background to-background py-20 sm:py-32">
        <div className="container mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-3xl text-center">
            <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
              Transparência no{" "}
              <span className="text-primary">Senado Federal</span>
            </h1>
            <p className="mt-6 text-lg leading-8 text-muted-foreground sm:text-xl">
              Acompanhe a atuação dos senadores brasileiros com dados abertos.
              Ranking objetivo baseado em produtividade, presença, economia e
              participação.
            </p>
            <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
              <Button asChild size="lg" className="w-full sm:w-auto">
                <Link href="/ranking">Ver Ranking Completo</Link>
              </Button>
              <Button
                asChild
                variant="outline"
                size="lg"
                className="w-full sm:w-auto"
              >
                <Link href="/metodologia">Entender Metodologia</Link>
              </Button>
            </div>
          </div>
        </div>

        {/* Background decoration */}
        <div
          className="absolute inset-x-0 -top-40 -z-10 transform-gpu overflow-hidden blur-3xl sm:-top-80"
          aria-hidden="true"
        >
          <div
            className="relative left-[calc(50%-11rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 rotate-[30deg] bg-gradient-to-tr from-primary to-secondary opacity-20 sm:left-[calc(50%-30rem)] sm:w-[72.1875rem]"
            style={{
              clipPath:
                "polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)",
            }}
          />
        </div>
      </section>

      {/* Stats Section */}
      <section className="border-y border-border bg-muted/30 py-16">
        <div className="container mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <dl className="grid grid-cols-2 gap-8 lg:grid-cols-4">
            {stats.map((stat) => (
              <div
                key={stat.label}
                className="flex flex-col items-center gap-2 text-center"
              >
                <dt className="text-sm text-muted-foreground">{stat.label}</dt>
                <dd className="text-3xl font-bold tracking-tight text-primary sm:text-4xl">
                  {stat.value}
                </dd>
              </div>
            ))}
          </dl>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 sm:py-32">
        <div className="container mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-2xl text-center">
            <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
              Como funciona
            </h2>
            <p className="mt-4 text-lg text-muted-foreground">
              Uma plataforma criada para facilitar o acompanhamento cidadão da
              atuação dos senadores.
            </p>
          </div>

          <div className="mx-auto mt-16 grid max-w-5xl gap-8 lg:grid-cols-3">
            {features.map((feature) => (
              <Card
                key={feature.title}
                className="group relative overflow-hidden transition-all hover:shadow-lg hover:shadow-primary/5"
              >
                <CardHeader>
                  <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                    {feature.icon}
                  </div>
                  <CardTitle className="text-xl">{feature.title}</CardTitle>
                  <CardDescription className="text-base">
                    {feature.description}
                  </CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
