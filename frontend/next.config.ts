import type { NextConfig } from "next";

const FRONTEND_URL = process.env.FRONTEND_URL;
const BACKEND_URL =
  process.env.BACKEND_URL || "http://api:8080";

const isDev = process.env.NODE_ENV !== "production";
// O Google Analytics só entra na página quando NEXT_PUBLIC_GA_ID existe no
// build (ver app/layout.tsx); a CSP acompanha essa mesma condição para não
// liberar origens do Google quando o script nem é carregado.
const GA_ATIVO = Boolean(process.env.NEXT_PUBLIC_GA_ID);
const GA_SCRIPT = GA_ATIVO ? ["https://www.googletagmanager.com"] : [];
const GA_CONEXAO = GA_ATIVO
  ? [
      "https://www.googletagmanager.com",
      "https://*.google-analytics.com",
      "https://*.analytics.google.com",
    ]
  : [];
// Com assetPrefix, JS/CSS/fontes de /_next vêm dessa origem em vez da própria.
const ORIGEM_ASSETS =
  !isDev && FRONTEND_URL ? [new URL(FRONTEND_URL).origin] : [];

// Content-Security-Policy estática, sem nonce: nonce exige renderização
// dinâmica em toda requisição e acabaria com o HTML pré-renderizado (SSG/ISR)
// que queremos cachear na borda. O custo é aceitar 'unsafe-inline' em dois
// pontos, ambos exigidos pelo próprio Next/React:
// - script-src: o App Router injeta <script> inline com o payload RSC
//   (self.__next_f.push) e o next-themes injeta o script que aplica o tema
//   antes da pintura (evita o flash claro/escuro);
// - style-src: Recharts, Radix e framer-motion escrevem atributos style=""
//   (posição de tooltip, animações, variáveis CSS), e o next/font injeta um
//   <style> inline.
// Em desenvolvimento o React precisa de eval para o overlay e o HMR.
const CSP = [
  ["default-src", "'self'"],
  [
    "script-src",
    "'self'",
    "'unsafe-inline'",
    ...(isDev ? ["'unsafe-eval'"] : []),
    ...ORIGEM_ASSETS,
    ...GA_SCRIPT,
  ],
  ["style-src", "'self'", "'unsafe-inline'", ...ORIGEM_ASSETS],
  // Fotos oficiais: a API grava http://www.senado.leg.br/senadores/img/...,
  // que redireciona (301) para legis.senado.leg.br; a CSP vale também para o
  // destino do redirecionamento, por isso os dois hosts. Ficam sem esquema de
  // propósito: numa página https equivalem a https:// (o navegador promove a
  // URL http da foto antes de buscar) e, no teste local em http, aceitam a
  // URL como a API a entrega. data:/blob: cobrem ícones embutidos do CSS.
  [
    "img-src",
    "'self'",
    "data:",
    "blob:",
    "www.senado.leg.br",
    "legis.senado.leg.br",
    ...ORIGEM_ASSETS,
    ...GA_CONEXAO,
  ],
  // Inter vem do next/font, que baixa a fonte no build e serve de /_next.
  ["font-src", "'self'", ...ORIGEM_ASSETS],
  // O cliente só fala com /api/* na mesma origem (rewrite para a API Go).
  ["connect-src", "'self'", ...GA_CONEXAO],
  ["frame-src", "'none'"],
  ["object-src", "'none'"],
  ["base-uri", "'self'"],
  ["form-action", "'self'"],
  ["frame-ancestors", "'self'"],
]
  .map((diretiva) => diretiva.join(" "))
  .join("; ");

const CABECALHOS_SEGURANCA = [
  { key: "Content-Security-Policy", value: CSP },
  { key: "X-Content-Type-Options", value: "nosniff" },
  // Redundante com frame-ancestors, mas cobre navegadores sem CSP nível 2.
  { key: "X-Frame-Options", value: "SAMEORIGIN" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  // O site não usa nenhum destes recursos; negar tudo impede que um script
  // injetado os peça em nome do domínio.
  {
    key: "Permissions-Policy",
    value:
      "accelerometer=(), browsing-topics=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",
  },
  // HSTS fica de fora de propósito: o Nginx Proxy Manager já envia
  // Strict-Transport-Security e repetir aqui geraria cabeçalho duplicado.
];

const nextConfig: NextConfig = {
  output: "standalone",
  // Não anunciar o framework em todas as respostas.
  poweredByHeader: false,
  // Asset Prefix para garantir carregamento de CSS/JS (somente se FRONTEND_URL for definido explicitamente)
  assetPrefix: process.env.NODE_ENV === "production" && FRONTEND_URL ? FRONTEND_URL : undefined,
  reactCompiler: false,
  experimental: {
    workerThreads: false,
    cpus: 1,
  },
  // Links antigos de votacao apontavam para a sessao inteira ("codigoSessao_ano").
  // Desde a v3 cada votacao tem id proprio; o link antigo vai para a lista das
  // votacoes daquela sessao (decisao D2 do PLANO-MIGRACAO.md).
  async redirects() {
    return [
      // www.todeolho.org chega ao Next pelo mesmo proxy e servia o site
      // inteiro, duplicando conteúdo para buscadores. Redirect de config (e
      // não middleware/proxy) porque é estático: não roda código por
      // requisição e não obriga as páginas a virar dinâmicas. O Next repassa
      // a query string ao destino.
      {
        source: "/:path*",
        has: [{ type: "host", value: "www.todeolho.org" }],
        destination: "https://todeolho.org/:path*",
        permanent: true,
      },
      {
        source: "/votacoes/:sessao([0-9]+)_:ano([0-9]+)",
        destination: "/votacoes?sessao=:sessao",
        permanent: true,
      },
    ];
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND_URL}/api/:path*`,
      },
    ];
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "www.senado.leg.br",
      },
    ],
    minimumCacheTTL: 60 * 60 * 24, // 24 horas
  },
  async headers() {
    return [
      // Sem CORS aqui: as páginas não são lidas por outras origens, e /api/*
      // é atendido pela API Go (no proxy de produção ou via rewrite), que
      // emite o próprio Access-Control-*. O bloco antigo liberava
      // POST/PUT/DELETE em todo o site e somava cabeçalhos aos da API.
      {
        source: "/:path*",
        headers: CABECALHOS_SEGURANCA,
      },
      {
        source: "/favicon.ico",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        // Cachear assets estáticos em /public (se houver folder static)
        source: "/static/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
