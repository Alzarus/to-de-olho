"use client";

import { useId, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const TEXTO_CURADORIA =
  "Nome popular atribuído pela curadoria do Tô De Olho, com base em página oficial; o Senado não registra apelido para esta matéria.";

function dominio(url: string): string {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

/**
 * Marca discreta do apelido de curadoria. Em listas clicáveis (comLink falso)
 * é só texto, sem elemento interativo aninhado; a fonte aparece no detalhe.
 */
export function MarcaCuradoria({
  fonteUrl,
  comLink = false,
  className,
}: {
  fonteUrl?: string | null;
  comLink?: boolean;
  className?: string;
}) {
  if (comLink && fonteUrl) {
    return (
      <span className={cn("inline-flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground", className)}>
        <Badge variant="outline" className="font-normal">
          nome popular
        </Badge>
        <span>
          Curadoria do Tô De Olho, com base em{" "}
          <a
            href={fonteUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="underline underline-offset-2 hover:text-primary focus-visible:outline-2 focus-visible:outline-ring"
          >
            fonte oficial ({dominio(fonteUrl)})
            <span className="sr-only"> (abre em nova aba)</span>
          </a>
          .
        </span>
      </span>
    );
  }
  return (
    <Badge variant="outline" className={cn("font-normal text-muted-foreground", className)} title={TEXTO_CURADORIA}>
      nome popular
      <span className="sr-only">: {TEXTO_CURADORIA}</span>
    </Badge>
  );
}

/** Temas da matéria (classificações do Senado) como chips. */
export function TemasChips({
  temas,
  max = 3,
  className,
}: {
  temas?: string[] | null;
  max?: number;
  className?: string;
}) {
  if (!temas || temas.length === 0) return null;
  const visiveis = temas.slice(0, max);
  const restantes = temas.length - visiveis.length;
  return (
    <ul className={cn("flex flex-wrap gap-1.5", className)} aria-label="Temas da matéria">
      {visiveis.map((t) => (
        <li key={t}>
          <Badge variant="secondary" className="font-normal">
            {t}
          </Badge>
        </li>
      ))}
      {restantes > 0 && (
        <li>
          <Badge variant="secondary" className="font-normal" title={temas.slice(max).join(", ")}>
            +{restantes}
            <span className="sr-only"> temas: {temas.slice(max).join(", ")}</span>
          </Badge>
        </li>
      )}
    </ul>
  );
}

// Abaixo disso o texto cabe em duas linhas na maioria das larguras
const LIMITE_VER_MAIS = 140;

/**
 * Descrição com line-clamp de 2 linhas e botão "ver mais" (aria-expanded).
 * O clique no botão não propaga: as linhas das listas são clicáveis.
 */
export function DescricaoExpansivel({
  texto,
  className,
}: {
  texto: string;
  className?: string;
}) {
  const id = useId();
  const [aberta, setAberta] = useState(false);
  if (!texto) return null;
  const longo = texto.length > LIMITE_VER_MAIS;
  return (
    <div className={className}>
      <p id={id} className={cn("text-sm text-muted-foreground", !aberta && longo && "line-clamp-2")}>
        {texto}
      </p>
      {longo && (
        <button
          type="button"
          aria-expanded={aberta}
          aria-controls={id}
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            setAberta((v) => !v);
          }}
          onKeyDown={(e) => e.stopPropagation()}
          className="mt-0.5 text-xs font-medium text-primary underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-ring rounded-sm"
        >
          {aberta ? "ver menos" : "ver mais"}
          <span className="sr-only"> da descrição</span>
        </button>
      )}
    </div>
  );
}
