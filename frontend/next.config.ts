import type { NextConfig } from "next";

const BACKEND_URL =
  process.env.BACKEND_URL ||
  "https://todeolho-backend-7ynjjppp3aq-rj.a.run.app";

const nextConfig: NextConfig = {
  output: "standalone",
  reactCompiler: true,
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND_URL}/:path*`,
      },
    ];
  },
};

export default nextConfig;
