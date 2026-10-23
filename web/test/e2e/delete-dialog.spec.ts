import { expect, test } from "@playwright/test";

test("Nuxt dialogs cancel safely and retry only failed batch deletions", async ({ page }, testInfo) => {
  test.setTimeout(90_000);
  await page.setViewportSize({ width: testInfo.project.name === "mobile" ? 375 : 1280, height: 840 });
  const nativeDialogs: string[] = [];
  page.on("dialog", async dialog => { nativeDialogs.push(dialog.type()); await dialog.dismiss(); });
  await page.goto("/mine");
  const titles = [`删除确认 ${Date.now()}`, "很长的内容标题 / " + "用于检查窄屏确认框中的完整换行与删除范围。".repeat(4)];
  const codes: string[] = [];
  try {
    for (const title of titles) {
      const response = await page.request.post("/api/v1/pastes", { data: { title, files: [{ path: "note.txt", content: "Dialog acceptance" }] } });
      expect(response.status()).toBe(201);
      codes.push((await response.json()).paste.code);
    }
    await page.reload();
    await expect(page.locator(".paste-ledger-row")).toHaveCount(2);
    const dialog = page.getByRole("dialog");
    let deleteRequests = 0;
    page.on("request", request => { if (request.method() === "DELETE") deleteRequests++; });
    await page.locator(".paste-ledger-row").first().getByRole("button", { name: "删除代码片段" }).click();
    await expect(dialog.getByRole("button", { name: "取消", exact: true })).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
    expect(deleteRequests).toBe(0);
    for (const theme of ["dark", "light"]) {
      await page.evaluate(theme => localStorage.setItem("nuxt-color-mode", theme), theme);
      await page.reload();
      await expect(page.locator(".paste-ledger-row")).toHaveCount(2);
      await page.getByRole("checkbox", { name: "选择当前筛选结果" }).click();
      await page.getByRole("button", { name: "批量修改", exact: true }).click();
      await expect(dialog.getByRole("heading", { name: "批量修改", exact: true })).toBeVisible();
      await page.screenshot({ animations: "disabled", path: testInfo.outputPath(`batch-edit-${theme}.png`) });
      await dialog.getByRole("button", { name: "取消", exact: true }).click();
      await page.getByRole("button", { name: "删除", exact: true }).focus();
      await page.keyboard.press("Enter");
      await expect(dialog.getByRole("heading", { name: "删除 2 项内容？", exact: true })).toBeVisible();
      await expect(dialog.getByText(titles[1]!, { exact: true })).toBeVisible();
      await page.screenshot({ animations: "disabled", path: testInfo.outputPath(`delete-${theme}.png`) });
      await dialog.getByRole("button", { name: "取消", exact: true }).click();
      expect(deleteRequests).toBe(0);
    }
    let release!: () => void;
    const gate = new Promise<void>(resolve => { release = resolve; });
    let failed = false;
    await page.route(`**/api/v1/me/pastes/${codes[1]}?**`, async route => {
      if (route.request().method() !== "DELETE" || failed) return route.continue();
      failed = true;
      await gate;
      await route.abort("failed");
    });
    await page.getByRole("button", { name: "删除", exact: true }).click();
    await dialog.getByRole("button", { name: "确认删除", exact: true }).click();
    await expect(dialog.getByRole("button", { name: "取消", exact: true })).toBeDisabled();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeVisible();
    release();
    await expect(dialog.getByText("已删除 1 项，1 项失败。可以重试剩余内容。", { exact: true })).toBeVisible();
    await expect(dialog.getByText(titles[0]!, { exact: true })).toHaveCount(0);
    await dialog.getByRole("button", { name: "重试删除", exact: true }).click();
    await expect(dialog).toBeHidden();
    await expect(page.locator(".paste-ledger-row")).toHaveCount(0);
    expect(nativeDialogs).toEqual([]);
  } finally {
    await page.unrouteAll({ behavior: "ignoreErrors" });
    for (const code of codes) {
      const response = await page.request.get(`/api/v1/me/pastes/${code}`);
      if (response.ok()) await page.request.delete(`/api/v1/me/pastes/${code}?expectedRevision=${(await response.json()).paste.revision}`);
    }
  }
});
