import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Ranking dos senadores",
  description:
    "Ranking dos 81 senadores por produtividade legislativa, presença em votações, economia da cota parlamentar e participação em comissões, com metodologia pública.",
  alternates: { canonical: "/ranking" },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
