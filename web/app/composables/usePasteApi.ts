import type { JsonValue } from "@yueli/http-runtime";
import { useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import type {
  Paste,
  PastePatchInput,
  PasteSummary,
  PasteWriteInput,
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
    async listMine() {
      const response = await api.request<{ pastes: PasteSummary[] }>("/api/v1/me/pastes", {
        auth: "required",
      });
      return { pastes: (response.pastes || []).map(normalizeSummary) };
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
      return api.request<Record<string, never>>(
        `/api/v1/me/pastes/${encodeURIComponent(code)}`,
        {
          method: "DELETE",
          query: { expectedRevision },
          auth: "required",
        },
      );
    },
  };
}

export function usePasteTransfer() {
  return useState<Record<string, Paste>>("paste-created-transfer", () => ({}));
}
