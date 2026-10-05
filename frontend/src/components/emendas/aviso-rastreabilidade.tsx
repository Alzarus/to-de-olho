import { AlertTriangle } from "lucide-react";

// Primeiro ano com as regras de rastreabilidade (ADPF 854 e LC 210/2024)
export const ANO_RASTREABILIDADE = 2025;

/** ano 0 é o mandato inteiro (desde 2023), que inclui anos sem rastreabilidade */
export function periodoSemRastreabilidade(ano: number): boolean {
  return ano === 0 || ano < ANO_RASTREABILIDADE;
}

export function AvisoRastreabilidade({ ano }: { ano: number }) {
  if (!periodoSemRastreabilidade(ano)) return null;
  return (
    <div
      role="note"
      className="flex gap-3 rounded-lg border border-amber-300 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-200"
    >
      <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
      <div className="space-y-1">
        <p className="font-semibold">Rastreabilidade limitada antes de 2025</p>
        <p>
          Até 2024, as transferências especiais (&ldquo;emendas Pix&rdquo;)
          caíam direto no caixa de estados e municípios, sem plano de trabalho
          que mostrasse onde o dinheiro seria aplicado, e parte das indicações
          de parlamentares era executada sem o nome de quem indicou (como nas
          emendas de relator, o &ldquo;orçamento secreto&rdquo;, até 2022). As
          regras de rastreabilidade vieram com as decisões do STF na ADPF 854 e
          com a Lei Complementar 210/2024.
        </p>
        <p>
          {ano === 0 ? "Nos anos de 2023 e 2024, os" : `Em ${ano}, os`} dados
          mostram o destino declarado (estado ou município), não o beneficiário
          final, e podem não incluir tudo o que o senador indicou.
        </p>
      </div>
    </div>
  );
}
