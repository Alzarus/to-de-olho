
import * as React from "react"

const MOBILE_BREAKPOINT = 768
const MOBILE_QUERY = `(max-width: ${MOBILE_BREAKPOINT - 1}px)`

function assinar(aoMudar: () => void) {
  const mql = window.matchMedia(MOBILE_QUERY)
  mql.addEventListener("change", aoMudar)
  return () => mql.removeEventListener("change", aoMudar)
}

const lerNoCliente = () => window.matchMedia(MOBILE_QUERY).matches

// No servidor (e na hidratação) o layout é tratado como desktop
const lerNoServidor = () => false

export function useIsMobile() {
  return React.useSyncExternalStore(assinar, lerNoCliente, lerNoServidor)
}
