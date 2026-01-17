"use client";

import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  Legend,
  Tooltip
} from "recharts";
import { useTheme } from "next-themes";
import type { SenadorScore } from "@/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface ComparatorRadarChartProps {
  senators: (SenadorScore & { color: string })[];
  year: number;
}

export function ComparatorRadarChart({ senators, year }: ComparatorRadarChartProps) {
  const { theme } = useTheme();
  const isDark = theme === "dark";
  const yearLabel = year === 0 ? "Mandato Completo" : year.toString();

  // Transform data for Recharts
  // We need an array of objects like { subject: 'Produtividade', A: 120, B: 110, fullMark: 150 }
  const data = [
    {
      subject: "Produtividade",
      fullMark: 100,
    },
    {
      subject: "Presença",
      fullMark: 100,
    },
    {
      subject: "Economia",
      fullMark: 100,
    },
    {
      subject: "Comissões",
      fullMark: 100,
    },
  ];

  // Populate data with senator values
  const chartData = data.map((item) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const newItem: any = { ...item };
    senators.forEach((senator) => {
      let value = 0;
      switch (item.subject) {
        case "Produtividade":
          value = senator.produtividade;
          break;
        case "Presença":
          value = senator.presenca;
          break;
        case "Economia":
          value = senator.economia_cota;
          break;
        case "Comissões":
          value = senator.comissoes;
          break;
      }
      newItem[senator.nome] = value;
    });
    return newItem;
  });

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Comparativo de Desempenho - {yearLabel}</CardTitle>
      </CardHeader>
      <CardContent className="h-[450px]">
        <ResponsiveContainer width="100%" height="100%">
          <RadarChart cx="50%" cy="45%" outerRadius="60%" data={chartData}>
            <PolarGrid stroke={isDark ? "#374151" : "#e5e7eb"} />
            <PolarAngleAxis
              dataKey="subject"
              tick={{ fill: isDark ? "#9ca3af" : "#4b5563", fontSize: 12 }}
            />
            <PolarRadiusAxis
              angle={30}
              domain={[0, 100]}
              tick={{ fill: isDark ? "#9ca3af" : "#4b5563", fontSize: 10 }}
            />
            {senators.map((senator) => (
              <Radar
                key={senator.senador_id}
                name={senator.nome}
                dataKey={senator.nome}
                stroke={senator.color}
                fill={senator.color}
                fillOpacity={0.3}
              />
            ))}
            <Legend 
              wrapperStyle={{ paddingTop: "30px" }} 
              iconType="square"
              iconSize={10}
            />
             <Tooltip 
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
