import { MetadataRoute } from "next";
import { buscarApi, SITE_URL } from "@/lib/seo";

// Gerado a cada pedido: estático, ele saía do build, quando a API não está no
// ar, e ficava só com as rotas fixas (em produção, 6 URLs e nenhum senador).
// As chamadas à API ficam 6 horas em cache (a carga dos dados é diária).
export const dynamic = "force-dynamic";
const revalidate = 21600;

type ListaSenadores = { senadores?: { id: number }[] };
type PaginaVotacoes = { data?: { codigo_votacao: number; data?: string }[]; total?: number };

// Teto da API de votações por página
const POR_PAGINA = 100;

async function votacoes(): Promise<MetadataRoute.Sitemap> {
  const primeira = await buscarApi<PaginaVotacoes>(`/api/v1/votacoes?limit=${POR_PAGINA}&page=1`, revalidate);
  if (!primeira?.data) return [];
  const paginas = Math.ceil((primeira.total ?? 0) / POR_PAGINA);
  const resto = await Promise.all(
    Array.from({ length: Math.max(0, paginas - 1) }, (_, i) =>
      buscarApi<PaginaVotacoes>(`/api/v1/votacoes?limit=${POR_PAGINA}&page=${i + 2}`, revalidate),
    ),
  );
  const todas = [primeira, ...resto].flatMap((p) => p?.data ?? []);
  // A mesma votação aparece uma vez (o código identifica a votação)
  const vistas = new Set<number>();
  return todas
    .filter((v) => !vistas.has(v.codigo_votacao) && vistas.add(v.codigo_votacao))
    .map((v) => ({
      url: `${SITE_URL}/votacoes/${v.codigo_votacao}`,
      ...(v.data ? { lastModified: new Date(v.data) } : {}),
      changeFrequency: "yearly" as const,
      priority: 0.5,
    }));
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const estaticas = ["", "/ranking", "/emendas", "/comparar", "/votacoes", "/metodologia", "/sobre"].map(
    (rota) => ({
      url: `${SITE_URL}${rota}`,
      changeFrequency: (rota === "/metodologia" || rota === "/sobre" ? "monthly" : "daily") as
        | "monthly"
        | "daily",
      priority: rota === "" ? 1 : 0.8,
    }),
  );

  const [lista, paginasVotacoes] = await Promise.all([
    buscarApi<ListaSenadores>("/api/v1/senadores", revalidate),
    votacoes(),
  ]);
  const senadores = (lista?.senadores ?? []).map((s) => ({
    url: `${SITE_URL}/senador/${s.id}`,
    changeFrequency: "weekly" as const,
    priority: 0.7,
  }));

  return [...estaticas, ...senadores, ...paginasVotacoes];
}
