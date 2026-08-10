import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("editor-first composer works at desktop and mobile widths", async ({ page }, testInfo) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: /把代码放好/ })).toBeVisible();
  await expect(page.getByRole("textbox", { name: /代码编辑器/ })).toBeVisible();
  await expect(page.locator(".cm-editor")).toBeVisible();
  await page.getByRole("button", { name: "添加文件" }).click();
  await expect(page.getByRole("button", { name: /untitled-2.txt/ })).toBeVisible();
  await page.getByLabel("当前文件名").fill("README.md");
  await page.getByLabel("当前文件名").press("Tab");
  await expect(page.getByLabel("代码语言")).toHaveValue("markdown");
  await page.locator(".cm-content").fill("# Paste\n\nA real multi-file share.");

  if (testInfo.project.name === "mobile") {
    const workbenchWidth = await page.locator(".paste-workbench").evaluate((node) => node.getBoundingClientRect().width);
    expect(workbenchWidth).toBeLessThanOrEqual(await page.evaluate(() => innerWidth));
  }

  await page.screenshot({ path: testInfo.outputPath(`composer-${testInfo.project.name}.png`), fullPage: true });
  const results = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(results.violations, results.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);

  await page.getByRole("button", { name: "生成分享链接" }).click();
  await expect(page).toHaveURL(/\/p\/[A-Za-z0-9_-]{8}\?created=1$/);
  await expect(page.getByRole("heading", { level: 1, name: "未命名 Paste" })).toBeVisible();
  await expect(page.getByRole("button", { name: /README.md/ })).toBeVisible();
  await expect(page.getByRole("button", { name: "复制全部" })).toBeVisible();
});

test("password access stays out of the URL", async ({ page, request }) => {
  const created = await request.post("/api/v1/pastes", {
    data: {
      files: [{ path: "secret.txt", language: "text", content: "bounded secret" }],
      visibility: "unlisted",
      password: "safepass123",
    },
  });
  expect(created.status()).toBe(200);
  const payload = await created.json();
  const code = payload.paste.code as string;

  await page.goto(`/p/${code}`);
  await expect(page.getByRole("heading", { level: 1, name: "需要访问密码" })).toBeVisible();
  await page.getByRole("textbox", { name: "访问密码", exact: true }).fill("wrong-pass");
  await page.getByRole("button", { name: "打开 Paste" }).click();
  await expect(page.getByRole("alert")).toContainText("密码不正确");
  await page.getByRole("textbox", { name: "访问密码", exact: true }).fill("safepass123");
  await page.getByRole("button", { name: "打开 Paste" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "未命名 Paste" })).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`/p/${code}$`));
  await expect(page.getByText("bounded secret")).toBeVisible();
});
