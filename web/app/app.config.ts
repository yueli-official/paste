import { createUiPreset } from "@yueli/ui/theme";

export default defineAppConfig(
  createUiPreset(
    { primary: "lime", neutral: "slate" },
    { cardRoot: "rounded-xl" },
  ),
);
