import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const layout = readFileSync(resolve(process.cwd(), "app/layouts/admin.vue"), "utf8");
const page = readFileSync(resolve(process.cwd(), "app/pages/admin.vue"), "utf8");

describe("Paste 管理控制台合同", () => {
  it("服务端输出统一控制台壳、路由栏与内容页头", () => {
    expect(layout).toMatch(/<template>\s*<YAdminConsoleLayout/);
    expect(layout).toContain(':current-label="currentLabel"');
    expect(layout).not.toMatch(/<template>\s*<ClientOnly>[\s\S]*?<YAdminConsoleLayout/);
    expect(layout).not.toMatch(/正在打开[^\n]{0,16}控制台/);
    expect(page).toContain('layout: "admin"');
    expect(page).toContain("<PageHeader");
    expect(page).not.toContain("paste-admin-chrome");
  });
});
