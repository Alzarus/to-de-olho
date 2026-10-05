import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Votações nominais do Senado",
  description:
    "Todas as votações nominais do Plenário do Senado na legislatura atual, com o resultado e o voto de cada senador.",
  alternates: { canonical: "/votacoes" },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
