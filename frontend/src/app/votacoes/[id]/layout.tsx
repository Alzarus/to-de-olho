import type { Metadata } from "next";
import { buscarApi, NOME_SITE } from "@/lib/seo";

type Props = { params: Promise<{ id: string }> };

type VotacaoMeta = {
  votacao?: {
    data?: string;
    materia?: string | null;
    ementa?: string | null;
    descricao_votacao?: string | null;
    resultado?: string | null;
  };
};

const RESULTADO: Record<string, string> = { A: "aprovada", R: "rejeitada" };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const canonical = `/votacoes/${id}`;
  const v = (await buscarApi<VotacaoMeta>(`/api/v1/votacoes/${encodeURIComponent(id)}`))?.votacao;
  if (!v) return { title: "Votação no Senado", alternates: { canonical } };

  const data = v.data ? new Date(v.data).toLocaleDateString("pt-BR", { timeZone: "UTC" }) : "";
  const assunto = v.materia || v.descricao_votacao || "Votação nominal";
  const resultado = v.resultado && RESULTADO[v.resultado] ? `, ${RESULTADO[v.resultado]}` : "";
  const title = `${assunto}${data ? ` (${data})` : ""}`;
  const resumo = (v.ementa || v.descricao_votacao || "").trim();
  const description =
    `Votação nominal no Plenário do Senado${data ? ` em ${data}` : ""}${resultado}. Veja como votou cada senador.` +
    (resumo ? ` ${resumo.length > 160 ? `${resumo.slice(0, 157)}...` : resumo}` : "");

  return {
    // O template do layout raiz não chega aqui (o layout de /votacoes define
    // um título simples), então o sufixo vai à mão
    title: { absolute: `${title} | ${NOME_SITE}` },
    description,
    alternates: { canonical },
    // openGraph próprio substitui o herdado inteiro: a imagem precisa vir junto
    openGraph: { title, description, type: "article", images: ["/logo.png"] },
  };
}

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
