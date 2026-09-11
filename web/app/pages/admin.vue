<!--
THESIS: Site governance is a focused inspection desk, not a dashboard of summary cards.
OWN-WORLD: Cool porcelain and ink-navy editor planes use mineral-blue focus, flat rules, and compact Nuxt UI controls.
STORY: find records or users, select one or many, apply bounded governance, then tune public identity and abuse controls in one save flow.
FIRST VIEWPORT: one 40px application bar, one section rail, a selection-aware toolbar or settings command bar, the work plane, and a 28px status bar.
FORM: the fourth grounded master-detail explorer extended with contextual user batching and a flat VS Code-like settings catalog; seed eb959727.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import { useActionFeedback } from "@yueli/ui/feedback";
import { SettingSection } from "@yueli/ui/settings/pattern";
import { TabbedSurface } from "@yueli/ui/admin";
import { AuthorizationUser } from "@yueli/ui/admin";
import { mergePublicUser, type SessionDisplayUser } from "@yueli/identity-nuxt/server/utils/profile";
import { CollectionPaginationBar, CollectionHeaderTools } from "@yueli/ui/collection/pattern";
import { PageHeader } from "@yueli/ui/admin";
import type {
  AdministrationPaste,
  AdministrationUser,
  AdministrationUserState,
  GovernanceSettings,
  PastePatchInput,
  PasteVisibility,
} from "../types/paste";
import type { FailureFeedback } from "@yueli/http-runtime";
import { pasteFailureFeedback } from "../utils/pasteFailure";
import { displayTitle, pasteErrorMessage } from "../utils/paste";

definePageMeta({ layout: "admin", middleware: ["auth", "operator"] });

const settingsTab = ref("site");
const settingsTabs = [{ value: "site", label: "站点", icon: "i-tabler-world" }];
const api = usePasteApi();
const route = useRoute();
const toast = useToast();
const { settings, siteName, adopt } = useSiteSettings();

const settingsFailure = ref<FailureFeedback | null>(null);
const governanceFailure = ref<FailureFeedback | null>(null);
const section = computed<"pastes" | "users" | "settings">(() => {
  if (route.query.view === "users") return "users";
  if (route.query.view === "settings") return "settings";
  return "pastes";
});
const pageTitle = computed(() => ({
  pastes: "片段管理",
  users: "用户管理",
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
const limit = ref(20);
async function changePageSize(size: number) { limit.value = size; offset.value = 0; userOffset.value = 0; if (section.value === "users") await loadUsers(true); else await loadPastes(true); }
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
const saveFeedback = useActionFeedback({ resetMs: 1000 });
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
  get: () => Math.floor(offset.value / limit.value) + 1,
  set: (value: number) => {
    const pageCount = Math.max(1, Math.ceil(total.value / limit.value));
    const nextPage = Math.min(Math.max(1, value), pageCount);
    const nextOffset = (nextPage - 1) * limit.value;
    if (nextOffset === offset.value) return;
    offset.value = nextOffset;
    void loadPastes();
  },
});
const settingsDirty = computed(() =>
  settingsName.value.trim() !== settings.value.name ||
  settingsDescription.value.trim() !== settings.value.description,
);
const directory = usePublicUserDirectory(computed(() => section.value === "users" ? users.value.map(user => user.userKey) : []));
const accountOrigin = String(useRuntimeConfig().public.accountUrl).replace(/\/$/, "");
function userInfo(subject: string) {
 const user = directory.users.value[subject];
 return { subject, name:user?.displayName, handle:user?.handle, avatarUrl:user ? mergePublicUser<SessionDisplayUser>({sub:subject},user,accountOrigin).avatar : undefined, profileUrl:`${accountOrigin}/u/${encodeURIComponent(subject)}`, loading:directory.pending.value };
}
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
  get: () => Math.floor(userOffset.value / limit.value) + 1,
  set: (value: number) => {
    const pageCount = Math.max(1, Math.ceil(userTotal.value / limit.value));
    const nextPage = Math.min(Math.max(1, value), pageCount);
    const nextOffset = (nextPage - 1) * limit.value;
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
      size: limit.value,
      page: Math.floor(offset.value / limit.value) + 1,
    });
    if (sequence !== loadSequence) return;
    values.value = page.items;
    total.value = page.total;
    const available = new Set(values.value.map((value) => value.code));
    selectedCodes.value = selectedCodes.value.filter((code) => available.has(code));
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
      size: limit.value,
      page: Math.floor(userOffset.value / limit.value) + 1,
    });
    if (sequence !== loadUsersSequence) return;
    users.value = page.items || [];
    userTotal.value = page.total;
    const available = new Set(users.value.map((value) => value.userKey));
    selectedUserKeys.value = selectedUserKeys.value.filter((userKey) => available.has(userKey));
  } catch (caught) {
    if (sequence === loadUsersSequence) userError.value = pasteErrorMessage(caught);
  } finally {
    if (sequence === loadUsersSequence) userLoading.value = false;
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
    announce(`已更新 ${result.succeeded.length} 个用户的创建状态。`, "success");
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
  batchDeleting.value = false;
  announce(
    result.failed.length
      ? `已删除 ${result.succeeded.length} 项，${result.failed.length} 项失败，失败项仍保持选中。`
      : `已删除 ${result.succeeded.length} 个代码片段。`,
    result.failed.length ? "warning" : "success",
  );
}

async function saveSettings(): Promise<boolean> {
  if (settingsSaving.value) return false;
  if (!settingsDirty.value) { saveFeedback.success(); return true; }
  settingsSaving.value = true;
  saveFeedback.pending();
  settingsFailure.value = null;
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
    saveFeedback.success();
    return true;
  } catch (caught) {
    saveFeedback.reset();
    settingsFailure.value = pasteFailureFeedback(caught,"公开展示未保存，请检查后重试。",{"/name":"name","/description":"description"});
    return false;
  } finally {
    settingsSaving.value = false;
  }
}

function resetSettingsDraft() {
  settingsFailure.value = null;
  settingsName.value = settings.value.name;
  settingsDescription.value = settings.value.description;
  settingsError.value = "";
}


function numberInputValue(event: Event): number | null {
  const raw = (event.target as HTMLInputElement).value.trim().replaceAll(",", "").replaceAll(" ", "");
  if (!raw) return null;
  const value = Number(raw);
  return Number.isFinite(value) ? value : null;
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
      <template v-if="section !== 'settings'" #tools>
        <CollectionHeaderTools v-if="section === 'pastes'" v-model:search="query" search-id="admin-search" label="片段搜索与筛选" search-placeholder="搜索标题、短码、用户或标签" :filter-count="activeFilterCount"
          :controls="[{kind:'select',id:'visibility',label:'可见性',value:visibility,options:visibilityItems},{kind:'select',id:'state',label:'状态',value:state,options:stateItems},{kind:'select',id:'ownership',label:'归属',value:ownership,options:ownershipItems}]"
          @filters="values => { visibility = String(values.visibility) as typeof visibility; state = String(values.state) as typeof state; ownership = String(values.ownership) as typeof ownership }" />
        <CollectionHeaderTools v-else v-model:search="userQuery" search-id="admin-user-search" label="用户搜索与筛选" search-placeholder="搜索用户主体标识" :controls="[{kind:'select',id:'state',label:'创建状态',value:userState,options:userStateItems}]" @filters="values => { userState = String(values.state) as typeof userState }" />
      </template>
      <template v-if="section === 'settings'" #actions>
        <UButton type="submit" form="paste-settings-workspace"
          :label="saveFeedback.status.value === 'success' ? '已保存' : '保存'"
          :icon="saveFeedback.status.value === 'success' ? 'i-tabler-check' : 'i-tabler-device-floppy'"
          color="primary" variant="solid" class="min-w-20 justify-center"
          :loading="settingsSaving" :disabled="settingsSaving" />
      </template>
    </PageHeader>

    <UAlert v-if="section === 'users' && directory.error.value" color="error" title="用户资料加载失败" class="mb-3"><template #actions><UButton label="重试" color="neutral" variant="outline" @click="directory.refresh()" /></template></UAlert>
    <section v-if="section === 'pastes'" class="yueli-card flex flex-col overflow-hidden bg-[var(--paste-editor)] text-[var(--paste-ink)]" aria-label="片段管理">
      <div class="flex min-w-0 items-center gap-[5px] border-b border-[var(--paste-line)] bg-[var(--paste-editor)] py-[3px] pr-1.5 pl-2.5 max-[720px]:gap-[3px] max-[720px]:p-[4px_5px]" role="toolbar" aria-label="全站代码片段工具栏">
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

          <UTooltip text="刷新列表"><UButton type="button" icon="i-tabler-refresh" color="neutral" variant="ghost" size="sm" square aria-label="刷新列表" :loading="loading" @click="loadPastes()" /></UTooltip>
        </template>
      </div>

      <div class="grid min-h-0 min-w-0 grid-cols-[minmax(0,1fr)] overflow-hidden data-[inspecting=true]:grid-cols-[minmax(0,1fr)_minmax(290px,340px)] max-[900px]:data-[inspecting=true]:relative max-[900px]:data-[inspecting=true]:grid-cols-[minmax(0,1fr)]" >
        <div class="min-h-0 min-w-0 overflow-auto bg-[var(--paste-editor)]" :aria-busy="loading">
          <div v-if="loading" class="paste-admin-loading min-w-[760px] max-[720px]:min-w-0" role="status" aria-label="正在读取全站代码片段">
            <div v-for="index in 6" :key="index" class="grid min-h-[62px] grid-cols-[32px_minmax(200px,1fr)_minmax(100px,160px)_112px_72px_100px_48px] items-center border-b border-[var(--paste-line)] px-3 py-2 max-[720px]:grid-cols-[34px_minmax(0,1fr)] max-[720px]:[&>:nth-child(n+3)]:hidden"><USkeleton class="size-3.5" /><div><USkeleton class="h-3.5 w-52 max-w-full" /><USkeleton class="mt-2 h-2.5 w-72 max-w-full" /></div><USkeleton class="h-3 w-28" /><USkeleton class="h-3 w-16" /><USkeleton class="h-3 w-16" /></div>
          </div>
          <div v-else-if="error" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]" role="alert"><UIcon name="i-tabler-alert-circle" class="size-7 text-[var(--paste-red)]" /><strong>无法读取治理列表</strong><p>{{ error }}</p><UButton type="button" label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="loadPastes()" /></div>
          <div v-else-if="values.length === 0" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[52ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]"><UIcon name="i-tabler-file-search" class="size-7" /><strong>没有符合条件的代码片段</strong><p>调整搜索或筛选条件后再试。</p><UButton type="button" label="清除条件" icon="i-tabler-filter-off" color="neutral" variant="outline" @click="clearFilters" /></div>
          <div v-else class="min-w-[760px] max-[720px]:min-w-0">
            <div class="sticky top-0 z-[2] grid min-h-[30px] grid-cols-[32px_minmax(200px,1fr)_minmax(100px,160px)_112px_72px_100px_48px] items-center border-b border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-3 text-[11px] font-[630] text-[var(--paste-ink-dim)] max-[720px]:hidden" aria-hidden="true"><span /><span>内容</span><span>归属</span><span>访问</span><span>状态</span><span>更新</span><span class="text-right">操作</span></div>
            <div role="list" aria-label="全站代码片段">
              <div v-for="value in values" :key="value.code" class="paste-admin-row group grid min-h-[62px] min-w-0 grid-cols-[32px_minmax(200px,1fr)_minmax(100px,160px)_112px_72px_100px_48px] items-center border-b border-[var(--paste-line)] px-3 py-2 transition-colors duration-100 hover:bg-[var(--paste-editor-active)] focus-within:bg-[var(--paste-editor-active)] data-[selected=true]:bg-[color-mix(in_srgb,var(--paste-blue)_9%,var(--paste-editor))] data-[inspected=true]:[box-shadow:inset_2px_0_var(--paste-blue)] max-[720px]:min-h-28 max-[720px]:grid-cols-[44px_minmax(0,1fr)_auto] max-[720px]:gap-x-2 max-[720px]:gap-y-[7px] max-[720px]:px-2 max-[720px]:py-2.5 max-[720px]:[&>:first-child]:col-start-1 max-[720px]:[&>:first-child]:row-start-1 max-[720px]:[&>:first-child]:row-end-4 max-[720px]:[&>:first-child]:mt-[-7px] max-[720px]:[&>:first-child]:ml-[-8px] max-[720px]:[&>:first-child]:size-11 max-[720px]:[&>:first-child]:self-start max-[720px]:[&>:first-child]:justify-center" role="listitem" :data-selected="selectedCodeSet.has(value.code)" >
                <UCheckbox :model-value="selectedCodeSet.has(value.code)" size="sm" :aria-label="`选择 ${displayTitle(value.title)}`" :disabled="batchSaving || batchDeleting || value.state === 'deleted'" @update:model-value="toggleSelected(value.code, $event === true)" />
                <div class="grid min-w-0 gap-[5px] border-0 bg-transparent py-0 pr-3.5 pl-0 text-left text-inherit max-[720px]:col-start-2 max-[720px]:col-end-[-1] max-[720px]:pr-0 [&>span]:flex [&>span]:min-w-0 [&>span]:gap-2.5 [&>span]:overflow-hidden [&>span]:whitespace-nowrap [&>span]:text-[11px] [&>span]:text-[var(--paste-ink-soft)] [&_code]:font-[680] [&_code]:text-[var(--paste-blue)] [&_strong]:truncate [&_strong]:text-[13px]" >
                  <strong>{{ displayTitle(value.title) }}</strong>
                  <span><code>{{ value.code }}</code><span>{{ value.fileCount }} 个文件</span><span>{{ value.primaryLanguage }}</span><span v-for="tag in value.tags.slice(0, 2)" :key="tag">#{{ tag }}</span></span>
                </div>
                <div class="flex min-w-0 items-center gap-[5px] pr-2.5 text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-2 max-[720px]:row-start-2 [&_span]:truncate"><UIcon :name="value.ownerUserKey ? 'i-tabler-user' : 'i-tabler-user-off'" class="size-4" /><span>{{ value.ownerUserKey || "匿名" }}</span></div>
                <span class="paste-admin-access flex min-w-0 items-center gap-[5px] pr-2.5 text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-2 max-[720px]:row-start-3"><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
                <span class="w-fit rounded-[3px] bg-[var(--paste-green-soft)] px-[5px] py-0.5 text-[10px] font-bold text-[var(--paste-green-ink)] data-[state=deleted]:bg-[var(--paste-red-soft)] data-[state=deleted]:text-[var(--paste-red)] data-[state=suspended]:bg-[var(--paste-red-soft)] data-[state=suspended]:text-[var(--paste-red)] max-[720px]:col-start-3 max-[720px]:row-start-2" :data-state="value.state">{{ value.state === "deleted" ? "已删除" : value.expiresAt && new Date(value.expiresAt) <= new Date() ? "已过期" : "有效" }}</span>
                <time class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[720px]:col-start-3 max-[720px]:row-start-3 max-[720px]:justify-self-end" :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString("zh-CN") }}</time>
                <div class="flex justify-end max-[720px]:col-start-3 max-[720px]:row-start-4"><UTooltip text="打开分享链接"><UButton :to="value.shareUrl" target="_blank" rel="noopener noreferrer" icon="i-tabler-external-link" aria-label="打开分享链接" color="neutral" variant="ghost" size="sm" square :disabled="value.state === 'deleted'" /></UTooltip></div>
              </div>
            </div>
          </div>
        </div>

      </div>
      <footer class="border-t border-default p-3"><CollectionPaginationBar :page="currentPage" :page-size="limit" :page-sizes="[20, 50, 100]" :total="total" @page-change="currentPage = $event" @page-size-change="changePageSize" /><p v-if="operationMessage" role="status" class="mt-2 text-xs text-muted">{{ operationMessage }}</p></footer>
    </section>

    <section v-else-if="section === 'users'" class="yueli-card overflow-hidden" aria-label="用户管理">
      <div class="flex min-h-11 items-center gap-2 border-b border-default px-3" role="toolbar" aria-label="用户管理工具栏">
        <UCheckbox class="paste-admin-select-all shrink-0 p-2" :model-value="userSelectionState" :disabled="userLoading || !users.length || userBatchSaving" aria-label="选择当前页用户" @update:model-value="toggleUserPageSelection($event)" />
        <template v-if="selectedUsers.length">
          <span class="text-sm">{{ selectedUsers.length }} 个用户已选择</span>
          <UButton class="ml-auto" label="批量设置" icon="i-tabler-users" color="neutral" variant="ghost" :disabled="userBatchSaving" @click="openUserBatch" />
          <UButton label="清除" color="neutral" variant="ghost" @click="clearUserSelection" />
        </template>
        <UTooltip v-else text="刷新用户"><UButton icon="i-tabler-refresh" aria-label="刷新用户列表" color="neutral" variant="ghost" size="sm" square :loading="userLoading" @click="loadUsers()" /></UTooltip>
      </div>
      <div v-if="userLoading" class="space-y-3 p-5" role="status" aria-label="正在读取用户"><USkeleton v-for="n in 3" :key="n" class="h-12" /></div>
      <div v-else-if="userError" class="p-5"><UAlert color="error" variant="subtle" title="无法读取用户列表" :description="userError"><template #actions><UButton label="重新读取" @click="loadUsers()" /></template></UAlert></div>
      <div v-else-if="!users.length" class="p-8 text-center text-sm text-muted">没有符合条件的用户</div>
      <div v-else>
        <div class="hidden grid-cols-[32px_minmax(200px,1fr)_80px_100px_90px_112px_110px] items-center gap-2 border-b border-default px-3 py-2 text-xs text-muted min-[900px]:grid" aria-hidden="true"><span /><span>用户</span><span>状态</span><span>今日创建</span><span>片段</span><span>最近创建</span><span class="text-right">操作</span></div>
        <div role="list" aria-label="Paste 使用主体">
          <div v-for="user in users" :key="user.userKey" class="paste-admin-user-row grid grid-cols-[24px_minmax(0,1fr)_auto] items-center gap-3 border-b border-default px-3 py-4 last:border-b-0 min-[900px]:grid-cols-[32px_minmax(200px,1fr)_80px_100px_90px_112px_110px] min-[900px]:gap-2" role="listitem" :data-selected="selectedUserKeySet.has(user.userKey)">
            <UCheckbox :model-value="selectedUserKeySet.has(user.userKey)" :aria-label="`选择用户 ${user.userKey}`" :disabled="userBatchSaving" @update:model-value="toggleUserSelected(user.userKey, $event === true)" />
            <AuthorizationUser v-bind="userInfo(user.userKey)" />
            <UBadge :color="user.state === 'suspended' ? 'warning' : 'success'" variant="subtle" size="sm" class="w-fit">{{ user.state === 'suspended' ? '已暂停' : '正常' }}</UBadge>
            <span class="col-start-2 text-xs text-muted min-[900px]:col-start-auto"><span class="min-[900px]:hidden">今日创建：</span>{{ user.usedToday }} / {{ user.effectiveDailyLimit }}</span>
            <span class="text-xs text-muted"><span class="min-[900px]:hidden">片段：</span>{{ user.activePastes }} / {{ user.totalPastes }}</span>
            <time class="col-start-2 text-xs text-muted min-[900px]:col-start-auto">{{ user.lastCreatedAt ? new Date(user.lastCreatedAt).toLocaleDateString('zh-CN') : '—' }}</time>
            <div class="flex justify-end"><UButton :icon="user.state === 'suspended' ? 'i-tabler-player-play' : 'i-tabler-player-pause'" :label="user.state === 'suspended' ? '恢复创建' : '暂停创建'" :color="user.state === 'suspended' ? 'primary' : 'error'" variant="ghost" size="sm" :loading="userPolicySaving" @click="updateUserState(user, user.state === 'suspended' ? 'active' : 'suspended')" /></div>
          </div>
        </div>
      </div>
      <footer class="border-t border-default p-3"><CollectionPaginationBar :page="currentUserPage" :page-size="limit" :page-sizes="[20, 50, 100]" :total="userTotal" @page-change="currentUserPage = $event" @page-size-change="changePageSize" /><p v-if="operationMessage" role="status" class="mt-2 text-xs text-muted">{{ operationMessage }}</p></footer>
    </section>

    <TabbedSurface v-else v-model="settingsTab" :items="settingsTabs" navigation-label="设置分区" data-manage-surface="settings">
      <form id="paste-settings-workspace" class="paste-admin-settings-catalog space-y-5 p-4 sm:p-5" @submit.prevent="saveSettings()">
        <SettingSection title="站点信息">
          <div class="grid gap-4">
            <UFormField label="站点名称" name="name" required :error="settingsFailure?.fieldErrors.name?.join(' ')">
              <UInput id="paste-site-name" v-model="settingsName" name="siteName" maxlength="40" autocomplete="off" aria-label="站点名称" class="w-full" />
            </UFormField>
            <UFormField label="站点描述" name="description" :error="settingsFailure?.fieldErrors.description?.join(' ')">
              <UTextarea id="paste-site-description" v-model="settingsDescription" name="siteDescription" maxlength="160" :rows="3" aria-label="站点说明" class="w-full" />
            </UFormField>
            <FailureNotice :feedback="settingsFailure" />
          </div>
        </SettingSection>
      </form>
    </TabbedSurface>


    <USlideover v-if="batchOpen" v-model:open="batchOpen" title="批量修改片段" :description="`修改选中的 ${selectedValues.length} 个代码片段。`" :dismissible="!batchSaving" :close="!batchSaving" :ui="{ content: 'sm:max-w-sm' }">
      <template #body><div class="grid gap-[18px] [&_p]:text-xs [&_p]:leading-[1.65] [&_p]:text-[var(--paste-ink-soft)]"><UFormField label="可见性"><USelect v-model="batchVisibility" :items="batchVisibilityItems" value-key="value" label-key="label" class="w-full" /></UFormField><UFormField label="有效期"><USelect v-model="batchExpiry" :items="batchExpiryItems" value-key="value" label-key="label" class="w-full" /></UFormField><p>匿名片段没有所有者，因此不能改为“仅自己”；其他项目仍会继续执行并保留失败项选择。</p></div></template>
      <template #footer><div class="flex w-full justify-end gap-[7px]"><UButton type="button" label="取消" color="neutral" variant="ghost" :disabled="batchSaving" @click="closeBatch" /><UButton type="button" label="应用修改" icon="i-tabler-check" color="primary" :loading="batchSaving" :disabled="batchSaving || (batchVisibility === 'keep' && batchExpiry === 'keep')" @click="applyBatch" /></div></template>
    </USlideover>
    <USlideover v-if="userBatchOpen" v-model:open="userBatchOpen" title="批量管理用户" :description="`修改选中的 ${selectedUsers.length} 个用户。`" :dismissible="!userBatchSaving" :close="!userBatchSaving" :ui="{ content: 'sm:max-w-sm' }">
      <template #body><div class="grid gap-[18px] [&_p]:text-xs [&_p]:leading-[1.65] [&_p]:text-[var(--paste-ink-soft)]"><UFormField label="创建状态"><USelect v-model="userBatchState" :items="userBatchStateItems" value-key="value" label-key="label" class="w-full" /></UFormField><UAlert v-if="userBatchState === 'suspended'" color="warning" variant="subtle" icon="i-tabler-alert-triangle" title="只暂停创建" description="不会停用 Identity 账号，也不会删除或禁用已有分享。" /><p>未成功的用户会保留选择，方便重试。</p></div></template>
      <template #footer><div class="flex w-full justify-end gap-[7px]"><UButton type="button" label="取消" color="neutral" variant="ghost" :disabled="userBatchSaving" @click="closeUserBatch" /><UButton type="button" :label="`应用到 ${selectedUsers.length} 个用户`" icon="i-tabler-check" color="primary" :loading="userBatchSaving" :disabled="userBatchSaving || (userBatchState === 'keep' && userBatchLimitMode === 'keep') || (userBatchLimitMode === 'custom' && (userBatchLimit === null || userBatchLimit < 1 || userBatchLimit > 10000))" @click="applyUserBatch" /></div></template>
    </USlideover>
    <p class="sr-only" aria-live="polite">{{ operationMessage }}</p>
  </div>
</template>
