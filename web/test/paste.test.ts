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
    const caught = remote("paste.password_required",423);
    expect(pasteProblemCode(caught)).toBe("paste.password_required");
    expect(pasteErrorMessage(caught)).toContain("需要密码");
  });

  it("turns file-size violations into a useful recovery message", () => {
    const caught = {
      failure: {
        kind: "remote",
        status: 400,
        code: "common.validation_failed",
        params: {},
        violations: [{ pointer: "/files/content", code: "validation.max_bytes", params: { maxBytes: 1048576 } }],
        traceId: "test-trace",
        reauth: "not-attempted",
      },
    };
    expect(pasteErrorMessage(caught)).toContain("单个文件不能超过 1 MiB");
  });

  it("explains product-local creation controls", () => {
    expect(pasteErrorMessage(remote("paste.creation_suspended",403))).toContain("创建权限已被暂停");
    expect(pasteErrorMessage(remote("paste.anonymous_creation_disabled",403))).toContain("停止匿名创建");
    expect(pasteErrorMessage(remote("paste.daily_limit_reached",429))).toContain("创建额度已经用完");
  });
});

it("does not interpret thrown messages or disclose transport detail", () => {
 const unsafe={message:"SQL password=secret",data:{detail:"private /server/path"}};
 expect(pasteProblemCode({message:"paste.password_required"})).toBe("");
 expect(pasteErrorMessage(unsafe)).toBe("请求没有完成，请稍后重试。");
 expect(pasteErrorMessage(remote("common.rate_limited",429))).toContain("操作过于频繁");
});

function remote(code:string,status:number){return {failure:{kind:"remote",status,code,params:{},violations:[],traceId:"test",reauth:"not-attempted"}};}
