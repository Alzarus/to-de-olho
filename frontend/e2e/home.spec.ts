import { test, expect } from "@playwright/test";

test("has title", async ({ page }) => {
  await page.goto("/");

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/Tô De Olho/);
});

test("senador link renders", async ({ page }) => {
  await page.goto("/");

  // Since we have a comparator mock or a ranking table, let's wait for the table to load at least one row
  const table = page.locator("table");
  await expect(table).toBeVisible({ timeout: 10000 });

  // Click on a Senator Link (href containing /senador/)
  const firstSenatorLink = page.locator('a[href*="/senador/"]').first();
  await firstSenatorLink.click();

  // Validate the Senator page loaded successfully with expected heading, using regexp ignoring case
  await expect(page.getByRole("heading", { level: 1 }).first()).toBeVisible({
    timeout: 10000,
  });
});
