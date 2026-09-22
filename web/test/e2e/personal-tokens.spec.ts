import { expect, request, test, type APIRequestContext, type Page } from "@playwright/test";

// PAT plaintext stays in memory, including when a test fails.
test.use({ trace: "off", video: "off", screenshot: "off" });

async function login(page: Page, password: string) {
  await page.goto("/mine");
  await expect(page).toHaveURL(/\/login(?:\?|$)/);
  await page.getByLabel("邮箱").waitFor({ state: "visible" });
  await page.waitForFunction(() => {
    const input = document.querySelector('form input[name="email"]');
    return Boolean((input?.closest("form") as HTMLElement & { __vueParentComponent?: unknown })?.__vueParentComponent);
  });
  await page.getByLabel("邮箱").fill(process.env.PASTE_E2E_EMAIL || "test@example.com");
  await page.getByLabel("密码", { exact: true }).fill(password);
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/mine(?:\?|$)/, { timeout: 30_000 });
}

async function json(client: APIRequestContext, method: string, path: string, status: number, data?: unknown) {
  const response = await client.fetch(path, { method, data });
  expect(response.status(), `${method} ${path}: ${response.status() === status ? "" : await response.text()}`).toBe(status);
  if (status === 204) {
    expect(await response.text()).toBe("");
    return;
  }
  const body = await response.json();
  if (status >= 400) {
    expect(response.headers()["content-type"]).toContain("application/problem+json");
    expect(body.status).toBe(status);
    expect(body.traceId).toBeTruthy();
  }
  return body;
}

test("developer token catalog and Paste lifecycle keep explicit scope boundaries", async ({ page, baseURL }, testInfo) => {
  test.setTimeout(180_000);
  const account = process.env.PASTE_E2E_ACCOUNT_URL || "http://account-paste.dev.yuelili.test:3803";
  const api = process.env.PASTE_E2E_API_URL || "http://127.0.0.1:8293";
  const password = process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026";
  const marker = `${Date.now().toString(36)}-${testInfo.project.name}`;
  const scope = (key: string) => `site:${Buffer.from("paste-yueli-web").toString("base64url")}:${key}`;
  const keys = [
    "paste.paste.create",
    "paste.paste.read",
    "paste.paste.update",
    "paste.paste.delete",
    "paste.paste.moderate",
    "paste.settings.manage",
  ];
  const tokens: { id: number; owner: APIRequestContext }[] = [];
  const clients: APIRequestContext[] = [];
  let createdCode = "";
  let createdRevision = 0;
  let originalSettings: { name: string; description: string; revision: number } | undefined;

  async function clientFor(token: string, target = baseURL!) {
    const client = await request.newContext({ baseURL: target, extraHTTPHeaders: { authorization: `Bearer ${token}` } });
    clients.push(client);
    return client;
  }

  async function tokenFor(...selected: string[]) {
    const token = await json(page.request, "POST", `${account}/api/v1/pat`, 201, {
      name: `Paste PAT ${marker}`,
      scopes: selected.map(scope),
      expiresInDays: 1,
    });
    tokens.push({ id: token.id, owner: page.request });
    return { ...token, client: await clientFor(token.token) };
  }

  await login(page, password);
  try {
    const catalog = await json(page.request, "GET", `${account}/api/v1/pat/scopes`, 200);
    const pasteScopes = catalog.items.filter((item: { site: string }) => item.site === "paste-yueli-web");
    expect(pasteScopes.map((item: { key: string }) => item.key).sort()).toEqual(keys.map(scope).sort());
    expect(catalog.unavailableSites).not.toContain("Paste");

    await page.goto(`${account}/developer-tokens`);
    const create = page.getByRole("button", { name: "创建令牌", exact: true });
    await expect.poll(() => create.evaluate((node) => Boolean((node as HTMLElement & { __vueParentComponent?: unknown }).__vueParentComponent))).toBe(true);
    await create.click();
    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("名称", { exact: true }).fill(`Paste 完整令牌 ${marker}`);
    for (const permission of pasteScopes) {
      await dialog.getByRole("checkbox", { name: permission.label, exact: true }).check();
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath("developer-token-permissions.png"), fullPage: true });
    const createdTokenResponse = page.waitForResponse((response) => response.url().endsWith("/api/v1/pat") && response.request().method() === "POST");
    await dialog.getByRole("button", { name: "创建令牌", exact: true }).click();
    const createdToken = await createdTokenResponse;
    expect(createdToken.status()).toBe(201);
    const fullToken = await createdToken.json();
    tokens.push({ id: fullToken.id, owner: page.request });
    const full = await clientFor(fullToken.token);
    const direct = await clientFor(fullToken.token, api);
    const createOnly = (await tokenFor("paste.paste.create")).client;

    await json(createOnly, "GET", "/api/v1/me/pastes", 403);
    await json(createOnly, "GET", "/api/v1/admin/pastes", 403);
    for (const path of [
      "/api/v1/admin/session",
      "/api/v1/admin/users",
      "/api/v1/admin/governance-settings",
      "/api/v1/internal/personal-token/permissions?userKey=TestA123",
    ]) {
      await json(full, "GET", path, 403);
    }
    await json(full, "POST", "/api/v1/authorization/setup/claim", 403);

    const created = await json(full, "POST", "/api/v1/pastes", 201, {
      title: `令牌片段 ${marker}`,
      visibility: "private",
      files: [{ path: "main.go", language: "go", content: `package token_${marker.replaceAll("-", "_")}` }],
    });
    createdCode = created.paste.code;
    createdRevision = created.paste.revision;
    expect((await json(full, "GET", `/api/v1/me/pastes?q=${encodeURIComponent(marker)}`, 200)).items.map((item: { code: string }) => item.code)).toContain(createdCode);
    expect((await json(direct, "GET", `/api/v1/me/pastes/${createdCode}`, 200)).paste.code).toBe(createdCode);
    const updated = await json(full, "PATCH", `/api/v1/me/pastes/${createdCode}`, 200, {
      title: `已更新令牌片段 ${marker}`,
      expectedRevision: createdRevision,
    });
    createdRevision = updated.paste.revision;
    expect((await json(full, "GET", `/api/v1/admin/pastes?q=${encodeURIComponent(marker)}`, 200)).items.map((item: { code: string }) => item.code)).toContain(createdCode);

    originalSettings = (await json(full, "GET", "/api/v1/settings", 200)).settings;
    const changed = await json(full, "PATCH", "/api/v1/admin/settings", 200, {
      name: originalSettings!.name,
      description: `PAT 验收 ${marker}`,
      expectedRevision: originalSettings!.revision,
    });
    const restored = await json(full, "PATCH", "/api/v1/admin/settings", 200, {
      name: originalSettings!.name,
      description: originalSettings!.description,
      expectedRevision: changed.settings.revision,
    });
    originalSettings = { ...originalSettings!, revision: restored.settings.revision };

    await json(full, "DELETE", `/api/v1/me/pastes/${createdCode}?expectedRevision=${createdRevision}`, 204);
    createdCode = "";
    await json(page.request, "DELETE", `${account}/api/v1/pat/${fullToken.id}`, 204);
    tokens.splice(tokens.findIndex((item) => item.id === fullToken.id), 1);
    await json(full, "GET", "/api/v1/me/pastes", 401);
    await json(direct, "GET", "/api/v1/me/pastes", 401);
    const noFallback = await page.request.get(`${baseURL}/api/v1/me/pastes`, { headers: { authorization: `Bearer ${fullToken.token}` } });
    expect(noFallback.status()).toBe(401);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  } finally {
    if (createdCode) {
      const current = await page.request.get(`${baseURL}/api/v1/me/pastes/${createdCode}`);
      if (current.ok()) {
        const payload = await current.json();
        await page.request.delete(`${baseURL}/api/v1/me/pastes/${createdCode}?expectedRevision=${payload.paste.revision}`);
      }
    }
    for (const token of tokens) {
      await token.owner.delete(`${account}/api/v1/pat/${token.id}`);
    }
    for (const client of clients) {
      await client.dispose();
    }
  }
});
