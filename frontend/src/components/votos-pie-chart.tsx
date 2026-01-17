"use client";

import { useMemo } from "react";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from "recharts";
import { VotosPorTipo } from "@/types/api";

const COLORS: Record<string, string> = {
  Sim: "#22c55e",      // green-500
  Nao: "#ef4444",      // red-500
  Abstencao: "#64748b", // slate-500
  Obstrucao: "#f97316", // orange-500
  Outros: "#a1a1aa",   // zinc-400
  NCom: "#e5e7eb",     // gray-200
};

interface VotosPieChartProps {
  data: VotosPorTipo[];
}

export function VotosPieChart({ data }: VotosPieChartProps) {
  const chartData = useMemo(() => {
    return data.map((item) => ({
      name: item.voto,
      value: item.total,
      color: COLORS[item.voto] || COLORS.Outros,
    })).filter(item => item.value > 0);
  }, [data]);

  if (chartData.length === 0) {
    return (
      <div className="flex h-[300px] items-center justify-center text-muted-foreground">
        Sem dados de votação para exibir gráfico.
      </div>
    );
  }

  return (
    <div className="h-[300px] w-full">
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
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color} strokeWidth={0} />
            ))}
          </Pie>
          <Tooltip 
             formatter={(value: any) => [value, "Votos"]}
             contentStyle={{ borderRadius: "8px", border: "None", boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)" }}
          />
          <Legend verticalAlign="bottom" height={36} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
