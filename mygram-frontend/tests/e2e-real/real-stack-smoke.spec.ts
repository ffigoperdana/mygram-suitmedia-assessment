import { expect, test } from "@playwright/test";

test("registers, logs in, and loads the feed through the real Go/PostgreSQL stack", async ({
  page,
  request,
}) => {
  const readiness = await request.get("http://127.0.0.1:8080/health/ready");
  expect(readiness.ok()).toBeTruthy();
  await expect(readiness.json()).resolves.toMatchObject({
    status: "ready",
    database: "connected",
  });

  const unique = `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  const email = `ci-${unique}@example.com`;

  await page.goto("/register");
  await page.getByLabel("Username").fill(`ci-${unique}`);
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Age").fill("25");
  await page.getByLabel("Password").fill("browser-test-password");
  await page.getByRole("button", { name: "Create account" }).click();

  await expect(page.getByText("Login to MyGram")).toBeVisible();
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill("browser-test-password");
  await page.getByRole("button", { name: "Sign in" }).click();

  await expect(page.getByRole("heading", { name: "Feed" })).toBeVisible();
  await expect(page.getByText(email)).toBeVisible();
});
