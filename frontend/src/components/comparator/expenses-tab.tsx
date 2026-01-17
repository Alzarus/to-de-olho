"use client";

import { useQueries } from "@tanstack/react-query";
import { getSenadorScore, getDespesasAgregado, getDespesas } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  LineChart,
  Line,
  Cell,
} from "recharts";
import { useTheme } from "next-themes";
import { formatCurrency } from "@/lib/utils";


interface ExpensesTabProps {
  selectedIds: number[];
}

const COLORS = [
  "#3b82f6", // Blue
  "#22c55e", // Green
  "#eab308", // Yellow
  "#ef4444", // Red
  "#a855f7", // Purple
];

export function ExpensesTab({ selectedIds }: ExpensesTabProps) {
  const { theme } = useTheme();
  const isDark = theme === "dark";

  // 1. Fetch Scores (for Totals vs Teto)
  const scoreQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-score", id],
      queryFn: () => getSenadorScore(id),
    })),
  });

  // 2. Fetch Aggregated (for Categories)
  const aggregatedQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-despesas-agregado", id],
      queryFn: () => getDespesasAgregado(id),
    })),
  });

  // 3. Fetch Detailed (for Evolution)
  const detailedQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-despesas", id],
      queryFn: () => getDespesas(id),
    })),
  });

  const isLoading = 
    scoreQueries.some(q => q.isLoading) || 
    aggregatedQueries.some(q => q.isLoading) || 
    detailedQueries.some(q => q.isLoading);

  const hasError = scoreQueries.every(q => q.isError) && aggregatedQueries.every(q => q.isError);

  if (isLoading) {
    return <Skeleton className="h-[600px] w-full rounded-lg" />;
  }

  if (hasError) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>
          Não foi possível carregar os dados de despesas.
        </AlertDescription>
      </Alert>
    );
  }

  // --- Process Data ---

  // 1. Total vs Cap
  const totalVsCapData = selectedIds.map((id, index) => {
      const q = scoreQueries[index];
      if (!q.data) return null;
      return {
          name: q.data.nome,
          Gasto: q.data.detalhes?.gasto_ceaps || 0,
          Teto: q.data.detalhes?.teto_ceaps || 0,
          color: COLORS[index % COLORS.length]
      };
  }).filter(Boolean);

  // 2. Categories
  const allCategories = new Set<string>();
  aggregatedQueries.forEach(q => {
      if (q.data?.por_tipo) {
          q.data.por_tipo.forEach((t: { tipo: string }) => allCategories.add(t.tipo));
      }
  });

  const categoryVolumes = Array.from(allCategories).map(cat => {
      const total = aggregatedQueries.reduce((acc, q) => {
          const catData = q.data?.por_tipo?.find((t: { tipo: string }) => t.tipo === cat);
          return acc + (catData?.total || 0);
      }, 0);
      return { category: cat, total };
  }).sort((a, b) => b.total - a.total).slice(0, 5);

  const topCategories = categoryVolumes.map(c => c.category);

   
  const categoryData = topCategories.map((cat: string) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const item: any = { category: cat };
      selectedIds.forEach((id, index) => {
          const q = aggregatedQueries[index];
          const scoreQ = scoreQueries[index];
          const name = scoreQ.data?.nome || `Senador ${id}`;
          
          if (q.data?.por_tipo) {
              const catData = q.data.por_tipo.find((t: { tipo: string }) => t.tipo === cat);
              item[name] = catData?.total || 0;
          } else {
              item[name] = 0;
          }
      });
      return item;
  });

  // 3. Evolution
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const evolutionMap = new Map<string, any>(); 

  detailedQueries.forEach((q, index) => {
      if (!q.data?.despesas) return;
      const scoreQ = scoreQueries[index];
      const name = scoreQ.data?.nome || `Senador ${selectedIds[index]}`;

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      q.data.despesas.forEach((d: any) => {
          const key = `${d.ano}-${d.mes.toString().padStart(2, '0')}`;
          if (!evolutionMap.has(key)) {
              evolutionMap.set(key, { 
                  date: `${d.mes.toString().padStart(2, '0')}/${d.ano}`,
                  sortKey: key 
              });
          }
          const item = evolutionMap.get(key);
          item[name] = (item[name] || 0) + d.valor_reembolsado;
      });
  });

  const evolutionData = Array.from(evolutionMap.values())
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .sort((a: any, b: any) => a.sortKey.localeCompare(b.sortKey))
    .slice(-12);

  return (
    <div className="space-y-8">
      {/* Total vs Cap */}
      <Card>
        <CardHeader>
          <CardTitle>Uso da Cota (Gasto vs Teto - 2024)</CardTitle>
        </CardHeader>
        <CardContent className="h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              data={totalVsCapData as any[]}
              layout="vertical"
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} stroke={isDark ? "#374151" : "#e5e7eb"} />
              <XAxis type="number" tickFormatter={(val) => `R$ ${(val/1000).toFixed(0)}k`} tick={{ fill: isDark ? "#9ca3af" : "#4b5563" }} />
              <YAxis dataKey="name" type="category" width={150} tick={{ fill: isDark ? "#9ca3af" : "#4b5563" }} />
              <Tooltip 
                 // eslint-disable-next-line @typescript-eslint/no-explicit-any
                 formatter={(value: any) => formatCurrency(value)}
                 contentStyle={{ 
                    backgroundColor: isDark ? "#1f2937" : "#ffffff",
                    borderColor: isDark ? "#374151" : "#e5e7eb",
                    color: isDark ? "#f3f4f6" : "#111827"
                }}
              />
              <Legend />
              <Bar dataKey="Gasto" fill="#8884d8" name="Gasto Total">
                 {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                 {(totalVsCapData as any[]).map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                 ))}
              </Bar>
              <Bar dataKey="Teto" fill="#82ca9d" name="Teto Disponível" opacity={0.3} />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Top Categories */}
      <Card>
        <CardHeader>
          <CardTitle>Top 5 Categorias de Despesa</CardTitle>
        </CardHeader>
        <CardContent className="h-[400px]">
           <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={categoryData}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={isDark ? "#374151" : "#e5e7eb"} />
              <XAxis dataKey="category" tick={{ fill: isDark ? "#9ca3af" : "#4b5563", fontSize: 10 }} interval={0} />
              <YAxis tickFormatter={(val) => `R$ ${(val/1000).toFixed(0)}k`} tick={{ fill: isDark ? "#9ca3af" : "#4b5563" }} />
              <Tooltip 
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                formatter={(value: any) => formatCurrency(value)}
                contentStyle={{ 
                    backgroundColor: isDark ? "#1f2937" : "#ffffff",
                    borderColor: isDark ? "#374151" : "#e5e7eb",
                    color: isDark ? "#f3f4f6" : "#111827"
                }}
              />
              <Legend />
              {selectedIds.map((id, index) => {
                  const scoreQ = scoreQueries[index];
                  const name = scoreQ.data?.nome || `Senador ${id}`;
                  return (
                    <Bar key={id} dataKey={name} fill={COLORS[index % COLORS.length]} />
                  );
              })}
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Evolution Line Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Evolução de Gastos (Últimos 12 meses)</CardTitle>
        </CardHeader>
        <CardContent className="h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={evolutionData}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={isDark ? "#374151" : "#e5e7eb"} />
              <XAxis dataKey="date" tick={{ fill: isDark ? "#9ca3af" : "#4b5563" }} />
              <YAxis tickFormatter={(val) => `R$ ${(val/1000).toFixed(0)}k`} tick={{ fill: isDark ? "#9ca3af" : "#4b5563" }} />
              <Tooltip 
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                formatter={(value: any) => formatCurrency(value)}
                contentStyle={{ 
                    backgroundColor: isDark ? "#1f2937" : "#ffffff",
                    borderColor: isDark ? "#374151" : "#e5e7eb",
                    color: isDark ? "#f3f4f6" : "#111827"
                }}
              />
              <Legend />
              {selectedIds.map((id, index) => {
                   const scoreQ = scoreQueries[index];
                   const name = scoreQ.data?.nome || `Senador ${id}`;
                   return (
                      <Line 
                        key={id} 
                        type="monotone" 
                        dataKey={name} 
                        stroke={COLORS[index % COLORS.length]} 
                        strokeWidth={2}
                        dot={{ r: 4 }}
                      />
                   );
              })}
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  );
}
