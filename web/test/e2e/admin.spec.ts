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
  await expect(page.locator("[data-paste-admin-shell]")).toBeVisible();
  await expect(
    page.getByRole("heading", { level: 1, name: "片段管理" }),
  ).toBeVisible();
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0, {
    timeout: 30_000,
  });
}

async function navigateAdmin(page: Page, name: string | RegExp) {
  const openSidebar = page.getByRole("button", { name: "打开侧边栏" });
  if (await openSidebar.isVisible().catch(() => false)) {
    await page.waitForFunction(() => {
      const toggle = document.querySelector(
        'button[aria-label="打开侧边栏"]',
      ) as (HTMLButtonElement & { __vueParentComponent?: unknown }) | null;
      return Boolean(toggle?.__vueParentComponent);
    });
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
  await page.getByRole("button", { name: "保存", exact: true }).click();
  expect((await response).status()).toBe(200);
}

async function dismissToasts(page: Page) {
  const closeButtons = page.locator('[data-slot="close"]:visible');
  while (await closeButtons.count()) await closeButtons.first().click();
}

test.beforeAll(async ({ browser }, testInfo) => {
  test.setTimeout(60_000);
  if (testInfo.project.name !== "desktop") return;
  governanceContext = await browser.newContext({ baseURL });
  const page = await governanceContext.newPage();
  await loginToAdmin(page);
  const sessionResponse = await page.request.get("/api/v1/admin/session");
  expect(sessionResponse.status()).toBe(200);
  governedUserKey = ((await sessionResponse.json()) as { userKey: string }).userKey;
  const userResponse = await page.request.get(`/api/v1/admin/users?q=${encodeURIComponent(governedUserKey)}`);
  expect(userResponse.status()).toBe(200);
  originalPolicy = ((await userResponse.json()).items as UserPolicySnapshot[])[0]
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
    const current = ((await currentResponse.json()).items as UserPolicySnapshot[])[0];
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
  await expect(page.getByRole("heading", { level: 1, name: "片段管理" })).toBeVisible();
  await expect(page.locator("[data-admin-console-breadcrumb]")).toContainText("片段管理");
  await expect(page.locator(".paste-public-header")).toHaveCount(0);

  const createdResponse = await page.request.post("/api/v1/pastes", {
    data: {
      title,
      files: [{ path: "admin-check.go", language: "go", content: `package admin_${marker}\n` }],
      visibility: "unlisted",
    },
  });
  expect(createdResponse.status()).toBe(201);

  const listResponse = page.waitForResponse((response) =>
    response.url().includes("/api/v1/admin/pastes") && response.request().method() === "GET",
  );
  await page.getByRole("textbox", { name: "搜索标题、短码、用户或标签" }).fill(title);
  expect((await listResponse).status()).toBe(200);
  const row = page.locator(".paste-admin-row").filter({ hasText: title });
  await expect(row).toBeVisible();
  const share = row.getByRole("link", { name: "打开分享链接" });
  await expect(share).toHaveAttribute("href", /\/p\//);
  const opened = page.waitForEvent("popup");
  await share.click();
  const publicPage = await opened;
  await publicPage.waitForLoadState("domcontentloaded");
  expect(publicPage.url()).toContain("/p/");
  await publicPage.close();
  await expect(page.locator(".paste-admin-inspector")).toHaveCount(0);

  await page.getByRole("checkbox", { name: `选择 ${title}` }).click();
  await page.getByRole("button", { name: "批量修改" }).click();
  await page.getByLabel("可见性").click();
  await page.getByRole("option", { name: "仅自己", exact: true }).click();
  await page.getByRole("button", { name: "应用修改" }).click();
  await expect(page.getByRole("heading", { name: "批量修改片段" })).toHaveCount(0);
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

test("user management offers state actions without policy inspectors", async ({ page }, testInfo) => {
  test.setTimeout(60_000);
  await loginToAdmin(page);
  const session = await (await page.request.get("/api/v1/admin/session")).json();
  await navigateAdmin(page, "用户管理");
  const row = page.locator(".paste-admin-user-row").filter({hasText:session.userKey});
  await expect(row).toBeVisible();
  await expect(row.getByText("创建策略", {exact:true})).toHaveCount(0);
  await expect(page.locator(".paste-admin-user-inspector")).toHaveCount(0);
  page.once("dialog", dialog => dialog.accept());
  await row.getByRole("button", {name:"暂停创建",exact:true}).click();
  try {
    await expect(row.getByRole("button", {name:"恢复创建",exact:true})).toBeVisible();
    await row.getByRole("button", {name:"恢复创建",exact:true}).click();
    await expect(row.getByRole("button", {name:"暂停创建",exact:true})).toBeVisible();
    await row.getByRole("checkbox").click();
    await page.getByRole("button", {name:"批量设置",exact:true}).click();
    const modal=page.getByRole("dialog");
    await expect(modal.getByText("每日额度",{exact:true})).toHaveCount(0);
    await modal.getByRole("button", {name:"取消",exact:true}).click();
    expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
    await page.screenshot({path:testInfo.outputPath("user-management.png"),fullPage:true});
  } finally {
    const current=(await (await page.request.get(`/api/v1/admin/users?q=${session.userKey}`)).json()).items[0];
    if(current.state!=="active") await page.request.patch(`/api/v1/admin/users/${session.userKey}`,{data:{state:"active",expectedRevision:current.revision,dailyLimitOverride:current.dailyLimitOverride,clearDailyLimit:current.dailyLimitOverride===undefined,reason:current.reason||""}});
  }
});

test("administrator filters apply drafts and cancel without changing the query", async ({page},info)=>{
 test.skip(info.project.name!=="desktop");await loginToAdmin(page);
 await page.getByRole("button",{name:"筛选",exact:true}).click();
 const dialog=page.getByRole("dialog");await expect(dialog).toBeVisible();
 await dialog.getByLabel("可见性").click();await page.getByRole("option",{name:"仅自己",exact:true}).click();
 await dialog.getByRole("button",{name:"取消",exact:true}).click();await expect(dialog).toBeHidden();
 await expect(page.getByRole("button",{name:"筛选",exact:true})).toBeVisible();
 await page.getByRole("button",{name:"筛选",exact:true}).click();await dialog.getByLabel("可见性").click();await page.getByRole("option",{name:"仅自己",exact:true}).click();
 const response=page.waitForResponse(r=>r.url().includes("/api/v1/admin/pastes")&&r.url().includes("visibility=private"));
 await dialog.getByRole("button",{name:"应用筛选"}).click();expect((await response).status()).toBe(200);await expect(dialog).toBeHidden();
 await expect(page.getByRole("button",{name:"筛选 · 1",exact:true})).toBeVisible();
});

test("administrator pagination uses bounded pages and supports page size changes",async({page},info)=>{
 const sizes:number[]=[];const pages:number[]=[];
 await page.route("**/api/v1/admin/pastes?**",async route=>{
  const url=new URL(route.request().url());const size=Number(url.searchParams.get("size")||20);const number=Number(url.searchParams.get("page")||1);sizes.push(size);pages.push(number);
  const items=Array.from({length:size},(_,i)=>({code:`p${number}-${i}`,title:`分页片段 ${(number-1)*size+i+1}`,ownerUserKey:"TestA123",tags:[],fileCount:1,primaryLanguage:"text",totalBytes:10,visibility:"unlisted",passwordProtected:false,state:"active",revision:1,createdAt:"2026-09-09T00:00:00Z",updatedAt:"2026-09-09T00:00:00Z"}));
  await route.fulfill({json:{items,total:2000,size,page:number}});
 });
 await loginToAdmin(page);await page.reload();await expect(page.getByText("分页片段 1",{exact:true})).toBeVisible();
 const bar=page.locator("[data-collection-pagination-bar]");expect(await bar.getByRole("button").count()).toBeLessThan(15);
 await bar.getByRole("button",{name:"第 2 页",exact:true}).click();await expect.poll(()=>pages).toContain(2);
 await bar.getByRole("button",{name:"最后一页",exact:true}).click();await expect.poll(()=>pages).toContain(100);
 await bar.getByRole("button",{name:"第一页",exact:true}).click();await expect(page.getByText("分页片段 1",{exact:true})).toBeVisible();
 await bar.getByRole("combobox",{name:"每页数量"}).click();await page.getByRole("option",{name:"50 条 / 页",exact:true}).click();await expect.poll(()=>sizes).toContain(50);expect(pages.at(-1)).toBe(1);
 const box=await bar.getByRole("button",{name:"第 1 页",exact:true}).boundingBox();expect(box?.height).toBe(28);
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 await page.screenshot({path:info.outputPath("admin-pagination.png"),fullPage:true});
});

test("administrator workspace adapts to a mobile viewport", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "mobile", "mobile administration contract");
  test.setTimeout(60_000);
  await page.emulateMedia({ colorScheme: "dark" });
  await loginToAdmin(page);
  await expect(page.getByRole("heading", { level: 1, name: "片段管理" })).toBeVisible();
  await expect(page.locator(".paste-admin-loading")).toHaveCount(0, {
    timeout: 30_000,
  });
  const dimensions = await page.locator("[data-paste-admin-shell]").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    documentScrollWidth: document.documentElement.scrollWidth,
  }));
  expect(dimensions.width).toBeLessThanOrEqual(dimensions.viewport);
  expect(dimensions.documentScrollWidth).toBeLessThanOrEqual(dimensions.viewport);
  const firstRow = page.locator(".paste-admin-row").first();
  await expect(firstRow.locator(".paste-admin-access")).toBeVisible();
  for (const [label, control] of [
    ["select all", page.locator(".paste-admin-select-all")],
    ["search", page.getByRole("textbox", { name: "搜索标题、短码、用户或标签" })],
    ["filters", page.getByRole("button", { name: "筛选", exact: true })],
    ["refresh", page.getByRole("button", { name: "刷新列表" })],
    ["sidebar", page.getByRole("button", { name: "打开侧边栏" })],
  ] as const) {
    const box = await control.boundingBox();
    expect(box?.height, `${label} target`).toBeGreaterThanOrEqual(label === "search" || label === "filters" ? 36 : 28);
  }
  await page.screenshot({ path: testInfo.outputPath("admin-mobile.png"), fullPage: true });
  const accessibility = await new AxeBuilder({ page })
    .exclude("[data-nuxt-devtools]")
    .exclude("[data-admin-console-breadcrumb]")
    .analyze();
  expect(accessibility.violations, accessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);

  await expect(firstRow.getByRole("link",{name:"打开分享链接"})).toBeVisible();
  await expect(page.locator(".paste-admin-inspector")).toHaveCount(0);
  await navigateAdmin(page, "站点设置");
  await expect(page.locator(".paste-admin-settings-catalog")).toBeVisible();
  await expect(page.locator(".paste-admin-brand-preview")).toHaveCount(0);
  await expect(page.getByRole("button", {name:"放弃修改",exact:true})).toHaveCount(0);
  const save = page.getByRole("button", {name:"保存",exact:true});
  await expect(save).toBeEnabled();
  await save.click();
  await expect(page.getByRole("button", {name:"已保存",exact:true})).toBeVisible();
  await expect(save).toBeVisible({timeout:2500});
  const siteName = page.locator('input[name="siteName"]');
  const originalSiteName = await siteName.inputValue();
  await siteName.fill(`${originalSiteName} mobile draft`);
  await page.reload();
  await expect(siteName).toHaveValue(originalSiteName);
  await expect(page.getByRole("link", {name:"返回首页",exact:true})).toHaveCount(0);
  await expect(page.getByRole("link", {name:"我的片段",exact:true})).toHaveCount(0);
  await page.screenshot({ path: testInfo.outputPath("admin-settings-mobile.png"), fullPage: true });
});


test("administrator loads public profiles and keeps checkboxes square", async ({ page }, testInfo) => {
  test.setTimeout(60_000);
  await loginToAdmin(page);
  const session = await (await page.request.get("/api/v1/admin/session")).json();
  const response = await page.request.get(`/identity-api/api/v1/users?ids=${encodeURIComponent(session.userKey)}`);
  expect(response.status()).toBe(200);
  const profile = (await response.json()).items[0];
  expect(profile.displayName).toBeTruthy();
  await navigateAdmin(page, "用户管理");
  const profileRow = page.locator(".paste-admin-user-row").filter({ hasText: session.userKey });
  await expect(profileRow.getByText(profile.displayName, { exact: true })).toBeVisible();
  await expect(page.getByText("用户资料加载失败", { exact: true })).toHaveCount(0);
  const checkbox = page.getByRole("checkbox", { name: "选择当前页用户" });
  const box = await checkbox.boundingBox();
  expect(box).not.toBeNull();
  expect(Math.abs(box!.width - box!.height)).toBeLessThanOrEqual(1);
  await page.screenshot({ path: testInfo.outputPath("admin-profiles.png"), fullPage: true });

  // A controlled media fixture exercises the avatar branch; profile data above is real.
  await page.route("**/identity-api/api/v1/users?**", async route => {
    const original = await route.fetch();
    const body = await original.json();
    for (const item of body.items) item.avatar = { mediaKey: "avatar-test" };
    await route.fulfill({ response: original, json: body });
  });
  await page.route("**/media/avatar-test?**", route => route.fulfill({ contentType: "image/png", body: Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=", "base64") }));
  await page.reload();
  const avatar = profileRow.getByRole("img", { name: profile.displayName, exact: true });
  await expect(avatar).toBeVisible();
  await expect.poll(() => avatar.evaluate(node => (node as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
});


test("initial setup keeps failed claims retryable", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "setup state fixture");
  test.setTimeout(60_000);
  await loginToAdmin(page);
  let claimed = false;
  let attempts = 0;
  await page.route("**/api/v1/authorization/setup", route => route.fulfill({ json: { claimed, canClaim: !claimed } }));
  await page.route("**/api/v1/authorization/setup/claim", route => {
    attempts += 1;
    if (attempts === 1) return route.fulfill({ status: 503, contentType: "application/problem+json", json: { type: "about:blank", title: "Unavailable", status: 503, code: "common.internal" } });
    claimed = true;
    return route.fulfill({ json: { claimed: true } });
  });
  // Client navigation permits a controlled setup state without changing the running site's authorization.
  await page.evaluate(async () => {
    const root = document.getElementById("__nuxt") as HTMLElement & {
      __vue_app__: { config: { globalProperties: { $router: { push: (path: string) => Promise<unknown> } } } };
    };
    await root.__vue_app__.config.globalProperties.$router.push("/setup");
  });
  const claim = page.getByRole("button", { name: "初始化站点并成为管理员", exact: true });
  await expect(claim).toBeEnabled();
  await page.screenshot({ path: testInfo.outputPath("setup-unclaimed.png"), fullPage: true });
  await claim.click();
  await expect(page.getByText("初始化未完成", { exact: true })).toBeVisible();
  await expect(claim).toBeEnabled();
  await claim.click();
  await expect(page).toHaveURL(/\/admin$/);
  expect(attempts).toBe(2);
});
