import Link from "next/link";

const footerLinks = {
  projeto: [
    { name: "Sobre", href: "/sobre" },
    { name: "Metodologia", href: "/metodologia" },
    { name: "Código Fonte", href: "https://github.com/Alzarus/to-de-olho", external: true },
  ],
  dados: [
    { name: "API Legislativa do Senado", href: "https://legis.senado.leg.br/dadosabertos", external: true },
    { name: "API Administrativa do Senado", href: "https://adm.senado.gov.br/adm-dadosabertos/swagger-ui", external: true },
    { name: "Portal da Transparência", href: "https://portaldatransparencia.gov.br", external: true },
  ],
};

export function Footer() {
  return (
    <footer className="border-t border-border bg-muted/30">
      <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid gap-8 md:grid-cols-3">
          {/* Brand */}
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src="/logo.png"
                alt="Tô De Olho"
                className="h-10 w-10 rounded-lg bg-white object-contain p-1"
              />
              <span className="text-lg font-bold tracking-tight">Tô De Olho</span>
            </div>
            <p className="max-w-xs text-sm text-muted-foreground">
              Plataforma de transparência e acompanhamento da atuação dos
              senadores brasileiros com dados abertos.
            </p>
            <p className="max-w-xs text-xs text-muted-foreground mt-2">
               Dados atualizados em tempo real via APIs oficiais.
            </p>
          </div>

          {/* Links - Projeto */}
          <div>
            <h3 className="mb-4 text-sm font-semibold uppercase tracking-wider text-foreground">
              Projeto
            </h3>
            <ul className="space-y-3">
              {footerLinks.projeto.map((link) => (
                <li key={link.name}>
                  <Link
                    href={link.href}
                    className="text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded"
                    {...(link.external && {
                      target: "_blank",
                      rel: "noopener noreferrer",
                    })}
                  >
                    {link.name}
                    {link.external && (
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="ml-1 inline-block h-3 w-3"
                        aria-hidden="true"
                      >
                        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                        <polyline points="15 3 21 3 21 9" />
                        <line x1="10" x2="21" y1="14" y2="3" />
                      </svg>
                    )}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Links - Dados */}
          <div>
            <h3 className="mb-4 text-sm font-semibold uppercase tracking-wider text-foreground">
              Fontes de Dados
            </h3>
            <ul className="space-y-3">
              {footerLinks.dados.map((link) => (
                <li key={link.name}>
                  <Link
                    href={link.href}
                    className="text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {link.name}
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="ml-1 inline-block h-3 w-3"
                      aria-hidden="true"
                    >
                      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                      <polyline points="15 3 21 3 21 9" />
                      <line x1="10" x2="21" y1="14" y2="3" />
                    </svg>
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom */}
        <div className="mt-12 border-t border-border pt-8 text-center">
             {/* Empty or minimal if requested */}
        </div>
      </div>
    </footer>
  );
}
