import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Comparar senadores",
  description:
    "Compare até quatro senadores lado a lado: votos, gastos da cota parlamentar, gabinete e emendas.",
  alternates: { canonical: "/comparar" },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
