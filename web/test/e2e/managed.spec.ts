import AxeBuilder from "@axe-core/playwright";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";

const email = process.env.PASTE_E2E_EMAIL || "test@example.com";
const password =
  process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026";
const baseURL = process.env.PASTE_E2E_BASE_URL || "http://localhost:3010";

interface UserPolicySnapshot {
  state: "active" | "suspended";
  dailyLimitOverride?: number;
  reason?: string;
  revision: number;
}

let governanceContext: BrowserContext | undefined;
let governedUserKey = "";
let originalPolicy: UserPolicySnapshot | undefined;

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

async function createOwnedPaste(page: Page, title: string, content: string) {
  await page.goto("/");
  await expect(page.getByRole("button", { name: /打开.+的用户菜单/ })).toBeVisible();
  await page.waitForLoadState("networkidle");
  await page.getByRole("textbox", { name: "main.go 代码编辑器" }).fill(content);
  await page.getByRole("button", { name: "打开分享设置" }).click();
  await page.getByLabel("标题").fill(title);
  await page.getByRole("button", { name: "生成分享链接" }).click();
  await expect(page.getByRole("heading", { level: 1, name: title })).toBeVisible();
}

test.beforeAll(async ({ browser }, testInfo) => {
  if (testInfo.project.name !== "desktop") return;
  governanceContext = await browser.newContext({ baseURL });
  const page = await governanceContext.newPage();
  await login(page);
  const sessionResponse = await page.request.get("/api/v1/admin/session");
  expect(sessionResponse.status()).toBe(200);
  governedUserKey = ((await sessionResponse.json()) as { userKey: string }).userKey;
  const userResponse = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(governedUserKey)}`);
  expect(userResponse.status()).toBe(200);
  originalPolicy = ((await userResponse.json()).users as UserPolicySnapshot[])[0]
    || { state: "active", revision: 0 };
  const capacityResponse = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(governedUserKey)}`, {
    data: {
      state: "active",
      dailyLimitOverride: 10_000,
      reason: originalPolicy.reason || "",
      expectedRevision: originalPolicy.revision,
    },
  });
  expect(capacityResponse.status()).toBe(200);
});

test.afterAll(async () => {
  if (!governanceContext || !governedUserKey || !originalPolicy) return;
  const page = governanceContext.pages()[0] || await governanceContext.newPage();
  const currentResponse = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(governedUserKey)}`);
  if (currentResponse.ok()) {
    const current = ((await currentResponse.json()).users as UserPolicySnapshot[])[0];
    if (current) {
      const restored = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(governedUserKey)}`, {
        data: {
          state: originalPolicy.state,
          dailyLimitOverride: originalPolicy.dailyLimitOverride,
          clearDailyLimit: originalPolicy.dailyLimitOverride === undefined,
          reason: originalPolicy.reason || "",
          expectedRevision: current.revision,
        },
      });
      expect(restored.status()).toBe(200);
    }
  }
  await governanceContext.close();
});

test("signed-in owner can create, find, edit, and delete a private Paste", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one real OIDC lifecycle is sufficient");
  const marker = Date.now().toString(36);
  const originalTitle = `私有接口示例 ${marker}`;
  const updatedTitle = `已复核接口示例 ${marker}`;

  await login(page);
  await expect(page.getByRole("heading", { level: 1, name: "我的片段" })).toBeVisible();
  await expect(page.locator(".paste-public-header")).toHaveCount(0);
  const mineStatusFontSize = await page.locator(".paste-mine-statusbar").evaluate((node) =>
    Number.parseFloat(getComputedStyle(node).fontSize),
  );
  expect(mineStatusFontSize).toBeGreaterThanOrEqual(12);
  await page.goto("/");
  await expect(page.getByRole("button", { name: /打开.+的用户菜单/ })).toBeVisible();
  await page.waitForLoadState("networkidle");
  await page.screenshot({ path: testInfo.outputPath("composer-owner.png"), fullPage: true });
  await page.getByRole("button", { name: "打开分享设置" }).click();
  await page.getByLabel("标题").fill(originalTitle);
  await page.getByLabel("可见性").click();
  await page.getByRole("option", { name: "仅自己可访问" }).click();
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await expect(page.getByLabel("可见性")).toContainText("仅自己可访问");
  await page.locator(".cm-content").fill(`package private_${marker}\n`);
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await expect(page.getByLabel("可见性")).toContainText("仅自己可访问");
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
  await page.getByRole("link", { name: "我的片段" }).click();
  const listed = await listResponse;
  expect(listed.status()).toBe(200);
  const listedPayload = await listed.json();
  expect(listedPayload.pastes.map((value: { title: string }) => value.title)).toContain(originalTitle);
  const row = page.locator(".paste-ledger-row").filter({ hasText: originalTitle });
  await expect(row).toBeVisible();
  await expect(row).toContainText("仅自己");
  await row.getByRole("link", { name: originalTitle }).focus();
  await page.screenshot({ path: testInfo.outputPath("mine-owner-ledger.png"), fullPage: true });
  await expect(page.locator('span[aria-hidden="true"][tabindex="0"]')).toHaveCount(0);
  const accessibility = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);
  await row.getByRole("link", { name: "编辑代码片段" }).click();

  await expect(page.getByRole("heading", { level: 1, name: "编辑片段" })).toBeVisible();
  await page.getByRole("button", { name: "打开分享设置" }).click();
  await expect(page.getByLabel("标题")).toHaveValue(originalTitle);
  await page.getByLabel("标题").fill(updatedTitle);
  await page.getByRole("button", { name: "保存修改" }).click();
  await expect(page.getByRole("heading", { level: 1, name: updatedTitle })).toBeVisible();

  await page.getByRole("link", { name: "我的片段" }).click();
  const updatedRow = page.locator(".paste-ledger-row").filter({ hasText: updatedTitle });
  await expect(updatedRow).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await updatedRow.getByRole("button", { name: "删除代码片段" }).click();
  await expect(updatedRow).toHaveCount(0);
});

test("owner can batch update and delete selected Pastes", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one real OIDC batch lifecycle is sufficient");
  test.setTimeout(90_000);
  const marker = Date.now().toString(36);
  const titles = [`批量示例 A ${marker}`, `批量示例 B ${marker}`];

  await login(page);
  await createOwnedPaste(page, titles[0], `const batchA = "${marker}";\n`);
  await createOwnedPaste(page, titles[1], `const batchB = "${marker}";\n`);
  await page.getByRole("link", { name: "我的片段" }).click();

  const rows = titles.map((title) => page.locator(".paste-ledger-row").filter({ hasText: title }));
  for (const [index, title] of titles.entries()) {
    await expect(rows[index]).toBeVisible();
    await page.getByRole("checkbox", { name: `选择 ${title}` }).click();
  }
  await expect(page.getByText("已选择 2 项", { exact: true }).first()).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  const narrowDimensions = await page.locator(".paste-mine-shell").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    scrollWidth: document.documentElement.scrollWidth,
  }));
  expect(narrowDimensions.width).toBeLessThanOrEqual(narrowDimensions.viewport);
  expect(narrowDimensions.scrollWidth).toBeLessThanOrEqual(narrowDimensions.viewport);
  await expect(page.getByRole("button", { name: "批量修改" })).toBeVisible();
  await expect(page.getByRole("button", { name: "删除", exact: true })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("mine-batch-selected-mobile.png"), fullPage: true });
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.getByRole("button", { name: "批量修改" }).click();
  await expect(page.getByRole("heading", { name: "批量修改" })).toBeVisible();
  await page.getByLabel("批量可见性").click();
  await page.getByRole("option", { name: "仅自己", exact: true }).click();

  const patchBodies: Array<Record<string, unknown>> = [];
  page.on("request", (request) => {
    if (request.method() === "PATCH" && request.url().includes("/api/v1/me/pastes/")) {
      patchBodies.push(request.postDataJSON());
    }
  });
  const accessibility = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);
  await page.getByRole("button", { name: "应用修改" }).click();
  await expect(page.getByRole("heading", { name: "批量修改" })).toHaveCount(0);
  for (const row of rows) await expect(row).toContainText("仅自己");
  expect(patchBodies).toHaveLength(2);
  for (const body of patchBodies) {
    expect(body.visibility).toBe("private");
    expect(body).not.toHaveProperty("password");
    expect(body).not.toHaveProperty("files");
  }

  for (const title of titles) {
    await page.getByRole("checkbox", { name: `选择 ${title}` }).click();
  }
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByRole("button", { name: "删除", exact: true }).click();
  for (const row of rows) await expect(row).toHaveCount(0);
  await page.reload();
  await expect(page.locator(".paste-ledger-loading")).toHaveCount(0);
  for (const row of rows) await expect(row).toHaveCount(0);
});

test("my Paste keeps the editor workspace at mobile width", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "mobile", "mobile shell contract");
  await page.emulateMedia({ colorScheme: "dark" });
  await login(page);
  await expect(page.getByRole("heading", { level: 1, name: "我的片段" })).toBeVisible();
  await expect(page.locator(".paste-public-header")).toHaveCount(0);
  await expect(page.locator(".paste-ledger-loading")).toHaveCount(0);
  const dimensions = await page.locator(".paste-mine-shell").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    scrollWidth: document.documentElement.scrollWidth,
  }));
  expect(dimensions.width).toBeLessThanOrEqual(dimensions.viewport);
  expect(dimensions.scrollWidth).toBeLessThanOrEqual(dimensions.viewport);
  await page.screenshot({ path: testInfo.outputPath("mine-mobile.png"), fullPage: true });
  const accessibility = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);
});
