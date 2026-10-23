<script setup lang="ts">
import { CollectionPaginationBar } from "@yueli/ui/collection/pattern";
import type {
  Paste,
  PastePatchInput,
  PasteSummary,
  PasteVisibility,
} from "../types/paste";
import { displayTitle, pasteErrorMessage, pasteProblemCode } from "../utils/paste";

const { loggedIn } = useAuth();

const api = usePasteApi();
const libraryVersion = useState("paste-library-version", () => 0);
const { siteName } = useSiteSettings();
const toast = useToast();
const values = ref<PasteSummary[]>([]);
const route = useRoute();
const query = ref(typeof route.query.q === "string" ? route.query.q : "");
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
let loadSequence = 0;
const loading = ref(true);
const loadingVisible = useMinimumLoading(loading);
const deleting = ref("");
const error = ref("");
const copied = ref("");
const selectedCodes = ref<string[]>([]);
const batchEditOpen = ref(false);
const batchVisibility = ref<"keep" | PasteVisibility>("keep");
const batchExpiry = ref<"keep" | "1h" | "1d" | "7d" | "30d" | "never">("keep");
const batchSaving = ref(false);
const batchDeleting = ref(false);
const batchError = ref("");
const deleteOpen = ref(false);
const deleteTargets = ref<PasteSummary[]>([]);
const deleteError = ref("");
const operationMessage = ref("");

const visibilityItems = computed(() => [
  { label: "保持原设置", value: "keep" },
  { label: "持链访问", value: "unlisted" },
  { label: "仅自己", value: "private", disabled: !loggedIn.value },
]);
const expiryItems: Array<{
  label: string;
  value: "keep" | "1h" | "1d" | "7d" | "30d" | "never";
}> = [
  { label: "保持原设置", value: "keep" },
  { label: "1 小时后", value: "1h" },
  { label: "1 天后", value: "1d" },
  { label: "7 天后", value: "7d" },
  { label: "30 天后", value: "30d" },
  { label: "不过期", value: "never" },
];

const filtered = computed(() => values.value);
const selectedCodeSet = computed(() => new Set(selectedCodes.value));
const selectedValues = computed(() =>
  values.value.filter((value) => selectedCodeSet.value.has(value.code)),
);
const selectionState = computed<boolean | "indeterminate">(() => {
  if (filtered.value.length === 0) return false;
  const count = filtered.value.filter((value) =>
    selectedCodeSet.value.has(value.code),
  ).length;
  if (count === 0) return false;
  return count === filtered.value.length ? true : "indeterminate";
});
const batchBusy = computed(() => batchSaving.value || batchDeleting.value);
const hasBatchChanges = computed(
  () => batchVisibility.value !== "keep" || batchExpiry.value !== "keep",
);

let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(query, () => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    if (page.value !== 1) page.value = 1;
    else void load();
  }, 250);
});
watch(page, () => {
  selectedCodes.value = [];
  void load();
});
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});
useSeoMeta({
  title: computed(() => `内容库 · ${siteName.value}`),
  description: "回看、编辑和删除你创建的代码片段。",
});

async function load() {
  const sequence = ++loadSequence;
  loading.value = true;
  error.value = "";
  try {
    const result = await api.listMine({
      page: page.value,
      size: pageSize.value,
      q: query.value.trim() || undefined,
    });
    if (sequence !== loadSequence) return;
    const lastPage = Math.max(1, Math.ceil(result.total / pageSize.value));
    if (page.value > lastPage) {
      page.value = lastPage;
      return;
    }
    values.value = result.items;
    total.value = result.total;
    const available = new Set(values.value.map((value) => value.code));
    selectedCodes.value = selectedCodes.value.filter((code) =>
      available.has(code),
    );
  } catch (caught) {
    if (sequence !== loadSequence) return;
    values.value = [];
    total.value = 0;
    selectedCodes.value = [];
    if (loggedIn.value || pasteProblemCode(caught) !== "paste.not_authenticated")
      error.value = pasteErrorMessage(caught);
  } finally {
    if (sequence === loadSequence) loading.value = false;
  }
}

watch(loggedIn, () => { page.value = 1; void load(); });

async function copyLink(value: PasteSummary) {
  try {
    await copyPasteText(value.shareUrl);
    copied.value = value.code;
    window.setTimeout(() => {
      if (copied.value === value.code) copied.value = "";
    }, 1800);
  } catch {
    announce("复制失败，请手动复制分享链接。", "error");
  }
}

function requestDelete(targets: PasteSummary[]) {
  if (!targets.length || batchBusy.value || deleting.value) return;
  deleteTargets.value = [...targets];
  deleteError.value = "";
  deleteOpen.value = true;
}

async function remove(value: PasteSummary) {
  deleting.value = value.code;
  deleteError.value = "";
  try {
    await api.remove(value.code, value.revision);
    libraryVersion.value++;
    await load();
    selectedCodes.value = selectedCodes.value.filter(
      (code) => code !== value.code,
    );
    deleteOpen.value = false;
  } catch (caught) {
    deleteError.value = pasteErrorMessage(caught);
  } finally {
    deleting.value = "";
  }
}

function announce(message: string, color: "success" | "warning" | "error") {
  operationMessage.value = message;
  if (color === "success") return;
  toast.add({
    title: color === "warning" ? "部分操作未完成" : "操作失败",
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

function toggleFilteredSelection(checked: boolean | "indeterminate") {
  const next = new Set(selectedCodes.value);
  for (const value of filtered.value) {
    if (checked === true) next.add(value.code);
    else next.delete(value.code);
  }
  selectedCodes.value = [...next];
}

function clearSelection() {
  selectedCodes.value = [];
}

function openBatchEdit() {
  batchVisibility.value = "keep";
  batchExpiry.value = "keep";
  batchError.value = "";
  batchEditOpen.value = true;
}

function closeBatchEdit() {
  batchEditOpen.value = false;
  batchError.value = "";
}

async function runBounded<T, R>(
  items: T[],
  operation: (item: T) => Promise<R>,
) {
  const succeeded: Array<{ item: T; value: R }> = [];
  const failed: T[] = [];
  let cursor = 0;
  const workers = Array.from(
    { length: Math.min(4, items.length) },
    async () => {
      while (cursor < items.length) {
        const item = items[cursor++];
        if (!item) continue;
        try {
          succeeded.push({ item, value: await operation(item) });
        } catch {
          failed.push(item);
        }
      }
    },
  );
  await Promise.all(workers);
  return { succeeded, failed };
}

function updateSummary(summary: PasteSummary, paste: Paste): PasteSummary {
  return {
    ...summary,
    title: paste.title,
    tags: paste.tags,
    visibility: paste.visibility,
    passwordProtected: paste.passwordProtected,
    revision: paste.revision,
    updatedAt: paste.updatedAt,
    expiresAt: paste.expiresAt,
  };
}

async function applyBatchEdit() {
  const targets = [...selectedValues.value];
  if (targets.length === 0 || !hasBatchChanges.value) return;

  const patch: PastePatchInput = {};
  if (batchVisibility.value !== "keep")
    patch.visibility = batchVisibility.value;
  if (batchExpiry.value === "never") {
    patch.clearExpiry = true;
  } else if (batchExpiry.value !== "keep") {
    const hours = { "1h": 1, "1d": 24, "7d": 168, "30d": 720 }[
      batchExpiry.value
    ];
    patch.expiresAt = new Date(
      Date.now() + hours * 60 * 60 * 1000,
    ).toISOString();
  }

  batchSaving.value = true;
  batchError.value = "";
  const result = await runBounded(
    targets,
    async (value) => (await api.patch(value.code, value.revision, patch)).paste,
  );
  const updated = new Map(
    result.succeeded.map(({ item, value }) => [item.code, value]),
  );
  values.value = values.value.map((value) => {
    const paste = updated.get(value.code);
    return paste ? updateSummary(value, paste) : value;
  });
  selectedCodes.value = result.failed.map((value) => value.code);
  batchSaving.value = false;

  if (result.failed.length > 0) {
    batchError.value = `已修改 ${result.succeeded.length} 项，${result.failed.length} 项失败。失败项仍保持选中，可以重试。`;
    return;
  }
  batchEditOpen.value = false;
  announce(`已修改 ${result.succeeded.length} 个代码片段。`, "success");
}

async function removeSelected() {
  const targets = [...deleteTargets.value];
  if (targets.length === 0 || batchBusy.value || deleting.value) return;
  if (targets.length === 1) return remove(targets[0]!);

  batchDeleting.value = true;
  deleteError.value = "";
  const result = await runBounded(targets, (value) =>
    api.remove(value.code, value.revision),
  );
  const removed = new Set(result.succeeded.map(({ item }) => item.code));
  values.value = values.value.filter((value) => !removed.has(value.code));
  selectedCodes.value = result.failed.map((value) => value.code);
  if (removed.size) {
    libraryVersion.value++;
    await load();
  }

  if (result.failed.length > 0) {
    deleteTargets.value = result.failed;
    deleteError.value = `已删除 ${result.succeeded.length} 项，${result.failed.length} 项失败。可以重试剩余内容。`;
  } else {
    deleteOpen.value = false;
    announce(`已删除 ${result.succeeded.length} 个代码片段。`, "success");
  }
  batchDeleting.value = false;
}

function focusSearch(event: KeyboardEvent) {
  if (event.key !== "/" || event.metaKey || event.ctrlKey || event.altKey)
    return;
  const target = event.target as HTMLElement | null;
  if (target?.matches("input, textarea, [contenteditable='true']")) return;
  event.preventDefault();
  document.querySelector<HTMLInputElement>("#mine-search")?.focus();
}

function clearSearch() {
  query.value = "";
}

watch(
  () => route.query.q,
  (value) => {
    query.value = typeof value === "string" ? value : "";
    page.value = 1;
    load();
  },
);
onMounted(() => {
  load();
  window.addEventListener("keydown", focusSearch);
});
onBeforeUnmount(() => window.removeEventListener("keydown", focusSearch));
</script>

<template>
  <PasteWorkspace>
    <section class="paste-mine-shell flex h-full min-h-0 flex-col overflow-hidden" aria-labelledby="mine-title">
      <header class="paste-command-row flex min-h-[88px] shrink-0 items-center gap-4 border-b border-[var(--paste-line)] px-[30px] py-5 max-[760px]:px-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <UIcon name="i-tabler-folders" class="size-5 text-[var(--paste-blue)]" />
          <h1 id="mine-title" class="text-[21px] font-semibold tracking-tight max-[760px]:text-lg">内容库</h1>
          <span v-if="!loggedIn" class="text-xs text-[var(--paste-ink-dim)]">当前浏览器</span>
        </div>
        <UButton to="/" label="新建内容" icon="i-tabler-plus" class="paste-primary min-h-10 rounded-[7px] bg-[var(--paste-blue)] px-4 font-semibold text-[#17200e] hover:bg-[var(--paste-blue-hover)] not-dark:text-white" />
      </header>

      <div class="flex min-h-16 shrink-0 items-center gap-3 border-b border-[var(--paste-line)] px-[30px] py-3 max-[760px]:gap-2 max-[760px]:px-4" role="toolbar" aria-label="代码片段列表工具栏">
        <UCheckbox :model-value="selectionState" size="sm" class="flex size-8 shrink-0 items-center justify-center" aria-label="选择当前筛选结果" :disabled="loading || filtered.length === 0 || batchBusy" @update:model-value="toggleFilteredSelection" />
        <template v-if="selectedValues.length > 0">
          <strong class="min-w-0 flex-1 text-xs text-[var(--paste-ink)]">已选择 {{ selectedValues.length }} 项</strong>
          <UButton icon="i-tabler-adjustments" color="neutral" variant="ghost" size="sm" :disabled="batchBusy" @click="openBatchEdit">批量修改</UButton>
          <UButton icon="i-tabler-trash" color="error" variant="ghost" size="sm" :loading="batchDeleting" :disabled="batchSaving" @click="requestDelete(selectedValues)">删除</UButton>
          <UButton icon="i-tabler-x" color="neutral" variant="ghost" size="sm" square aria-label="取消全部选择" :disabled="batchBusy" @click="clearSelection" />
        </template>
        <template v-else>
          <UInput id="mine-search" v-model="query" type="search" icon="i-tabler-search" placeholder="搜索内容…" aria-label="搜索代码片段" variant="none" class="min-w-0 max-w-[460px] flex-1" :ui="{ base: 'h-9 rounded-lg bg-[var(--paste-surface)] ring-1 ring-[var(--paste-line)] focus:ring-[var(--paste-blue)]' }" @keyup.esc="clearSearch">
            <template #trailing><kbd class="rounded border border-[var(--paste-line)] px-1.5 text-[10px] text-[var(--paste-ink-dim)] max-[760px]:hidden">/</kbd></template>
          </UInput>
          <span class="ml-auto whitespace-nowrap text-xs text-[var(--paste-ink-dim)] max-[760px]:hidden">{{ total }} 项{{ query.trim() ? '匹配' : '内容' }}</span>
          <UButton icon="i-tabler-refresh" color="neutral" variant="ghost" square aria-label="刷新列表" :loading="loading" @click="load" />
        </template>
      </div>

      <div class="min-h-0 flex-1 overflow-auto bg-[var(--paste-page)] p-6 max-[760px]:p-3" :aria-busy="loading">
        <div v-if="loadingVisible" class="paste-ledger-loading mx-auto grid max-w-[1180px] gap-3" role="status" aria-label="正在读取代码片段">
          <div v-for="index in 4" :key="index" class="flex h-28 items-center gap-4 rounded-xl border border-[var(--paste-line)] bg-[var(--paste-editor)] p-5">
            <USkeleton class="size-11 shrink-0 rounded-lg" /><div class="flex-1"><USkeleton class="h-4 w-40 max-w-full" /><USkeleton class="mt-3 h-3 w-64 max-w-full" /></div>
          </div>
        </div>
        <div v-else-if="error" class="grid min-h-full place-content-center justify-items-center gap-3 p-6 text-center text-sm" role="alert">
          <UIcon name="i-tabler-alert-circle" class="size-7 text-[var(--paste-red)]" /><strong>内容读取失败</strong><p class="text-[var(--paste-ink-soft)]">{{ error }}</p>
          <UButton label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="load" />
        </div>
        <div v-else-if="filtered.length === 0" class="grid min-h-full place-content-center justify-items-center gap-4 p-6 text-center text-sm">
          <span class="grid size-14 place-items-center rounded-xl border border-[var(--paste-line)] bg-[var(--paste-surface)] text-[var(--paste-ink-dim)]"><UIcon :name="query.trim() ? 'i-tabler-file-search' : 'i-tabler-folders'" class="size-7" /></span>
          <strong>{{ query.trim() ? '没有找到匹配的内容' : '内容库还是空的' }}</strong>
          <UButton v-if="query.trim()" label="清除搜索" icon="i-tabler-x" color="neutral" variant="outline" @click="clearSearch" />
          <UButton v-else to="/" label="新建内容" icon="i-tabler-plus" color="primary" variant="outline" />
        </div>
        <div v-else class="mx-auto max-w-[1180px]">
          <div class="mb-3 flex items-center gap-2 px-1 text-[11px] text-[var(--paste-ink-dim)]"><UIcon name="i-tabler-clock" class="size-3.5" />最近创建</div>
          <div class="paste-ledger-rows grid gap-3" role="list" aria-label="内容库">
            <div v-for="value in filtered" :key="value.code" class="paste-ledger-row group grid min-w-0 grid-cols-[24px_44px_minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 rounded-xl border border-[var(--paste-line)] bg-[var(--paste-editor)] p-5 transition-colors hover:border-[var(--paste-line-strong)] focus-within:border-[var(--paste-line-strong)] data-[selected=true]:border-[var(--paste-blue)] data-[selected=true]:bg-[var(--paste-editor-active)] max-[760px]:grid-cols-[20px_minmax(0,1fr)_auto] max-[760px]:gap-x-2.5 max-[760px]:p-3.5" role="listitem" :data-selected="selectedCodeSet.has(value.code)" :aria-busy="deleting === value.code || batchBusy">
              <UCheckbox :model-value="selectedCodeSet.has(value.code)" size="sm" class="flex min-h-9 items-center" :aria-label="`选择 ${displayTitle(value.title)}`" :disabled="deleting === value.code || batchBusy" @update:model-value="toggleSelected(value.code, $event === true)" />
              <span class="grid size-11 place-items-center rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface)] text-[var(--paste-blue)] max-[760px]:hidden"><UIcon :name="value.primaryLanguage === 'markdown' ? 'i-tabler-markdown' : value.primaryLanguage === 'text' ? 'i-tabler-file-text' : 'i-tabler-code'" class="size-5" /></span>
              <div class="min-w-0">
                <NuxtLink :to="`/p/${value.code}`" :title="displayTitle(value.title)" class="block truncate py-0.5 text-sm font-semibold leading-6 text-[var(--paste-ink)] hover:text-[var(--paste-blue)]">{{ displayTitle(value.title) }}</NuxtLink>
                <div class="mt-1 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] text-[var(--paste-ink-dim)]">
                  <span class="rounded border border-[var(--paste-line)] px-1.5 py-0.5 font-mono text-[10px] text-[var(--paste-ink-soft)]">{{ value.primaryLanguage === 'text' ? 'TEXT' : value.primaryLanguage === 'markdown' ? 'Markdown' : value.primaryLanguage }}</span>
                  <span>{{ value.fileCount }} 个文件</span><code class="max-[760px]:hidden">/{{ value.code }}</code>
                  <span v-for="tag in value.tags.slice(0, 2)" :key="tag" class="max-w-28 truncate max-[760px]:hidden">#{{ tag }}</span>
                </div>
              </div>
              <div class="flex items-center gap-1 max-[760px]:col-start-2 max-[760px]:col-end-4 max-[760px]:row-start-3 max-[760px]:justify-end">
                <UTooltip :text="copied === value.code ? '已复制' : '复制链接'"><UButton :icon="copied === value.code ? 'i-tabler-check' : 'i-tabler-link'" color="neutral" variant="ghost" square aria-label="复制链接" class="size-9" :disabled="deleting === value.code || batchBusy" @click="copyLink(value)" /></UTooltip>
                <UTooltip text="编辑"><UButton :to="`/?edit=${value.code}`" icon="i-tabler-edit" color="neutral" variant="ghost" square aria-label="编辑代码片段" class="size-9" :disabled="deleting === value.code || batchBusy" /></UTooltip>
                <UTooltip text="删除"><UButton icon="i-tabler-trash" color="neutral" variant="ghost" square aria-label="删除代码片段" class="size-9 hover:text-[var(--paste-red)]" :loading="deleting === value.code" :disabled="batchBusy" @click="requestDelete([value])" /></UTooltip>
              </div>
              <div class="col-start-3 col-end-5 flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-[var(--paste-line)] pt-3 text-[11px] text-[var(--paste-ink-dim)] max-[760px]:col-start-2 max-[760px]:col-end-4 max-[760px]:row-start-2">
                <span class="inline-flex items-center gap-1.5"><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-3.5" />{{ value.visibility === 'private' ? '仅自己' : '链接访问' }}{{ value.passwordProtected ? ' · 密码保护' : '' }}</span>
                <span v-if="value.expiresAt" class="inline-flex items-center gap-1.5"><UIcon name="i-tabler-hourglass" class="size-3.5" />{{ new Date(value.expiresAt).getTime() < Date.now() ? '已过期' : `${new Date(value.expiresAt).toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })} 到期` }}</span>
                <time class="ml-auto" :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) }} 更新</time>
              </div>
            </div>
          </div>
        </div>
      </div>
      <footer class="paste-mine-statusbar shrink-0 border-t border-[var(--paste-line)] bg-[var(--paste-editor)] px-5 py-2 text-xs text-[var(--paste-ink-dim)] max-[760px]:px-3" aria-label="工作区状态">
        <CollectionPaginationBar :page-size-option="(size) => `${size} / 页`" :page="page" :page-size="pageSize" :page-sizes="[20, 50, 100]" :total="total" @page-size-change="pageSize = $event; page = 1; load();" @page-change="page = $event" />
        <span class="sr-only" aria-live="polite">{{ operationMessage }}</span>
      </footer>

      <UModal
        v-model:open="batchEditOpen"
        title="批量修改"
        :description="`为选中的 ${selectedValues.length} 个代码片段统一修改访问设置。`"
        :dismissible="!batchSaving"
        :close="!batchSaving"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div class="grid gap-[18px]">
            <UFormField label="可见性">
              <USelect
                v-model="batchVisibility"
                :items="visibilityItems"
                value-key="value"
                label-key="label"
                class="w-full"
                aria-label="批量可见性"
                :disabled="batchSaving"
              />
            </UFormField>
            <UFormField label="有效期">
              <USelect
                v-model="batchExpiry"
                :items="expiryItems"
                value-key="value"
                label-key="label"
                class="w-full"
                aria-label="批量有效期"
                :disabled="batchSaving"
              />
            </UFormField>
            <p
              class="mt-[-4px] text-xs leading-[1.65] text-[var(--paste-ink-soft)]"
            >
              “保持原设置”不会覆盖每个代码片段当前不同的值；已有密码和文件内容不会改变。
            </p>
            <UAlert
              v-if="batchError"
              color="warning"
              variant="subtle"
              icon="i-tabler-alert-triangle"
              title="部分修改未完成"
              :description="batchError"
              role="alert"
            />
          </div>
        </template>
        <template #footer>
          <div class="flex w-full items-center justify-end gap-1.5">
            <UButton
              type="button"
              label="取消"
              color="neutral"
              variant="ghost"
              :disabled="batchSaving"
              @click="closeBatchEdit"
            />
            <UButton
              type="button"
              label="应用修改"
              icon="i-tabler-check"
              color="primary"
              :loading="batchSaving"
              :disabled="
                batchSaving || !hasBatchChanges || selectedValues.length === 0
              "
              @click="applyBatchEdit"
            />
          </div>
        </template>
      </UModal>
      <PasteDeleteDialog v-model:open="deleteOpen" :titles="deleteTargets.map(value => displayTitle(value.title))" :busy="batchDeleting || Boolean(deleting)" :error="deleteError" @confirm="removeSelected" />
      <p class="sr-only" aria-live="polite">
        {{ copied ? "分享链接已复制" : operationMessage }}
      </p>
    </section>
  </PasteWorkspace>
</template>
