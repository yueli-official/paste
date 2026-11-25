import { createUiPreset } from "@yueli/ui/theme";

export default defineAppConfig(
  createUiPreset(
    { primary: "lime", neutral: "slate" },
    { cardRoot: "rounded-xl", icons: { dark: "i-tabler-moon-stars", light: "i-tabler-sun-high" } },
  ),
);
