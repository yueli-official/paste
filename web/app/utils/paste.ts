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
  const normalized = title.trim();
  if (normalized && normalized !== "未命名 Paste") return normalized;
  return path?.trim() || "未命名片段";
}

export function pasteProblemCode(caught: unknown): string { return getApiFailure(caught)?.code || ""; }
export function pasteErrorMessage(caught: unknown, fallback = "请求没有完成，请稍后重试。"): string {
 const feedback=pasteFailureFeedback(caught,fallback);
 return [feedback.message,...feedback.summary].join(" ");
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
import { pasteFailureFeedback } from "./pasteFailure";
