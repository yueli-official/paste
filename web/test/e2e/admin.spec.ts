import AxeBuilder from "@axe-core/playwright";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";

const email = process.env.PASTE_E2E_EMAIL || "test@example.com";
const password = process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026";
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

async function loginToAdmin(page: Page) {
  await page.goto("/admin");
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
  await page.waitForURL(/\/admin(?:\?|$)/, { timeout: 30_000 });
  await page.waitForFunction(() => {
    const toggle = document.querySelector(
      'button[aria-label="打开侧边栏"]',
    ) as (HTMLButtonElement & { __vueParentComponent?: unknown }) | null;
    return Boolean(toggle?.__vueParentComponent);
  });
}

async function navigateAdmin(page: Page, name: string | RegExp) {
  const openSidebar = page.getByRole("button", { name: "打开侧边栏" });
  if (await openSidebar.isVisible().catch(() => false)) {
    await openSidebar.click();
    const drawer = page.getByRole("dialog");
    await expect(drawer).toBeVisible();
    await drawer.getByRole("link", { name, exact: typeof name === "string" }).click();
    return;
  }
  await page.getByRole("link", { name, exact: typeof name === "string" }).click();
}

async function waitForSettingsForm(page: Page) {
  await page.locator('input[name="siteName"]').waitFor({ state: "visible" });
  await page.waitForFunction(() => {
    const input = document.querySelector('input[name="siteName"]');
    const form = input?.closest("form") as
      | (HTMLFormElement & { __vueParentComponent?: unknown })
      | null;
    return Boolean(form?.__vueParentComponent);
  });
}

async function saveSettings(page: Page) {
  const response = page.waitForResponse((candidate) =>
    candidate.url().includes("/api/v1/admin/settings") && candidate.request().method() === "PATCH",
  );
  await page.getByRole("button", { name: "保存更改" }).click();
  expect((await response).status()).toBe(200);
}

async function dismissToasts(page: Page) {
  const closeButtons = page.locator('[data-slot="close"]:visible');
  while (await closeButtons.count()) await closeButtons.first().click();
}

test.beforeAll(async ({ browser }, testInfo) => {
  if (testInfo.project.name !== "desktop") return;
  governanceContext = await browser.newContext({ baseURL });
  const page = await governanceContext.newPage();
  await loginToAdmin(page);
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

test("administrator governs snippets and updates public site settings", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one real administrator lifecycle is sufficient");
  test.setTimeout(90_000);

  const marker = Date.now().toString(36);
  const title = `后台治理示例 ${marker}`;
  await loginToAdmin(page);
  await expect(page.getByRole("heading", { level: 1, name: "片段治理" })).toBeVisible();
  await expect(page.locator("[data-admin-console-breadcrumb]")).toContainText("片段治理");
  await expect(page.locator(".paste-public-header")).toHaveCount(0);

  const createdResponse = await page.request.post("/api/v1/pastes", {
    data: {
      title,
      files: [{ path: "admin-check.go", language: "go", content: `package admin_${marker}\n` }],
      visibility: "unlisted",
    },
  });
  expect(createdResponse.status()).toBe(200);

  const listResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/pastes") && response.request().method() === "GET",
  );
  await page.getByRole("searchbox", { name: "搜索全站代码片段" }).fill(title);
  expect((await listResponse).status()).toBe(200);
  const row = page.locator(".paste-admin-row").filter({ hasText: title });
  await expect(row).toBeVisible();
  await row.getByRole("button", { name: `检查 ${title}` }).click();
  await expect(page.getByRole("heading", { level: 2, name: title })).toBeVisible();
  await expect(page.getByText("治理摘要不返回代码正文或密码材料。")).toBeVisible();

  await page.getByRole("checkbox", { name: `选择 ${title}` }).click();
  await page.getByRole("button", { name: "批量修改" }).click();
  await page.getByLabel("可见性").click();
  await page.getByRole("option", { name: "仅自己", exact: true }).click();
  await page.getByRole("button", { name: "应用修改" }).click();
  await expect(page.getByRole("heading", { name: "批量治理" })).toHaveCount(0);
  await expect(row).toContainText("仅自己");
  await page.screenshot({ path: testInfo.outputPath("admin-governance-desktop.png"), fullPage: true });

  await expect(page.locator('span[aria-hidden="true"][tabindex="0"]')).toHaveCount(0);
  const accessibility = await new AxeBuilder({ page })
    .exclude("[data-nuxt-devtools]")
    .exclude("[data-admin-console-breadcrumb]")
    .analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);

  await page.getByRole("checkbox", { name: `选择 ${title}` }).click();
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByRole("button", { name: "删除", exact: true }).click();
  await expect(row).toHaveCount(0);

  await navigateAdmin(page, "站点设置");
  await waitForSettingsForm(page);
  await expect(page.locator(".paste-admin-brand-preview")).toHaveCount(0);
  const nameInput = page.locator('input[name="siteName"]');
  const descriptionInput = page.locator('textarea[name="siteDescription"]');
  const originalName = await nameInput.inputValue();
  const originalDescription = await descriptionInput.inputValue();
  const temporaryName = `代码片段 ${marker}`;
  await nameInput.fill(temporaryName);
  await descriptionInput.fill("后台设置可以安全更新所有公开页面的展示信息。");
  await saveSettings(page);
  await expect(page.locator("[data-admin-console-brand-icon]").locator("..")).toContainText(temporaryName);
  await dismissToasts(page);
  await page.screenshot({ path: testInfo.outputPath("admin-settings-desktop.png"), fullPage: true });

  await page.goto("/");
  await expect(page.locator(".paste-editor-brand-label")).toHaveText(temporaryName);
  await page.goto("/admin?view=settings");
  await waitForSettingsForm(page);
  await nameInput.fill(originalName);
  await descriptionInput.fill(originalDescription);
  await saveSettings(page);
  await expect(page.locator("[data-admin-console-brand-icon]").locator("..")).toContainText(originalName);
});

test("admin endpoints reject anonymous requests", async ({ request }) => {
  const response = await request.get("/api/v1/admin/pastes");
  expect(response.status()).toBe(401);
  const session = await request.get("/api/v1/admin/session");
  expect(session.status()).toBe(401);
});

test("administrator governs Paste users and creation limits", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one real user-governance lifecycle is sufficient");
  test.setTimeout(90_000);
  const marker = Date.now().toString(36);
  await loginToAdmin(page);
  const sessionResponse = await page.request.get("/api/v1/admin/session");
  expect(sessionResponse.status()).toBe(200);
  const userKey = ((await sessionResponse.json()) as { userKey: string }).userKey;
  const original = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(userKey)}`);
  const originalUser = ((await original.json()).users as Array<{
    state: "active" | "suspended";
    dailyLimitOverride?: number;
    reason?: string;
    revision: number;
  }>)[0] || { state: "active" as const, revision: 0 };
  if (originalUser.state === "suspended") {
    const activated = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(userKey)}`, {
      data: {
        state: "active",
        dailyLimitOverride: originalUser.dailyLimitOverride,
        clearDailyLimit: originalUser.dailyLimitOverride === undefined,
        reason: originalUser.reason || "",
        expectedRevision: originalUser.revision,
      },
    });
    expect(activated.status()).toBe(200);
  }
  const createdResponse = await page.request.post("/api/v1/pastes", {
    data: {
      title: `用户治理示例 ${marker}`,
      files: [{ path: "governance.txt", language: "text", content: marker }],
      visibility: "unlisted",
    },
  });
  expect(createdResponse.status()).toBe(200);
  const created = (await createdResponse.json()).paste as { code: string; revision: number };
  const baselineResponse = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(userKey)}`);
  const baselineUser = ((await baselineResponse.json()).users as Array<{
    state: "active" | "suspended";
    dailyLimitOverride?: number;
    usedToday: number;
    reason?: string;
    revision: number;
  }>)[0]!;
  const limitedResponse = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(userKey)}`, {
    data: {
      state: "active",
      dailyLimitOverride: baselineUser.usedToday,
      reason: baselineUser.reason || "",
      expectedRevision: baselineUser.revision,
    },
  });
  expect(limitedResponse.status()).toBe(200);
  const limitedRevision = ((await limitedResponse.json()).user as { revision: number }).revision;
  const quotaBlocked = await page.request.post("/api/v1/pastes", {
    data: { title: "quota blocked", files: [{ path: "quota.txt", content: "quota" }], visibility: "unlisted" },
  });
  expect(quotaBlocked.status()).toBe(429);
  expect((await quotaBlocked.json()).code).toBe("common.rate_limited");
  const baselineRestored = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(userKey)}`, {
    data: {
      state: "active",
      dailyLimitOverride: baselineUser.dailyLimitOverride,
      clearDailyLimit: baselineUser.dailyLimitOverride === undefined,
      reason: baselineUser.reason || "",
      expectedRevision: limitedRevision,
    },
  });
  expect(baselineRestored.status()).toBe(200);

  await navigateAdmin(page, /^用户治理/);
  await expect(page).toHaveURL(/\/admin\?view=users$/);
  await expect(page.getByRole("searchbox", { name: "搜索 Paste 用户" })).toBeVisible();
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0);
  const userResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/users") && new URL(response.url()).searchParams.get("q") === userKey,
  );
  await page.getByRole("searchbox", { name: "搜索 Paste 用户" }).fill(userKey);
  expect((await userResponse).status()).toBe(200);
  const row = page.locator(".paste-admin-user-row").filter({ hasText: userKey });
  await expect(row).toBeVisible();
  await row.getByRole("checkbox", { name: `选择用户 ${userKey}` }).click();
  await page.getByRole("button", { name: "批量设置" }).click();
  await expect(page.getByRole("heading", { name: "批量设置创建策略" })).toBeVisible();
  await page.waitForTimeout(250);
  await page.screenshot({ path: testInfo.outputPath("admin-users-batch-desktop.png"), fullPage: true });
  await page.getByRole("button", { name: "取消", exact: true }).click();
  await page.getByRole("button", { name: "清除", exact: true }).click();

  await row.getByRole("button", { name: `检查用户 ${userKey}` }).click();
  await expect(page.getByRole("button", { name: "保存策略" })).toBeDisabled();
  await page.getByLabel("创建权限").click();
  await page.getByRole("option", { name: "暂停创建", exact: true }).click();
  if (!(await page.getByRole("checkbox", { name: "为此用户设置单独上限" }).isChecked())) {
    await page.getByRole("checkbox", { name: "为此用户设置单独上限" }).click();
  }
  await page.getByRole("spinbutton", { name: "每日创建上限" }).fill("9876");
  await page.getByLabel("治理说明").fill(`自动化验收 ${marker}`);
  await expect(page.getByRole("button", { name: "保存策略" })).toBeEnabled();
  const savePolicy = page.waitForResponse((response) =>
    response.url().includes(`/api/v1/admin/users/${encodeURIComponent(userKey)}`) && response.request().method() === "PATCH",
  );
  await page.getByRole("button", { name: "保存策略" }).click();
  expect((await savePolicy).status()).toBe(200);
  await expect(row).toContainText("已暂停");
  const blocked = await page.request.post("/api/v1/pastes", {
    data: { title: "blocked", files: [{ path: "blocked.txt", content: "blocked" }], visibility: "unlisted" },
  });
  expect(blocked.status()).toBe(403);
  expect((await blocked.json()).code).toBe("paste.creation_suspended");
  await page.screenshot({ path: testInfo.outputPath("admin-users-desktop.png"), fullPage: true });

  const current = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(userKey)}`);
  const currentUser = ((await current.json()).users as Array<{ revision: number }>)[0]!;
  const restored = await page.request.patch(`/api/v1/admin/users/${encodeURIComponent(userKey)}`, {
    data: {
      state: originalUser.state,
      dailyLimitOverride: originalUser.dailyLimitOverride,
      clearDailyLimit: originalUser.dailyLimitOverride === undefined,
      reason: originalUser.reason || "",
      expectedRevision: currentUser.revision,
    },
  });
  expect(restored.status()).toBe(200);
  const removed = await page.request.delete(`/api/v1/admin/pastes/${created.code}?expectedRevision=${created.revision}`);
  expect(removed.status()).toBe(200);

  const governanceResponse = await page.request.get("/api/v1/admin/governance-settings");
  expect(governanceResponse.status()).toBe(200);
  const originalSettings = (await governanceResponse.json()).settings as {
    userDailyLimit: number;
    anonymousDailyLimit: number;
    revision: number;
  };
  await navigateAdmin(page, "站点设置");
  await expect(page.getByRole("spinbutton", { name: "登录用户每日默认上限" })).toBeVisible();
  const userLimitInput = page.getByRole("spinbutton", { name: "登录用户每日默认上限" });
  await expect(page.getByRole("button", { name: "已保存" })).toBeDisabled();
  await userLimitInput.fill(String(originalSettings.userDailyLimit + 1));
  await expect(page.getByRole("button", { name: "保存更改" })).toBeEnabled();
  const saveLimits = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/governance-settings") && response.request().method() === "PATCH",
  );
  await page.getByRole("button", { name: "保存更改" }).click();
  const changedSettings = await saveLimits;
  expect(changedSettings.status()).toBe(200);
  const changedRevision = ((await changedSettings.json()).settings as { revision: number }).revision;
  const restoredSettings = await page.request.patch("/api/v1/admin/governance-settings", {
    data: { ...originalSettings, expectedRevision: changedRevision },
  });
  expect(restoredSettings.status()).toBe(200);
});

test("user governance stays operable at 320px", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "mobile", "320px user-governance contract");
  await page.setViewportSize({ width: 320, height: 568 });
  const userBatchBodies: Array<Record<string, unknown>> = [];
  await page.route("**/api/v1/admin/users**", async (route) => {
    if (route.request().method() === "PATCH") {
      const body = route.request().postDataJSON() as Record<string, unknown>;
      userBatchBodies.push(body);
      const userKey = decodeURIComponent(new URL(route.request().url()).pathname.split("/").pop() || "");
      if (userKey === "usr_SUSPEND") {
        await route.fulfill({ status: 409, json: { type: "https://errors.yuelili.com/problems/paste.conflict", status: 409, code: "paste.conflict", traceId: "mock-conflict" } });
        return;
      }
      await route.fulfill({ json: { user: { userKey, ...body, revision: Number(body.expectedRevision) + 1, updatedAt: "2026-08-11T10:00:00Z" } } });
      return;
    }
    if (route.request().method() !== "GET") return route.continue();
    await route.fulfill({
      json: {
        users: [
          { userKey: "usr_ABUSE01", state: "active", effectiveDailyLimit: 50, usedToday: 48, totalPastes: 91, activePastes: 72, lastCreatedAt: "2026-08-11T09:30:00Z", revision: 0 },
          { userKey: "usr_SUSPEND", state: "suspended", dailyLimitOverride: 4, effectiveDailyLimit: 4, usedToday: 4, totalPastes: 14, activePastes: 9, lastCreatedAt: "2026-08-11T08:00:00Z", reason: "异常批量创建", revision: 2 },
        ],
        total: 2,
        limit: 50,
        offset: 0,
      },
    });
  });
  await page.route("**/api/v1/admin/governance-settings", async (route) => {
    if (route.request().method() !== "GET") return route.continue();
    await route.fulfill({ json: { settings: { userDailyLimit: 50, anonymousDailyLimit: 200, revision: 1, updatedAt: "2026-08-11T00:00:00Z" } } });
  });
  await loginToAdmin(page);
  await navigateAdmin(page, /^用户治理/);
  await expect(page).toHaveURL(/\/admin\?view=users$/);
  await expect(page.locator(".paste-admin-user-row")).toHaveCount(2);
  const dimensions = await page.locator("[data-paste-admin-shell]").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    documentScrollWidth: document.documentElement.scrollWidth,
  }));
  expect(dimensions.width).toBeLessThanOrEqual(dimensions.viewport);
  expect(dimensions.documentScrollWidth).toBeLessThanOrEqual(dimensions.viewport);
  const pause = page.getByRole("button", { name: "暂停创建", exact: true }).first();
  const pauseBox = await pause.boundingBox();
  expect(pauseBox?.height).toBeGreaterThanOrEqual(44);
  const selectPage = page.getByRole("checkbox", { name: "选择当前页用户" });
  await selectPage.click();
  await expect(page.getByText("2 个用户已选择", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "批量设置" }).click();
  await page.getByLabel("创建权限").click();
  await page.getByRole("option", { name: "正常创建", exact: true }).click();
  await page.getByLabel("每日额度").click();
  await page.getByRole("option", { name: "设置统一上限", exact: true }).click();
  await page.getByRole("spinbutton", { name: "统一每日上限" }).fill("80");
  await page.screenshot({ path: testInfo.outputPath("admin-users-batch-mobile.png"), fullPage: true });
  await page.getByRole("button", { name: "应用到 2 个用户" }).click();
  await expect.poll(() => userBatchBodies.length).toBe(2);
  expect(userBatchBodies.every((body) => body.state === "active" && body.dailyLimitOverride === 80)).toBe(true);
  await expect(page.getByRole("button", { name: "应用到 1 个用户" })).toBeVisible();
  await page.getByRole("button", { name: "取消", exact: true }).click();
  await expect(page.locator('span[aria-hidden="true"][tabindex="0"]')).toHaveCount(0);
  await expect(page.getByText("1 个用户已选择", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "清除", exact: true }).click();
  await page.locator(".paste-admin-user-row").first().getByRole("button", { name: /检查用户/ }).click();
  await expect(page.getByRole("heading", { level: 2, name: "usr_ABUSE01" })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("admin-users-mobile.png"), fullPage: true });
  const accessibility = await new AxeBuilder({ page })
    .exclude("[data-nuxt-devtools]")
    .exclude("[data-admin-console-breadcrumb]")
    .analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);
  await page.getByRole("button", { name: "关闭用户检查器" }).click();
  await navigateAdmin(page, "站点设置");
  const anonymousLimit = page.getByRole("spinbutton", { name: "匿名创建每日全站总额" });
  await expect(anonymousLimit).toBeVisible();
  await anonymousLimit.scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath("admin-limits-mobile.png"), fullPage: true });
});

test("administrator filters open without runtime errors", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one browser regression signal is sufficient");
  await loginToAdmin(page);
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0);
  const runtimeErrors: string[] = [];
  page.on("pageerror", (error) => runtimeErrors.push(error.message));
  page.on("console", (message) => {
    if (message.type() === "error") runtimeErrors.push(message.text());
  });

  await page.getByRole("button", { name: "筛选", exact: true }).click();
  await page.waitForTimeout(50);
  expect(runtimeErrors).toEqual([]);
  const panel = page.locator(".paste-admin-filter-panel");
  await expect(panel).toBeVisible();

  const privateResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/pastes") && response.url().includes("visibility=private"),
  );
  await panel.getByLabel("可见性").click();
  await page.getByRole("option", { name: "仅自己", exact: true }).click();
  expect((await privateResponse).status()).toBe(200);
  await expect(page.getByRole("button", { name: "筛选 1", exact: true })).toBeVisible();

  await expect(panel).toBeVisible();
  const clearResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/pastes") && !response.url().includes("visibility="),
  );
  await panel.getByLabel("可见性").click();
  await page.getByRole("option", { name: "全部可见性", exact: true }).click();
  expect((await clearResponse).status()).toBe(200);
  expect(runtimeErrors).toEqual([]);
});

test("administrator can jump between numbered result pages", async ({ page }, testInfo) => {
  const constrainedMobile = testInfo.project.name === "mobile";
  if (constrainedMobile) await page.setViewportSize({ width: 320, height: 568 });
  const offsets: number[] = [];
  await page.route("**/api/v1/admin/pastes**", async (route) => {
    if (route.request().method() !== "GET") return route.continue();
    const requestURL = new URL(route.request().url());
    const offset = Number(requestURL.searchParams.get("offset") || 0);
    offsets.push(offset);
    const count = Math.min(50, Math.max(0, 121 - offset));
    const pastes = Array.from({ length: count }, (_, index) => {
      const number = offset + index + 1;
      return {
        code: `P${String(number).padStart(7, "0")}`,
        shareUrl: `/p/P${String(number).padStart(7, "0")}`,
        title: `分页片段 ${number}`,
        tags: [],
        fileCount: 1,
        primaryLanguage: "text",
        visibility: "unlisted",
        passwordProtected: false,
        state: "active",
        revision: 1,
        createdAt: "2026-08-11T00:00:00Z",
        updatedAt: "2026-08-11T00:00:00Z",
      };
    });
    await route.fulfill({ json: { pastes, total: 121, limit: 50, offset } });
  });

  await loginToAdmin(page);
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "第 2 页", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "第 2 页", exact: true }).click();
  await expect.poll(() => offsets).toContain(50);
  await expect(page.getByRole("button", { name: "第 2 页", exact: true })).toHaveAttribute("aria-current", "page");
  await expect(page.locator(".paste-admin-page-status")).toContainText("51-100 / 121");
  const statusbar = await page.locator(".paste-admin-statusbar").evaluate((node) => ({
    width: node.clientWidth,
    scrollWidth: node.scrollWidth,
  }));
  expect(statusbar.scrollWidth).toBeLessThanOrEqual(statusbar.width);
  if (constrainedMobile) {
    for (const control of [
      page.getByRole("button", { name: "上一页" }),
      page.getByRole("button", { name: "第 2 页", exact: true }),
      page.getByRole("button", { name: "下一页" }),
    ]) {
      const box = await control.boundingBox();
      expect(box?.width).toBeGreaterThanOrEqual(44);
      expect(box?.height).toBeGreaterThanOrEqual(44);
    }
    await page.getByRole("button", { name: "第 3 页", exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect.poll(() => offsets).toContain(100);
    await expect(page.getByRole("button", { name: "第 3 页", exact: true })).toHaveAttribute("aria-current", "page");
  }
  await page.screenshot({ path: testInfo.outputPath("admin-pagination.png"), fullPage: true });
});

test("administrator workspace adapts to a mobile viewport", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "mobile", "mobile administration contract");
  await page.emulateMedia({ colorScheme: "dark" });
  await loginToAdmin(page);
  await expect(page.getByRole("heading", { level: 1, name: "片段治理" })).toBeVisible();
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0);
  const dimensions = await page.locator("[data-paste-admin-shell]").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    documentScrollWidth: document.documentElement.scrollWidth,
  }));
  expect(dimensions.width).toBeLessThanOrEqual(dimensions.viewport);
  expect(dimensions.documentScrollWidth).toBeLessThanOrEqual(dimensions.viewport);
  const firstRow = page.locator(".paste-admin-row").first();
  await expect(firstRow.locator(".paste-admin-access")).toBeVisible();
  for (const control of [
    page.locator(".paste-admin-select-all"),
    page.getByRole("searchbox", { name: "搜索全站代码片段" }),
    page.getByRole("button", { name: "筛选", exact: true }),
    page.getByRole("button", { name: "刷新列表" }),
    page.getByRole("button", { name: "打开侧边栏" }),
  ]) {
    const box = await control.boundingBox();
    expect(box?.height).toBeGreaterThanOrEqual(44);
  }
  await page.screenshot({ path: testInfo.outputPath("admin-mobile.png"), fullPage: true });
  const accessibility = await new AxeBuilder({ page })
    .exclude("[data-nuxt-devtools]")
    .exclude("[data-admin-console-breadcrumb]")
    .analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);

  await firstRow.getByRole("button", { name: /^检查/ }).click();
  await expect(page.locator(".paste-admin-inspector")).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("admin-inspector-mobile.png"), fullPage: true });
  await page.getByRole("button", { name: "关闭检查器" }).click();
  await navigateAdmin(page, "站点设置");
  await expect(page.locator(".paste-admin-settings-catalog")).toBeVisible();
  await expect(page.locator(".paste-admin-brand-preview")).toHaveCount(0);
  const resetSettings = page.getByRole("button", { name: "放弃修改" });
  const siteName = page.locator('input[name="siteName"]');
  const originalSiteName = await siteName.inputValue();
  const resetBox = await resetSettings.boundingBox();
  expect(resetBox?.width).toBeGreaterThanOrEqual(44);
  expect(resetBox?.height).toBeGreaterThanOrEqual(44);
  await siteName.fill(`${originalSiteName} mobile draft`);
  await expect(resetSettings).toBeEnabled();
  await resetSettings.click();
  await expect(siteName).toHaveValue(originalSiteName);
  await expect(resetSettings).toBeDisabled();
  await page.screenshot({ path: testInfo.outputPath("admin-settings-mobile.png"), fullPage: true });
});
