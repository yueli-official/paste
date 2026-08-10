import { createUiPreset } from "@yueli/ui/theme";

export default defineAppConfig(
  createUiPreset(
    { primary: "sky", neutral: "stone" },
    { cardRoot: "rounded-xl" },
  ),
);
