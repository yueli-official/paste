import { expect, test, type Page } from "@playwright/test";

async function login(page: Page) {
  await page.goto("/mine");
  await page.waitForFunction(() =>
    Boolean(
      (
        document
          .querySelector("input[name=email]")
          ?.closest("form") as HTMLFormElement & {
          __vueParentComponent?: unknown;
        }
      )?.__vueParentComponent,
    ),
  );
  await page
    .getByLabel("邮箱")
    .fill(process.env.PASTE_E2E_EMAIL || "test@example.com");
  await page
    .getByLabel("密码", { exact: true })
    .fill(process.env.PASTE_E2E_PASSWORD || "Yueli-local-development-2026");
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL((url) => url.pathname === "/mine");
}

test("HTTP Result uses 201 Location, bounded pages, 204 and structured Problems", async ({
  page,
}, info) => {
  test.skip(info.project.name !== "desktop", "one real HTTP contract journey");
  await login(page);
  const title = `HTTP Result ${Date.now()}`;
  const created = await page.request.post("/api/v1/pastes", {
    data: {
      title,
      files: [{ path: "a.txt", language: "text", content: "contract" }],
      visibility: "private",
    },
  });
  expect(created.status()).toBe(201);
  const value = (await created.json()).paste;
  expect(created.headers().location).toBe(`/api/v1/pastes/${value.code}`);
  const listed = await page.request.get("/api/v1/me/pastes", {
    params: { q: title, page: 1, size: 1 },
  });
  const result = await listed.json();
  expect(result).toMatchObject({ page: 1, size: 1, total: 1 });
  expect(result.items).toHaveLength(1);
  expect(result).not.toHaveProperty("pastes");
  const bad = await page.request.get("/api/v1/me/pastes?size=101");
  expect(bad.status()).toBe(400);
  expect(bad.headers()["content-type"]).toContain("application/problem+json");
  expect(await bad.json()).toMatchObject({
    status: 400,
    code: "common.validation_failed",
    traceId: bad.headers()["x-trace-id"],
  });
  const deleted = await page.request.delete(
    `/api/v1/me/pastes/${value.code}?expectedRevision=${value.revision}`,
  );
  expect(deleted.status()).toBe(204);
  expect(await deleted.text()).toBe("");
  for (let i = 0; i < 3; i++) {
    const response = await page.goto("/admin");
    expect(await response!.text()).toContain("data-admin-shell");
    await expect(page.locator("[data-admin-shell]")).toBeVisible();
  }
});

test("field feedback keeps editor content and hides raw server detail", async ({
  page,
}, info) => {
  await page.goto("/");
  const editor = page.getByRole("textbox", { name: "main.go 代码编辑器" });
  await expect(editor).toBeVisible();
  await editor.fill("package main // keep this draft");
  await page.getByRole("button", { name: "打开分享设置" }).click();
  await page.getByLabel("标题", { exact: true }).fill("保留标题");
  await page.route("**/api/v1/pastes", (route) =>
    route.fulfill({
      status: 400,
      contentType: "application/problem+json",
      headers: { "x-trace-id": "paste-field-trace" },
      json: {
        type: "https://errors.yueli.dev/problems/common.validation_failed",
        status: 400,
        code: "common.validation_failed",
        traceId: "paste-field-trace",
        detail: "SQL password=secret",
        violations: [
          { pointer: "/title", code: "validation.invalid" },
          { pointer: "/unknown", code: "validation.required" },
        ],
      },
    }),
  );
  await page.getByRole("button", { name: "生成分享链接", exact: true }).click();
  await expect(page.getByLabel("标题", { exact: true })).toHaveAttribute(
    "aria-invalid",
    "true",
  );
  await expect(page.locator("[data-failure-notice]")).toContainText(
    "请填写此项",
  );
  await expect(page.getByText("SQL password=secret")).toHaveCount(0);
  await page.getByText("技术详情", { exact: true }).click();
  await expect(page.locator("[data-failure-notice]")).toContainText(
    "paste-field-trace",
  );
  await page.screenshot({ path: info.outputPath("field-failure.png") });
  await page.keyboard.press("Escape");
  await expect(editor).toContainText("keep this draft");
});
