#!/usr/bin/env bash
#
# Portas 80/443 da VPS so aceitam conexoes vindas da Cloudflare.
#
# PROBLEMA (auditoria de 24/09/2026)
#   HTTPS direto no IP da VPS respondia 200 com o site. Quem acha a origem
#   (historico de DNS, certificado indexado) contorna o WAF, o cache e a
#   protecao contra DDoS da Cloudflare, e ainda forja o CF-Connecting-IP que a
#   API e o nginx do Quem Votar usam como IP do visitante.
#
# POR QUE NAO E UFW
#   As portas 80/443 sao publicadas pelo Docker (contêiner do Nginx Proxy
#   Manager). O Docker faz DNAT na PREROUTING e o pacote segue pela FORWARD:
#   ele nunca passa pelas regras de entrada do ufw. O lugar documentado pelo
#   Docker para filtrar isso e a cadeia DOCKER-USER, que ele nao apaga.
#   Casamos pela porta ORIGINAL (conntrack --ctorigdstport), que continua 80/443
#   mesmo depois do DNAT, e so conexoes NOVAS vindas da interface externa;
#   trafego entre contêineres nao e afetado. Porta 22 (SSH) nao e tocada.
#   IPv6: o mesmo filtro na DOCKER-USER do ip6tables (se existir) e na INPUT
#   (quando o docker-proxy escuta em [::] e atende IPv6 em espaco de usuario).
#
# SEGURANCA DA OPERACAO
#   - "verificar" nao altera nada: confere Docker, cadeia, interface, faixas da
#     Cloudflare e se TODO dominio do NPM resolve para a Cloudflare. Um dominio
#     "DNS only" (nuvem cinza) cairia com este filtro; nesse caso aborta.
#   - "aplicar" agenda a REMOCAO automatica em 5 min. Se o site cair, espere ou
#     rode "remover". Se estiver tudo certo, rode "confirmar" (instala o servico
#     que reaplica no boot e o timer semanal que atualiza as faixas).
#   - A lista da Cloudflare vem de api.cloudflare.com/client/v4/ips. Se a busca
#     falhar ou vier incompleta, usa a ultima lista salva; nunca aplica vazia.
#
# USO (NA VPS, como root)
#   bash firewall-cloudflare.sh verificar
#   bash firewall-cloudflare.sh aplicar      # testar o site em seguida
#   bash firewall-cloudflare.sh confirmar    # em ate 5 min
#   bash firewall-cloudflare.sh status | remover | atualizar
#
set -euo pipefail

CADEIA=TDO-CF
DIR=/etc/todeolho-firewall
NPM_DIR="${NPM_DIR:-/opt/proxy}"
INSTALADO=/usr/local/sbin/todeolho-firewall-cf
MIN_FAIXAS_V4=10

[ "$(id -u)" = 0 ] || { echo "[ERRO] rode como root" >&2; exit 1; }

interface_externa() { ip route get 1.1.1.1 | grep -o 'dev [^ ]*' | cut -d' ' -f2; }
tem_cadeia() { "$1" -S "$2" >/dev/null 2>&1; }

ip_para_int() { local IFS=. a b c d; read -r a b c d <<<"$1"; echo $(( (a<<24) + (b<<16) + (c<<8) + d )); }
ip_na_faixa() { # ip_na_faixa 104.21.90.141 104.16.0.0/13
  local ip rede bits mascara
  ip=$(ip_para_int "$1"); rede=$(ip_para_int "${2%/*}"); bits=${2#*/}
  mascara=$(( bits == 0 ? 0 : (0xFFFFFFFF << (32 - bits)) & 0xFFFFFFFF ))
  (( (ip & mascara) == (rede & mascara) ))
}

# Baixa as faixas; so substitui o arquivo salvo se a resposta for completa.
buscar_faixas() {
  local json v4 v6; mkdir -p "$DIR"
  json=$(curl -fsS -m 15 --retry 3 https://api.cloudflare.com/client/v4/ips 2>/dev/null || true)
  v4=$(grep -o '"ipv4_cidrs":\[[^]]*' <<<"$json" | grep -oE '[0-9.]+/[0-9]+' || true)
  v6=$(grep -o '"ipv6_cidrs":\[[^]]*' <<<"$json" | grep -oE '[0-9a-f:]+/[0-9]+' || true)
  if [ "$(wc -l <<<"$v4")" -ge "$MIN_FAIXAS_V4" ] && [ -n "$v6" ]; then
    printf '%s\n' "$v4" > "$DIR/cf-v4.txt"; printf '%s\n' "$v6" > "$DIR/cf-v6.txt"
    echo "[OK] faixas da Cloudflare: $(wc -l < "$DIR/cf-v4.txt") IPv4, $(wc -l < "$DIR/cf-v6.txt") IPv6"
  elif [ -s "$DIR/cf-v4.txt" ]; then
    echo "[AVISO] busca das faixas falhou; usando a lista salva em $DIR"
  else
    echo "[ERRO] sem faixas da Cloudflare (busca falhou e nao ha lista salva)" >&2; return 1
  fi
}

dominios_npm() {
  grep -hoE 'server_name[^;]*' "$NPM_DIR"/data/nginx/{proxy_host,redirection_host,dead_host}/*.conf 2>/dev/null \
    | sed 's/server_name//' | tr ' ' '\n' | grep -E '^[a-z0-9.-]+\.[a-z]+$' | sort -u
}

# Todo dominio servido pelo NPM precisa resolver para a Cloudflare; senao, o
# filtro derruba aquele site. Tambem avisa de portas publicadas em 0.0.0.0 que
# NAO sao 80/443 (continuariam abertas, fora deste filtro).
verificar() {
  local falhou=0 d ip ok faixa
  docker info >/dev/null 2>&1 || { echo "[ERRO] Docker nao responde" >&2; return 1; }
  tem_cadeia iptables DOCKER-USER || { echo "[ERRO] cadeia DOCKER-USER ausente (backend nftables do Docker?)" >&2; return 1; }
  echo "[OK] Docker e DOCKER-USER (IPv4); interface externa: $(interface_externa)"
  tem_cadeia ip6tables DOCKER-USER && echo "[OK] DOCKER-USER (IPv6) presente" || echo "[INFO] sem DOCKER-USER no IPv6"
  buscar_faixas
  echo "-- dominios do NPM"
  while read -r d; do
    [ -n "$d" ] || continue
    for ip in $(getent ahostsv4 "$d" | awk '{print $1}' | sort -u); do
      ok=0; while read -r faixa; do ip_na_faixa "$ip" "$faixa" && { ok=1; break; }; done < "$DIR/cf-v4.txt"
      if [ $ok = 1 ]; then echo "   [OK]    $d -> $ip (Cloudflare)"; else echo "   [FALHA] $d -> $ip NAO e Cloudflare (DNS only?)"; falhou=1; fi
    done
  done < <(dominios_npm)
  ls "$NPM_DIR"/data/nginx/stream/*.conf >/dev/null 2>&1 && echo "[AVISO] o NPM tem streams TCP/UDP: nao sao filtrados aqui"
  echo "-- portas publicadas pelo Docker fora do loopback"
  docker ps --format '{{.Names}} {{.Ports}}' | tr ',' '\n' | grep -E '0\.0\.0\.0|\[::\]|:::' | sed 's/^/   /' || true
  [ $falhou = 0 ] || { echo "[ERRO] ha dominios fora da Cloudflare; ligue o proxy (nuvem laranja) ou nao aplique" >&2; return 1; }
}

montar_cadeia() { # montar_cadeia iptables cf-v4.txt
  local bin="$1" arq="$2" faixa
  "$bin" -N "$CADEIA" 2>/dev/null || "$bin" -F "$CADEIA"
  while read -r faixa; do [ -n "$faixa" ] && "$bin" -A "$CADEIA" -s "$faixa" -j RETURN; done < "$DIR/$arq"
  "$bin" -A "$CADEIA" -j DROP
}

regra_salto() { # regra_salto iptables DOCKER-USER -C|-I|-D porta
  local bin="$1" cad="$2" op="$3" porta="$4" iface
  iface=$(interface_externa)
  if [ "$cad" = INPUT ]; then
    "$bin" "$op" "$cad" $([ "$op" = -I ] && echo 1) -i "$iface" -p tcp --dport "$porta" -m conntrack --ctstate NEW -j "$CADEIA"
  else
    "$bin" "$op" "$cad" $([ "$op" = -I ] && echo 1) -i "$iface" -p tcp -m conntrack --ctstate NEW --ctorigdstport "$porta" --ctdir ORIGINAL -j "$CADEIA"
  fi
}

ligar() { # ligar iptables cf-v4.txt DOCKER-USER [INPUT]
  local bin="$1" arq="$2"; shift 2
  montar_cadeia "$bin" "$arq"
  for cad in "$@"; do
    tem_cadeia "$bin" "$cad" || continue
    for porta in 80 443; do regra_salto "$bin" "$cad" -C "$porta" 2>/dev/null || regra_salto "$bin" "$cad" -I "$porta"; done
  done
}

desligar() { # desligar iptables DOCKER-USER [INPUT]
  local bin="$1"; shift
  for cad in "$@"; do
    tem_cadeia "$bin" "$cad" || continue
    for porta in 80 443; do while regra_salto "$bin" "$cad" -D "$porta" 2>/dev/null; do :; done; done
  done
  "$bin" -F "$CADEIA" 2>/dev/null || true; "$bin" -X "$CADEIA" 2>/dev/null || true
}

aplicar_regras() {
  buscar_faixas
  ligar iptables cf-v4.txt DOCKER-USER
  ligar ip6tables cf-v6.txt DOCKER-USER INPUT
  echo "[OK] filtro ativo: 80/443 so a partir da Cloudflare ($(interface_externa))"
}

cmd_aplicar() {
  verificar
  aplicar_regras
  install -m 700 "$0" "$INSTALADO"
  systemctl stop tdo-cf-rollback.timer 2>/dev/null || true
  systemd-run --quiet --unit=tdo-cf-rollback --on-active=300 "$INSTALADO" remover
  cat <<'EOF'
[ATENCAO] remocao automatica agendada para daqui a 5 min.
  Teste agora, de fora:  curl -sI https://todeolho.org/ | head -1           (deve dar 200)
                         curl -skI --resolve todeolho.org:443:<IP> https://todeolho.org/ (deve dar timeout)
  Tudo certo?            bash firewall-cloudflare.sh confirmar
  Algo quebrou?          bash firewall-cloudflare.sh remover   (ou espere 5 min)
EOF
}

cmd_confirmar() {
  systemctl stop tdo-cf-rollback.timer 2>/dev/null || true
  install -m 700 "$0" "$INSTALADO"
  # PartOf/WantedBy docker.service: reaplica quando o Docker sobe ou reinicia.
  cat > /etc/systemd/system/todeolho-firewall-cf.service <<EOF
[Unit]
Description=Filtro 80/443 so para a Cloudflare (DOCKER-USER)
After=docker.service network-online.target
PartOf=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=$INSTALADO atualizar
ExecStop=$INSTALADO remover

[Install]
WantedBy=docker.service
EOF
  cat > /etc/systemd/system/todeolho-firewall-cf-atualizar.timer <<'EOF'
[Unit]
Description=Atualiza semanalmente as faixas de IP da Cloudflare

[Timer]
OnCalendar=Mon *-*-* 05:00:00 UTC
Persistent=true
Unit=todeolho-firewall-cf.service

[Install]
WantedBy=timers.target
EOF
  systemctl daemon-reload
  systemctl enable todeolho-firewall-cf.service todeolho-firewall-cf-atualizar.timer >/dev/null
  systemctl start todeolho-firewall-cf-atualizar.timer
  echo "[OK] confirmado: remocao automatica cancelada; reaplica no boot e atualiza as faixas toda segunda"
}

cmd_remover() {
  desligar iptables DOCKER-USER
  desligar ip6tables DOCKER-USER INPUT
  echo "[OK] filtro removido: 80/443 abertas para qualquer origem"
}

cmd_status() {
  for bin in iptables ip6tables; do
    echo "-- $bin"
    for cad in DOCKER-USER INPUT; do "$bin" -S "$cad" 2>/dev/null | grep -- "-j $CADEIA" | sed 's/^/   /'; done
    "$bin" -L "$CADEIA" -nv 2>/dev/null | tail -1 | sed 's/^/   ultima regra (DROP): /' || echo "   (sem cadeia $CADEIA)"
  done
  systemctl is-active tdo-cf-rollback.timer >/dev/null 2>&1 && echo "[ATENCAO] remocao automatica pendente"
  systemctl is-enabled todeolho-firewall-cf.service 2>/dev/null | sed 's/^/servico no boot: /' || true
}

case "${1:-}" in
  verificar) verificar ;;
  aplicar)   cmd_aplicar ;;
  confirmar) cmd_confirmar ;;
  remover)   cmd_remover ;;
  atualizar) aplicar_regras ;;
  status)    cmd_status ;;
  *) sed -n '2,40p' "$0"; exit 1 ;;
esac
