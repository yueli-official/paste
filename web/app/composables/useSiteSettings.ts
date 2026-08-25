import type { SiteSettings } from "../types/paste";

const defaultSettings: SiteSettings = {
  name: "代码片段",
  description: "轻量、专注的多文件代码分享。",
  revision: 1,
  updatedAt: "",
};

export function useSiteSettings() {
  const api = usePasteApi();
  const settings = useState<SiteSettings>("paste-site-settings", () => ({ ...defaultSettings }));
  const loaded = useState<boolean>("paste-site-settings-loaded", () => false);

  async function load(force = false) {
    if (loaded.value && !force) return settings.value;
    try {
      const response = await api.getSettings();
      settings.value = response.settings;
      loaded.value = true;
    } catch {
      // Public branding has safe local defaults; failed settings reads must not
      // block creating or opening a code snippet.
    }
    return settings.value;
  }

  function adopt(value: SiteSettings) {
    settings.value = value;
    loaded.value = true;
  }

  return {
    settings,
    siteName: computed(() => settings.value.name || defaultSettings.name),
    siteDescription: computed(() => settings.value.description || defaultSettings.description),
    load,
    adopt,
  };
}
