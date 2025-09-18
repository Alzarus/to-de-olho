# 🚀 Decisão Técnica: Deployment "Tô De Olho" no Google Cloud Platform

**Data da Decisão**: 17-18 de Setembro de 2025  
**Responsável**: Equipe de Desenvolvimento  
**Status**: ✅ APROVADO

---

## 🎯 Executive Summary

**DECISÃO FINAL**: Deploy em **Google Cloud VM (Compute Engine)** com arquitetura Docker Compose preservada.

**Custo Estimado**: ~$35-40/mês (produção) | $4/mês (desenvolvimento)  
**Timeline**: Migration imediata possível  
**ROI**: Excelente - sem refatoração de código necessária

---

## 📊 Análise Comparativa: Cloud VM vs Cloud Run

### 🏆 **Cloud VM (Compute Engine) - ESCOLHIDO**

#### ✅ **Vantagens Decisivas**
```yaml
Compatibilidade Arquitetural:
  - Docker Compose funciona 100% sem modificações
  - PostgreSQL + Redis nativos preservados
  - Scheduler/Ingestor mantêm funcionalidade completa
  - Zero refatoração de código necessária

Custos Previsíveis:
  - e2-medium: $24.67/mês (1 vCPU, 4GB RAM)
  - SSD 50GB: $8.00/mês
  - Network: $2-5/mês
  - TOTAL: ~$35-40/mês

Controle Operacional:
  - SSH access completo
  - Volumes persistentes nativos
  - Debug facilitado
  - Monitoramento granular (Prometheus/Grafana)
  - Backup customizado
```

#### ⚠️ **Limitações Aceitáveis**
- Gerenciamento manual da VM
- Scaling manual (adequado para fase atual)
- Responsabilidade por updates do SO

### ❌ **Cloud Run - REJEITADO**

#### 🚫 **Limitações Críticas**
```yaml
Incompatibilidade Arquitetural:
  - Não suporta Docker Compose
  - Serviços stateless apenas
  - PostgreSQL → Cloud SQL obrigatório
  - Redis → Memorystore obrigatório
  - Scheduler → Cloud Scheduler + Pub/Sub
  - Refatoração completa necessária (3-6 meses)

Custos Explosivos:
  - Cloud Run: $15-30/mês
  - Cloud SQL: $25+/mês (mínimo)
  - Memorystore: $15+/mês
  - Load Balancer: $18/mês
  - TOTAL: $75-90+/mês (2-3x mais caro)

Complexidade Operacional:
  - Cold starts impactam UX
  - Debugging complexo
  - Vendor lock-in severo
```

---

## 🏗️ Arquitetura de Deployment

### **Configuração da VM Escolhida**
```bash
Machine Type: e2-medium
vCPUs: 1 (com burst até 2)
RAM: 4GB
Storage: 50GB SSD Persistent Disk
Zone: us-central1-a (baixa latência Brasil)
OS: Ubuntu 22.04 LTS
```

### **Stack Tecnológica na VM**
```yaml
Container Runtime: Docker + Docker Compose
Services Deploy:
  - Backend Go (API)
  - Frontend Next.js
  - PostgreSQL 16
  - Redis 7
  - Scheduler (cron jobs)
  - Ingestor (batch jobs)

Monitoramento:
  - Prometheus (métricas)
  - Grafana (dashboards)
  - Logs via GCP Cloud Logging

Backup:
  - PostgreSQL: dump diário + Cloud Storage
  - Redis: snapshot + persistência
  - Código: GitHub
```

---

## 💰 Análise Financeira Detalhada

### **Fase 1: Desenvolvimento (Free Tier)**
```yaml
VM e2-micro (sempre gratuita):
  - Compute: $0/mês (730h incluídas)
  - Storage 30GB: $4.80/mês
  - Network: $0-2/mês
  TOTAL: ~$5/mês

Duração: 3-6 meses (adequado para MVP)
```

### **Fase 2: Produção (e2-medium)**
```yaml
Custos Mensais:
  - VM e2-medium: $24.67/mês
  - SSD 50GB: $8.00/mês
  - Network egress: $3-5/mês
  - Cloud Storage (backup): $1-2/mês
  TOTAL: $36-40/mês

Sustained Use Discount: -30% após 25% do mês
Committed Use: -20-30% adicional (1-3 anos)
```

### **Comparação com Alternativas**
```yaml
Cloud Run (managed):     $75-90/mês
AWS EC2 t3.medium:       $40-50/mês
DigitalOcean Droplet:    $25/mês (menos recursos)
Heroku (equivalent):     $75-100/mês
```

**Conclusão**: GCP Cloud VM oferece melhor custo-benefício para nossa arquitetura.

---

## 🌐 Domínio e DNS

### **✅ SIM, Google Cloud oferece domínios!**

#### **Google Domains (integrado ao GCP)**
```yaml
Disponível no Brasil: ✅ SIM
Integração GCP: Automática
Preços (anuais):
  - .com.br: R$ 40-60/ano
  - .com: $12/ano (~R$ 60/ano)
  - .org: $12/ano
  - .app: $20/ano (HTTPS obrigatório)

Recursos Inclusos:
  - DNS gerenciado (Cloud DNS)
  - Certificados SSL gratuitos
  - Proteção WHOIS
  - Email forwarding
  - Integração com Load Balancer
```

#### **Recomendação de Domínio**
```yaml
Opções para "Tô De Olho":
  1. todeolho.com.br (RECOMENDADO)
     - Credibilidade nacional
     - SEO local otimizado
     - ~R$ 50/ano
  
  2. todeolho.app
     - Moderno, tech-friendly
     - HTTPS nativo
     - $20/ano (~R$ 100/ano)
  
  3. todeolho.gov.br (FUTURO)
     - Máxima credibilidade oficial
     - Processo burocrático
     - Gratuito (se aprovado)
```

#### **Setup DNS Recomendado**
```yaml
Cloud DNS Configuration:
  - A record: VM external IP
  - CNAME www: redirect to apex
  - MX records: Google Workspace (futuro)
  - TXT: SPF, DKIM, DMARC
  
SSL/TLS:
  - Let's Encrypt (gratuito)
  - Auto-renewal via certbot
  - HTTPS redirect obrigatório
```

---

## 🛡️ Monitoramento e Observabilidade

### **Stack de Monitoramento (Self-Hosted)**
```yaml
Prometheus:
  - Métricas: Go runtime, PostgreSQL, Redis, HTTP
  - Retenção: 15 dias local + backup Cloud Storage
  - Custo: $0 (software) + $2/mês (storage)

Grafana:
  - Dashboards: System, Application, Business
  - Alerting: Slack/Email integration
  - Users: 3-5 desenvolvedores
  - Custo: $0 (self-hosted)

GCP Cloud Monitoring:
  - VM health + basic metrics
  - Integration com Prometheus
  - Alerting redundante
  - Custo: Dentro do free tier
```

### **Métricas Críticas**
```yaml
Infrastructure:
  - CPU/Memory utilization
  - Disk usage/IOPS
  - Network throughput
  - VM uptime

Application:
  - API response times
  - Error rates por endpoint
  - Database connection pool
  - Cache hit ratio

Business:
  - Deputados sincronizados
  - Proposições atualizadas
  - Usuários ativos
  - API calls external (Câmara)
```

---

## 🚀 Roadmap de Implementação

### **Sprint 1: Setup Básico (1-2 semanas)**
```yaml
✅ Decisão arquitetural (CONCLUÍDO)
⏳ Criação da VM e2-micro (dev)
⏳ Setup Docker + Docker Compose
⏳ Deploy inicial com docker-compose atual
⏳ Configuração de domínio
⏳ SSL/HTTPS setup
```

### **Sprint 2: Monitoramento (1 semana)**
```yaml
⏳ Prometheus + Grafana setup
⏳ Dashboards básicos
⏳ Alerting crítico
⏳ Backup automatizado
```

### **Sprint 3: CI/CD (1-2 semanas)**
```yaml
⏳ GitHub Actions workflow
⏳ Deploy automatizado via SSH
⏳ Testing pipeline
⏳ Rollback strategy
```

### **Sprint 4: Produção (1 semana)**
```yaml
⏳ Upgrade para e2-medium
⏳ Otimizações de performance
⏳ Security hardening
⏳ Documentation completa
```

---

## 🔒 Considerações de Segurança

### **VM Security**
```yaml
Network:
  - Firewall rules específicas (22, 80, 443, 8080)
  - VPC privada
  - No SSH root access
  - Key-based authentication apenas

Application:
  - Container isolation
  - Secrets via environment variables
  - PostgreSQL user limitado
  - Rate limiting ativo
  
Backup & Recovery:
  - Daily automated backups
  - Point-in-time recovery (PostgreSQL)
  - Disaster recovery plan
  - RTO: 2 horas | RPO: 24 horas
```

---

## 📈 Escalabilidade Futura

### **Cenários de Crescimento**
```yaml
Fase 1 (0-1K usuários): e2-medium atual
Fase 2 (1K-10K): e2-standard-2 (2 vCPU, 8GB)
Fase 3 (10K+): Múltiplas VMs + Load Balancer
Fase 4 (100K+): Microservices + managed services
```

### **Migration Path (se necessário)**
```yaml
Para Kubernetes (GKE):
  - Docker containers já prontos
  - Helm charts from docker-compose
  - Gradual migration por service
  
Para Cloud Run (futuro):
  - Refatoração planejada
  - Stateless conversion
  - Managed services adoption
```

---

## ✅ Conclusão e Próximos Passos

### **Decisão Justificada**
1. **Compatibilidade Total**: Zero refatoração necessária
2. **Custo Otimizado**: $35-40/mês vs $75-90/mês alternatives
3. **Controle Operacional**: Flexibilidade máxima para debugging/customização
4. **Learning Curve**: Mínima - equipe já domina Docker
5. **Time to Market**: Deploy possível em 1-2 semanas

### **Action Items Imediatos**
- [ ] Criar conta GCP e configurar billing
- [ ] Registrar domínio `todeolho.com.br`
- [ ] Provisionar VM e2-micro (desenvolvimento)
- [ ] Setup inicial Docker Compose na VM
- [ ] Configurar HTTPS com Let's Encrypt

### **Métricas de Sucesso**
- Deploy em produção: < 3 semanas
- Uptime: > 99.5%
- Response time: < 500ms (p95)
- Custo mensal: < $45

---

**📋 Este documento será atualizado conforme a implementação progride.**

---

## 📚 Referências

- [GCP Compute Engine Pricing](https://cloud.google.com/compute/pricing)
- [GCP Free Tier](https://cloud.google.com/free)
- [Google Domains Pricing](https://domains.google/pricing/)
- [Prometheus Monitoring Stack](https://prometheus.io/docs/introduction/overview/)
- [Docker Compose Production Best Practices](https://docs.docker.com/compose/production/)

---

*Documento gerado em 18/09/2025 - Projeto "Tô De Olho"*