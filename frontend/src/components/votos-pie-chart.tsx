
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

export function VotosPieChart({ data, onSliceClick }: VotosPieChartProps) {
  const [activeIndex, setActiveIndex] = useState<number | undefined>(undefined);

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
    <div className="space-y-4">
      <div 
        className="h-[300px] w-full" 
        role="img" 
        aria-label={`Gráfico de distribuição de votos. Total: ${totalVotos}.`}
      >
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={80}
              paddingAngle={2}
              dataKey="value"
              onMouseEnter={handlePieEnter}
              onMouseLeave={handlePieLeave}
              onClick={(_, index) => handleClick(chartData[index])}
              style={{ cursor: onSliceClick ? "pointer" : "default" }}
              isAnimationActive={!useIsMobile()}
            >
              {chartData.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={entry.color} 
                  strokeWidth={activeIndex === index ? 3 : 0}
                  stroke={activeIndex === index ? "#1e293b" : undefined}
                  style={{
                    filter: activeIndex === index ? "brightness(1.1)" : undefined,
                    transition: "all 0.2s ease-in-out"
                  }}
                />
              ))}
            </Pie>
            <RechartsTooltip 
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              formatter={(value: any, name: any, props: any) => {
                const item = props.payload;
                if (item.isGroup) {
                  return [
                    <div key="tooltip-content" className="flex flex-col gap-1">
                      <span>{value} votos</span>
                      <div className="border-t pt-2 mt-1">
                        <p className="text-xs font-semibold mb-1 text-foreground">Composição:</p>
                        <p className="text-xs text-muted-foreground font-normal">
                          {item.breakdown}
                        </p>
                      </div>
                    </div>,
                    item.label
                  ];
                }
                return [value, item.label];
              }}
              contentStyle={{ 
                borderRadius: "8px", 
                border: "none", 
                boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
                padding: "12px",
                backgroundColor: "rgba(255, 255, 255, 0.98)"
              }}
            />
            <Legend 
              verticalAlign="bottom" 
              height={36}
              wrapperStyle={{ paddingTop: "20px" }}
              formatter={(value: string) => {
                const item = chartData.find(d => d.name === value);
                const label = item?.label || value;
                
                if (value === "Outros" && outrosDetails) {
                   return (
                    <span className="inline-flex items-center gap-1">
                      {label}
                      <TooltipProvider delayDuration={0}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help hover:text-foreground transition-colors" />
                          </TooltipTrigger>
                          <TooltipContent side="top" className="max-w-[200px]">
                            <p className="text-xs font-semibold mb-1">Composição:</p>
                            <p className="text-xs opacity-90 whitespace-pre-line">
                              {outrosDetails}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </span>
                  );
                }
                return label;
              }}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>

      {onSliceClick && (
        <p className="text-center text-xs text-muted-foreground">
          Clique nas fatias para filtrar a lista abaixo
        </p>
      )}
    </div>
  );
}
