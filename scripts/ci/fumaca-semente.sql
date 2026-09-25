-- Semente mínima do teste de fumaça (verificacao.yml, job "fumaca").
-- Roda depois que a API subiu e o AutoMigrate criou as tabelas. Com o banco
-- vazio, /senador/1 responde 200 mas mostra "Senador não encontrado", e o
-- gráfico que o Playwright confere nem aparece.
INSERT INTO senadores (id, codigo_parlamentar, nome, nome_completo, partido, uf, em_exercicio, created_at, updated_at)
VALUES (1, 5012, 'Senadora Fumaça', 'Senadora do Teste de Fumaça', 'PX', 'BA', true, now(), now());

INSERT INTO despesas_ceaps (senador_id, ano, mes, tipo_despesa, fornecedor, valor, valor_centavos, id_origem, created_at, updated_at)
SELECT 1, 2025, m, 'Passagens aéreas', 'Fornecedor ' || m, 1000 + m, (1000 + m) * 100, 900000 + m, now(), now()
FROM generate_series(1, 12) AS m;
