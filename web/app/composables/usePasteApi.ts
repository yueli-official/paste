import type { JsonValue } from "@yueli/http-runtime";
import { useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import type {
  AdministrationPaste,
  AdministrationPastePage,
  AdministrationPasteQuery,
  AdministrationUserPage,
  AdministrationUserPolicy,
  AdministrationUserPolicyInput,
  AdministrationUserQuery,
  GovernanceSettings,
  GovernanceSettingsInput,
  Paste,
  PastePatchInput,
  PasteSummary,
  PasteWriteInput,
  SiteSettings,
  SiteSettingsInput,
} from "../types/paste";

function jsonBody(value: unknown): JsonValue {
  return value as JsonValue;
}

function normalizePaste(value: Paste): Paste {
  return {
    ...value,
    tags: Array.isArray(value.tags) ? value.tags : [],
    files: Array.isArray(value.files) ? value.files : [],
  };
}

function normalizeSummary(value: PasteSummary): PasteSummary {
  return { ...value, tags: Array.isArray(value.tags) ? value.tags : [] };
}

export function usePasteApi() {
  const api = useFoundationApi("platform");

  return {
    getAdministrationSession() {
      return api.request<{ allowed: boolean; userKey: string }>("/api/v1/admin/session", { auth: "required" });
    },
    async create(input: PasteWriteInput) {
      const response = await api.request<{ paste: Paste }>("/api/v1/pastes", {
        method: "POST",
        body: jsonBody(input),
        auth: "optional",
      });
      return { paste: normalizePaste(response.paste) };
    },
    async get(code: string) {
      const response = await api.request<{ paste: Paste }>(
        `/api/v1/pastes/${encodeURIComponent(code)}`,
        { auth: "optional" },
      );
      return { paste: normalizePaste(response.paste) };
    },
    async access(code: string, password: string) {
      const response = await api.request<{ paste: Paste }>(
        `/api/v1/pastes/${encodeURIComponent(code)}/access`,
        {
          method: "POST",
          body: jsonBody({ password }),
          auth: "optional",
        },
      );
      return { paste: normalizePaste(response.paste) };
    },
    async listMine(query: { page?: number; size?: number; q?: string } = {}) {
      const response = await api.request<{ items: PasteSummary[]; page: number; size: number; total: number }>("/api/v1/me/pastes", {
        query,
        auth: "required",
      });
      return { ...response, items: response.items.map(normalizeSummary) };
    },
    async getSettings() {
      return api.request<{ settings: SiteSettings }>("/api/v1/settings", {
        auth: "optional",
      });
    },
    async getMine(code: string) {
      const response = await api.request<{ paste: Paste }>(
        `/api/v1/me/pastes/${encodeURIComponent(code)}`,
        { auth: "required" },
      );
      return { paste: normalizePaste(response.paste) };
    },
    async update(code: string, expectedRevision: number, input: PasteWriteInput) {
      const response = await api.request<{ paste: Paste }>(
        `/api/v1/me/pastes/${encodeURIComponent(code)}`,
        {
          method: "PATCH",
          body: jsonBody({ expectedRevision, ...input }),
          auth: "required",
        },
      );
      return { paste: normalizePaste(response.paste) };
    },
    async patch(code: string, expectedRevision: number, input: PastePatchInput) {
      const response = await api.request<{ paste: Paste }>(
        `/api/v1/me/pastes/${encodeURIComponent(code)}`,
        {
          method: "PATCH",
          body: jsonBody({ expectedRevision, ...input }),
          auth: "required",
        },
      );
      return { paste: normalizePaste(response.paste) };
    },
    remove(code: string, expectedRevision: number) {
      return api.request<void>(
        `/api/v1/me/pastes/${encodeURIComponent(code)}`,
        {
          method: "DELETE",
          query: { expectedRevision },
          auth: "required",
        },
      );
    },
    async listAdministration(query: AdministrationPasteQuery = {}) {
      const response = await api.request<AdministrationPastePage>("/api/v1/admin/pastes", {
        auth: "required",
        query: query as Record<string, string | number>,
      });
      return {
        ...response,
        items: response.items.map((value) => normalizeSummary(value) as AdministrationPaste),
      };
    },
    async govern(code: string, expectedRevision: number, input: PastePatchInput) {
      const response = await api.request<{ paste: AdministrationPaste }>(
        `/api/v1/admin/pastes/${encodeURIComponent(code)}`,
        {
          method: "PATCH",
          body: jsonBody({ expectedRevision, ...input }),
          auth: "required",
        },
      );
      return { paste: normalizeSummary(response.paste) as AdministrationPaste };
    },
    removeAdministration(code: string, expectedRevision: number) {
      return api.request<void>(
        `/api/v1/admin/pastes/${encodeURIComponent(code)}`,
        {
          method: "DELETE",
          query: { expectedRevision },
          auth: "required",
        },
      );
    },
    listAdministrationUsers(query: AdministrationUserQuery = {}) {
      return api.request<AdministrationUserPage>("/api/v1/admin/users", {
        auth: "required",
        query: query as Record<string, string | number>,
      });
    },
    updateAdministrationUser(userKey: string, input: AdministrationUserPolicyInput) {
      return api.request<{ user: AdministrationUserPolicy }>(
        `/api/v1/admin/users/${encodeURIComponent(userKey)}`,
        {
          method: "PATCH",
          body: jsonBody(input),
          auth: "required",
        },
      );
    },
    getGovernanceSettings() {
      return api.request<{ settings: GovernanceSettings }>("/api/v1/admin/governance-settings", {
        auth: "required",
      });
    },
    updateGovernanceSettings(input: GovernanceSettingsInput) {
      return api.request<{ settings: GovernanceSettings }>("/api/v1/admin/governance-settings", {
        method: "PATCH",
        body: jsonBody(input),
        auth: "required",
      });
    },
    updateSettings(input: SiteSettingsInput) {
      return api.request<{ settings: SiteSettings }>("/api/v1/admin/settings", {
        method: "PATCH",
        body: jsonBody(input),
        auth: "required",
      });
    },
  };
}

export function usePasteTransfer() {
  return useState<Record<string, Paste>>("paste-created-transfer", () => ({}));
}
