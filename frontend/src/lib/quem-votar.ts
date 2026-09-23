// Quem Votar: consulta de candidaturas do TSE, servida em /quemvotar/ pelo
// mesmo proxy do todeolho.org (repositório quem-votar).
export const QUEM_VOTAR_URL = "/quemvotar/";

// Destaque de "Eleições 2026" até o dia seguinte ao 2º turno (25/10/2026).
// Depois disso o link continua, sem o selo.
const FIM_PERIODO_ELEITORAL = new Date("2026-10-26T03:00:00Z");

export function emPeriodoEleitoral(agora: Date = new Date()): boolean {
  return agora < FIM_PERIODO_ELEITORAL;
}
