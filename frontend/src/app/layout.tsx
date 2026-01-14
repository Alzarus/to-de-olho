import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Header, Footer } from "@/components/layout";

const inter = Inter({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "Tô De Olho - Transparência no Senado",
    template: "%s | Tô De Olho",
  },
  description:
    "Plataforma de transparência e acompanhamento da atuação dos senadores brasileiros. Ranking baseado em produtividade, presença e economia.",
  keywords: [
    "senado",
    "transparência",
    "senadores",
    "ranking",
    "política",
    "brasil",
    "dados abertos",
  ],
  authors: [{ name: "Pedro Eli" }],
  openGraph: {
    type: "website",
    locale: "pt_BR",
    siteName: "Tô De Olho",
    title: "Tô De Olho - Transparência no Senado",
    description:
      "Plataforma de transparência e acompanhamento da atuação dos senadores brasileiros.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <div className="flex min-h-screen flex-col">
          <Header />
          <main className="flex-1">{children}</main>
          <Footer />
        </div>
      </body>
    </html>
  );
}
