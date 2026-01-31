import type { NextConfig } from "next";

const FRONTEND_URL = "https://todeolho-frontend-7ynjppp3aq-rj.a.run.app";
const BACKEND_URL =
  process.env.BACKEND_URL || "https://todeolho-backend-7ynjppp3aq-rj.a.run.app";

const nextConfig: NextConfig = {
  output: "standalone",
  // Asset Prefix para garantir carregamento de CSS/JS via Cloud Run direto
  assetPrefix: process.env.NODE_ENV === "production" ? FRONTEND_URL : undefined,
  reactCompiler: false,
  experimental: {
    workerThreads: false,
    cpus: 1,
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
      {
        source: "/:path*",
        headers: [
          { key: "Access-Control-Allow-Origin", value: "*" },
          {
            key: "Access-Control-Allow-Methods",
            value: "GET,OPTIONS,PATCH,DELETE,POST,PUT",
          },
          {
            key: "Access-Control-Allow-Headers",
            value:
              "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version",
          },
        ],
      },
      {
        source: "/(api|static|favicon.ico|_next/static).*",
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
