#!/usr/bin/env bash
#
# Endurecimento basico da VPS: SSH so por chave, fail2ban e atualizacoes
# automaticas de seguranca.
#
# O QUE MUDA
#   - SSH: senha desligada (so chave), root so por chave (prohibit-password: o
#     deploy do GitHub Actions entra como root por chave e continua funcionando),
#     no maximo 3 tentativas por conexao, sem X11.
#     Vai num arquivo 10-todeolho.conf em sshd_config.d. No sshd vale a PRIMEIRA
#     definicao encontrada, e os includes sao lidos em ordem alfabetica: o
#     prefixo 10- vence o 50-cloud-init.conf que muitas imagens trazem com
#     "PasswordAuthentication yes".
#   - fail2ban na jail sshd, com backend systemd (Debian 12 e Ubuntu recentes
#     nao gravam /var/log/auth.log; com o backend padrao o fail2ban sobe e nao
#     vigia nada).
#   - unattended-upgrades so com o repositorio de seguranca, SEM reinicio
#     automatico (reiniciar derruba os sites; o script avisa quando houver
#     reinicio pendente).
#
# PROTECOES
#   - Aborta se nao houver chave em /root/.ssh/authorized_keys (trancaria o
#     acesso) ou se o sshd_config nao incluir sshd_config.d.
#   - Valida com "sshd -t" antes de recarregar; recarregar nao derruba sessoes.
#   - MANTENHA OUTRA SESSAO SSH ABERTA enquanto aplica e teste um login novo
#     antes de fechar as duas.
#
# USO (NA VPS, como root)
#   bash endurecer-ssh.sh verificar
#   bash endurecer-ssh.sh aplicar
#   bash endurecer-ssh.sh desfazer     # remove o arquivo do SSH e recarrega
#
set -euo pipefail

[ "$(id -u)" = 0 ] || { echo "[ERRO] rode como root" >&2; exit 1; }
DROPIN=/etc/ssh/sshd_config.d/10-todeolho.conf

unidade_ssh() { systemctl list-unit-files ssh.service >/dev/null 2>&1 && echo ssh || echo sshd; }

verificar() {
  local chaves=0
  [ -f /root/.ssh/authorized_keys ] && chaves=$(grep -cE '^(ssh-|ecdsa-|sk-)' /root/.ssh/authorized_keys || true)
  [ "$chaves" -gt 0 ] || { echo "[ERRO] /root/.ssh/authorized_keys sem chaves: aplicar trancaria o acesso" >&2; return 1; }
  echo "[OK] root tem $chaves chave(s) autorizada(s)"
  grep -qE '^\s*Include\s+/etc/ssh/sshd_config\.d/\*\.conf' /etc/ssh/sshd_config \
    || { echo "[ERRO] sshd_config nao inclui sshd_config.d/*.conf" >&2; return 1; }
  echo "[OK] sshd_config inclui sshd_config.d"
  ls /etc/ssh/sshd_config.d/ 2>/dev/null | sed 's/^/   sshd_config.d: /'
  echo "-- configuracao efetiva hoje"
  sshd -T 2>/dev/null | grep -Ei '^(permitrootlogin|passwordauthentication|kbdinteractiveauthentication|maxauthtries|x11forwarding) ' | sed 's/^/   /'
  for u in fail2ban unattended-upgrades; do printf '   %s: %s\n' "$u" "$(systemctl is-active "$u" 2>/dev/null || true)"; done
}

aplicar_ssh() {
  cat > "$DROPIN" <<'EOF'
# Endurecimento To De Olho (scripts/vps/endurecer-ssh.sh). Prefixo 10- para
# vencer outros arquivos deste diretorio: no sshd vale a primeira definicao.
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitEmptyPasswords no
PermitRootLogin prohibit-password
MaxAuthTries 3
LoginGraceTime 30
X11Forwarding no
EOF
  chmod 644 "$DROPIN"
  sshd -t || { rm -f "$DROPIN"; echo "[ERRO] sshd -t falhou; arquivo removido, nada mudou" >&2; return 1; }
  systemctl reload "$(unidade_ssh)"
  # Sem "grep -q": com pipefail, o grep saindo cedo mata o sshd com SIGPIPE e o
  # teste falha mesmo com a configuracao certa.
  local efetiva; efetiva=$(sshd -T 2>/dev/null)
  grep -qx 'passwordauthentication no' <<<"$efetiva" || { echo "[ERRO] senha continua ligada: outro arquivo vence o 10-todeolho.conf" >&2; return 1; }
  echo "[OK] SSH: so chave; root por chave; MaxAuthTries 3"
}

aplicar_fail2ban() {
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq fail2ban >/dev/null
  cat > /etc/fail2ban/jail.d/todeolho-sshd.local <<'EOF'
[sshd]
enabled = true
backend = systemd
maxretry = 5
findtime = 10m
bantime = 1h
# Reincidente fica bloqueado por mais tempo a cada vez (ate 1 semana).
bantime.increment = true
bantime.maxtime = 1w
EOF
  systemctl enable --now fail2ban >/dev/null
  systemctl restart fail2ban
  sleep 2
  fail2ban-client status sshd | grep -E 'Currently|Total' | sed 's/^/   /'
  echo "[OK] fail2ban ativo na jail sshd"
}

aplicar_atualizacoes() {
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq unattended-upgrades >/dev/null
  cat > /etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF
  # So o repositorio de seguranca (o padrao do pacote) e sem reinicio sozinho.
  cat > /etc/apt/apt.conf.d/52todeolho-unattended <<'EOF'
Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
EOF
  systemctl enable --now unattended-upgrades >/dev/null
  echo "[OK] atualizacoes automaticas de seguranca ligadas (sem reinicio automatico)"
  [ -f /var/run/reboot-required ] && echo "[AVISO] ha reinicio pendente: agende uma janela (docker sobe tudo com restart: unless-stopped)"
  return 0
}

case "${1:-}" in
  verificar) verificar ;;
  aplicar)
    verificar
    apt-get update -qq
    aplicar_ssh
    aplicar_fail2ban
    aplicar_atualizacoes
    echo "[PRONTO] ABRA UMA SESSAO SSH NOVA para confirmar o login antes de fechar esta." ;;
  desfazer) rm -f "$DROPIN"; sshd -t && systemctl reload "$(unidade_ssh)"; echo "[OK] configuracao SSH anterior restaurada" ;;
  *) sed -n '2,32p' "$0"; exit 1 ;;
esac
