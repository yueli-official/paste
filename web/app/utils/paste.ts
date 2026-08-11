const languageByExtension: Readonly<Record<string, string>> = {
  c: "cpp",
  cc: "cpp",
  cpp: "cpp",
  css: "css",
  go: "go",
  h: "cpp",
  hpp: "cpp",
  html: "html",
  java: "java",
  js: "javascript",
  json: "json",
  jsx: "javascript",
  md: "markdown",
  php: "php",
  py: "python",
  rs: "rust",
  sql: "sql",
  ts: "typescript",
  tsx: "typescript",
  txt: "text",
  vue: "vue",
  yaml: "yaml",
  yml: "yaml",
};

export const languageItems = [
  { label: "纯文本", value: "text" },
  { label: "Go", value: "go" },
  { label: "JavaScript", value: "javascript" },
  { label: "TypeScript", value: "typescript" },
  { label: "Vue", value: "vue" },
  { label: "Python", value: "python" },
  { label: "Rust", value: "rust" },
  { label: "Java", value: "java" },
  { label: "C / C++", value: "cpp" },
  { label: "HTML", value: "html" },
  { label: "CSS", value: "css" },
  { label: "JSON", value: "json" },
  { label: "Markdown", value: "markdown" },
  { label: "SQL", value: "sql" },
  { label: "YAML", value: "yaml" },
  { label: "PHP", value: "php" },
] as const;

export function languageFromPath(path: string): string {
  const extension = path.split(".").pop()?.toLowerCase() || "";
  return languageByExtension[extension] || "text";
}

export function displayTitle(title: string, path?: string): string {
  return title.trim() || path?.trim() || "未命名 Paste";
}

export function pasteProblemCode(caught: unknown): string {
  const failure = getApiFailure(caught);
  if (caught && typeof caught === "object") {
    const value = caught as {
      data?: { detail?: string; message?: string; code?: string; failure?: { code?: string } };
      failure?: { code?: string };
      message?: string;
      statusCode?: number;
    };
    const candidates = [
      failure?.code,
      value.data?.code,
      value.data?.failure?.code,
      value.failure?.code,
      value.message,
    ];
    return candidates.find((candidate) =>
      typeof candidate === "string" && /^[a-z][a-z0-9]*(?:[._-][a-z0-9]+)+$/.test(candidate),
    ) || "";
  }
  return "";
}

export function pasteErrorMessage(caught: unknown): string {
  const failure = getApiFailure(caught);
  if (caught && typeof caught === "object") {
    const value = caught as {
      data?: { detail?: string; message?: string; code?: string };
      message?: string;
      statusCode?: number;
    };
    const code = pasteProblemCode(caught);
    if (code === "paste.password_required") return "此 Paste 需要密码才能打开。";
    if (code === "paste.password_invalid") return "密码不正确，请重新输入。";
    if (code === "paste.gone") return "此 Paste 已过期或已被删除。";
    if (code === "paste.not_authenticated") return "请先登录，再继续管理你的 Paste。";
    if (code === "paste.conflict") return "内容已在别处更新，请刷新后重试。";
    if (code === "validation.failed") {
      const violation = failure?.kind === "remote" ? failure.violations[0] : undefined;
      if (violation?.pointer === "/files/content" && violation.code === "validation.max_bytes") {
        return "单个文件不能超过 1 MiB，请拆分文件或删减内容后重试。";
      }
      if (violation?.pointer === "/files/content") {
        return "全部文件合计不能超过 1 MiB，请删减内容后重试。";
      }
      if (violation?.pointer === "/files/path") return "文件名不符合要求，请检查长度、重复项和相对路径。";
      if (violation?.pointer === "/password") return "访问密码需要包含 8–128 个字符。";
      if (violation?.pointer === "/tags") return "标签最多 8 个，每个标签不超过 32 个字符。";
      return "填写的分享信息不符合要求，请检查后重试。";
    }
    if (value.data?.detail) return value.data.detail;
    if (value.data?.message) return value.data.message;
    if (value.message && !value.message.startsWith("[") && !value.message.startsWith("paste.")) return value.message;
  }
  return "请求没有完成，请稍后重试。";
}

export function expiryISO(value: string): string | undefined {
  if (value === "never" || value === "keep") return undefined;
  const now = Date.now();
  const durations: Readonly<Record<string, number>> = {
    "1h": 60 * 60 * 1000,
    "1d": 24 * 60 * 60 * 1000,
    "7d": 7 * 24 * 60 * 60 * 1000,
    "30d": 30 * 24 * 60 * 60 * 1000,
  };
  return new Date(now + (durations[value] || durations["7d"]!)).toISOString();
}
import { getApiFailure } from "@yueli/http-runtime";
