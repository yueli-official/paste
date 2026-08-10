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
    expect(displayTitle("", "")).toBe("未命名 Paste");
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
});
