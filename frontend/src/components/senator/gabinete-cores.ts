// Cores fixas por vínculo (ordem categórica fixa: a cor acompanha o vínculo,
// não a posição). Tons validados para contraste e daltonismo.
const CORES_VINCULO: Record<string, string> = {
  Comissionado: "#2a78d6",
  Efetivo: "#eb6834",
  Requisitado: "#1baf7a",
};
const COR_OUTROS = "#eda100";

export function corVinculo(vinculo: string): string {
  return CORES_VINCULO[vinculo] ?? COR_OUTROS;
}

// Ordem estável dos vínculos nas legendas e tabelas
export function ordenarVinculos(vinculos: Iterable<string>): string[] {
  const conhecidos = Object.keys(CORES_VINCULO);
  return [...new Set(vinculos)].sort((a, b) => {
    const ia = conhecidos.indexOf(a);
    const ib = conhecidos.indexOf(b);
    if (ia !== -1 || ib !== -1) return (ia === -1 ? 99 : ia) - (ib === -1 ? 99 : ib);
    return a.localeCompare(b, "pt-BR");
  });
}

// Texto da nota para quem ocupa cargo na Mesa Diretora
export function notaMesa(cargo: string): string {
  const cargoFmt = cargo.charAt(0) + cargo.slice(1).toLowerCase();
  if (cargo.toUpperCase() === "PRESIDENTE") {
    return "Atualmente preside o Senado: parte da equipe fica lotada nos órgãos da Presidência e da Mesa, fora do gabinete, e não entra nestes números.";
  }
  return `Atualmente ocupa o cargo de ${cargoFmt} na Mesa Diretora: parte da equipe pode estar lotada nos órgãos da Mesa e não entrar nestes números.`;
}

export const NOTA_PRIVACIDADE =
  "Por decisão de privacidade, mostramos apenas números agregados (sem nomes nem salários individuais). O detalhe está na página oficial de transparência do Senado.";
