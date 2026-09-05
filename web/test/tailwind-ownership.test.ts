import { readFileSync, readdirSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const appRoot = resolve(process.cwd(), "app");
const mainCSS = readFileSync(resolve(appRoot, "assets/css/main.css"), "utf8");

function vueStyleSources(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = `${directory}/${entry.name}`;
    if (entry.isDirectory()) return vueStyleSources(path);
    if (!entry.isFile() || !entry.name.endsWith(".vue")) return [];
    const source = readFileSync(path, "utf8");
    return [...source.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/gu)].map(
      (match) => match[1] || "",
    );
  });
}

describe("Paste Tailwind ownership", () => {
  it("keeps CSS ownership monotonically decreasing", () => {
    const componentStyles = vueStyleSources(appRoot);
    const componentCSS = componentStyles.join("\n");
    const combined = `${mainCSS}\n${componentCSS}`;

    expect(mainCSS.trimEnd().split(/\r?\n/u).length).toBeLessThanOrEqual(149);
    expect(componentStyles).toEqual([]);
    expect((combined.match(/\{/gu) || []).length).toBeLessThanOrEqual(13);
    expect((combined.match(/@media\b/gu) || []).length).toBeLessThanOrEqual(1);
    expect(combined).not.toContain("@apply");
    expect(
      [...mainCSS.matchAll(/^\s*(\.paste-[^{\r\n]+)\s*\{/gmu)].map(
        (match) => match[1]?.trim(),
      ),
    ).toEqual([
      ".paste-code-editor .cm-editor",
      ".paste-code-editor .cm-scroller",
      ".paste-code-editor .cm-tooltip",
    ]);
  });
});
