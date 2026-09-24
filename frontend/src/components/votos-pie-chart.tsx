
import { useMemo, useState } from "react";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip as RechartsTooltip, Legend } from "recharts";
import { Info } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import { VotosPorTipo } from "@/types/api";
import { ChartTooltipContent } from "@/components/ui/chart-tooltip";
import { cn } from "@/lib/utils";

// Cores atualizadas para melhor distinção e contraste
const COLORS: Record<string, string> = {
  Sim: "#16a34a",       // green-600
  Nao: "#dc2626",       // red-600
  Abstencao: "#f59e0b", // amber-500 (Amarelo/Laranja distinto)
  Obstrucao: "#9333ea", // purple-600 (Roxo para diferenciar)
  Outros: "#64748b",    // slate-500 (Cinza neutro)
};

const LABELS: Record<string, string> = {
  Sim: "Sim",
  Nao: "Não",
  Abstencao: "Abstenção",
  Obstrucao: "Obstrução",
  Outros: "Outros",
};

const MAIN_TYPES = ["Sim", "Nao", "Abstencao", "Obstrucao"];

interface VotosPieChartProps {
  data: VotosPorTipo[];
  onSliceClick?: (voteType: string) => void;
  /** Tipo de voto atualmente filtrado (para `aria-pressed` na legenda). */
  activeType?: string;
}

const VOTE_DESCRIPTIONS: Record<string, string> = {
  AP: "Atividade Parlamentar/Partidária",
  LP: "Licença Particular",
  LS: "Licença Saúde",
  LG: "Licença Gestante",
  LC: "Licença Conjunta",
  MIS: "Missão Oficial",
  NCom: "Não Compareceu",
  "P-NR": "Presidente (Não Votou)",
  "P-OD": "Presidente (Obstrução)",
};

export function VotosPieChart({ data, onSliceClick, activeType }: VotosPieChartProps) {
  const [activeIndex, setActiveIndex] = useState<number | undefined>(undefined);
  const isMobile = useIsMobile();

  const { chartData, outrosDetails } = useMemo(() => {
    const mainItems: { name: string; value: number; color: string; label: string; isGroup: boolean; breakdown?: string }[] = [];
    let outrosTotal = 0;
    const outrosTypes = new Set<string>();

    // Ordenar para consistência visual (Sim, Não, ... resto)
    const sortedData = [...data].sort((a, b) => {
      const order = ["Sim", "Nao", "Abstencao", "Obstrucao"];
      const idxA = order.indexOf(a.voto);
      const idxB = order.indexOf(b.voto);
      if (idxA !== -1 && idxB !== -1) return idxA - idxB;
      if (idxA !== -1) return -1;
      if (idxB !== -1) return 1;
      return b.total - a.total;
    });

    sortedData.forEach((item) => {
      if (MAIN_TYPES.includes(item.voto)) {
        mainItems.push({
          name: item.voto,
          value: item.total,
          color: COLORS[item.voto],
          label: LABELS[item.voto],
          isGroup: false
        });
      } else {
        outrosTotal += item.total;
        outrosTypes.add(item.voto);
      }
    });

    const outrosBreakdown = Array.from(outrosTypes).map(type => {
        const desc = VOTE_DESCRIPTIONS[type] || "Outros";
        return `${type}: ${desc}`;
    });

    if (outrosTotal > 0) {
      mainItems.push({
        name: "Outros",
        value: outrosTotal,
        color: COLORS.Outros,
        label: "Outros",
        isGroup: true,
        breakdown: outrosBreakdown.join(" | ")
      });
    }

    return { chartData: mainItems, outrosDetails: outrosBreakdown.join("\n") };
  }, [data]);

  const totalVotos = useMemo(() => {
    return chartData.reduce((acc, item) => acc + item.value, 0);
  }, [chartData]);

  if (chartData.length === 0) {
    return (
      <div className="flex h-[300px] items-center justify-center text-muted-foreground" role="status">
        Sem dados de votação para exibir gráfico.
      </div>
    );
  }

  const handlePieEnter = (_: unknown, index: number) => {
    setActiveIndex(index);
  };

  const handlePieLeave = () => {
    setActiveIndex(undefined);
  };

  const handleClick = (entry: { name: string }) => {
    if (onSliceClick) {
      // Se for grupo "Outros", passamos "Outros" para filtrar todos os tipos mapeados como Outros
      // O componente pai precisará saber lidar com "Outros" ou passamos null para limpar
      onSliceClick(entry.name);
    }
  };

  return (
    <div className="space-y-4 w-full min-w-0 overflow-hidden">
      <div 
        className="h-[250px] sm:h-[350px] w-full min-w-0 relative" 
        role="img" 
        aria-label={`Gráfico de distribuição de votos. Total: ${totalVotos}.`}
      >
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="45%"
              innerRadius={isMobile ? 60 : 80}
              outerRadius={isMobile ? 100 : 130}
              paddingAngle={2}
              dataKey="value"
              onMouseEnter={handlePieEnter}
              onMouseLeave={handlePieLeave}
              onClick={(_, index) => handleClick(chartData[index])}
              style={{ cursor: onSliceClick ? "pointer" : "default" }}
              isAnimationActive={!isMobile}
            >
              {chartData.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={entry.color} 
                  strokeWidth={activeIndex === index ? 3 : 0}
                  stroke={activeIndex === index ? "var(--foreground)" : undefined}
                  style={{
                    filter: activeIndex === index ? "brightness(1.1)" : undefined,
                    transition: "all 0.2s ease-in-out"
                  }}
                />
              ))}
            </Pie>
            <RechartsTooltip
              content={({ active, payload }) => (
                <ChartTooltipContent
                  active={active}
                  payload={payload}
                  hideLabel
                  nameFormatter={(name, entry) => entry.payload?.label ?? String(name)}
                  valueFormatter={(value) => `${value} votos`}
                  colorFormatter={(entry) => entry.payload?.color}
                  extra={(items) => {
                    const item = items[0]?.payload;
                    if (!item?.isGroup || !item.breakdown) return null;
                    return (
                      <>
                        <p className="mb-1 font-semibold text-popover-foreground">Composição:</p>
                        <p>{item.breakdown}</p>
                      </>
                    );
                  }}
                />
              )}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <ul className="flex flex-wrap justify-center gap-x-4 gap-y-2 pt-2 w-full px-2">
        {chartData.map((entry, index) => {
          const value = entry.name;
          const label = entry.label || value;
          const isActive = activeType === value;
          const swatch = (
            <span
              aria-hidden="true"
              className="w-3 h-3 block flex-shrink-0"
              style={{ backgroundColor: entry.color }}
            />
          );

          return (
            <li key={`item-${index}`} className="flex items-center gap-1 text-sm">
              {onSliceClick ? (
                <button
                  type="button"
                  onClick={() => handleClick(entry)}
                  aria-pressed={isActive}
                  aria-label={`Filtrar votos: ${label} (${entry.value})`}
                  className={cn(
                    "inline-flex items-center gap-1.5 rounded-sm px-1 py-0.5 transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
                    isActive ? "font-semibold text-foreground underline underline-offset-4" : "text-muted-foreground",
                  )}
                >
                  {swatch}
                  {label}
                </button>
              ) : (
                <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                  {swatch}
                  {label}
                </span>
              )}
              {value === "Outros" && outrosDetails && (
                <TooltipProvider delayDuration={0}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        type="button"
                        aria-label="Composição de Outros"
                        className="inline-flex rounded-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-help transition-colors"
                      >
                        <Info className="h-3.5 w-3.5" aria-hidden="true" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent side="top" className="max-w-[200px]">
                      <p className="text-xs font-semibold mb-1">Composição:</p>
                      <p className="text-xs whitespace-pre-line">
                        {outrosDetails}
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </li>
          );
        })}
      </ul>

      {onSliceClick && (
        <p className="text-center text-xs text-muted-foreground">
          Clique nas fatias ou na legenda para filtrar a lista abaixo
        </p>
      )}
    </div>
  );
}
