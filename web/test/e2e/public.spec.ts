import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("editor-first composer works at desktop and mobile widths", async ({ page }, testInfo) => {
  if (testInfo.project.name === "mobile") {
    await page.emulateMedia({ colorScheme: "dark" });
  }
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: "新建片段" })).toBeVisible();
  await expect(page.getByRole("textbox", { name: /代码编辑器/ })).toBeVisible();
  await expect(page.locator(".cm-editor")).toBeVisible();
  const statusFontSize = await page.locator(".paste-statusbar").evaluate((node) =>
    Number.parseFloat(getComputedStyle(node).fontSize),
  );
  expect(statusFontSize).toBeGreaterThanOrEqual(12);
  await page.getByRole("button", { name: "添加文件" }).click();
  const untitledTab = page.getByRole("button", { name: /^untitled-2\.txt，/ });
  await expect(untitledTab).toBeVisible();
  await untitledTab.dblclick();
  await expect(page.getByLabel("重命名文件")).toHaveValue("untitled-2.txt");
  await page.screenshot({ path: testInfo.outputPath(`rename-${testInfo.project.name}.png`), fullPage: true });
  await page.getByLabel("重命名文件").fill("README.md");
  await page.getByLabel("重命名文件").press("Enter");
  const newFileTab = page.getByRole("button", { name: /^README\.md，/ });
  await expect(newFileTab).toBeVisible();
  await newFileTab.hover();
  await expect(page.getByRole("button", { name: "删除 README.md" })).toBeVisible();
  const mainFileTab = page.getByRole("button", { name: /^main\.go，/ });
  await newFileTab.dragTo(mainFileTab);
  await expect(page.locator(".paste-file-tab-button").first()).toContainText("README.md");
  await newFileTab.press("Alt+ArrowRight");
  await expect(page.locator(".paste-file-tab-button").nth(1)).toContainText("README.md");
  await newFileTab.dblclick();
  await expect(page.getByLabel("重命名文件")).toHaveValue("README.md");
  await page.getByLabel("重命名文件").press("Escape");
  await expect(page.getByLabel("代码语言")).toContainText("Markdown");
  await page.getByLabel("代码语言").click();
  const markdownOption = page.getByRole("option", { name: "Markdown" });
  await page.screenshot({ path: testInfo.outputPath(`language-menu-${testInfo.project.name}.png`), fullPage: true });
  await markdownOption.click();
  await expect(markdownOption).toBeHidden();
  await page.locator(".cm-content").fill("# Paste\n\nA real multi-file share.");

  if (testInfo.project.name === "mobile") {
    const workbenchWidth = await page.locator(".paste-workbench").evaluate((node) => node.getBoundingClientRect().width);
    expect(workbenchWidth).toBeLessThanOrEqual(await page.evaluate(() => innerWidth));
  }

  await page.screenshot({ path: testInfo.outputPath(`composer-${testInfo.project.name}.png`), fullPage: true });
  const results = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(results.violations, results.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);

  await page.getByRole("button", { name: "打开分享设置" }).click();
  await expect(page.getByRole("heading", { name: "分享代码片段" })).toBeVisible();
  await page.getByRole("button", { name: "生成分享链接" }).click();
  await expect(page).toHaveURL(/\/p\/[A-Za-z0-9_-]{8}\?created=1$/);
  await expect(page.getByRole("heading", { level: 1, name: "未命名片段" })).toBeVisible();
  await expect(page.locator(".paste-public-header")).toHaveCount(0);
  await expect(page.locator(".paste-reader-shell")).toBeVisible();
  const readerStatusFontSize = await page.locator(".paste-reader-statusbar").evaluate((node) =>
    Number.parseFloat(getComputedStyle(node).fontSize),
  );
  expect(readerStatusFontSize).toBeGreaterThanOrEqual(12);
  const mainTab = page.getByRole("tab", { name: "main.go" });
  const readmeTab = page.getByRole("tab", { name: "README.md" });
  await expect(readmeTab).toBeVisible();
  await expect(mainTab).toHaveAttribute("aria-selected", "true");
  await mainTab.press("ArrowRight");
  await expect(readmeTab).toHaveAttribute("aria-selected", "true");
  const readerDimensions = await page.locator(".paste-reader-shell").evaluate((node) => ({
    width: node.getBoundingClientRect().width,
    viewport: innerWidth,
    scrollWidth: document.documentElement.scrollWidth,
  }));
  expect(readerDimensions.width).toBeLessThanOrEqual(readerDimensions.viewport);
  expect(readerDimensions.scrollWidth).toBeLessThanOrEqual(readerDimensions.viewport);
  await page.screenshot({ path: testInfo.outputPath(`reader-${testInfo.project.name}.png`), fullPage: true });
  await page.getByRole("button", { name: "查看片段信息" }).click();
  await expect(page.getByRole("heading", { name: "片段信息" })).toBeVisible();
  await expect(page.getByRole("button", { name: "复制当前文件" })).toBeVisible();
  await expect(page.getByRole("button", { name: "复制全部文件" })).toBeVisible();
  const infoPanel = page.getByRole("dialog", { name: "片段信息" });
  await expect.poll(async () => {
    const box = await infoPanel.boundingBox();
    return box ? Math.round(box.x + box.width) : 0;
  }).toBe(await page.evaluate(() => innerWidth));
  await page.screenshot({ path: testInfo.outputPath(`reader-info-${testInfo.project.name}.png`), fullPage: true });
  const readerAccessibility = await new AxeBuilder({ page }).exclude("[data-nuxt-devtools]").analyze();
  expect(readerAccessibility.violations, readerAccessibility.violations.map((item) => `${item.id}: ${item.help}`).join("\n")).toEqual([]);
});

test("composer shares one file larger than the old 256 KiB boundary", async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== "desktop", "one boundary check is sufficient");
  await page.goto("/");
  await page.locator(".cm-content").fill("a".repeat((256 << 10) + 1));
  await expect(page.locator(".paste-status-size")).toHaveAttribute("data-over-limit", "false");
  await expect(page.locator(".paste-status-size")).toContainText("257/1024 KiB");
  await page.getByRole("button", { name: "打开分享设置" }).click();
  await page.getByRole("button", { name: "生成分享链接" }).click();
  await expect(page).toHaveURL(/\/p\/[A-Za-z0-9_-]{8}\?created=1$/);
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
  await page.getByRole("button", { name: "打开片段" }).click();
  await expect(page.getByRole("alert")).toContainText("密码不正确");
  await page.getByRole("textbox", { name: "访问密码", exact: true }).fill("safepass123");
  await page.getByRole("button", { name: "打开片段" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "未命名片段" })).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`/p/${code}$`));
  await expect(page.getByText("bounded secret")).toBeVisible();
});

test("missing shares keep the read-only editor shell", async ({ page }) => {
  await page.goto("/p/Missing1");
  await expect(page.locator(".paste-public-header")).toHaveCount(0);
  await expect(page.locator(".paste-reader-shell")).toBeVisible();
  await expect(page.getByRole("heading", { level: 1, name: "无法打开这个片段" })).toBeVisible();
  await expect(page.getByRole("button", { name: "重新打开" })).toBeVisible();
  await expect(page.getByRole("link", { name: "新建片段" })).toBeVisible();
});
