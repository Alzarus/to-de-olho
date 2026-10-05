import { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  const baseUrl = process.env.FRONTEND_URL || "https://todeolho.org";

  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // /_next/ fica liberado: bloquear o JS e o CSS impede o Google de
      // renderizar as páginas, que carregam os dados no navegador
      disallow: ["/api/"],
    },
    sitemap: `${baseUrl}/sitemap.xml`,
  };
}
