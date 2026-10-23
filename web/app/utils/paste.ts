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
  { label: "纯文本", value: "text", extension: "txt" },
  { label: "Go", value: "go", extension: "go" },
  { label: "JavaScript", value: "javascript", extension: "js" },
  { label: "TypeScript", value: "typescript", extension: "ts" },
  { label: "Vue", value: "vue", extension: "vue" },
  { label: "Python", value: "python", extension: "py" },
  { label: "Rust", value: "rust", extension: "rs" },
  { label: "Java", value: "java", extension: "java" },
  { label: "C / C++", value: "cpp", extension: "cpp" },
  { label: "HTML", value: "html", extension: "html" },
  { label: "CSS", value: "css", extension: "css" },
  { label: "JSON", value: "json", extension: "json" },
  { label: "Markdown", value: "markdown", extension: "md" },
  { label: "SQL", value: "sql", extension: "sql" },
  { label: "YAML", value: "yaml", extension: "yaml" },
  { label: "PHP", value: "php", extension: "php" },
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
