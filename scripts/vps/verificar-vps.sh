#!/bin/bash
# Checagem SOMENTE LEITURA da VPS (auditoria 24/09). Nao altera nada.
s(){ echo; echo "===== $*"; }
s SO; . /etc/os-release; echo "$PRETTY_NAME"; uname -r; uptime
s RECURSOS; nproc; free -h; swapon --show; df -h / /var/lib/docker 2>/dev/null
s SSHD; sshd -T 2>/dev/null | grep -Ei '^(port|permitrootlogin|passwordauthentication|kbdinteractiveauthentication|pubkeyauthentication|maxauthtries|allowusers) '
s CHAVES; for f in /root/.ssh/authorized_keys /home/*/.ssh/authorized_keys; do [ -f "$f" ] && echo "$f: $(wc -l <"$f") chave(s): $(awk '{print $NF}' "$f" | tr '\n' ' ')"; done
s USUARIOS_COM_SHELL; awk -F: '$7 ~ /(bash|sh)$/ {print $1}' /etc/passwd | tr '\n' ' '; echo; getent group docker sudo
s FIREWALL; ufw status verbose 2>&1; echo "-- nft/iptables DOCKER-USER:"; iptables -S DOCKER-USER 2>&1; ip6tables -S DOCKER-USER 2>&1 | head -5
s PORTAS_ESCUTANDO; ss -tlnpH | awk '{print $4, $6}' | sort -u
s SERVICOS; for u in fail2ban unattended-upgrades ssh docker; do printf '%s: %s\n' "$u" "$(systemctl is-active $u 2>&1)"; done
s ATUALIZACOES_PENDENTES; apt list --upgradable 2>/dev/null | tail -n +2 | wc -l; [ -f /var/run/reboot-required ] && echo "REBOOT PENDENTE"
s DOCKER; docker --version; docker ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'; echo; docker stats --no-stream --format '{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'
cat /etc/docker/daemon.json 2>/dev/null
s CRON; crontab -l 2>&1; ls /etc/cron.d /etc/cron.daily 2>/dev/null; systemctl list-timers --all --no-pager 2>/dev/null | head -20
s BACKUPS; ls -la /opt/todeolho/backups 2>&1; ls -la /opt/*/backup* 2>/dev/null | head; command -v rclone restic aws 2>/dev/null
s BANCO; U=$(grep ^POSTGRES_USER= /opt/todeolho/.env|cut -d= -f2); D=$(grep ^POSTGRES_DB= /opt/todeolho/.env|cut -d= -f2)
docker exec todeolho-db psql -U "$U" -d "$D" -Atc "select 'max_connections='||current_setting('max_connections'), 'shared_buffers='||current_setting('shared_buffers'), 'tamanho='||pg_size_pretty(pg_database_size(current_database())), 'conexoes_agora='||(select count(*) from pg_stat_activity)"
s QUEMVOTAR; ls -la /opt/quemvotar 2>/dev/null | head; docker inspect -f '{{range .Mounts}}{{.Name}} -> {{.Destination}}{{"\n"}}{{end}}' quemvotar-api 2>/dev/null
s NPM_DOMINIOS; grep -h 'server_name' /opt/proxy/data/nginx/proxy_host/*.conf 2>/dev/null | sort -u; ls /opt/proxy 2>/dev/null
s NPM_CERTS; ls /opt/proxy/letsencrypt/live 2>/dev/null
s FIM
