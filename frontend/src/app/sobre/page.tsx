
import { Metadata } from "next";
import { Badge } from "@/components/ui/badge";
import { Mail } from "lucide-react";
import { GithubIcon, LinkedinIcon } from "@/components/ui/brand-icons";

export const metadata: Metadata = {
  title: "Sobre | Tô de Olho",
  description: "Conheça o projeto de TCC Tô de Olho.",
};

export default function SobrePage() {
  return (
    <div className="container mx-auto px-4 py-12 max-w-4xl">
      <div className="space-y-4 text-center mb-16">
        <h1 className="text-4xl font-bold tracking-tight sm:text-5xl text-foreground">
          Sobre o Projeto
        </h1>
        <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
          Transparência política acessível para o cidadão brasileiro.
        </p>
      </div>

      <div className="grid gap-12 md:grid-cols-2">
        {/* Coluna Projeto */}
        <div className="space-y-6">
          <div className="space-y-2">
            <h2 className="text-2xl font-semibold text-primary">A Iniciativa</h2>
            <div className="h-1 w-20 bg-primary/20 rounded-full" />
          </div>
          
          <p className="leading-relaxed text-muted-foreground">
            O <strong>Tô de Olho</strong> é uma plataforma de monitoramento legislativo desenvolvida como Trabalho de Conclusão de Curso (TCC). 
            Seu objetivo é simplificar o acesso aos dados abertos do Senado Federal, permitindo que qualquer cidadão acompanhe 
            a atuação de seus representantes sem burocracia.
          </p>

          <p className="leading-relaxed text-muted-foreground">
            Utilizando índices de produtividade, transparência de gastos e histórico de votações, o projeto busca 
            fomentar o voto consciente e o controle social no Brasil.
          </p>

          <div className="flex flex-wrap gap-2 pt-4">
            <Badge variant="secondary">Fiscalização</Badge>
            <Badge variant="secondary">Dados Abertos</Badge>
            <Badge variant="secondary">Cidadania</Badge>
          </div>
        </div>

        {/* Coluna Autor */}
        <div className="relative rounded-2xl border bg-card p-8 shadow-sm">
          <div className="space-y-6">
            <div className="space-y-2">
              <h2 className="text-2xl font-semibold text-primary">O Autor</h2>
              <div className="h-1 w-20 bg-primary/20 rounded-full" />
            </div>

            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <div className="h-20 w-20 rounded-full overflow-hidden border-2 border-primary/20">
                  <img 
                    src="/pedro-almeida.png" 
                    alt="Pedro Almeida" 
                    className="h-full w-full object-cover"
                  />
                </div>
                <div>
                  <h3 className="font-bold text-lg">Pedro Batista de Almeida Filho</h3>
                  <p className="text-sm text-muted-foreground">Desenvolvedor Full Stack</p>
                </div>
              </div>

              <div className="space-y-2 text-sm text-muted-foreground">
                <p>🎓 Tecnólogo em Análise e Desenvolvimento de Sistemas (ADS)</p>
                <p>🏫 Instituto Federal da Bahia (IFBA), 2017–2026</p>
                <p>
                  📄 TCC:{" "}
                  <a
                    href="https://ads.ifba.edu.br/item1259"
                    target="_blank"
                    rel="noreferrer"
                    className="text-primary hover:underline"
                  >
                    Tô De Olho: Democratizando a Transparência do Senado Federal
                    através de Dados Abertos
                  </a>
                </p>
              </div>
            </div>
            
            <div className="pt-6 border-t mt-6 flex gap-4">
                <a href="https://github.com/Alzarus" target="_blank" rel="noreferrer" className="text-muted-foreground hover:text-primary transition-colors">
                    <GithubIcon className="w-5 h-5" />
                    <span className="sr-only">GitHub</span>
                </a>
                <a href="https://www.linkedin.com/in/pedroalmei/" target="_blank" rel="noreferrer" className="text-muted-foreground hover:text-primary transition-colors">
                    <LinkedinIcon className="w-5 h-5" />
                    <span className="sr-only">LinkedIn</span>
                </a>
                 <a href="mailto:pedro.almei@hotmail.com" className="text-muted-foreground hover:text-primary transition-colors">
                    <Mail className="w-5 h-5" />
                     <span className="sr-only">Email</span>
                </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
