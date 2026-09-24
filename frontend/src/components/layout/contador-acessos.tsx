"use client";

import { useEffect } from "react";

const CHAVE = "tdo-acesso-registrado";

// Avisa a API de uma visita por aba aberta. A API conta um visitante por dia
// (hash de IP e navegador com sal diario, sem cookie); aqui so evitamos
// repetir o aviso a cada troca de pagina.
export function ContadorAcessos() {
  useEffect(() => {
    try {
      if (sessionStorage.getItem(CHAVE)) return;
      sessionStorage.setItem(CHAVE, "1");
    } catch {
      // sem sessionStorage (modo privado restrito): avisa assim mesmo
    }
    const url = "/api/v1/acessos";
    if (typeof navigator.sendBeacon === "function" && navigator.sendBeacon(url)) return;
    fetch(url, { method: "POST", keepalive: true }).catch(() => {});
  }, []);

  return null;
}
