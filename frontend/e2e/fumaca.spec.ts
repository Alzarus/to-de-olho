import { test, expect, type Page } from "@playwright/test";

// Teste de fumaça das imagens (verificacao.yml, job "fumaca"): roda contra as
// imagens recém-construídas, com o banco semeado por scripts/ci/fumaca-semente.sql.
// Foi a checagem manual que validou o deploy do #60: o gráfico desenha, o
// tooltip aparece ao passar o mouse e o console fica limpo. Um erro de
// hidratação, de CSP ou de tipo do Recharts (#53) aparece aqui antes da VPS.

function coletarErros(page: Page): string[] {
  const erros: string[] = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") erros.push(msg.text());
  });
  page.on("pageerror", (err) => erros.push(err.message));
  return erros;
}

test("gráfico do senador desenha e mostra tooltip, sem erro no console", async ({
  page,
}) => {
  const erros = coletarErros(page);
  await page.goto("/senador/1");

  const grafico = page.locator("svg.recharts-surface").first();
  await expect(grafico).toBeVisible({ timeout: 20_000 });

  // O tooltip do radar segue o ângulo do mouse: qualquer ponto dentro do
  // gráfico, fora do centro exato, ativa um eixo.
  const caixa = await grafico.boundingBox();
  expect(caixa).not.toBeNull();
  await page.mouse.move(
    caixa!.x + caixa!.width / 2,
    caixa!.y + caixa!.height * 0.25,
  );
  await expect(
    page.locator(".recharts-tooltip-wrapper").first(),
  ).toBeVisible();

  expect(erros).toEqual([]);
});
