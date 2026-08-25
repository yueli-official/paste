import { describe, expect, it } from "vitest";
import { displayTitle, expiryISO, languageFromPath, pasteErrorMessage, pasteProblemCode } from "../app/utils/paste";

describe("Paste presentation rules", () => {
  it("infers supported languages from case-insensitive paths", () => {
    expect(languageFromPath("src/main.GO")).toBe("go");
    expect(languageFromPath("app.vue")).toBe("vue");
    expect(languageFromPath("LICENSE")).toBe("text");
  });

  it("uses the path before the generic untitled label", () => {
    expect(displayTitle("", "main.go")).toBe("main.go");
    expect(displayTitle("", "")).toBe("未命名片段");
    expect(displayTitle("未命名 Paste", "")).toBe("未命名片段");
  });

  it("keeps never-expiring Pastes explicit", () => {
    expect(expiryISO("never")).toBeUndefined();
    expect(new Date(expiryISO("1h")!).getTime()).toBeGreaterThan(Date.now());
  });

  it("recognizes transport-safe Problem codes", () => {
    const caught = { message: "paste.password_required" };
    expect(pasteProblemCode(caught)).toBe("paste.password_required");
    expect(pasteErrorMessage(caught)).toContain("需要密码");
  });

  it("turns file-size violations into a useful recovery message", () => {
    const caught = {
      failure: {
        kind: "remote",
        status: 400,
        code: "validation.failed",
        params: {},
        violations: [{ pointer: "/files/content", code: "validation.max_bytes", params: { maxBytes: 1048576 } }],
        traceId: "test-trace",
        reauth: "not-attempted",
      },
    };
    expect(pasteErrorMessage(caught)).toContain("单个文件不能超过 1 MiB");
  });

  it("explains product-local creation controls", () => {
    expect(pasteErrorMessage({ message: "paste.creation_suspended" })).toContain("创建权限已被暂停");
    expect(pasteErrorMessage({ message: "paste.anonymous_creation_disabled" })).toContain("停止匿名创建");
    expect(pasteErrorMessage({ message: "common.rate_limited" })).toContain("创建额度已经用完");
  });
});
