<!--
THESIS: Site governance is a focused inspection desk, not a dashboard of summary cards.
OWN-WORLD: Cool porcelain and ink-navy editor planes use mineral-blue focus, flat rules, and compact Nuxt UI controls.
STORY: find records or users, select one or many, apply bounded governance, then tune public identity and abuse controls in one save flow.
FIRST VIEWPORT: one 40px application bar, one section rail, a selection-aware toolbar or settings command bar, the work plane, and a 28px status bar.
FORM: the fourth grounded master-detail explorer extended with contextual user batching and a flat VS Code-like settings catalog; seed eb959727.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import { PageHeader } from "@yueli/ui/admin";
import type {
  AdministrationPaste,
  AdministrationUser,
  AdministrationUserState,
  GovernanceSettings,
  PastePatchInput,
  PasteVisibility,
} from "../types/paste";
import { displayTitle, pasteErrorMessage } from "../utils/paste";

definePageMeta({ layout: "admin", middleware: ["auth", "operator"] });

const api = usePasteApi();
const route = useRoute();
const toast = useToast();
const { settings, siteName, adopt } = useSiteSettings();

const section = computed<"pastes" | "users" | "settings">(() => {
  if (route.query.view === "users") return "users";
  if (route.query.view === "settings") return "settings";
  return "pastes";
});
const pageTitle = computed(() => ({
  pastes: "片段治理",
  users: "用户治理",
  settings: "站点设置",
})[section.value]);
const pageIcon = computed(() => ({
  pastes: "i-tabler-files",
  users: "i-tabler-users",
  settings: "i-tabler-adjustments-horizontal",
})[section.value]);
const values = ref<AdministrationPaste[]>([]);
const total = ref(0);
const offset = ref(0);
const limit = 50;
const query = ref("");
const visibility = ref<"all" | PasteVisibility>("all");
const state = ref<"all" | "active" | "deleted">("active");
const ownership = ref<"all" | "anonymous" | "owned">("all");
const loading = ref(true);
const error = ref("");
const selectedCodes = ref<string[]>([]);
const inspectedCode = ref("");
const batchOpen = ref(false);
const batchVisibility = ref<"keep" | PasteVisibility>("keep");
const batchExpiry = ref<"keep" | "1d" | "7d" | "30d" | "never">("keep");
const batchSaving = ref(false);
const batchDeleting = ref(false);
const operationMessage = ref("");
const settingsName = ref(settings.value.name);
const settingsDescription = ref(settings.value.description);
const settingsSaving = ref(false);
const settingsError = ref("");
const users = ref<AdministrationUser[]>([]);
const userTotal = ref(0);
const userOffset = ref(0);
const userQuery = ref("");
const userState = ref<"all" | AdministrationUserState>("all");
const userLoading = ref(true);
const userError = ref("");
const inspectedUserKey = ref("");
const userPolicySaving = ref(false);
const selectedUserKeys = ref<string[]>([]);
const userBatchOpen = ref(false);
const userBatchState = ref<"keep" | AdministrationUserState>("keep");
const userBatchLimitMode = ref<"keep" | "default" | "custom">("keep");
const userBatchLimit = ref<number | null>(null);
const userBatchSaving = ref(false);
const userDraftState = ref<AdministrationUserState>("active");
const userDraftCustomLimit = ref(false);
const userDraftLimit = ref<number | null>(null);
const userDraftReason = ref("");
const governanceSettings = ref<GovernanceSettings | null>(null);
const governanceUserLimit = ref<number | null>(null);
const governanceAnonymousLimit = ref<number | null>(null);
const governanceSettingsLoading = ref(true);
const governanceSettingsSaving = ref(false);
const governanceSettingsError = ref("");

const visibilityItems = [
  { label: "全部可见性", value: "all" },
  { label: "持链访问", value: "unlisted" },
  { label: "仅自己", value: "private" },
];
const stateItems = [
  { label: "全部状态", value: "all" },
  { label: "有效", value: "active" },
  { label: "已删除", value: "deleted" },
];
const userStateItems = [
  { label: "全部创建状态", value: "all" },
  { label: "正常创建", value: "active" },
  { label: "已暂停创建", value: "suspended" },
];
const userPolicyStateItems = [
  { label: "正常创建", value: "active" },
  { label: "暂停创建", value: "suspended" },
];
const userBatchStateItems = [
  { label: "保持原状态", value: "keep" },
  { label: "正常创建", value: "active" },
  { label: "暂停创建", value: "suspended" },
];
const userBatchLimitItems = [
  { label: "保持原额度", value: "keep" },
  { label: "跟随全站默认", value: "default" },
  { label: "设置统一上限", value: "custom" },
];
const ownershipItems = [
  { label: "全部归属", value: "all" },
  { label: "登录用户", value: "owned" },
  { label: "匿名", value: "anonymous" },
];
const batchVisibilityItems = [
  { label: "保持原设置", value: "keep" },
  { label: "持链访问", value: "unlisted" },
  { label: "仅自己", value: "private" },
];
const batchExpiryItems = [
  { label: "保持原设置", value: "keep" },
  { label: "1 天后", value: "1d" },
  { label: "7 天后", value: "7d" },
  { label: "30 天后", value: "30d" },
  { label: "不过期", value: "never" },
];

const selectedCodeSet = computed(() => new Set(selectedCodes.value));
const selectedValues = computed(() => values.value.filter((value) => selectedCodeSet.value.has(value.code)));
const inspected = computed(() => values.value.find((value) => value.code === inspectedCode.value));
const selectionState = computed<boolean | "indeterminate">(() => {
  if (values.value.length === 0) return false;
  const count = values.value.filter((value) => selectedCodeSet.value.has(value.code)).length;
  if (count === 0) return false;
  return count === values.value.length ? true : "indeterminate";
});
const activeFilterCount = computed(() =>
  [visibility.value !== "all", state.value !== "active", ownership.value !== "all"].filter(Boolean).length,
);
const pageStart = computed(() => total.value === 0 ? 0 : offset.value + 1);
const pageEnd = computed(() => Math.min(offset.value + values.value.length, total.value));
const currentPage = computed({
  get: () => Math.floor(offset.value / limit) + 1,
  set: (value: number) => {
    const pageCount = Math.max(1, Math.ceil(total.value / limit));
    const nextPage = Math.min(Math.max(1, value), pageCount);
    const nextOffset = (nextPage - 1) * limit;
    if (nextOffset === offset.value) return;
    offset.value = nextOffset;
    void loadPastes();
  },
});
const settingsDirty = computed(() =>
  settingsName.value.trim() !== settings.value.name ||
  settingsDescription.value.trim() !== settings.value.description,
);
const inspectedUser = computed(() => users.value.find((value) => value.userKey === inspectedUserKey.value));
const selectedUserKeySet = computed(() => new Set(selectedUserKeys.value));
const selectedUsers = computed(() => users.value.filter((value) => selectedUserKeySet.value.has(value.userKey)));
const userSelectionState = computed<boolean | "indeterminate">(() => {
  if (users.value.length === 0) return false;
  const count = users.value.filter((value) => selectedUserKeySet.value.has(value.userKey)).length;
  if (count === 0) return false;
  return count === users.value.length ? true : "indeterminate";
});
const userPageStart = computed(() => userTotal.value === 0 ? 0 : userOffset.value + 1);
const userPageEnd = computed(() => Math.min(userOffset.value + users.value.length, userTotal.value));
const currentUserPage = computed({
  get: () => Math.floor(userOffset.value / limit) + 1,
  set: (value: number) => {
    const pageCount = Math.max(1, Math.ceil(userTotal.value / limit));
    const nextPage = Math.min(Math.max(1, value), pageCount);
    const nextOffset = (nextPage - 1) * limit;
    if (nextOffset === userOffset.value) return;
    userOffset.value = nextOffset;
    void loadUsers();
  },
});
const userPolicyDirty = computed(() => {
  if (!inspectedUser.value) return false;
  const currentLimit = inspectedUser.value.dailyLimitOverride;
  return userDraftState.value !== inspectedUser.value.state ||
    userDraftCustomLimit.value !== (currentLimit !== undefined) ||
    (userDraftCustomLimit.value && userDraftLimit.value !== currentLimit) ||
    userDraftReason.value.trim() !== (inspectedUser.value.reason || "");
});
const governanceSettingsDirty = computed(() =>
  Boolean(governanceSettings.value) && (
    governanceUserLimit.value !== governanceSettings.value?.userDailyLimit ||
    governanceAnonymousLimit.value !== governanceSettings.value?.anonymousDailyLimit
  ),
);
const settingsWorkspaceDirty = computed(() => settingsDirty.value || governanceSettingsDirty.value);
const settingsWorkspaceSaving = computed(() => settingsSaving.value || governanceSettingsSaving.value);

useSeoMeta({
  title: computed(() => `管理后台 · ${siteName.value}`),
  description: computed(() => `治理全站代码片段并维护 ${siteName.value} 的展示设置。`),
});

let loadSequence = 0;
async function loadPastes(reset = false) {
  if (reset) offset.value = 0;
  const sequence = ++loadSequence;
  loading.value = true;
  error.value = "";
  try {
    const page = await api.listAdministration({
      q: query.value.trim() || undefined,
      visibility: visibility.value === "all" ? undefined : visibility.value,
      state: state.value === "all" ? undefined : state.value,
      ownership: ownership.value === "all" ? undefined : ownership.value,
      limit,
      offset: offset.value,
    });
    if (sequence !== loadSequence) return;
    values.value = page.pastes;
    total.value = page.total;
    const available = new Set(values.value.map((value) => value.code));
    selectedCodes.value = selectedCodes.value.filter((code) => available.has(code));
    if (inspectedCode.value && !available.has(inspectedCode.value)) inspectedCode.value = "";
  } catch (caught) {
    if (sequence === loadSequence) error.value = pasteErrorMessage(caught);
  } finally {
    if (sequence === loadSequence) loading.value = false;
  }
}

let loadUsersSequence = 0;
async function loadUsers(reset = false) {
  if (reset) userOffset.value = 0;
  const sequence = ++loadUsersSequence;
  userLoading.value = true;
  userError.value = "";
  try {
    const page = await api.listAdministrationUsers({
      q: userQuery.value.trim() || undefined,
      state: userState.value === "all" ? undefined : userState.value,
      limit,
      offset: userOffset.value,
    });
    if (sequence !== loadUsersSequence) return;
    users.value = page.users || [];
    userTotal.value = page.total;
    const available = new Set(users.value.map((value) => value.userKey));
    selectedUserKeys.value = selectedUserKeys.value.filter((userKey) => available.has(userKey));
    if (inspectedUserKey.value && !users.value.some((value) => value.userKey === inspectedUserKey.value)) {
      inspectedUserKey.value = "";
    }
  } catch (caught) {
    if (sequence === loadUsersSequence) userError.value = pasteErrorMessage(caught);
  } finally {
    if (sequence === loadUsersSequence) userLoading.value = false;
  }
}

async function loadGovernanceSettings() {
  governanceSettingsLoading.value = true;
  governanceSettingsError.value = "";
  try {
    const response = await api.getGovernanceSettings();
    governanceSettings.value = response.settings;
    governanceUserLimit.value = response.settings.userDailyLimit;
    governanceAnonymousLimit.value = response.settings.anonymousDailyLimit;
  } catch (caught) {
    governanceSettingsError.value = pasteErrorMessage(caught);
  } finally {
    governanceSettingsLoading.value = false;
  }
}

function announce(message: string, color: "success" | "warning" | "error") {
  operationMessage.value = message;
  if (color === "success") return;
  toast.add({
    title: color === "warning" ? "部分项目未完成" : "操作未完成",
    description: message,
    color,
  });
}

function toggleSelected(code: string, checked: boolean) {
  const next = new Set(selectedCodes.value);
  if (checked) next.add(code);
  else next.delete(code);
  selectedCodes.value = [...next];
}

function togglePageSelection(checked: boolean | "indeterminate") {
  selectedCodes.value = checked === true ? values.value.map((value) => value.code) : [];
}

function clearSelection() {
  selectedCodes.value = [];
}

function toggleUserSelected(userKey: string, checked: boolean) {
  const next = new Set(selectedUserKeys.value);
  if (checked) next.add(userKey);
  else next.delete(userKey);
  selectedUserKeys.value = [...next];
}

function toggleUserPageSelection(checked: boolean | "indeterminate") {
  selectedUserKeys.value = checked === true ? users.value.map((value) => value.userKey) : [];
}

function clearUserSelection() {
  selectedUserKeys.value = [];
}

function openUserBatch() {
  userBatchState.value = "keep";
  userBatchLimitMode.value = "keep";
  userBatchLimit.value = governanceSettings.value?.userDailyLimit ?? 50;
  userBatchOpen.value = true;
}

function closeUserBatch() {
  userBatchOpen.value = false;
}

function clearFilters() {
  query.value = "";
  visibility.value = "all";
  state.value = "active";
  ownership.value = "all";
}

function openBatch() {
  batchVisibility.value = "keep";
  batchExpiry.value = "keep";
  batchOpen.value = true;
}

function closeBatch() {
  batchOpen.value = false;
}

function inspect(code: string) {
  inspectedCode.value = code;
}

function closeInspector() {
  inspectedCode.value = "";
}

function inspectUser(userKey: string) {
  const user = users.value.find((value) => value.userKey === userKey);
  if (!user) return;
  inspectedUserKey.value = userKey;
  userDraftState.value = user.state;
  userDraftCustomLimit.value = user.dailyLimitOverride !== undefined;
  userDraftLimit.value = user.dailyLimitOverride ?? user.effectiveDailyLimit;
  userDraftReason.value = user.reason || "";
}

function closeUserInspector() {
  inspectedUserKey.value = "";
}

function viewUserPastes(userKey: string) {
  query.value = userKey;
}

async function updateUserState(user: AdministrationUser, nextState: AdministrationUserState) {
  if (nextState === "suspended" && !window.confirm(`暂停 ${user.userKey} 创建新的代码片段？已有链接不会自动删除。`)) return;
  userPolicySaving.value = true;
  try {
    await api.updateAdministrationUser(user.userKey, {
      state: nextState,
      dailyLimitOverride: user.dailyLimitOverride,
      clearDailyLimit: user.dailyLimitOverride === undefined,
      reason: user.reason || "",
      expectedRevision: user.revision,
    });
    await loadUsers();
    announce(nextState === "suspended" ? `已暂停 ${user.userKey} 创建代码片段。` : `已恢复 ${user.userKey} 创建代码片段。`, "success");
  } catch (caught) {
    announce(pasteErrorMessage(caught), "error");
  } finally {
    userPolicySaving.value = false;
  }
}

async function saveUserPolicy() {
  const user = inspectedUser.value;
  if (!user || !userPolicyDirty.value) return;
  if (userDraftCustomLimit.value && (userDraftLimit.value === null || userDraftLimit.value < 1)) {
    announce("单独上限必须至少为 1。", "error");
    return;
  }
  userPolicySaving.value = true;
  try {
    await api.updateAdministrationUser(user.userKey, {
      state: userDraftState.value,
      dailyLimitOverride: userDraftCustomLimit.value ? userDraftLimit.value! : undefined,
      clearDailyLimit: !userDraftCustomLimit.value,
      reason: userDraftReason.value,
      expectedRevision: user.revision,
    });
    await loadUsers();
    if (inspectedUserKey.value) inspectUser(inspectedUserKey.value);
    announce(`已更新 ${user.userKey} 的创建策略。`, "success");
  } catch (caught) {
    announce(pasteErrorMessage(caught), "error");
  } finally {
    userPolicySaving.value = false;
  }
}

async function applyUserBatch() {
  const targets = [...selectedUsers.value];
  if (targets.length === 0) return;
  if (userBatchState.value === "keep" && userBatchLimitMode.value === "keep") return;
  if (userBatchLimitMode.value === "custom" && (userBatchLimit.value === null || userBatchLimit.value < 1 || userBatchLimit.value > 10_000)) {
    announce("统一上限必须在 1–10,000 之间。", "error");
    return;
  }
  userBatchSaving.value = true;
  const result = await runBounded(targets, async (user) => {
    const clearDailyLimit = userBatchLimitMode.value === "default"
      || (userBatchLimitMode.value === "keep" && user.dailyLimitOverride === undefined);
    const dailyLimitOverride = userBatchLimitMode.value === "custom"
      ? userBatchLimit.value!
      : userBatchLimitMode.value === "keep"
        ? user.dailyLimitOverride
        : undefined;
    return (await api.updateAdministrationUser(user.userKey, {
      state: userBatchState.value === "keep" ? user.state : userBatchState.value,
      dailyLimitOverride,
      clearDailyLimit,
      reason: user.reason || "",
      expectedRevision: user.revision,
    })).user;
  });
  userBatchSaving.value = false;
  await loadUsers();
  selectedUserKeys.value = result.failed.map((user) => user.userKey);
  if (result.failed.length) {
    announce(`已更新 ${result.succeeded.length} 个用户，${result.failed.length} 个失败；失败项仍保持选中。`, "warning");
  } else {
    userBatchOpen.value = false;
    announce(`已更新 ${result.succeeded.length} 个用户的创建策略。`, "success");
  }
}

async function runBounded<T, R>(items: T[], operation: (item: T) => Promise<R>) {
  const succeeded: Array<{ item: T; value: R }> = [];
  const failed: T[] = [];
  let cursor = 0;
  const workers = Array.from({ length: Math.min(4, items.length) }, async () => {
    while (cursor < items.length) {
      const item = items[cursor++];
      if (!item) continue;
      try {
        succeeded.push({ item, value: await operation(item) });
      } catch {
        failed.push(item);
      }
    }
  });
  await Promise.all(workers);
  return { succeeded, failed };
}

function governancePatch(): PastePatchInput {
  const patch: PastePatchInput = {};
  if (batchVisibility.value !== "keep") patch.visibility = batchVisibility.value;
  if (batchExpiry.value === "never") patch.clearExpiry = true;
  else if (batchExpiry.value !== "keep") {
    const hours = { "1d": 24, "7d": 168, "30d": 720 }[batchExpiry.value];
    patch.expiresAt = new Date(Date.now() + hours * 60 * 60 * 1000).toISOString();
  }
  return patch;
}

async function applyBatch() {
  const targets = [...selectedValues.value];
  if (targets.length === 0) return;
  const patch = governancePatch();
  if (Object.keys(patch).length === 0) return;
  batchSaving.value = true;
  const result = await runBounded(targets, async (value) =>
    (await api.govern(value.code, value.revision, patch)).paste,
  );
  const updates = new Map(result.succeeded.map(({ item, value }) => [item.code, value]));
  values.value = values.value.map((value) => updates.get(value.code) || value);
  selectedCodes.value = result.failed.map((value) => value.code);
  batchSaving.value = false;
  if (result.failed.length) {
    announce(`已修改 ${result.succeeded.length} 项，${result.failed.length} 项失败；匿名片段不能改为仅自己。`, "warning");
  } else {
    batchOpen.value = false;
    announce(`已修改 ${result.succeeded.length} 个代码片段。`, "success");
  }
}

async function deleteTargets(targets: AdministrationPaste[]) {
  if (targets.length === 0 || !window.confirm(`删除选中的 ${targets.length} 个代码片段？分享链接将永久失效。`)) return;
  batchDeleting.value = true;
  const result = await runBounded(targets, (value) => api.removeAdministration(value.code, value.revision));
  const removed = new Set(result.succeeded.map(({ item }) => item.code));
  values.value = values.value.filter((value) => !removed.has(value.code));
  total.value = Math.max(0, total.value - removed.size);
  selectedCodes.value = result.failed.map((value) => value.code);
  if (inspectedCode.value && removed.has(inspectedCode.value)) inspectedCode.value = "";
  batchDeleting.value = false;
  announce(
    result.failed.length
      ? `已删除 ${result.succeeded.length} 项，${result.failed.length} 项失败，失败项仍保持选中。`
      : `已删除 ${result.succeeded.length} 个代码片段。`,
    result.failed.length ? "warning" : "success",
  );
}

async function saveSettings(silent = false): Promise<boolean> {
  if (!settingsDirty.value) return true;
  settingsSaving.value = true;
  settingsError.value = "";
  try {
    const response = await api.updateSettings({
      name: settingsName.value,
      description: settingsDescription.value,
      expectedRevision: settings.value.revision,
    });
    adopt(response.settings);
    settingsName.value = response.settings.name;
    settingsDescription.value = response.settings.description;
    if (!silent) announce("站点展示设置已更新。", "success");
    return true;
  } catch (caught) {
    settingsError.value = pasteErrorMessage(caught);
    return false;
  } finally {
    settingsSaving.value = false;
  }
}

function resetSettingsDraft() {
  settingsName.value = settings.value.name;
  settingsDescription.value = settings.value.description;
  settingsError.value = "";
}

function resetGovernanceSettingsDraft() {
  if (!governanceSettings.value) return;
  governanceUserLimit.value = governanceSettings.value.userDailyLimit;
  governanceAnonymousLimit.value = governanceSettings.value.anonymousDailyLimit;
  governanceSettingsError.value = "";
}

function numberInputValue(event: Event): number | null {
  const raw = (event.target as HTMLInputElement).value.trim().replaceAll(",", "").replaceAll(" ", "");
  if (!raw) return null;
  const value = Number(raw);
  return Number.isFinite(value) ? value : null;
}

async function saveGovernanceSettings(silent = false): Promise<boolean> {
  if (!governanceSettingsDirty.value) return true;
  if (!governanceSettings.value || governanceUserLimit.value === null || governanceAnonymousLimit.value === null) return false;
  governanceSettingsSaving.value = true;
  governanceSettingsError.value = "";
  try {
    const response = await api.updateGovernanceSettings({
      userDailyLimit: governanceUserLimit.value,
      anonymousDailyLimit: governanceAnonymousLimit.value,
      expectedRevision: governanceSettings.value.revision,
    });
    governanceSettings.value = response.settings;
    resetGovernanceSettingsDraft();
    await loadUsers();
    if (!silent) announce("创建限制已更新。", "success");
    return true;
  } catch (caught) {
    governanceSettingsError.value = pasteErrorMessage(caught);
    return false;
  } finally {
    governanceSettingsSaving.value = false;
  }
}

async function saveSettingsWorkspace() {
  if (!settingsWorkspaceDirty.value || settingsWorkspaceSaving.value) return;
  const publicDirty = settingsDirty.value;
  const governanceDirty = governanceSettingsDirty.value;
  const results: boolean[] = [];
  if (publicDirty) results.push(await saveSettings(true));
  if (governanceDirty) results.push(await saveGovernanceSettings(true));
  const saved = results.filter(Boolean).length;
  if (saved === results.length) {
    announce(saved === 1 ? "已保存 1 组站点设置。" : `已保存 ${saved} 组站点设置。`, "success");
  } else if (saved > 0) {
    announce(`已保存 ${saved} 组设置，其余设置需要修正后重试。`, "warning");
  } else {
    announce("站点设置未能保存，请检查页面中的错误。", "error");
  }
}

function resetSettingsWorkspace() {
  resetSettingsDraft();
  resetGovernanceSettingsDraft();
}

function focusSearch(event: KeyboardEvent) {
  if (event.key !== "/" || event.metaKey || event.ctrlKey || event.altKey || section.value === "settings") return;
  const target = event.target as HTMLElement | null;
  if (target?.matches("input, textarea, [contenteditable='true']")) return;
  event.preventDefault();
  document.querySelector<HTMLInputElement>(section.value === "users" ? "#admin-user-search" : "#admin-search")?.focus();
}

let queryTimer: number | undefined;
let userQueryTimer: number | undefined;
watch(query, () => {
  window.clearTimeout(queryTimer);
  queryTimer = window.setTimeout(() => loadPastes(true), 240);
});
watch([visibility, state, ownership], () => loadPastes(true));
watch(userQuery, () => {
  window.clearTimeout(userQueryTimer);
  userQueryTimer = window.setTimeout(() => loadUsers(true), 240);
});
watch(userState, () => loadUsers(true));
watch(settings, resetSettingsDraft);
onMounted(() => {
  loadPastes();
  loadUsers();
  loadGovernanceSettings();
  window.addEventListener("keydown", focusSearch);
});
onBeforeUnmount(() => {
  window.clearTimeout(queryTimer);
  window.clearTimeout(userQueryTimer);
  window.removeEventListener("keydown", focusSearch);
});
</script>

<template>
  <div class="w-full space-y-5">
    <PageHeader :title="pageTitle" :icon="pageIcon">
      <template v-if="section === 'settings'" #actions>
        <UButton
          class="max-[720px]:min-h-11 max-[720px]:w-11 max-[720px]:min-w-11 max-[720px]:px-0 max-[720px]:[&_[data-slot=label]]:sr-only"
          type="button"
          label="放弃修改"
          icon="i-tabler-arrow-back-up"
          color="neutral"
          variant="ghost"
          :disabled="!settingsWorkspaceDirty || settingsWorkspaceSaving"
          @click="resetSettingsWorkspace"
        />
        <UButton
          type="submit"
          form="paste-settings-workspace"
          :label="settingsWorkspaceSaving ? '保存中' : settingsWorkspaceDirty ? '保存更改' : '已保存'"
          icon="i-tabler-device-floppy"
          :color="settingsWorkspaceDirty ? 'primary' : 'neutral'"
          :variant="settingsWorkspaceDirty ? 'solid' : 'ghost'"
          :loading="settingsWorkspaceSaving"
          :disabled="!settingsWorkspaceDirty || settingsWorkspaceSaving || governanceUserLimit === null || governanceAnonymousLimit === null"
        />
      </template>
    </PageHeader>

    <section v-if="section === 'pastes'" class="yueli-card flex h-[calc(100svh-14.5rem)] min-h-[32rem] flex-col overflow-hidden bg-[var(--paste-editor)] text-[var(--paste-ink)] max-[720px]:h-[calc(100svh-12rem)] max-[720px]:min-h-[30rem]" aria-label="片段治理">
      <div class="flex min-w-0 items-center gap-[5px] border-b border-[var(--paste-line)] bg-[var(--paste-editor)] py-[3px] pr-1.5 pl-2.5 max-[720px]:gap-[3px] max-[720px]:p-[4px_5px] max-[720px]:[&_button]:min-h-11" role="toolbar" aria-label="全站代码片段工具栏">
        <UCheckbox
          :model-value="selectionState"
          size="sm"
          class="paste-admin-select-all h-7 w-6 min-w-6 shrink-0 items-center justify-start max-[720px]:h-11 max-[720px]:w-11 max-[720px]:min-w-11 max-[720px]:justify-center max-[720px]:[&_[role=checkbox]]:relative max-[720px]:[&_[role=checkbox]]:before:absolute max-[720px]:[&_[role=checkbox]]:before:-inset-[15px] max-[720px]:[&_[role=checkbox]]:before:content-['']"
          aria-label="选择当前页"
          :disabled="loading || values.length === 0 || batchSaving || batchDeleting"
          @update:model-value="togglePageSelection"
        />
        <template v-if="selectedValues.length">
          <strong>{{ selectedValues.length }} 项已选择</strong>
          <div class="ml-auto flex items-center gap-0.5">
            <UButton type="button" icon="i-tabler-adjustments" color="neutral" variant="ghost" size="sm" :disabled="batchDeleting" @click="openBatch">批量修改</UButton>
            <UButton type="button" icon="i-tabler-trash" color="error" variant="ghost" size="sm" :loading="batchDeleting" :disabled="batchSaving" @click="deleteTargets(selectedValues)">删除</UButton>
            <UTooltip text="取消选择"><UButton type="button" icon="i-tabler-x" color="neutral" variant="ghost" size="sm" square aria-label="取消全部选择" @click="clearSelection" /></UTooltip>
          </div>
        </template>
        <template v-else>
          <UInput id="admin-search" v-model="query" type="search" icon="i-tabler-search" placeholder="搜索标题、短码、用户或标签" aria-label="搜索全站代码片段" variant="none" size="sm" class="min-w-40 max-w-[620px] flex-1 max-[720px]:min-w-0 max-[720px]:max-w-none" :ui="{ base: 'rounded-none ring-0 text-xs min-[721px]:text-xs max-[720px]:min-h-11' }">
            <template #trailing><kbd class="rounded-[3px] border border-[var(--paste-line)] px-[5px] text-[10px] leading-[18px] text-[var(--paste-ink-dim)] max-[720px]:hidden">/</kbd></template>
          </UInput>
          <UPopover :content="{ align: 'end' }">
            <UButton type="button" icon="i-tabler-filter" color="neutral" variant="ghost" size="sm" :label="activeFilterCount ? `筛选 ${activeFilterCount}` : '筛选'" />
            <template #content>
              <div class="paste-admin-filter-panel grid w-[260px] gap-3.5 p-3.5">
                <UFormField label="可见性"><USelect v-model="visibility" :items="visibilityItems" value-key="value" label-key="label" class="w-full" /></UFormField>
                <UFormField label="状态"><USelect v-model="state" :items="stateItems" value-key="value" label-key="label" class="w-full" /></UFormField>
                <UFormField label="归属"><USelect v-model="ownership" :items="ownershipItems" value-key="value" label-key="label" class="w-full" /></UFormField>
              </div>
            </template>
          </UPopover>
          <UTooltip text="刷新列表"><UButton type="button" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" square aria-label="刷新列表" :loading="loading" @click="loadPastes()" /></UTooltip>
        </template>
      </div>

      <div class="grid min-h-0 min-w-0 grid-cols-[minmax(0,1fr)] overflow-hidden data-[inspecting=true]:grid-cols-[minmax(0,1fr)_minmax(290px,340px)] max-[900px]:data-[inspecting=true]:relative max-[900px]:data-[inspecting=true]:grid-cols-[minmax(0,1fr)]" :data-inspecting="Boolean(inspected)">
        <div class="min-h-0 min-w-0 overflow-auto bg-[var(--paste-editor)]" :aria-busy="loading">
          <div v-if="loading" class="paste-admin-loading min-w-[760px] max-[720px]:min-w-0" role="status" aria-label="正在读取全站代码片段">
            <div v-for="index in 6" :key="index" class="grid min-h-[62px] grid-cols-[32px_minmax(240px,1fr)_minmax(130px,180px)_112px_82px_110px] items-center border-b border-[var(--paste-line)] px-3 py-2 max-[720px]:grid-cols-[34px_minmax(0,1fr)] max-[720px]:[&>:nth-child(n+3)]:hidden"><USkeleton class="size-3.5" /><div><USkeleton class="h-3.5 w-52 max-w-full" /><USkeleton class="mt-2 h-2.5 w-72 max-w-full" /></div><USkeleton class="h-3 w-28" /><USkeleton class="h-3 w-16" /><USkeleton class="h-3 w-16" /></div>
          </div>
          <div v-else-if="error" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]" role="alert"><UIcon name="i-tabler-alert-circle" class="size-7 text-[var(--paste-red)]" /><strong>无法读取治理列表</strong><p>{{ error }}</p><UButton type="button" label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="loadPastes()" /></div>
          <div v-else-if="values.length === 0" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]"><UIcon name="i-tabler-file-search" class="size-7" /><strong>没有符合条件的代码片段</strong><p>调整搜索或筛选条件后再试。</p><UButton type="button" label="清除条件" icon="i-tabler-filter-off" color="neutral" variant="outline" @click="clearFilters" /></div>
          <div v-else class="min-w-[760px] max-[720px]:min-w-0">
            <div class="sticky top-0 z-[2] grid min-h-[30px] grid-cols-[32px_minmax(240px,1fr)_minmax(130px,180px)_112px_82px_110px] items-center border-b border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-3 text-[11px] font-[630] text-[var(--paste-ink-dim)] max-[720px]:hidden" aria-hidden="true"><span /><span>内容</span><span>归属</span><span>访问</span><span>状态</span><span>更新</span></div>
            <div role="list" aria-label="全站代码片段">
              <div v-for="value in values" :key="value.code" class="paste-admin-row group grid min-h-[62px] min-w-0 grid-cols-[32px_minmax(240px,1fr)_minmax(130px,180px)_112px_82px_110px] items-center border-b border-[var(--paste-line)] px-3 py-2 transition-colors duration-100 hover:bg-[var(--paste-editor-active)] focus-within:bg-[var(--paste-editor-active)] data-[selected=true]:bg-[color-mix(in_srgb,var(--paste-blue)_9%,var(--paste-editor))] data-[inspected=true]:[box-shadow:inset_2px_0_var(--paste-blue)] max-[720px]:min-h-28 max-[720px]:grid-cols-[44px_minmax(0,1fr)_auto] max-[720px]:gap-x-2 max-[720px]:gap-y-[7px] max-[720px]:px-2 max-[720px]:py-2.5 max-[720px]:[&>:first-child]:col-start-1 max-[720px]:[&>:first-child]:row-start-1 max-[720px]:[&>:first-child]:row-end-4 max-[720px]:[&>:first-child]:mt-[-7px] max-[720px]:[&>:first-child]:ml-[-8px] max-[720px]:[&>:first-child]:size-11 max-[720px]:[&>:first-child]:self-start max-[720px]:[&>:first-child]:justify-center" role="listitem" :data-selected="selectedCodeSet.has(value.code)" :data-inspected="inspectedCode === value.code">
                <UCheckbox :model-value="selectedCodeSet.has(value.code)" size="sm" :aria-label="`选择 ${displayTitle(value.title)}`" :disabled="batchSaving || batchDeleting || value.state === 'deleted'" @update:model-value="toggleSelected(value.code, $event === true)" />
                <button type="button" class="grid min-w-0 gap-[5px] border-0 bg-transparent py-0 pr-3.5 pl-0 text-left text-inherit max-[720px]:col-start-2 max-[720px]:col-end-[-1] max-[720px]:pr-0 [&>span]:flex [&>span]:min-w-0 [&>span]:gap-2.5 [&>span]:overflow-hidden [&>span]:whitespace-nowrap [&>span]:text-[11px] [&>span]:text-[var(--paste-ink-soft)] [&_code]:font-[680] [&_code]:text-[var(--paste-blue)] [&_strong]:truncate [&_strong]:text-[13px]" :aria-label="`检查 ${displayTitle(value.title)}`" @click="inspect(value.code)">
                  <strong>{{ displayTitle(value.title) }}</strong>
                  <span><code>{{ value.code }}</code><span>{{ value.fileCount }} 个文件</span><span>{{ value.primaryLanguage }}</span><span v-for="tag in value.tags.slice(0, 2)" :key="tag">#{{ tag }}</span></span>
                </button>
                <div class="flex min-w-0 items-center gap-[5px] pr-2.5 text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-2 max-[720px]:row-start-2 [&_span]:truncate"><UIcon :name="value.ownerUserKey ? 'i-tabler-user' : 'i-tabler-user-off'" class="size-4" /><span>{{ value.ownerUserKey || "匿名" }}</span></div>
                <span class="paste-admin-access flex min-w-0 items-center gap-[5px] pr-2.5 text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-2 max-[720px]:row-start-3"><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
                <span class="w-fit rounded-[3px] bg-[var(--paste-green-soft)] px-[5px] py-0.5 text-[10px] font-bold text-[var(--paste-green-ink)] data-[state=deleted]:bg-[var(--paste-red-soft)] data-[state=deleted]:text-[var(--paste-red)] data-[state=suspended]:bg-[var(--paste-red-soft)] data-[state=suspended]:text-[var(--paste-red)] max-[720px]:col-start-3 max-[720px]:row-start-2" :data-state="value.state">{{ value.state === "deleted" ? "已删除" : value.expiresAt && new Date(value.expiresAt) <= new Date() ? "已过期" : "有效" }}</span>
                <time class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-3 max-[720px]:row-start-3 max-[720px]:justify-self-end" :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString("zh-CN") }}</time>
              </div>
            </div>
          </div>
        </div>

        <aside v-if="inspected" class="paste-admin-inspector min-w-0 overflow-y-auto border-l border-[var(--paste-line)] bg-[var(--paste-surface-muted)] max-[900px]:absolute max-[900px]:inset-y-0 max-[900px]:right-0 max-[900px]:z-[5] max-[900px]:w-[min(100%,380px)] max-[900px]:border-l-[var(--paste-line-strong)] max-[900px]:[box-shadow:-18px_0_38px_-30px_rgb(15_23_42/.62)] max-[720px]:w-full max-[720px]:border-l-0 [&>header]:flex [&>header]:min-h-[60px] [&>header]:items-start [&>header]:justify-between [&>header]:gap-2.5 [&>header]:border-b [&>header]:border-[var(--paste-line)] [&>header]:bg-[var(--paste-chrome)] [&>header]:p-3 [&_small]:text-[10px] [&_small]:text-[var(--paste-ink-dim)] [&_h2]:mt-[3px] [&_h2]:text-sm [&_h2]:leading-[1.35] [&_h2]:[overflow-wrap:anywhere] [&_dl]:m-0 [&_dl>div]:grid [&_dl>div]:grid-cols-[76px_minmax(0,1fr)] [&_dl>div]:gap-2.5 [&_dl>div]:border-b [&_dl>div]:border-[var(--paste-line)] [&_dl>div]:px-3 [&_dl>div]:py-2.5 [&_dl>div]:text-[11px] [&_dt]:text-[var(--paste-ink-dim)] [&_dd]:m-0 [&_dd]:min-w-0 [&_dd]:text-[var(--paste-ink)] [&_dd]:[overflow-wrap:anywhere]" aria-labelledby="inspector-title">
          <header><div><small>检查代码片段</small><h2 id="inspector-title">{{ displayTitle(inspected.title) }}</h2></div><UButton type="button" icon="i-tabler-x" color="neutral" variant="ghost" size="sm" square aria-label="关闭检查器" @click="closeInspector" /></header>
          <dl>
            <div><dt>短码</dt><dd><code>{{ inspected.code }}</code></dd></div>
            <div><dt>归属</dt><dd>{{ inspected.ownerUserKey || "匿名创建" }}</dd></div>
            <div><dt>访问</dt><dd>{{ inspected.visibility === "private" ? "仅所有者" : "知道链接即可访问" }}{{ inspected.passwordProtected ? " · 另有密码" : "" }}</dd></div>
            <div><dt>内容摘要</dt><dd>{{ inspected.fileCount }} 个文件 · {{ inspected.primaryLanguage }}</dd></div>
            <div><dt>创建时间</dt><dd>{{ new Date(inspected.createdAt).toLocaleString("zh-CN") }}</dd></div>
            <div><dt>有效期</dt><dd>{{ inspected.expiresAt ? new Date(inspected.expiresAt).toLocaleString("zh-CN") : "不过期" }}</dd></div>
          </dl>
          <p class="m-3 text-[11px] leading-[1.65] text-[var(--paste-ink-soft)]">治理摘要不返回代码正文或密码材料。管理员打开分享链接时仍受原访问边界约束。</p>
          <div class="flex flex-wrap gap-1.5 px-3 pb-4"><UButton :to="`/p/${inspected.code}`" target="_blank" label="打开分享链接" icon="i-tabler-external-link" color="neutral" variant="outline" /><UButton v-if="inspected.state !== 'deleted'" type="button" label="选择此项" icon="i-tabler-square-check" color="primary" variant="soft" @click="toggleSelected(inspected.code, true)" /></div>
        </aside>
      </div>
    </section>

    <section v-else-if="section === 'users'" class="yueli-card flex h-[calc(100svh-14.5rem)] min-h-[32rem] flex-col overflow-hidden bg-[var(--paste-editor)] text-[var(--paste-ink)] max-[720px]:h-[calc(100svh-12rem)] max-[720px]:min-h-[30rem]" aria-label="用户治理">
      <div class="flex min-w-0 items-center gap-[5px] border-b border-[var(--paste-line)] bg-[var(--paste-editor)] py-[3px] pr-1.5 pl-2.5 max-[720px]:gap-[3px] max-[720px]:p-[4px_5px] max-[720px]:[&_button]:min-h-11" role="toolbar" aria-label="Paste 用户治理工具栏">
        <UCheckbox class="paste-admin-select-all h-7 w-6 min-w-6 shrink-0 items-center justify-start max-[720px]:h-11 max-[720px]:w-11 max-[720px]:min-w-11 max-[720px]:justify-center" :model-value="userSelectionState" :disabled="userLoading || users.length === 0 || userBatchSaving" aria-label="选择当前页用户" @update:model-value="toggleUserPageSelection($event)" />
        <template v-if="selectedUsers.length">
          <strong>{{ selectedUsers.length }} 个用户已选择</strong>
          <div class="ml-auto flex items-center gap-0.5"><UButton type="button" icon="i-tabler-x" label="清除" color="neutral" variant="ghost" size="sm" :disabled="userBatchSaving" @click="clearUserSelection" /><UButton type="button" icon="i-tabler-user-cog" label="批量设置" color="primary" variant="soft" size="sm" :disabled="userBatchSaving" @click="openUserBatch" /></div>
        </template>
        <template v-else>
          <UInput id="admin-user-search" v-model="userQuery" type="search" icon="i-tabler-search" placeholder="搜索用户主体标识" aria-label="搜索 Paste 用户" variant="none" size="sm" class="min-w-40 max-w-[620px] flex-1 max-[720px]:min-w-0 max-[720px]:max-w-none" :ui="{ base: 'rounded-none ring-0 text-xs max-[720px]:min-h-11' }"><template #trailing><kbd class="rounded-[3px] border border-[var(--paste-line)] px-[5px] text-[10px] leading-[18px] text-[var(--paste-ink-dim)] max-[720px]:hidden">/</kbd></template></UInput>
          <USelect v-model="userState" :items="userStateItems" value-key="value" label-key="label" aria-label="筛选创建状态" size="sm" class="w-[154px] shrink-0 max-[720px]:w-32" />
          <UTooltip text="刷新用户列表"><UButton type="button" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" square aria-label="刷新用户列表" :loading="userLoading" @click="loadUsers()" /></UTooltip>
        </template>
      </div>

      <div class="grid min-h-0 min-w-0 grid-cols-[minmax(0,1fr)] overflow-hidden data-[inspecting=true]:grid-cols-[minmax(0,1fr)_minmax(290px,340px)] max-[900px]:data-[inspecting=true]:relative max-[900px]:data-[inspecting=true]:grid-cols-[minmax(0,1fr)]" :data-inspecting="Boolean(inspectedUser)">
        <div class="min-h-0 min-w-0 overflow-auto bg-[var(--paste-editor)]" :aria-busy="userLoading">
          <div v-if="userLoading" class="paste-admin-loading min-w-[800px] max-[720px]:min-w-0" role="status" aria-label="正在读取 Paste 用户">
            <div v-for="index in 6" :key="index" class="grid min-h-[62px] grid-cols-[32px_minmax(230px,1fr)_92px_110px_90px_112px_116px] items-center border-b border-[var(--paste-line)] px-3 py-2 max-[720px]:min-h-28 max-[720px]:grid-cols-[minmax(0,1fr)_70px] max-[720px]:gap-2.5 max-[720px]:p-3 max-[720px]:[&>:nth-child(n+3)]:hidden"><USkeleton class="h-3.5 w-40 max-w-full" /><USkeleton class="h-3 w-20" /><USkeleton class="h-3 w-16" /><USkeleton class="h-3 w-14" /></div>
          </div>
          <div v-else-if="userError" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]" role="alert"><UIcon name="i-tabler-alert-circle" class="size-7 text-[var(--paste-red)]" /><strong>无法读取用户治理列表</strong><p>{{ userError }}</p><UButton type="button" label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="loadUsers()" /></div>
          <div v-else-if="users.length === 0" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]"><UIcon name="i-tabler-user-search" class="size-7" /><strong>没有符合条件的使用主体</strong><p>这里仅列出创建过代码片段或已有显式策略的 Identity 用户。</p><UButton v-if="userQuery || userState !== 'all'" type="button" label="清除条件" icon="i-tabler-filter-off" color="neutral" variant="outline" @click="userQuery = ''; userState = 'all'" /></div>
          <div v-else class="min-w-[800px] max-[720px]:min-w-0">
            <div class="sticky top-0 z-[2] grid min-h-[30px] grid-cols-[32px_minmax(230px,1fr)_92px_110px_90px_112px_116px] items-center border-b border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-3 text-[11px] font-[630] text-[var(--paste-ink-dim)] max-[720px]:hidden" aria-hidden="true"><span /><span>使用主体</span><span>创建权限</span><span>今日额度</span><span>片段</span><span>最近创建</span><span /></div>
            <div role="list" aria-label="Paste 使用主体">
              <div v-for="user in users" :key="user.userKey" class="paste-admin-user-row group grid min-h-[62px] min-w-0 grid-cols-[32px_minmax(230px,1fr)_92px_110px_90px_112px_116px] items-center border-b border-[var(--paste-line)] px-3 py-2 transition-colors duration-100 hover:bg-[var(--paste-editor-active)] focus-within:bg-[var(--paste-editor-active)] data-[selected=true]:bg-[color-mix(in_srgb,var(--paste-blue)_9%,var(--paste-editor))] data-[inspected=true]:[box-shadow:inset_2px_0_var(--paste-blue)] max-[720px]:min-h-[150px] max-[720px]:grid-cols-[44px_minmax(0,1fr)_auto] max-[720px]:gap-x-2.5 max-[720px]:gap-y-2 max-[720px]:pt-[11px] max-[720px]:pr-2.5 max-[720px]:pb-[11px] max-[720px]:pl-1" role="listitem" :data-selected="selectedUserKeySet.has(user.userKey)" :data-inspected="inspectedUserKey === user.userKey">
                <UCheckbox class="w-6 justify-start max-[720px]:col-start-1 max-[720px]:row-start-1 max-[720px]:row-end-5 max-[720px]:mt-[-7px] max-[720px]:size-11 max-[720px]:self-start max-[720px]:justify-center max-[720px]:[&_[role=checkbox]]:relative max-[720px]:[&_[role=checkbox]]:before:absolute max-[720px]:[&_[role=checkbox]]:before:-inset-3.5 max-[720px]:[&_[role=checkbox]]:before:content-['']" :model-value="selectedUserKeySet.has(user.userKey)" :aria-label="`选择用户 ${user.userKey}`" :disabled="userBatchSaving" @update:model-value="toggleUserSelected(user.userKey, $event === true)" />
                <button type="button" class="grid min-w-0 gap-[5px] border-0 bg-transparent py-0 pr-3.5 pl-0 text-left text-inherit max-[720px]:col-start-2 max-[720px]:col-end-[-1] max-[720px]:pr-0 [&_strong]:flex [&_strong]:min-w-0 [&_strong]:items-center [&_strong]:gap-1.5 [&_strong]:truncate [&_strong]:font-mono [&_strong]:text-xs [&_strong]:text-[var(--paste-blue)] [&>span]:truncate [&>span]:text-[11px] [&>span]:text-[var(--paste-ink-soft)]" :aria-label="`检查用户 ${user.userKey}`" @click="inspectUser(user.userKey)">
                  <strong><UIcon name="i-tabler-user" class="size-4" />{{ user.userKey }}</strong>
                  <span>{{ user.dailyLimitOverride === undefined ? "跟随全站默认上限" : `单独上限 ${user.dailyLimitOverride}` }}<template v-if="user.reason"> · {{ user.reason }}</template></span>
                </button>
                <span class="w-fit rounded-[3px] bg-[var(--paste-green-soft)] px-[5px] py-0.5 text-[10px] font-bold text-[var(--paste-green-ink)] data-[state=suspended]:bg-[var(--paste-red-soft)] data-[state=suspended]:text-[var(--paste-red)] max-[720px]:col-start-2 max-[720px]:row-start-2" :data-state="user.state">{{ user.state === "suspended" ? "已暂停" : "正常" }}</span>
                <span class="font-mono text-[11px] text-[var(--paste-ink-soft)] data-[exhausted=true]:font-[720] data-[exhausted=true]:text-[var(--paste-red)] max-[720px]:col-start-3 max-[720px]:row-start-2 max-[720px]:justify-self-end" :data-exhausted="user.usedToday >= user.effectiveDailyLimit">{{ user.usedToday }} / {{ user.effectiveDailyLimit }}</span>
                <span class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-2 max-[720px]:row-start-3">{{ user.activePastes }} / {{ user.totalPastes }}</span>
                <time v-if="user.lastCreatedAt" class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-3 max-[720px]:row-start-3 max-[720px]:justify-self-end" :datetime="user.lastCreatedAt">{{ new Date(user.lastCreatedAt).toLocaleDateString("zh-CN") }}</time><span v-else class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-3 max-[720px]:row-start-3 max-[720px]:justify-self-end">—</span>
                <UButton class="max-[720px]:col-start-2 max-[720px]:col-end-[-1] max-[720px]:row-start-4 max-[720px]:min-h-11 max-[720px]:justify-self-end" type="button" :icon="user.state === 'suspended' ? 'i-tabler-player-play' : 'i-tabler-player-pause'" :label="user.state === 'suspended' ? '恢复创建' : '暂停创建'" :color="user.state === 'suspended' ? 'primary' : 'error'" variant="ghost" size="sm" :loading="userPolicySaving" @click="updateUserState(user, user.state === 'suspended' ? 'active' : 'suspended')" />
              </div>
            </div>
          </div>
        </div>

        <aside v-if="inspectedUser" class="paste-admin-inspector min-w-0 overflow-y-auto border-l border-[var(--paste-line)] bg-[var(--paste-surface-muted)] max-[900px]:absolute max-[900px]:inset-y-0 max-[900px]:right-0 max-[900px]:z-[5] max-[900px]:w-[min(100%,380px)] max-[900px]:border-l-[var(--paste-line-strong)] max-[900px]:[box-shadow:-18px_0_38px_-30px_rgb(15_23_42/.62)] max-[720px]:w-full max-[720px]:border-l-0 [&>header]:flex [&>header]:min-h-[60px] [&>header]:items-start [&>header]:justify-between [&>header]:gap-2.5 [&>header]:border-b [&>header]:border-[var(--paste-line)] [&>header]:bg-[var(--paste-chrome)] [&>header]:p-3 [&_small]:text-[10px] [&_small]:text-[var(--paste-ink-dim)] [&_h2]:mt-[3px] [&_h2]:text-sm [&_h2]:leading-[1.35] [&_h2]:[overflow-wrap:anywhere] [&_dl]:m-0 [&_dl]:border-b [&_dl]:border-[var(--paste-line)] [&_dl>div]:grid [&_dl>div]:grid-cols-[76px_minmax(0,1fr)] [&_dl>div]:gap-2.5 [&_dl>div]:border-b [&_dl>div]:border-[var(--paste-line)] [&_dl>div]:px-3 [&_dl>div]:py-2.5 [&_dl>div]:text-[11px] [&_dt]:text-[var(--paste-ink-dim)] [&_dd]:m-0 [&_dd]:min-w-0 [&_dd]:text-[var(--paste-ink)] [&_dd]:[overflow-wrap:anywhere]" aria-labelledby="user-inspector-title">
          <header><div><small>使用主体策略</small><h2 id="user-inspector-title">{{ inspectedUser.userKey }}</h2></div><UButton type="button" icon="i-tabler-x" color="neutral" variant="ghost" size="sm" square aria-label="关闭用户检查器" @click="closeUserInspector" /></header>
          <dl>
            <div><dt>今日额度</dt><dd>{{ inspectedUser.usedToday }} / {{ inspectedUser.effectiveDailyLimit }}</dd></div>
            <div><dt>片段记录</dt><dd>{{ inspectedUser.activePastes }} 个有效 · {{ inspectedUser.totalPastes }} 个累计</dd></div>
            <div><dt>最近创建</dt><dd>{{ inspectedUser.lastCreatedAt ? new Date(inspectedUser.lastCreatedAt).toLocaleString("zh-CN") : "尚无创建记录" }}</dd></div>
          </dl>
          <form class="grid gap-4 px-3 pt-3.5 pb-[18px] [&>footer]:flex [&>footer]:flex-wrap [&>footer]:justify-end [&>footer]:gap-1.5" @submit.prevent="saveUserPolicy">
            <UFormField label="创建权限"><USelect v-model="userDraftState" :items="userPolicyStateItems" value-key="value" label-key="label" class="w-full" /></UFormField>
            <UCheckbox v-model="userDraftCustomLimit" label="为此用户设置单独上限" />
            <UFormField v-if="userDraftCustomLimit" label="每日创建上限" hint="1–10,000；按 UTC 自然日重置"><UInputNumber v-model="userDraftLimit" :min="1" :max="10000" class="w-full" @input="userDraftLimit = numberInputValue($event)" /></UFormField>
            <UFormField label="治理说明" hint="管理员可见，最多 240 个字符"><UTextarea v-model="userDraftReason" maxlength="240" :rows="3" autoresize class="w-full" /></UFormField>
            <UAlert v-if="userDraftState === 'suspended'" color="warning" variant="subtle" icon="i-tabler-alert-triangle" title="暂停后不能创建新片段" description="既有分享链接不会自动删除；如需下架内容，请回到片段治理执行删除。" />
            <footer><UButton to="/admin" label="查看该用户片段" icon="i-tabler-files" color="neutral" variant="ghost" @click="viewUserPastes(inspectedUser.userKey)" /><UButton type="submit" label="保存策略" icon="i-tabler-device-floppy" color="primary" :loading="userPolicySaving" :disabled="!userPolicyDirty || userPolicySaving" /></footer>
          </form>
        </aside>
      </div>
    </section>

    <div v-else class="yueli-card min-w-0 overflow-hidden bg-[var(--paste-editor)] text-[var(--paste-ink)]">
      <form id="paste-settings-workspace" class="paste-admin-settings-catalog w-full min-w-0" @submit.prevent="saveSettingsWorkspace">
        <section class="grid grid-cols-[220px_minmax(0,1fr)] border-b border-[var(--paste-line-strong)] max-[720px]:grid-cols-[minmax(0,1fr)]" aria-labelledby="public-settings-title">
          <header class="flex items-start gap-2.5 border-r border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-[22px] py-7 max-[720px]:border-r-0 max-[720px]:border-b max-[720px]:px-3.5 max-[720px]:py-[18px] [&_h2]:text-[15px] [&_h2]:tracking-[-0.02em] [&_p]:mt-[7px] [&_p]:mb-2.5 [&_p]:text-[11px] [&_p]:leading-[1.6] [&_p]:text-[var(--paste-ink-soft)] [&_code]:text-[10px] [&_code]:text-[var(--paste-ink-dim)]"><span class="grid size-8 shrink-0 place-items-center rounded-[7px] border border-[color-mix(in_srgb,var(--paste-blue)_22%,var(--paste-line))] bg-[var(--paste-blue-soft)] text-[var(--paste-blue)]"><UIcon name="i-tabler-world" class="size-5" /></span><div><h2 id="public-settings-title">公开展示</h2><p>控制访客在编辑器、浏览器标题和分享页面看到的名称与说明。</p><code>r{{ settings.revision }}</code></div></header>
          <div class="min-w-0 px-[26px] pt-3 pb-6 max-[720px]:px-3.5 max-[720px]:pt-0 max-[720px]:pb-5 [&>div]:grid [&>div]:grid-cols-[minmax(230px,1fr)_minmax(220px,360px)] [&>div]:items-center [&>div]:gap-7 [&>div]:border-b [&>div]:border-[var(--paste-line)] [&>div]:py-[18px] max-[720px]:[&>div]:grid-cols-[minmax(0,1fr)] max-[720px]:[&>div]:gap-2.5 max-[720px]:[&>div]:py-[17px] [&_label]:text-xs [&_label]:font-[680] [&_p]:mt-[5px] [&_p]:max-w-[58ch] [&_p]:text-[11px] [&_p]:leading-[1.55] [&_p]:text-[var(--paste-ink-soft)]">
            <div><div><label for="paste-site-name">站点名称</label><p>最多 40 个字符；不会改变服务地址、API 或已有链接。</p></div><UInput id="paste-site-name" v-model="settingsName" name="siteName" maxlength="40" autocomplete="off" aria-label="站点名称" class="w-full" /></div>
            <div><div><label for="paste-site-description">站点说明</label><p>最多 160 个字符，用于搜索摘要和分享页面说明。</p></div><UTextarea id="paste-site-description" v-model="settingsDescription" name="siteDescription" maxlength="160" :rows="3" autoresize aria-label="站点说明" class="w-full" /></div>
            <UAlert v-if="settingsError" class="mt-4" color="error" variant="subtle" icon="i-tabler-alert-circle" title="公开展示未保存" :description="settingsError" role="alert" />
          </div>
        </section>
        <section class="grid grid-cols-[220px_minmax(0,1fr)] border-b border-[var(--paste-line-strong)] max-[720px]:grid-cols-[minmax(0,1fr)]" aria-labelledby="creation-settings-title">
          <header class="flex items-start gap-2.5 border-r border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-[22px] py-7 max-[720px]:border-r-0 max-[720px]:border-b max-[720px]:px-3.5 max-[720px]:py-[18px] [&_h2]:text-[15px] [&_h2]:tracking-[-0.02em] [&_p]:mt-[7px] [&_p]:mb-2.5 [&_p]:text-[11px] [&_p]:leading-[1.6] [&_p]:text-[var(--paste-ink-soft)] [&_code]:text-[10px] [&_code]:text-[var(--paste-ink-dim)]"><span class="grid size-8 shrink-0 place-items-center rounded-[7px] border border-[color-mix(in_srgb,var(--paste-blue)_22%,var(--paste-line))] bg-[var(--paste-blue-soft)] text-[var(--paste-blue)]"><UIcon name="i-tabler-shield-bolt" class="size-5" /></span><div><h2 id="creation-settings-title">创建策略</h2><p>用清晰的每日边界保护服务；只有成功创建才会占用额度。</p><code>r{{ governanceSettings?.revision || "—" }}</code></div></header>
          <div class="min-w-0 px-[26px] pt-3 pb-6 max-[720px]:px-3.5 max-[720px]:pt-0 max-[720px]:pb-5 [&>div]:grid [&>div]:grid-cols-[minmax(230px,1fr)_minmax(220px,360px)] [&>div]:items-center [&>div]:gap-7 [&>div]:border-b [&>div]:border-[var(--paste-line)] [&>div]:py-[18px] max-[720px]:[&>div]:grid-cols-[minmax(0,1fr)] max-[720px]:[&>div]:gap-2.5 max-[720px]:[&>div]:py-[17px] [&_label]:text-xs [&_label]:font-[680] [&_p]:mt-[5px] [&_p]:max-w-[58ch] [&_p]:text-[11px] [&_p]:leading-[1.55] [&_p]:text-[var(--paste-ink-soft)]">
            <div v-if="governanceSettingsLoading" class="grid gap-3 py-[18px]" role="status" aria-label="正在读取创建限制"><USkeleton class="h-14 w-full" /><USkeleton class="h-14 w-full" /></div>
            <template v-else>
              <div><div><label for="paste-user-limit">登录用户默认额度</label><p>每位登录用户在一个 UTC 自然日内可成功创建 1–10,000 个片段；可在用户治理中单独覆盖。</p></div><UInputNumber id="paste-user-limit" v-model="governanceUserLimit" :min="1" :max="10000" aria-label="登录用户每日默认上限" class="w-full" @input="governanceUserLimit = numberInputValue($event)" /></div>
              <div><div><label for="paste-anonymous-limit">匿名全站额度</label><p>所有匿名访客共享 0–100,000 的每日总额；设为 0 会暂停匿名创建。</p></div><UInputNumber id="paste-anonymous-limit" v-model="governanceAnonymousLimit" :min="0" :max="100000" aria-label="匿名创建每日全站总额" class="w-full" @input="governanceAnonymousLimit = numberInputValue($event)" /></div>
              <p class="mt-4 flex items-start gap-[7px] text-[11px] leading-[1.55] text-[var(--paste-ink-soft)]"><UIcon name="i-tabler-clock" class="size-4" />每天按 UTC 自然日重置。匿名额度是全站保护阀，不会把 IP 地址当成用户身份。</p>
            </template>
            <UAlert v-if="governanceSettingsError" class="mt-4" color="error" variant="subtle" icon="i-tabler-alert-circle" title="创建策略未保存" :description="governanceSettingsError" role="alert" />
          </div>
        </section>
      </form>
    </div>

    <footer v-if="section !== 'settings'" class="paste-admin-statusbar flex min-w-0 items-center justify-between gap-3 rounded-xl border-t border-[var(--paste-status-border)] bg-[var(--paste-status)] px-2 font-mono text-xs text-[var(--paste-status-text-muted)] max-[720px]:min-h-11 max-[720px]:justify-end max-[720px]:overflow-hidden max-[720px]:px-1 max-[720px]:text-[11px]">
      <span class="flex min-w-0 items-center gap-[7px] max-[720px]:hidden"><span class="size-[7px] shrink-0 rounded-full bg-[var(--paste-green)]" />{{ section === "users" ? "用户治理" : "片段治理" }}</span>
      <span v-if="section === 'pastes'" class="paste-admin-page-status">
        <span class="max-w-[190px] truncate max-[720px]:hidden">{{ selectedValues.length ? `已选择 ${selectedValues.length} 项` : operationMessage || `${pageStart}-${pageEnd} / ${total}` }}</span>
        <UPagination
          v-if="total > limit"
          v-model:page="currentPage"
          class="shrink-0 [&_[data-slot=list]]:gap-0.5 [&_button]:min-h-6 [&_button]:min-w-6 [&_button]:rounded-[3px] [&_button]:bg-transparent [&_button]:px-1.5 [&_button]:text-[var(--paste-status-text-muted)] [&_button]:shadow-none [&_button:hover:not(:disabled)]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button:hover:not(:disabled)]:text-[var(--paste-status-text)] [&_button:focus-visible]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button:focus-visible]:text-[var(--paste-status-text)] [&_button[data-selected=true]]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button[data-selected=true]]:font-[750] [&_button[data-selected=true]]:text-[var(--paste-status-text)] [&_button[data-selected=true]]:[box-shadow:inset_0_-2px_var(--paste-status-text)] [&_button:disabled]:bg-transparent [&_button:disabled]:text-[var(--paste-status-text-muted)] [&_button:disabled]:opacity-40 [&_button:disabled]:shadow-none max-[720px]:[&_button]:min-h-11 max-[720px]:[&_button]:min-w-11 max-[720px]:[&_button]:px-[7px]"
          :total="total"
          :items-per-page="limit"
          :sibling-count="1"
          :show-edges="false"
          :show-controls="false"
          :disabled="loading"
          color="neutral"
          variant="ghost"
          active-color="neutral"
          active-variant="ghost"
          size="xs"
        >
          <template #prev><UButton type="button" icon="i-tabler-chevron-left" color="neutral" variant="ghost" size="xs" square aria-label="上一页" /></template>
          <template #item="{ item, page }"><UButton v-if="item.type === 'page'" type="button" :label="String(item.value)" color="neutral" variant="ghost" size="xs" :aria-label="`第 ${item.value} 页`" :data-selected="page === item.value ? 'true' : undefined" /></template>
          <template #next><UButton type="button" icon="i-tabler-chevron-right" color="neutral" variant="ghost" size="xs" square aria-label="下一页" /></template>
        </UPagination>
      </span>
      <span v-else-if="section === 'users'" class="paste-admin-page-status">
        <span class="max-w-[190px] truncate max-[720px]:hidden">{{ selectedUsers.length ? `已选择 ${selectedUsers.length} 个用户` : operationMessage || `${userPageStart}-${userPageEnd} / ${userTotal}` }}</span>
        <UPagination
          v-if="userTotal > limit"
          v-model:page="currentUserPage"
          class="shrink-0 [&_[data-slot=list]]:gap-0.5 [&_button]:min-h-6 [&_button]:min-w-6 [&_button]:rounded-[3px] [&_button]:bg-transparent [&_button]:px-1.5 [&_button]:text-[var(--paste-status-text-muted)] [&_button]:shadow-none [&_button:hover:not(:disabled)]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button:hover:not(:disabled)]:text-[var(--paste-status-text)] [&_button:focus-visible]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button:focus-visible]:text-[var(--paste-status-text)] [&_button[data-selected=true]]:bg-[color-mix(in_srgb,var(--paste-status-text)_14%,transparent)] [&_button[data-selected=true]]:font-[750] [&_button[data-selected=true]]:text-[var(--paste-status-text)] [&_button[data-selected=true]]:[box-shadow:inset_0_-2px_var(--paste-status-text)] [&_button:disabled]:bg-transparent [&_button:disabled]:text-[var(--paste-status-text-muted)] [&_button:disabled]:opacity-40 [&_button:disabled]:shadow-none max-[720px]:[&_button]:min-h-11 max-[720px]:[&_button]:min-w-11 max-[720px]:[&_button]:px-[7px]"
          :total="userTotal"
          :items-per-page="limit"
          :sibling-count="1"
          :show-edges="false"
          :show-controls="false"
          :disabled="userLoading"
          color="neutral"
          variant="ghost"
          active-color="neutral"
          active-variant="ghost"
          size="xs"
        >
          <template #prev><UButton type="button" icon="i-tabler-chevron-left" color="neutral" variant="ghost" size="xs" square aria-label="上一页用户" /></template>
          <template #item="{ item, page }"><UButton v-if="item.type === 'page'" type="button" :label="String(item.value)" color="neutral" variant="ghost" size="xs" :aria-label="`第 ${item.value} 页用户`" :data-selected="page === item.value ? 'true' : undefined" /></template>
          <template #next><UButton type="button" icon="i-tabler-chevron-right" color="neutral" variant="ghost" size="xs" square aria-label="下一页用户" /></template>
        </UPagination>
      </span>
      <span v-else aria-live="polite">{{ settingsDirty || governanceSettingsDirty ? "有未保存修改" : operationMessage || `当前名称：${siteName}` }}</span>
    </footer>

    <USlideover v-if="batchOpen" v-model:open="batchOpen" title="批量治理" :description="`修改选中的 ${selectedValues.length} 个代码片段。`" :dismissible="!batchSaving" :close="!batchSaving" :ui="{ content: 'sm:max-w-sm' }">
      <template #body><div class="grid gap-[18px] [&_p]:text-xs [&_p]:leading-[1.65] [&_p]:text-[var(--paste-ink-soft)]"><UFormField label="可见性"><USelect v-model="batchVisibility" :items="batchVisibilityItems" value-key="value" label-key="label" class="w-full" /></UFormField><UFormField label="有效期"><USelect v-model="batchExpiry" :items="batchExpiryItems" value-key="value" label-key="label" class="w-full" /></UFormField><p>匿名片段没有所有者，因此不能改为“仅自己”；其他项目仍会继续执行并保留失败项选择。</p></div></template>
      <template #footer><div class="flex w-full justify-end gap-[7px]"><UButton type="button" label="取消" color="neutral" variant="ghost" :disabled="batchSaving" @click="closeBatch" /><UButton type="button" label="应用修改" icon="i-tabler-check" color="primary" :loading="batchSaving" :disabled="batchSaving || (batchVisibility === 'keep' && batchExpiry === 'keep')" @click="applyBatch" /></div></template>
    </USlideover>
    <USlideover v-if="userBatchOpen" v-model:open="userBatchOpen" title="批量设置创建策略" :description="`把同一组策略应用到选中的 ${selectedUsers.length} 个用户。`" :dismissible="!userBatchSaving" :close="!userBatchSaving" :ui="{ content: 'sm:max-w-sm' }">
      <template #body><div class="grid gap-[18px] [&_p]:text-xs [&_p]:leading-[1.65] [&_p]:text-[var(--paste-ink-soft)]"><UFormField label="创建权限"><USelect v-model="userBatchState" :items="userBatchStateItems" value-key="value" label-key="label" class="w-full" /></UFormField><UFormField label="每日额度"><USelect v-model="userBatchLimitMode" :items="userBatchLimitItems" value-key="value" label-key="label" class="w-full" /></UFormField><UFormField v-if="userBatchLimitMode === 'custom'" label="统一每日上限" hint="1–10,000；按 UTC 自然日重置"><UInputNumber v-model="userBatchLimit" :min="1" :max="10000" class="w-full" @input="userBatchLimit = numberInputValue($event)" /></UFormField><UAlert v-if="userBatchState === 'suspended'" color="warning" variant="subtle" icon="i-tabler-alert-triangle" title="只暂停创建" description="不会停用 Identity 账号，也不会删除或禁用已有分享。" /><p>每个用户原有的治理说明都会保留；发生冲突或失败的用户会继续保持选择，方便重试。</p></div></template>
      <template #footer><div class="flex w-full justify-end gap-[7px]"><UButton type="button" label="取消" color="neutral" variant="ghost" :disabled="userBatchSaving" @click="closeUserBatch" /><UButton type="button" :label="`应用到 ${selectedUsers.length} 个用户`" icon="i-tabler-check" color="primary" :loading="userBatchSaving" :disabled="userBatchSaving || (userBatchState === 'keep' && userBatchLimitMode === 'keep') || (userBatchLimitMode === 'custom' && (userBatchLimit === null || userBatchLimit < 1 || userBatchLimit > 10000))" @click="applyUserBatch" /></div></template>
    </USlideover>
    <p class="sr-only" aria-live="polite">{{ operationMessage }}</p>
  </div>
</template>
