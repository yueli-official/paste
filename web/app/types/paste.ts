export type PasteVisibility = "unlisted" | "private";

export interface PasteFile {
  path: string;
  language: string;
  content: string;
  order?: number;
}

export interface Paste {
  id: string;
  code: string;
  shareUrl: string;
  title: string;
  description?: string;
  tags: string[];
  files: PasteFile[];
  visibility: PasteVisibility;
  passwordProtected: boolean;
  state: "active" | "deleted";
  revision: number;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface PasteSummary {
  code: string;
  shareUrl: string;
  title: string;
  tags: string[];
  fileCount: number;
  primaryLanguage: string;
  visibility: PasteVisibility;
  passwordProtected: boolean;
  state: "active" | "deleted";
  revision: number;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface PasteWriteInput {
  title: string;
  description: string;
  tags: string[];
  files: PasteFile[];
  visibility: PasteVisibility;
  password: string;
  expiresAt?: string;
  clearExpiry?: boolean;
}
