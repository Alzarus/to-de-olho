import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Emendas parlamentares dos senadores",
  description:
    "Emendas individuais de cada senador: valores empenhados e pagos, transferências especiais (emendas Pix) e destino dos recursos, com dados do Portal da Transparência.",
  alternates: { canonical: "/emendas" },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
