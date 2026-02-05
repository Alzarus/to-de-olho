import type { Metadata } from "next";
import SenadorClient from "./senador-client";

type Props = {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
};

async function getSenador(id: string) {
  // Em produção, usa o URL interno ou público.
  // Como estamos no server-side, idealmente usaríamos o URL interno do container se possível,
  // mas aqui vamos usar o BACKEND_URL público ou localhost
  const baseUrl = process.env.BACKEND_URL || "http://localhost:8080";
  
  try {
    const res = await fetch(`${baseUrl}/api/v1/senadores/${id}`, {
      next: { revalidate: 3600 }, // Cache de 1 hora
    });
    
    if (!res.ok) return null;
    return res.json();
  } catch (error) {
    console.error("Erro ao buscar senador para metadata:", error);
    return null;
  }
}

export async function generateMetadata(
  { params }: Props
): Promise<Metadata> {
  const { id } = await params;
  const senador = await getSenador(id);

  if (!senador) {
    return {
      title: "Senador não encontrado | Tô De Olho",
      description: "Informações detalhadas sobre senadores brasileiros.",
    };
  }

  const title = `Senador ${senador.nome_civil} (${senador.partido}-${senador.uf}) | Tô De Olho`;
  const description = `Veja o desempenho de ${senador.nome_civil} no Senado: Produtividade, Presença, Gastos e Emendas. Ranking: ${senador.score_ranking?.posicao}º lugar.`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      images: [
        {
          url: senador.url_foto || "/logo.png",
          width: 800,
          height: 600,
          alt: `Foto de ${senador.nome_civil}`,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [senador.url_foto || "/logo.png"],
    },
  };
}

export default function Page() {
  return <SenadorClient />;
}
