import { expect, test, type Page } from "@playwright/test";

const email = process.env.PASTE_E2E_EMAIL || "test@example.com";
const password =
  process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026";

async function login(page: Page) {
  await page.goto("/mine");
  await expect(page).toHaveURL(/\/login(?:\?|$)/);
  await page.getByLabel("邮箱").waitFor({ state: "visible" });
  await page.waitForFunction(() => {
    const input = document.querySelector('form input[name="email"]');
    const form = input?.closest("form") as
      | (HTMLFormElement & { __vueParentComponent?: unknown })
      | null;
    return Boolean(form?.__vueParentComponent);
  });
  await page.getByLabel("邮箱").fill(email);
  await page.getByLabel("密码", { exact: true }).fill(password);
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/mine(?:\?|$)/, { timeout: 30_000 });
}

test("signed-in owner can create, find, edit, and delete a private Paste", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one real OIDC lifecycle is sufficient");
  const marker = Date.now().toString(36);
  const originalTitle = `私有接口示例 ${marker}`;
  const updatedTitle = `已复核接口示例 ${marker}`;

  await login(page);
  await expect(page.getByRole("heading", { level: 1, name: "我的 Paste" })).toBeVisible();
  await page.goto("/");
  await expect(page.getByRole("button", { name: /打开.+的用户菜单/ })).toBeVisible();
  await page.waitForLoadState("networkidle");
  await page.getByLabel("标题").fill(originalTitle);
  await page.getByLabel("可见性").selectOption("private");
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await expect(page.getByLabel("可见性")).toHaveValue("private");
  await page.locator(".cm-content").fill(`package private_${marker}\n`);
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await expect(page.getByLabel("可见性")).toHaveValue("private");
  const createRequest = page.waitForRequest((request) =>
    request.url().endsWith("/api/v1/pastes") && request.method() === "POST",
  );
  await page.getByRole("button", { name: "生成分享链接" }).click();
  const submitted = (await createRequest).postDataJSON();
  expect(submitted.title).toBe(originalTitle);
  expect(submitted.visibility).toBe("private");
  await expect(page.getByRole("heading", { level: 1, name: originalTitle })).toBeVisible();

  const listResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/me/pastes") && response.request().method() === "GET",
  );
  await page.getByRole("link", { name: "我的 Paste" }).click();
  const listed = await listResponse;
  expect(listed.status()).toBe(200);
  const listedPayload = await listed.json();
  expect(listedPayload.pastes.map((value: { title: string }) => value.title)).toContain(originalTitle);
  const row = page.locator(".paste-ledger-row").filter({ hasText: originalTitle });
  await expect(row).toBeVisible();
  await expect(row).toContainText("仅自己");
  await row.getByRole("link", { name: originalTitle }).focus();
  await page.screenshot({ path: testInfo.outputPath("mine-owner-ledger.png"), fullPage: true });
  await row.getByRole("link", { name: "编辑 Paste" }).click();

  await expect(page.getByRole("heading", { level: 1, name: "继续编辑" })).toBeVisible();
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await page.getByLabel("标题").fill(updatedTitle);
  await page.getByRole("button", { name: "保存修改" }).click();
  await expect(page.getByRole("heading", { level: 1, name: updatedTitle })).toBeVisible();

  await page.getByRole("link", { name: "我的 Paste" }).click();
  const updatedRow = page.locator(".paste-ledger-row").filter({ hasText: updatedTitle });
  await expect(updatedRow).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await updatedRow.getByRole("button", { name: "删除 Paste" }).click();
  await expect(updatedRow).toHaveCount(0);
});
