import type { Metadata } from "next";
import SenadorClient from "./senador-client";
import { buscarApi, jsonLd, NOME_SITE, SITE_URL } from "@/lib/seo";

type Props = {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
};

type SenadorMeta = {
  nome?: string;
  nome_completo?: string;
  partido?: string;
  uf?: string;
  foto_url?: string;
  em_exercicio?: boolean;
};

type ScoreMeta = { posicao?: number; score_final?: number; dados_insuficientes?: boolean };

// Mesma URL nas duas chamadas (metadata e página): o fetch do Next deduplica
const buscarSenador = (id: string) => buscarApi<SenadorMeta>(`/api/v1/senadores/${encodeURIComponent(id)}`);
const buscarScore = (id: string) => buscarApi<ScoreMeta>(`/api/v1/senadores/${encodeURIComponent(id)}/score`);

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const canonical = `/senador/${id}`;
  const [senador, score] = await Promise.all([buscarSenador(id), buscarScore(id)]);

  if (!senador?.nome) {
    return {
      title: "Senador não encontrado",
      description: "Informações detalhadas sobre senadores brasileiros.",
      alternates: { canonical },
    };
  }

  const title = `Senador ${senador.nome} (${senador.partido || "-"}-${senador.uf || "-"})`;
  // Antes lia score_ranking, que a API não tem: toda descrição saía com "undefinedº lugar"
  const posicao =
    score && !score.dados_insuficientes && score.posicao
      ? ` ${score.posicao}º no ranking do Tô De Olho, com nota ${score.score_final?.toFixed(1).replace(".", ",")}.`
      : "";
  const description = `Desempenho de ${senador.nome} no Senado: proposições, presença em votações, gastos da cota parlamentar, emendas e gabinete.${posicao}`;
  // A API devolve foto_url (o código lia url_foto, e a imagem era sempre o logo)
  const imagem = senador.foto_url || "/logo.png";

  return {
    title,
    description,
    alternates: { canonical },
    openGraph: {
      title,
      description,
      type: "profile",
      images: [{ url: imagem, alt: `Foto de ${senador.nome}` }],
    },
    twitter: {
      card: "summary",
      title,
      description,
      images: [imagem],
    },
  };
}

export default async function Page({ params }: Props) {
  const { id } = await params;
  const senador = await buscarSenador(id);
  const dados = senador?.nome
    ? {
        "@context": "https://schema.org",
        "@type": "Person",
        name: senador.nome,
        ...(senador.nome_completo && senador.nome_completo !== senador.nome
          ? { alternateName: senador.nome_completo }
          : {}),
        jobTitle: "Senador da República",
        ...(senador.foto_url ? { image: senador.foto_url } : {}),
        ...(senador.partido ? { affiliation: { "@type": "PoliticalParty", name: senador.partido } } : {}),
        ...(senador.uf ? { workLocation: { "@type": "State", name: senador.uf } } : {}),
        url: `${SITE_URL}/senador/${id}`,
        mainEntityOfPage: { "@type": "ProfilePage", name: `${senador.nome} | ${NOME_SITE}` },
      }
    : null;

  return (
    <>
      {dados && (
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: jsonLd(dados) }} />
      )}
      <SenadorClient />
    </>
  );
}
