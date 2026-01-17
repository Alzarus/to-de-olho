"use client";

import { useQueries } from "@tanstack/react-query";
import { getSenadorScore, getDespesas } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, Building2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";


interface SuppliersTabProps {
  selectedIds: number[];
}

const COLORS = [
  "#3b82f6", // Blue
  "#22c55e", // Green
  "#eab308", // Yellow
  "#ef4444", // Red
  "#a855f7", // Purple
];

export function SuppliersTab({ selectedIds }: SuppliersTabProps) {
  // Fetch Scores (for names)
  const scoreQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-score", id],
      queryFn: () => getSenadorScore(id),
    })),
  });

  // Fetch Detailed Expenses (contains supplier info)
  const detailedQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-despesas", id],
      queryFn: () => getDespesas(id),
    })),
  });

  const isLoading = 
    scoreQueries.some(q => q.isLoading) || 
    detailedQueries.some(q => q.isLoading);

  const hasError = detailedQueries.every(q => q.isError);

  if (isLoading) {
    return <Skeleton className="h-[600px] w-full rounded-lg" />;
  }

  if (hasError) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>
          Não foi possível carregar os dados de fornecedores.
        </AlertDescription>
      </Alert>
    );
  }

  // Process Data
  const senatorSuppliers = selectedIds.map((id, index) => {
    const q = detailedQueries[index];
    const scoreQ = scoreQueries[index];
    const name = scoreQ.data?.nome || `Senador ${id}`;
    const color = COLORS[index % COLORS.length];

    if (!q.data?.despesas) return { id, name, color, suppliers: new Map() };

    // Aggregate by supplier
    const suppliers = new Map<string, number>();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    q.data.despesas.forEach((d: any) => {
        const fornecedor = d.fornecedor || "NÃO INFORMADO";
        suppliers.set(fornecedor, (suppliers.get(fornecedor) || 0) + d.valor_reembolsado);
    });

    return { id, name, color, suppliers };
  });

  // 1. Common Suppliers
  const allSuppliers = new Set<string>();
  senatorSuppliers.forEach(s => {
      Array.from(s.suppliers.keys()).forEach(k => allSuppliers.add(k));
  });

  const commonSuppliers = Array.from(allSuppliers).filter(supplier => {
      // Must appear in at least 2 senators
      if (selectedIds.length < 2) return false;
      const count = senatorSuppliers.filter(s => s.suppliers.has(supplier)).length;
      return count >= 2;
  }).map(supplier => {
      const total = senatorSuppliers.reduce((acc, s) => acc + (s.suppliers.get(supplier) || 0), 0);
      return { name: supplier, total };
  }).sort((a, b) => b.total - a.total).slice(0, 10); // Top 10 common

  // 2. Top Suppliers per Senator
  const topSuppliersPerSenator = senatorSuppliers.map(s => {
      const top = Array.from(s.suppliers.entries())
          .map(([name, total]) => ({ name, total }))
          .sort((a, b) => b.total - a.total)
          .slice(0, 5);
      return { ...s, top };
  });

  return (
    <div className="space-y-8">
        {/* Common Suppliers */}
        {selectedIds.length > 1 && commonSuppliers.length > 0 && (
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Building2 className="h-5 w-5" />
                        Fornecedores em Comum
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        {commonSuppliers.map((supplier, idx) => (
                            <div key={idx} className="flex flex-col sm:flex-row sm:items-center justify-between border-b pb-2 last:border-0">
                                <span className="font-medium text-sm sm:text-base">{supplier.name}</span>
                                <div className="flex items-center gap-4 mt-2 sm:mt-0">
                                    <div className="flex -space-x-2">
                                        {senatorSuppliers.filter(s => s.suppliers.has(supplier.name)).map(s => (
                                            <div key={s.id} className="h-6 w-6 rounded-full border-2 border-background flex items-center justify-center text-[10px] text-white font-bold" style={{ backgroundColor: s.color }} title={s.name}>
                                                {s.name.charAt(0)}
                                            </div>
                                        ))}
                                    </div>
                                    <span className="font-bold text-sm">{formatCurrency(supplier.total)}</span>
                                </div>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        )}

        {/* Top Suppliers Grid */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {topSuppliersPerSenator.map((senator) => (
                <Card key={senator.id} className="overflow-hidden">
                    <div className="h-2 w-full" style={{ backgroundColor: senator.color }} />
                    <CardHeader>
                        <CardTitle className="text-base truncate" title={senator.name}>{senator.name}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3">
                            {senator.top.map((supplier, idx) => (
                                <div key={idx} className="flex justify-between items-start text-sm">
                                    <span className="text-muted-foreground line-clamp-1 w-[60%]" title={supplier.name}>{supplier.name}</span>
                                    <span className="font-medium">{formatCurrency(supplier.total)}</span>
                                </div>
                            ))}
                            {senator.top.length === 0 && (
                                <span className="text-muted-foreground text-sm">Nenhum dado encontrado.</span>
                            )}
                        </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    </div>
  );
}
