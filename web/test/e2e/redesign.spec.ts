import { expect, test, type Page } from "@playwright/test";

test("language selection updates file extensions and keeps names unique", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator(".cm-content")).toBeVisible();
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "Go", exact: true }).click();
  await expect(page.locator(".paste-file-tab-button").first()).toContainText("untitled.go");
  await page.locator(".paste-file-tab-button").first().dblclick();
  await page.getByLabel("重命名文件").fill("src/example.ts");
  await page.getByLabel("重命名文件").press("Enter");
  await expect(page.getByLabel("代码语言")).toContainText("TypeScript");
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "Go", exact: true }).click();
  await expect(page.locator(".paste-file-tab-button").first()).toContainText("src/example.go");
  await page.getByRole("button", { name: "添加文件" }).click();
  await page.locator(".paste-file-tab-button").nth(1).dblclick();
  await page.getByLabel("重命名文件").fill("src/example.txt");
  await page.getByLabel("重命名文件").press("Enter");
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "Go", exact: true }).click();
  await expect(page.locator(".paste-file-tab-button").nth(1)).toContainText("src/example-2.go");
});

// Exercise CodeMirror's clipboard path: contenteditable.fill() is not a paste
// and Chromium can normalize its multiline insertText before CodeMirror sees it.
async function pasteIntoEditor(page: Page, content: string) {
  await page.locator(".cm-content").evaluate((node, text) => {
    const clipboardData = new DataTransfer();
    clipboardData.setData("text/plain", text);
    node.dispatchEvent(new ClipboardEvent("paste", { bubbles: true, cancelable: true, clipboardData }));
  }, content);
}

test("Markdown supports live side-by-side editing and narrow-screen preview", async ({ page }, testInfo) => {
  await page.goto("/");
  await expect(page.locator(".cm-content")).toBeVisible();
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "Markdown", exact: true }).click();
  await expect(page.locator(".paste-file-tab-button").first()).toContainText("untitled.md");
  await pasteIntoEditor(page, "# 实时预览\n\n正文与 **强调**。\n\n## 下一步\n\n- 保存原文\n- 分享链接\n");
  const previewPane = page.locator(".paste-preview-pane");
  if (testInfo.project.name === "desktop") {
    await expect(previewPane.getByRole("heading", { name: "实时预览", exact: true })).toBeVisible();
    const source = await page.locator(".paste-source-pane").boundingBox();
    const preview = await previewPane.boundingBox();
    expect(source!.x + source!.width).toBeLessThanOrEqual(preview!.x + 1);
    expect(Math.abs(source!.y - preview!.y)).toBeLessThan(1);
    await page.locator(".cm-content").press("Control+End");
    await pasteIntoEditor(page, "\n## 即时更新\n");
    await expect(previewPane.getByRole("heading", { name: "即时更新", exact: true })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath("markdown-split.png"), fullPage: true });
  } else {
    await expect(previewPane).toBeHidden();
    await expect(page.getByRole("button", { name: "双栏", exact: true })).toBeHidden();
  }
  await page.getByRole("button", { name: "预览", exact: true }).click();
  await expect(page.locator(".cm-content")).toBeHidden();
  await expect(previewPane.getByRole("heading", { name: "实时预览", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "编辑", exact: true }).click();
  await expect(page.locator(".cm-content")).toContainText("# 实时预览");
  await expect(previewPane).toBeHidden();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});

test("direct sharing and Markdown modes preserve the original source", async ({
  page,
}, testInfo) => {
  await page.goto("/");
  const create = page.getByRole("button", {
    name: "生成分享链接",
    exact: true,
  });
  await expect(create).toBeDisabled();
  await expect(page.locator(".cm-content")).toBeFocused();
  await page.screenshot({
    path: testInfo.outputPath("a6-empty.png"),
    fullPage: true,
  });
  await page.getByLabel("标题", { exact: true }).fill("API 请求示例");
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "TypeScript", exact: true }).click();
  const code = 'interface ShareOptions {\n  expiresIn: number;\n  visibility: "unlisted" | "private";\n}\n\nexport async function createShare(content: string) {\n  const options: ShareOptions = {\n    expiresIn: 7 * 24 * 60 * 60,\n    visibility: "unlisted",\n  };\n\n  const response = await fetch("/api/v1/pastes", {\n    method: "POST",\n    body: JSON.stringify({ content, ...options }),\n  });\n\n  return response.json();\n}';
  await pasteIntoEditor(page, code);
  await expect(create).toBeEnabled();
  await page.getByRole("button", { name: "自动换行" }).click();
  await expect(page.getByRole("button", { name: "自动换行" })).toHaveAttribute(
    "aria-pressed",
    "false",
  );
  await expect(
    page.locator(".cm-line").filter({ hasText: "return response.json();" }),
  ).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath("a6-code.png"),
    fullPage: true,
  });
  await page.getByRole("button", { name: "添加文件" }).click();
  await page.getByRole("button", { name: /^untitled-2.txt，/ }).dblclick();
  await page.getByLabel("重命名文件").fill("README.md");
  await page.getByLabel("重命名文件").press("Enter");
  const markdown =
    "# 分享，让灵感继续\n\n一段代码，一份笔记。**值得留下的内容**，都可以从这里开始。\n\n## 一个链接，就够了\n\n- 写下想法，或直接粘贴\n- 选择访问范围与有效期\n- 生成链接，发送给需要的人\n\n```typescript\nconst share = await createShare(content);\nconsole.log(share.url);\n```\n\n<!-- preserve this custom comment -->\n\n";
  await pasteIntoEditor(page, markdown);
  await page.getByRole("button", { name: "预览", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "分享，让灵感继续" }),
  ).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath("a6-markdown-preview.png"),
    fullPage: true,
  });
  await page.getByRole("button", { name: "选择 Markdown 编辑方式" }).click();
  await page.getByRole("menuitem", { name: "排版编辑" }).click();
  await page.getByRole("button", { name: "使用排版编辑", exact: true }).click();
  await expect(page.locator(".tiptap")).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath("a6-markdown-rich.png"),
    fullPage: true,
  });
  await page.getByRole("button", { name: "选择 Markdown 编辑方式" }).click();
  await page.getByRole("menuitem", { name: "源码编辑" }).click();
  const submitted = page.waitForRequest(
    (request) =>
      request.method() === "POST" && request.url().endsWith("/api/v1/pastes"),
  );
  await create.click();
  const payload = (await submitted).postDataJSON();
  expect(payload.files[1].content).toBe(markdown);
  expect(payload.files[0].content).toBe(code);
  expect(payload.files[0].language).toBe("typescript");
  await expect(page).toHaveURL(/\/p\/[^/?]+\?created=1$/);
  await expect(page.getByLabel("分享链接", { exact: true })).toHaveValue(
    /\/p\//,
  );
  await page.getByRole("tab", { name: "README.md" }).click();
  await expect(
    page.getByRole("heading", { name: "分享，让灵感继续" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "复制分享链接", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "复制分享链接", exact: true }),
  ).toContainText("链接已复制");
  await page.screenshot({
    path: testInfo.outputPath("a6-shared.png"),
    fullPage: true,
  });
});

test("rich Markdown editing produces a shareable Markdown string", async ({
  page,
}) => {
  await page.goto("/");
  await expect(page.locator(".cm-content")).toBeVisible();
  await page.getByLabel("代码语言").click();
  await page.getByRole("option", { name: "Markdown", exact: true }).click();
  await page.getByRole("button", { name: "选择 Markdown 编辑方式" }).click();
  await page.getByRole("menuitem", { name: "排版编辑" }).click();
  await page.locator(".tiptap").fill("从排版编辑器创建的内容");
  const submitted = page.waitForRequest(
    (request) =>
      request.method() === "POST" && request.url().endsWith("/api/v1/pastes"),
  );
  await page.getByRole("button", { name: "生成分享链接", exact: true }).click();
  expect((await submitted).postDataJSON().files[0].content).toContain(
    "从排版编辑器创建的内容",
  );
  await expect(page).toHaveURL(/\/p\//);
  await expect(
    page.getByText("从排版编辑器创建的内容", { exact: true }),
  ).toBeVisible();
});
