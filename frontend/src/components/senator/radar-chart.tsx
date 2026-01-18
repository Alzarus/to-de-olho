"use client";

import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  Tooltip
} from "recharts";
import { useIsMobile } from "@/hooks/use-mobile";
import { useTheme } from "next-themes";
import type { SenadorScore } from "@/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface SenatorRadarChartProps {
  score: SenadorScore;
}

export function SenatorRadarChart({ score }: SenatorRadarChartProps) {
  const { theme } = useTheme();
  const isDark = theme === "dark";

  const data = [
    { subject: "Produtividade", value: score.produtividade, fullMark: 100 },
    { subject: "Presença", value: score.presenca, fullMark: 100 },
    { subject: "Economia", value: score.economia_cota, fullMark: 100 },
    { subject: "Comissões", value: score.comissoes, fullMark: 100 },
  ];

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>Perfil de Atuação</CardTitle>
      </CardHeader>
      <CardContent className="flex-1 min-h-[300px]">
        <ResponsiveContainer width="100%" height="100%">
          <RadarChart cx="50%" cy="50%" outerRadius="55%" data={data}>
            <PolarGrid stroke={isDark ? "#374151" : "#e5e7eb"} />
            <PolarAngleAxis
              dataKey="subject"
              tick={{ fill: isDark ? "#9ca3af" : "#4b5563", fontSize: 12 }}
            />
            <PolarRadiusAxis
              angle={30}
              domain={[0, 100]}
              tick={false} 
              axisLine={false}
            />
            <Radar
              name={score.nome}
              dataKey="value"
              stroke="#d4af37"
              fill="#d4af37"
              fillOpacity={0.5}
              isAnimationActive={!useIsMobile()}
            />
            <Tooltip 
                cursor={false}
                contentStyle={{ 
                    backgroundColor: isDark ? "#1f2937" : "#ffffff",
                    borderColor: isDark ? "#374151" : "#e5e7eb",
                    borderRadius: "8px",
                    color: isDark ? "#f3f4f6" : "#111827"
                }}
             />
          </RadarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
