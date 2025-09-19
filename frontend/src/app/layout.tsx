import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Tô De Olho - Transparência Política",
  description: "Plataforma de transparência política que democratiza o acesso aos dados da Câmara dos Deputados",
  keywords: ["transparência", "política", "câmara", "deputados", "brasil"],
  authors: [{ name: "Pedro Batista de Almeida Filho" }],
  openGraph: {
    title: "Tô De Olho - Transparência Política", 
    description: "Democratizando o acesso aos dados da Câmara dos Deputados",
    type: "website",
    locale: "pt_BR",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-gray-50 text-gray-900`}
      >
        <main className="min-h-screen">
          {children}
        </main>
        
        <footer className="bg-gray-800 text-white py-8 mt-16">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <p className="text-sm">
              © 2025 Tô De Olho - TCC IFBA | Desenvolvido por Pedro Batista
            </p>
            <p className="text-xs text-gray-400 mt-2">
              Dados fornecidos pela API oficial da Câmara dos Deputados
            </p>
          </div>
        </footer>
      </body>
    </html>
  );
}
