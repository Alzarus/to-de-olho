import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { Toaster } from "sonner";
import "./globals.css";
import { Header, Footer } from "@/components/layout";
import { ComparatorDock } from "@/components/comparator/comparator-dock";
import { Providers } from "@/lib/providers";

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
      <body className={`${inter.variable} font-sans antialiased`} suppressHydrationWarning>
        <Providers>
          <div className="flex min-h-screen flex-col overflow-x-hidden">
            <Header />
            <main className="flex-1">{children}</main>
            <ComparatorDock />
            <Footer />
            <Toaster richColors position="top-right" />
          </div>
        </Providers>
      </body>
    </html>
  );
}
