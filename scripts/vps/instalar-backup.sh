#!/usr/bin/env bash
#
# Instala o backup diario (backup.sh) na VPS. Idempotente: pode rodar de novo.
#
# USO (NA VPS, como root, a partir de uma copia deste diretorio):
#   bash instalar-backup.sh              # instala age, rclone, script e timer
#   nano /etc/todeolho-backup.env        # preenche R2 e a chave publica age
#   bash instalar-backup.sh testar       # roda um backup agora e mostra o log
#
# Pre-requisitos fora da VPS (passo a passo no README.md deste diretorio):
#   - bucket R2 privado + token R2 "Object Read & Write" restrito a ele;
#   - par de chaves age gerado na SUA maquina (a privada nunca vem para ca).
#
set -euo pipefail

[ "$(id -u)" = 0 ] || { echo "[ERRO] rode como root" >&2; exit 1; }
AQUI="$(cd "$(dirname "$0")" && pwd)"
CONFIG=/etc/todeolho-backup.env
RCLONE_MIN="1.63.0"

versao_ok() { [ "$(printf '%s\n%s\n' "$RCLONE_MIN" "$1" | sort -V | head -1)" = "$RCLONE_MIN" ]; }

instalar_age() {
  command -v age >/dev/null && { echo "[OK] age $(age --version)"; return; }
  apt-get update -qq && apt-get install -y -qq age && echo "[OK] age instalado"
}

# rclone do apt costuma ser antigo (Ubuntu 22.04 traz 1.53, sem o provedor
# Cloudflare). Baixa o .deb oficial e confere o SHA256 publicado no release,
# em vez de "curl | bash".
instalar_rclone() {
  if command -v rclone >/dev/null && versao_ok "$(rclone version | head -1 | grep -o '[0-9.]*' | head -1)"; then
    echo "[OK] $(rclone version | head -1)"; return
  fi
  local arq ver tmp; tmp=$(mktemp -d)
  arq=$(dpkg --print-architecture)
  ver=$(curl -fsS https://downloads.rclone.org/version.txt | grep -o 'v[0-9.]*')
  curl -fsSL -o "$tmp/rclone.deb" "https://downloads.rclone.org/${ver}/rclone-${ver}-linux-${arq}.deb"
  curl -fsSL -o "$tmp/SHA256SUMS" "https://downloads.rclone.org/${ver}/SHA256SUMS"
  (cd "$tmp" && grep " rclone-${ver}-linux-${arq}.deb\$" SHA256SUMS | sed "s/rclone-${ver}-linux-${arq}.deb/rclone.deb/" | sha256sum -c -) \
    || { echo "[ERRO] SHA256 do rclone nao confere" >&2; exit 1; }
  dpkg -i "$tmp/rclone.deb" >/dev/null && rm -rf "$tmp" && echo "[OK] rclone ${ver} instalado"
}

escrever_config() {
  [ -f "$CONFIG" ] && { echo "[OK] $CONFIG ja existe (mantido)"; chmod 600 "$CONFIG"; return; }
  install -m 600 /dev/null "$CONFIG"
  cat > "$CONFIG" <<'EOF'
# Configuracao do backup (lida por /usr/local/sbin/todeolho-backup). Modo 600.
# Chave PUBLICA age (linha "age1..."); a privada fica FORA da VPS.
AGE_DESTINATARIO=
# Cloudflare R2: bucket privado e token "Object Read & Write" SO desse bucket.
R2_BUCKET=
RCLONE_CONFIG_R2_ACCESS_KEY_ID=
RCLONE_CONFIG_R2_SECRET_ACCESS_KEY=
# https://<ID da conta>.r2.cloudflarestorage.com
RCLONE_CONFIG_R2_ENDPOINT=
# Opcional: URL de ping do healthchecks.io (alerta por e-mail se faltar backup).
HC_URL=
EOF
  echo "[PENDENTE] preencha $CONFIG"
}

instalar_timer() {
  install -m 700 "$AQUI/backup.sh" /usr/local/sbin/todeolho-backup
  cat > /etc/systemd/system/todeolho-backup.service <<'EOF'
[Unit]
Description=Backup diario To De Olho / Quem Votar / NPM para o R2
After=docker.service network-online.target
Wants=network-online.target
Requires=docker.service

[Service]
Type=oneshot
ExecStart=/usr/local/sbin/todeolho-backup
# Prioridade baixa de CPU e disco: o backup nao pode deixar o site lento.
Nice=10
IOSchedulingClass=idle
TimeoutStartSec=2h
EOF
  cat > /etc/systemd/system/todeolho-backup.timer <<'EOF'
[Unit]
Description=Backup diario (depois do sync das 06:00 UTC)

[Timer]
OnCalendar=*-*-* 07:30:00 UTC
RandomizedDelaySec=10min
# Se a VPS estava desligada no horario, roda assim que voltar.
Persistent=true

[Install]
WantedBy=timers.target
EOF
  systemctl daemon-reload
  systemctl enable --now todeolho-backup.timer >/dev/null
  echo "[OK] timer: $(systemctl list-timers todeolho-backup.timer --no-pager | sed -n 2p)"
}

case "${1:-instalar}" in
  instalar) instalar_age; instalar_rclone; escrever_config; instalar_timer ;;
  testar)   systemctl start todeolho-backup.service || echo "[ERRO] backup falhou; log abaixo"; journalctl -u todeolho-backup.service -n 30 --no-pager ;;
  *) sed -n '2,14p' "$0"; exit 1 ;;
esac
