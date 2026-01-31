"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useState, useEffect } from "react";

interface PaginationWithInputProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  className?: string;
}

export function PaginationWithInput({
  currentPage,
  totalPages,
  onPageChange,
  className = "",
}: PaginationWithInputProps) {
  const [inputPage, setInputPage] = useState((currentPage || 1).toString());

  useEffect(() => {
    if ((currentPage || 1).toString() !== inputPage) {
        setInputPage((currentPage || 1).toString());
    }
  }, [currentPage]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // Allow empty string to let user delete
    if (e.target.value === "") {
        setInputPage("");
        return;
    }
    const val = parseInt(e.target.value);
    if (!isNaN(val)) {
        setInputPage(e.target.value);
    }
  };

  const handleInputBlur = () => {
      handlePageSubmit();
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
          handlePageSubmit();
      }
  }

  const handlePageSubmit = () => {
    let p = parseInt(inputPage);
    if (isNaN(p)) {
        // Reset to current if invalid
        setInputPage(currentPage.toString());
        return;
    }

    // Clamp
    if (p < 1) p = 1;
    if (p > totalPages) p = totalPages;

    setInputPage(p.toString());
    if (p !== currentPage) {
        onPageChange(p);
    }
  };

  if (totalPages <= 1) return null;

  return (
    <div className={`flex items-center justify-between text-sm text-muted-foreground ${className}`}>
      <div>
        Página {currentPage} de {totalPages}
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.max(1, currentPage - 1))}
          disabled={currentPage === 1}
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Anterior
        </Button>
        
        <div className="flex items-center gap-1">
            <span className="text-xs">Ir para:</span>
            <Input 
                className="h-8 w-12 text-center p-0" 
                value={inputPage}
                onChange={handleInputChange}
                onBlur={handleInputBlur}
                onKeyDown={handleKeyDown}
            />
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
          disabled={currentPage >= totalPages}
        >
          Próximo
          <ChevronRight className="h-4 w-4 ml-1" />
        </Button>
      </div>
    </div>
  );
}
