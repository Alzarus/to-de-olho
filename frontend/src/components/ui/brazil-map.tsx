"use client";

import * as React from "react";
import { BRAZIL_STATES } from "./brazil-map-data";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface BrazilMapProps {
  data?: { uf: string; value: number }[];
  className?: string;
  onHover?: (uf: string | null) => void;
}

export function BrazilMap({ data = [], className, onHover }: BrazilMapProps) {
  // Normalize data for easier lookup
  const dataMap = React.useMemo(() => {
    return data.reduce(
      (acc, item) => {
        acc[item.uf] = item.value;
        return acc;
      },
      {} as Record<string, number>,
    );
  }, [data]);

  // Find max value for scaling opacity (simple linear scale for now)
  const maxValue = React.useMemo(() => {
    return Math.max(...data.map((d) => d.value), 1);
  }, [data]);

  const getColor = (uf: string) => {
    const value = dataMap[uf];
    if (!value) return "fill-muted-foreground/20"; // Default color for no data

    // Calculate opacity based on value relative to max or just use a fixed highlight
    // For this use case (destinations), usually we just want to highlight states that have data
    // But if we want choropleth:
    // const opacity = Math.max(0.3, value / maxValue);
    // return `fill-blue-600 dark:fill-blue-500 opacity-[${opacity}]`;

    // For now, let's just highlight states with traffic
    return "fill-blue-600 dark:fill-blue-500 hover:fill-blue-700 dark:hover:fill-blue-400";
  };

  return (
    <div className={cn("relative w-full h-full aspect-[1.1]", className)}>
      <TooltipProvider delayDuration={0}>
        <svg
          viewBox="0 0 1000 912"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-full"
        >
          <g stroke="#ffffff" strokeWidth="1" strokeLinejoin="round">
            {BRAZIL_STATES.map((state) => {
              const hasData = state.id in dataMap;
              const value = dataMap[state.id];
              const formattedValue = value
                ? new Intl.NumberFormat("pt-BR", {
                    style: "currency",
                    currency: "BRL",
                    notation: "compact",
                  }).format(value)
                : null;

              return (
                <Tooltip key={state.id}>
                  <TooltipTrigger asChild>
                    <path
                      d={state.path}
                      id={state.id}
                      tabIndex={0}
                      onClick={(e) => {
                        e.currentTarget.focus();
                        // Stop propagation if it's inside another clickable area
                        e.stopPropagation();
                      }}
                      className={cn(
                        "transition-colors duration-200 cursor-pointer outline-none focus:outline-none",
                        hasData
                          ? "fill-blue-600 dark:fill-blue-500 hover:opacity-80 focus:opacity-80"
                          : "fill-gray-200 dark:fill-gray-800 hover:fill-gray-300 dark:hover:fill-gray-700 focus:fill-gray-300 dark:focus:fill-gray-700",
                      )}
                      onMouseEnter={() => onHover?.(state.id)}
                      onMouseLeave={() => onHover?.(null)}
                    />
                  </TooltipTrigger>
                  <TooltipContent>
                    <div className="text-sm font-medium">
                      {state.name} ({state.id})
                    </div>
                    {formattedValue && (
                      <div className="text-xs font-medium opacity-90">
                        {formattedValue}
                      </div>
                    )}
                  </TooltipContent>
                </Tooltip>
              );
            })}
          </g>
        </svg>
      </TooltipProvider>
    </div>
  );
}
