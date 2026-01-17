"use client";

import { useQueries } from "@tanstack/react-query";
import { getEmendas } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend, CartesianGrid } from "recharts";
import { formatCurrency } from "@/lib/utils";
import { SenatorBasicProfile } from "@/contexts/comparator-context";

// As cores podem vir do perfil se tiver, ou fixed.
const COLORS = [
  "#3b82f6",
  "#22c55e", 
  "#eab308", 
  "#ef4444", 
  "#a855f7",
];

interface EmendasTabProps {
    senators: SenatorBasicProfile[];
    year: number;
}

export function EmendasTab({ senators, year }: EmendasTabProps) {
    const queries = useQueries({
        queries: senators.map(s => ({
            queryKey: ["senador-emendas", s.id, year],
            queryFn: () => getEmendas(s.id, year),
        }))
    });

    const isLoading = queries.some(q => q.isLoading);
    if (isLoading) return <Skeleton className="h-[500px] w-full" />;

    const chartData = senators.map((s, index) => {
        const query = queries[index];
        const emendasData = query.data;
        const resumo = emendasData?.resumo;
        
        return {
            name: s.nome,
            color: COLORS[index % COLORS.length], // s.color se existisse
            Pago: resumo?.total_pago || 0,
            Empenhado: resumo?.total_empenhado || 0,
            Pix: emendasData?.emendas
                ? emendasData.emendas
                    .filter(e => isEmendaEspecial(e.tipo))
                    .reduce((acc, curr) => acc + curr.valor_pago, 0)
                : 0
        };
    });

    return (
      <div className="space-y-8 animate-in fade-in duration-500">
          <Card>
              <CardHeader>
                  <CardTitle>Execução de Emendas ({year})</CardTitle>
                  <CardDescription>
                    Comparativo de valores pagos e emendas especiais (PIX).
                  </CardDescription>
              </CardHeader>
              <CardContent>
                  <div className="h-[500px] w-full">
                      <ResponsiveContainer width="100%" height="100%">
                          <BarChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                              <CartesianGrid strokeDasharray="3 3" vertical={false} />
                              <XAxis dataKey="name" />
                              <YAxis tickFormatter={(val) => `R$ ${(val/1000000).toFixed(1)}M`} />
                                                            <Tooltip
                                                                formatter={(value: number | string) => {
                                                                    const numericValue = typeof value === "number" ? value : Number(value);
                                                                    return formatCurrency(Number.isFinite(numericValue) ? numericValue : 0);
                                                                }}
                                                                cursor={{ fill: "transparent" }}
                                                            />
                              <Legend wrapperStyle={{ paddingTop: "20px" }} />
                              <Bar dataKey="Empenhado" fill="#94a3b8" name="Total Empenhado" radius={[4, 4, 0, 0]} />
                              <Bar dataKey="Pago" fill="#2563eb" name="Total Pago" radius={[4, 4, 0, 0]} />
                              <Bar dataKey="Pix" fill="#ea580c" name="Emendas PIX (Pago)" radius={[4, 4, 0, 0]} />
                          </BarChart>
                      </ResponsiveContainer>
                  </div>
              </CardContent>
          </Card>
      </div>
    );
}

function isEmendaEspecial(tipo: string): boolean {
    const normalizado = tipo.toLowerCase();
    // Buscar por "especia" para capturar tanto "especial" quanto "especiais"
    return normalizado.includes("especia") || normalizado.includes("pix");
}
