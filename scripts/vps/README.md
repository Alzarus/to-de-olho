# Operação da VPS e da Cloudflare

Scripts da auditoria de produção de 24/09/2026. Todos são idempotentes (podem
rodar de novo) e têm um modo que só lê (`status` ou `verificar`).

| script | onde roda | o que faz |
|---|---|---|
| `firewall-cloudflare.sh` | VPS (root) | 80/443 só aceitam a Cloudflare (cadeia `DOCKER-USER`) |
| `endurecer-ssh.sh` | VPS (root) | SSH só por chave, fail2ban, atualizações de segurança |
| `instalar-backup.sh` + `backup.sh` | VPS (root) | backup diário criptografado para o Cloudflare R2 |
| `restauracao-local.sh` | sua máquina (Docker) | gera a chave age e testa a restauração |
| `../cloudflare/regras-zona.sh` | sua máquina | TLS, cache de borda e limite de taxa da zona |

## Ordem recomendada

1. **Cloudflare** (na sua máquina). Crie um token com as permissões listadas no
   cabeçalho de `regras-zona.sh` e rode `status`, depois `aplicar`.
2. **SSH**, na VPS, com **outra sessão SSH aberta**: `verificar`, depois
   `aplicar`. Confirme que um login novo funciona antes de fechar as sessões.
3. **Firewall**, na VPS: `verificar` (aborta se algum domínio do NPM não
   estiver atrás da Cloudflare), depois `aplicar`, teste o site e rode
   `confirmar` em até 5 minutos. Sem o `confirmar`, as regras são removidas
   sozinhas.
4. **Backup**:
   1. Na sua máquina, rode `bash restauracao-local.sh chave`. Guarde a chave
      privada no gerenciador de senhas: sem ela nenhum backup abre.
   2. No painel da Cloudflare, abra R2 e crie o bucket `todeolho-backups`
      (privado). Em "Manage API tokens", crie um token **Object Read & Write**
      restrito a esse bucket e anote a access key, a secret key e o endpoint
      `https://<conta>.r2.cloudflarestorage.com`.
   3. Ainda no bucket, crie as regras de ciclo de vida: prefixo `diario/`,
      apagar após 35 dias; prefixo `mensal/`, apagar após 400 dias.
      Recomendado: um *bucket lock* de 30 dias. Assim nem quem invadir a VPS
      consegue apagar os backups recentes.
   4. Copie este diretório para a VPS, rode `bash instalar-backup.sh`,
      preencha `/etc/todeolho-backup.env` e rode `bash instalar-backup.sh testar`.
   5. Na sua máquina, rode `bash restauracao-local.sh testar r2` para provar que
      o backup restaura. Repita uma vez por mês.
   6. Opcional, mas recomendado: crie um check no healthchecks.io (grátis) com
      período de 1 dia e folga de 1 hora, e ponha a URL em `HC_URL`. Se o
      backup falhar ou não rodar, você recebe um e-mail.

## Como foi testado (24/09/2026)

- **Backup**: ciclo completo em contêineres locais. Postgres e a API do Quem
  Votar com os nomes de produção, um servidor S3 no lugar do R2 e um Debian 12
  no papel da VPS. Passou por backup, envio conferido pelo tamanho, download,
  descriptografia, sha256 contra o MANIFESTO, `pg_restore` com contagens
  conferidas e `integrity_check` do SQLite. Com o banco parado, o backup falha
  com exit 1 e mensagem.
- **Firewall**: Debian 12 privilegiado com iptables-nft. Testados a
  idempotência (aplicar duas vezes não duplica regras), o uso da lista salva
  quando a API da Cloudflare não responde, a remoção limpa e o aborto com um
  domínio fora da Cloudflare (nenhuma regra instalada). **Não testado**: o
  tráfego real passando pela `DOCKER-USER`, que só dá para ver na VPS. Por isso
  existe a remoção automática em 5 minutos.
- **SSH**: `10-todeolho.conf` vence um `50-cloud-init.conf` com
  `PasswordAuthentication yes`, o `desfazer` restaura a configuração e o
  fail2ban valida a jail com `backend=systemd`.
- **Cloudflare**: a mesclagem preserva regras criadas no painel e substitui as
  gerenciadas. **Não testado** contra a API real, porque depende do seu token.
