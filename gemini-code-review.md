In backend/internal/application/smart_backfill_service_test.go:

> @@ -0,0 +1,117 @@
+package application
high

Os testes para o SmartBackfillService são muito limitados e cobrem apenas alguns cenários de configuração e o gatilho de analytics. Dada a complexidade do serviço (lógica de decisão, execução em goroutine, paralelismo, tratamento de erros, etc.), é crucial adicionar mais testes unitários e de integração para garantir sua robustez. Recomendo testar a lógica de ShouldRunHistoricalBackfill em diferentes cenários, o fluxo de execução de runHistoricalBackfill com mocks para os serviços dependentes, e o tratamento de erros e retries.

In backend/internal/application/smart_scheduler_service_concurrency_test.go:

> @@ -0,0 +1,82 @@
+package application
high

Este teste cobre bem o cenário de concorrência, mas a cobertura de testes para o SmartSchedulerService como um todo é baixa. Recomendo adicionar mais testes para cobrir a lógica de ShouldRunScheduler, os diferentes fluxos de sincronização de entidades, e o tratamento de erros durante a execução do scheduler.

In .env.example:

> @@ -4,6 +4,64 @@
 # Este arquivo contém todas as variáveis de ambiente necessárias para o projeto.
 # Copie para .env e ajuste os valores conforme necessário.
 
+# =============================================================================
+# 🧠 CONFIGURAÇÕES INTELIGENTES DE BACKFILL
+# =============================================================================
+# O sistema decide automaticamente se precisa executar backfill
+BACKFILL_START_YEAR=2022
+BACKFILL_END_YEAR=0
medium

O valor 0 para BACKFILL_END_YEAR não é autoexplicativo. Sugiro adicionar um comentário para esclarecer seu significado, por exemplo, se 0 significa o ano atual, para melhorar a clareza da configuração.

BACKFILL_END_YEAR=0 # Usar 0 para definir o ano atual
In .env.example:

> +SCHEDULER_INCLUIR_DEPUTADOS=true
+SCHEDULER_INCLUIR_PROPOSICOES=true
+SCHEDULER_INCLUIR_DESPESAS=true
+SCHEDULER_INCLUIR_VOTACOES=true
medium

Há uma inconsistência na nomeação das variáveis de ambiente para controle de inclusão de entidades. Algumas usam o prefixo SCHEDULER_INCLUIR_ (em português), enquanto outras usam SCHEDULER_INCLUDE_ (em inglês, como nas linhas 56-57). Para manter a consistência e evitar confusão, sugiro padronizar para o inglês, que é o padrão para a maioria das outras variáveis.

SCHEDULER_INCLUDE_DEPUTADOS=true
SCHEDULER_INCLUDE_PROPOSICOES=true
SCHEDULER_INCLUDE_DESPESAS=true
SCHEDULER_INCLUDE_VOTACOES=true
In backend/cmd/ingestor/main.go:

> @@ -125,12 +190,30 @@ func runStrategicBackfill(
 
 	// Criar gerenciador de backfill e executor estratégico
 	backfillManager := ingestor.NewBackfillManager(pgPool)
+	votacaoRepo := repository.NewVotacaoRepository(pgPool)
+
+	// Construir VotacoesService e PartidosService localmente (precisa de client e cache)
+	clientLocal := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
+	cacheLocal := cache.NewFromConfig(&cfg.Redis)
+	votacoesSvcLocal := app.NewVotacoesService(votacaoRepo, clientLocal, cacheLocal)
+
+	partidoRepoLocal := repository.NewPartidoRepository(pgPool)
+	partidosSvcLocal := app.NewPartidosService(clientLocal, partidoRepoLocal)
+
+	// Criar analytics service para atualizar rankings após backfill
+	despesaRepoLocal := repository.NewDespesaRepository(pgPool)
+	analyticsSvcLocal := app.NewAnalyticsService(deputadoRepo, proposicaoRepo, votacaoRepo, despesaRepoLocal, cacheLocal, slog.New(slog.NewTextHandler(os.Stdout, nil)))
medium

Um novo logger está sendo criado aqui (slog.New(...)), enquanto um logger já foi configurado no início da função main. Para manter a consistência e centralizar a configuração de logs, o logger existente deveria ser passado para a função runStrategicBackfill e reutilizado aqui.