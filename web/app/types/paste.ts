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

export interface AdministrationPaste extends PasteSummary {
  ownerUserKey?: string;
}

export interface AdministrationPastePage {
  pastes: AdministrationPaste[];
  total: number;
  limit: number;
  offset: number;
}

export interface AdministrationPasteQuery {
  q?: string;
  visibility?: "" | PasteVisibility;
  state?: "" | "active" | "deleted";
  ownership?: "" | "anonymous" | "owned";
  limit?: number;
  offset?: number;
}

export type AdministrationUserState = "active" | "suspended";

export interface AdministrationUser {
  userKey: string;
  state: AdministrationUserState;
  dailyLimitOverride?: number;
  effectiveDailyLimit: number;
  usedToday: number;
  totalPastes: number;
  activePastes: number;
  lastCreatedAt?: string;
  reason?: string;
  revision: number;
  updatedAt?: string;
}

export interface AdministrationUserPage {
  users: AdministrationUser[];
  total: number;
  limit: number;
  offset: number;
}

export interface AdministrationUserQuery {
  q?: string;
  state?: "" | AdministrationUserState;
  limit?: number;
  offset?: number;
}

export interface AdministrationUserPolicyInput {
  state: AdministrationUserState;
  dailyLimitOverride?: number;
  clearDailyLimit?: boolean;
  reason: string;
  expectedRevision: number;
}

export interface AdministrationUserPolicy {
  userKey: string;
  state: AdministrationUserState;
  dailyLimitOverride?: number;
  reason?: string;
  revision: number;
  updatedAt: string;
}

export interface SiteSettings {
  name: string;
  description: string;
  revision: number;
  updatedAt: string;
}

export interface GovernanceSettings {
  userDailyLimit: number;
  anonymousDailyLimit: number;
  revision: number;
  updatedAt: string;
}

export interface GovernanceSettingsInput {
  userDailyLimit: number;
  anonymousDailyLimit: number;
  expectedRevision: number;
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

export interface PastePatchInput {
  visibility?: PasteVisibility;
  expiresAt?: string;
  clearExpiry?: boolean;
}

export interface SiteSettingsInput {
  name: string;
  description: string;
  expectedRevision: number;
}
