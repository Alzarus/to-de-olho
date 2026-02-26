import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { Toaster } from "sonner";
import "./globals.css";
import { Header, Footer } from "@/components/layout";
import { ComparatorDock } from "@/components/comparator/comparator-dock";
import { Providers } from "@/lib/providers";
import { GoogleAnalytics } from "@next/third-parties/google";

const inter = Inter({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL(process.env.FRONTEND_URL || "https://todeolho.org"),
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
    "fiscalização",
    "emendas parlamentares",
  ],
  authors: [{ name: "Pedro Eli" }],
  openGraph: {
    type: "website",
    locale: "pt_BR",
    siteName: "Tô De Olho",
    title: "Tô De Olho - Transparência no Senado",
    description:
      "Descubra quem são os senadores mais eficientes do Brasil. Ranking completo, gastos e emendas detalhadas.",
    images: [
      {
        url: "/logo.png",
        width: 1200,
        height: 630,
        alt: "Logo Tô De Olho",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Tô De Olho - Transparência no Senado",
    description: "Plataforma de transparência e acompanhamento da atuação dos senadores brasileiros.",
    images: ["/logo.png"],
  },
  robots: {
    index: true,
    follow: true,
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
          <div className="flex min-h-screen flex-col overflow-x-hidden pt-16">
            <Header />
            <main className="flex-1 min-w-0 w-full overflow-x-hidden">{children}</main>
            <ComparatorDock />
            <Footer />
            <Toaster richColors position="top-right" />
          </div>
        </Providers>
        {process.env.NEXT_PUBLIC_GA_ID && (
          <GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GA_ID} />
        )}
      </body>
    </html>
  );
}
