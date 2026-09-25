#!/usr/bin/env bash
#
# DNS redundante na VPS: resolvedores publicos ao lado dos da Contabo.
#
# POR QUE (25/09/2026)
#   O deploy do #58 falhou no docker login com "lookup ghcr.io on
#   127.0.0.53:53: i/o timeout". A VPS so tinha os dois resolvedores da Contabo
#   (via netplan, no link eth0), sem nenhum de outro provedor; num teste de 20
#   consultas, 2 passaram de 1 s. Quando os dois engasgam juntos, nao ha para
#   onde ir.
#
# COMO
#   Um drop-in em /etc/systemd/resolved.conf.d/ define servidores GLOBAIS
#   (Cloudflare e Quad9, IPv4 e IPv6). O systemd-resolved passa a ter dois
#   escopos com rota padrao (o do eth0, com os da Contabo, e o global) e consulta
#   os dois; vale a primeira resposta positiva. Nada muda no netplan.
#   O restart do resolved leva menos de 1 s; os conteineres usam o DNS embutido
#   do Docker, que repassa ao do host, e so perdem resolucao nesse intervalo.
#
# USO (NA VPS, como root)
#   bash dns-redundante.sh status     # configuracao atual + 20 consultas cronometradas
#   bash dns-redundante.sh aplicar
#   bash dns-redundante.sh desfazer
#
set -euo pipefail

[ "$(id -u)" = 0 ] || { echo "[ERRO] rode como root" >&2; exit 1; }
DROPIN=/etc/systemd/resolved.conf.d/10-todeolho-dns.conf

medir() { # 20 consultas a ghcr.io e ao host dos blobs do GHCR
  local ok=0 falha=0 lento=0 t0 ms i h
  for i in $(seq 1 10); do
    for h in ghcr.io pkg-containers.githubusercontent.com; do
      t0=$(date +%s%N)
      if getent ahosts "$h" >/dev/null 2>&1; then ok=$((ok + 1)); else falha=$((falha + 1)); fi
      ms=$(( ($(date +%s%N) - t0) / 1000000 )); [ "$ms" -gt 1000 ] && lento=$((lento + 1))
    done
  done
  echo "20 consultas: ok=$ok falha=$falha lentas(>1s)=$lento"
}

cmd_status() {
  resolvectl status | grep -E "^Global|Current DNS|DNS Servers" | sed 's/^/   /'
  [ -f "$DROPIN" ] && echo "[OK] $DROPIN presente" || echo "[INFO] sem $DROPIN"
  medir
}

cmd_aplicar() {
  mkdir -p "$(dirname "$DROPIN")"
  cat > "$DROPIN" <<'EOF'
# DNS redundante (scripts/vps/dns-redundante.sh). Servidores globais somados
# aos do link eth0 (Contabo, via netplan): o resolved consulta os dois escopos.
[Resolve]
DNS=1.1.1.1 9.9.9.9 2606:4700:4700::1111 2620:fe::fe
EOF
  systemctl restart systemd-resolved
  sleep 1
  getent ahosts ghcr.io >/dev/null || { echo "[ERRO] resolucao falhou apos aplicar; desfazendo" >&2; cmd_desfazer; exit 1; }
  echo "[OK] DNS redundante aplicado"
  cmd_status
}

cmd_desfazer() {
  rm -f "$DROPIN"
  systemctl restart systemd-resolved
  echo "[OK] drop-in removido; so os resolvedores da Contabo"
}

case "${1:-status}" in
  status)   cmd_status ;;
  aplicar)  cmd_aplicar ;;
  desfazer) cmd_desfazer ;;
  *) sed -n '2,24p' "$0"; exit 1 ;;
esac
