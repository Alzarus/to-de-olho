import { MetadataRoute } from "next";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = process.env.FRONTEND_URL || "https://todeolho.org";

  // Rotas estáticas
  const routes = [
    "",
    "/ranking",
    "/emendas",
    "/comparar",
    "/metodologia",
    "/votacoes",
  ].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: new Date(),
    changeFrequency: "daily" as const,
    priority: route === "" ? 1 : 0.8,
  }));

  // Buscar senadores para rotas dinâmicas
  const backendUrl = process.env.BACKEND_URL || "http://localhost:8080";
  try {
    const res = await fetch(`${backendUrl}/api/v1/senadores`);
    const data = await res.json();
    const senadores = data.senadores || [];

    const senatorRoutes = senadores.map((s: any) => ({
      url: `${baseUrl}/senador/${s.id}`,
      lastModified: new Date(),
      changeFrequency: "weekly" as const,
      priority: 0.6,
    }));
    return [...routes, ...senatorRoutes];
  } catch (e) {
    console.error("Erro Sitemap: ", e);
    return routes;
  }
}
