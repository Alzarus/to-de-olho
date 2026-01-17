"use client";

import { useComparator } from "@/contexts/comparator-context";
import { Button } from "@/components/ui/button";
import { Trash2, Download, X as XIcon, ArrowRight } from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { OverviewTab } from "@/components/comparator/overview-tab";
import { ExpensesTab } from "@/components/comparator/expenses-tab";
import { SuppliersTab } from "@/components/comparator/suppliers-tab";
import { EmendasTab } from "@/components/comparator/emendas-tab";
import { SenatorSelector } from "@/components/comparator/senator-selector";

export default function ComparatorPage() {
  const { selectedSenators, clearSelection, removeSenator } = useComparator();
  const searchParams = useSearchParams();
  const year = Number(searchParams.get("ano")) || 2024;


  // Empty state - show the selector
  if (selectedSenators.length === 0) {
    return (
      <div className="container mx-auto max-w-7xl px-4 py-8 sm:py-12 sm:px-6 lg:px-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight mb-2">
            Comparador de Senadores
          </h1>
          <p className="text-muted-foreground max-w-2xl mx-auto">
            Selecione até 5 senadores para comparar lado a lado seus indicadores de
            desempenho, gastos e votações.
          </p>
        </div>
        
        <SenatorSelector />
      </div>
    );
  }

  // Placeholder for export function
  const handleExport = () => {
    console.log("Exporting comparison...");
  };

  return (
    <div className="container mx-auto max-w-7xl px-4 py-8 sm:py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">
            Comparação
          </h1>
          <p className="mt-2 text-muted-foreground">
            Análise comparativa detalhada de {selectedSenators.length} senador{selectedSenators.length !== 1 ? 'es' : ''}.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            Exportar
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={clearSelection}
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Limpar
          </Button>
        </div>
      </div>

      {/* Senators Header / Legend */}
      <div className="mb-8 overflow-x-auto pb-4">
        <div className="flex min-w-max gap-4 px-1">
          {selectedSenators.map((senator, index) => (
            <Card key={senator.id} className="relative w-56 shrink-0 transition-all hover:shadow-md">
              {/* Remove Button */}
              <button 
                onClick={() => removeSenator(senator.id)}
                className="absolute right-2 top-2 rounded-full p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-colors"
              >
                <XIcon className="h-4 w-4" />
              </button>
                
              <CardContent className="flex flex-col items-center p-4 text-center">
                <div className="relative mb-3">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={senator.fotoUrl}
                    alt={senator.nome}
                    className="h-20 w-20 rounded-full object-cover shadow-sm"
                  />
                  <div className={`absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold shadow-sm ${
                    index === 0 ? "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300" :
                    index === 1 ? "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300" :
                    index === 2 ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300" :
                    index === 3 ? "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300" :
                    "bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300"
                  }`}>
                    {index + 1}
                  </div>
                </div>
                
                <Link href={`/senador/${senator.id}`} className="hover:underline">
                  <h3 className="font-bold text-sm leading-tight mb-1 truncate w-full">
                    {senator.nome}
                  </h3>
                </Link>
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Badge variant="secondary" className="text-xs px-1.5 py-0">{senator.partido}</Badge>
                  <span>{senator.uf}</span>
                </div>
              </CardContent>
            </Card>
          ))}
            
          {/* Add more button */}
          {selectedSenators.length < 5 && (
            <Link 
              href="#add-senators"
              onClick={(e) => {
                e.preventDefault();
                const element = document.getElementById('add-senators');
                element?.scrollIntoView({ behavior: 'smooth' });
              }}
              className="flex w-40 shrink-0 flex-col items-center justify-center rounded-lg border-2 border-dashed border-muted p-4 text-center bg-muted/20 hover:bg-muted/40 transition-colors"
            >
              <ArrowRight className="h-6 w-6 text-muted-foreground mb-2" />
              <span className="text-xs font-medium text-muted-foreground">
                Adicionar mais ({5 - selectedSenators.length} restantes)
              </span>
            </Link>
          )}
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="w-full justify-start overflow-x-auto">
          <TabsTrigger value="overview">Visão Geral</TabsTrigger>
          <TabsTrigger value="expenses">Despesas</TabsTrigger>
          <TabsTrigger value="cabinet">Gabinete</TabsTrigger>
          <TabsTrigger value="amendments">Emendas</TabsTrigger>
          <TabsTrigger value="suppliers">Fornecedores</TabsTrigger>
        </TabsList>

        <div className="mt-6">
          <TabsContent value="overview">
            <OverviewTab selectedIds={selectedSenators.map(s => s.id)} />
          </TabsContent>
          
          <TabsContent value="expenses">
            <ExpensesTab selectedIds={selectedSenators.map(s => s.id)} />
          </TabsContent>

          <TabsContent value="cabinet">
            <Card>
              <CardContent className="p-6">
                <h2 className="text-xl font-bold mb-4">Estrutura de Gabinete</h2>
                <div className="h-64 flex items-center justify-center border-dashed border-2 rounded-lg">
                  <span className="text-muted-foreground">Lista de Servidores em construção</span>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="amendments" className="mt-6 space-y-6">
            <EmendasTab senators={selectedSenators} year={year} />
          </TabsContent>
          
          <TabsContent value="suppliers">
            <SuppliersTab selectedIds={selectedSenators.map(s => s.id)} />
          </TabsContent>
        </div>
      </Tabs>

      {/* Add More Section */}
      {selectedSenators.length < 5 && (
        <div id="add-senators" className="mt-12 pt-8 border-t">
          <h2 className="text-xl font-bold mb-4">Adicionar mais senadores</h2>
          <SenatorSelector />
        </div>
      )}
    </div>
  );
}
