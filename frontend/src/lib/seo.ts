// Metadados e dados estruturados (SEO) montados no servidor

export const SITE_URL = process.env.FRONTEND_URL || "https://todeolho.org";
export const NOME_SITE = "Tô De Olho";

/**
 * GET na API a partir do servidor (generateMetadata, sitemap), com cache.
 * Falha vira null: a página sai com a metadata genérica, sem quebrar.
 */
export async function buscarApi<T>(caminho: string, revalidate = 3600): Promise<T | null> {
  const base = process.env.BACKEND_URL || "http://localhost:8080";
  try {
    const res = await fetch(`${base}${caminho}`, { next: { revalidate } });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch (error) {
    console.error(`SEO: falha ao buscar ${caminho}`, error);
    return null;
  }
}

/** JSON-LD seguro dentro de <script>: "<" escapado impede fechar a tag */
export function jsonLd(dados: object): string {
  return JSON.stringify(dados).replace(/</g, "\u003c");
}
