import { expect, test } from "@playwright/test";

test("owner library and reader support finding, cancelling, and deleting links", async ({ page, browser }, testInfo) => {
  test.setTimeout(90_000);
  await page.goto("/auth/login?return_to=%2Fmine");
  await page.getByLabel("邮箱").waitFor();
  await page.waitForFunction(() => Boolean((document.querySelector('form input[name="email"]')?.closest('form') as HTMLFormElement & { __vueParentComponent?: unknown })?.__vueParentComponent));
  await page.getByLabel("邮箱").fill(process.env.PASTE_E2E_EMAIL || "test@example.com");
  await page.getByLabel("密码", { exact: true }).fill(process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026");
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/mine(?:\?|$)/, { timeout: 30_000 });
  const marker = `library-${Date.now()}`;
  const created: Array<{ code: string; revision: number }> = [];
  const origin = new URL(page.url()).origin;
  try {
    for (const [index, title] of ["Go · 请求日志中间件", "Markdown · 部署检查清单", "一个很长的标题，用于确认内容库窄屏中的排版与操作仍然完整可用"].entries()) {
      const response = await page.request.post("/api/v1/pastes", { data: {
        title, description: "", tags: [marker], visibility: index === 1 ? "private" : "unlisted", password: "",
        files: [{ path: index === 1 ? "deploy.md" : "logging.go", language: index === 1 ? "markdown" : "go", content: index === 1 ? "# Deployment\n\n- Check health" : "package main\n\nfunc main() {}\n" }],
      } });
      expect(response.ok()).toBeTruthy();
      created.push((await response.json()).paste);
    }
    await page.goto(`/mine?q=${marker}`);
    await expect(page.locator(".paste-ledger-row")).toHaveCount(3);
    await page.getByRole("link", { name: "Go · 请求日志中间件", exact: true }).focus();
    for (const theme of ["dark", "light"] as const) {
      await page.evaluate((value) => { localStorage.setItem("nuxt-color-mode", value); }, theme);
      await page.reload();
      await expect(page.locator(".paste-ledger-row")).toHaveCount(3);
      await page.screenshot({ path: testInfo.outputPath(`library-${theme}.png`), fullPage: true });
    }
    await page.getByLabel("搜索代码片段").fill(`${marker}-missing`);
    await expect(page.getByText("没有找到匹配的内容", { exact: true })).toBeVisible();
    await page.getByLabel("搜索代码片段").fill(marker);
    await expect(page.locator(".paste-ledger-row")).toHaveCount(3);
    await page.goto(`/p/${created[0]!.code}`);
    await expect(page.getByRole("button", { name: "删除此链接" })).toBeVisible();
    await expect(page.getByRole("link", { name: "编辑此内容" })).toBeVisible();
    await expect(page.locator(".cm-content")).toContainText("func main()");
    await page.screenshot({ path: testInfo.outputPath("owner-reader.png"), fullPage: true });
    await page.locator(".paste-file-tabs").screenshot({ path: testInfo.outputPath("filename.png") });
    const guest = await browser.newContext();
    try {
      const other = await guest.newPage();
      await other.goto(`${origin}/p/${created[0]!.code}`);
      await expect(other.locator(".cm-content")).toBeVisible();
      await expect(other.getByRole("button", { name: "删除此链接" })).toHaveCount(0);
      const rejected = await other.request.delete(`${origin}/api/v1/me/pastes/${created[0]!.code}?expectedRevision=1`);
      expect([401, 403]).toContain(rejected.status());
    } finally { await guest.close(); }
    await page.getByRole("button", { name: "删除此链接" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "取消", exact: true }).click();
    await expect(page.locator(".cm-content")).toBeVisible();
    await page.getByRole("button", { name: "删除此链接" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "确认删除", exact: true }).click();
    await expect(page).toHaveURL(/\/mine$/);
    const removed = created.shift()!;
    await page.goto(`/p/${removed.code}`);
    await expect(page.getByRole("heading", { name: "无法打开这个片段" })).toBeVisible();
  } finally {
    for (const item of created) await page.request.delete(`/api/v1/me/pastes/${item.code}?expectedRevision=${item.revision}`);
  }
});
