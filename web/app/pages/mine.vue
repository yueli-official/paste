<!--
THESIS: Paste history is another editor workspace, not an administrative card page.
OWN-WORLD: The same flat Yueli editor chrome frames a dense, readable file ledger.
STORY: search the workspace, open a Paste, act on the row, and keep moving.
FIRST VIEWPORT: one 40px titlebar, a compact search toolbar, the ledger, and a readable statusbar.
FORM: fixed-height application shell with one internally scrolling content plane.
-->
<script setup lang="ts">
import type { Paste, PastePatchInput, PasteSummary, PasteVisibility } from "../types/paste";
import { displayTitle, pasteErrorMessage } from "../utils/paste";

definePageMeta({ middleware: "auth" });

const api = usePasteApi();
const toast = useToast();
const { isAdmin } = useAuth();
const { siteName } = useSiteSettings();
const accountActions = computed(() => [
  ...(isAdmin.value
    ? [{ label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" }]
    : []),
]);
const values = ref<PasteSummary[]>([]);
const query = ref("");
const loading = ref(true);
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
const operationMessage = ref("");

const visibilityItems: Array<{ label: string; value: "keep" | PasteVisibility }> = [
  { label: "保持原设置", value: "keep" },
  { label: "持链访问", value: "unlisted" },
  { label: "仅自己", value: "private" },
];
const expiryItems: Array<{ label: string; value: "keep" | "1h" | "1d" | "7d" | "30d" | "never" }> = [
  { label: "保持原设置", value: "keep" },
  { label: "1 小时后", value: "1h" },
  { label: "1 天后", value: "1d" },
  { label: "7 天后", value: "7d" },
  { label: "30 天后", value: "30d" },
  { label: "不过期", value: "never" },
];

const filtered = computed(() => {
  const needle = query.value.trim().toLowerCase();
  if (!needle) return values.value;
  return values.value.filter((value) =>
    [value.title, value.code, value.primaryLanguage, ...value.tags]
      .join(" ")
      .toLowerCase()
      .includes(needle),
  );
});
const selectedCodeSet = computed(() => new Set(selectedCodes.value));
const selectedValues = computed(() => values.value.filter((value) => selectedCodeSet.value.has(value.code)));
const selectionState = computed<boolean | "indeterminate">(() => {
  if (filtered.value.length === 0) return false;
  const count = filtered.value.filter((value) => selectedCodeSet.value.has(value.code)).length;
  if (count === 0) return false;
  return count === filtered.value.length ? true : "indeterminate";
});
const batchBusy = computed(() => batchSaving.value || batchDeleting.value);
const hasBatchChanges = computed(() => batchVisibility.value !== "keep" || batchExpiry.value !== "keep");

useSeoMeta({
  title: computed(() => `我的片段 · ${siteName.value}`),
  description: "回看、编辑和删除你创建的代码片段。",
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    values.value = (await api.listMine()).pastes || [];
    const available = new Set(values.value.map((value) => value.code));
    selectedCodes.value = selectedCodes.value.filter((code) => available.has(code));
  } catch (caught) {
    error.value = pasteErrorMessage(caught);
  } finally {
    loading.value = false;
  }
}

async function copyLink(value: PasteSummary) {
  try {
    await navigator.clipboard.writeText(value.shareUrl);
    copied.value = value.code;
    window.setTimeout(() => { if (copied.value === value.code) copied.value = ""; }, 1800);
  } catch {
    announce("复制失败，请手动复制分享链接。", "error");
  }
}

async function remove(value: PasteSummary) {
  if (!window.confirm(`删除“${displayTitle(value.title)}”？链接将永久返回已失效。`)) return;
  deleting.value = value.code;
  try {
    await api.remove(value.code, value.revision);
    values.value = values.value.filter((candidate) => candidate.code !== value.code);
    selectedCodes.value = selectedCodes.value.filter((code) => code !== value.code);
  } catch (caught) {
    announce(pasteErrorMessage(caught), "error");
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
  if (batchVisibility.value !== "keep") patch.visibility = batchVisibility.value;
  if (batchExpiry.value === "never") {
    patch.clearExpiry = true;
  } else if (batchExpiry.value !== "keep") {
    const hours = { "1h": 1, "1d": 24, "7d": 168, "30d": 720 }[batchExpiry.value];
    patch.expiresAt = new Date(Date.now() + hours * 60 * 60 * 1000).toISOString();
  }

  batchSaving.value = true;
  batchError.value = "";
  const result = await runBounded(targets, async (value) =>
    (await api.patch(value.code, value.revision, patch)).paste,
  );
  const updated = new Map(result.succeeded.map(({ item, value }) => [item.code, value]));
  values.value = values.value.map((value) => {
    const paste = updated.get(value.code);
    return paste ? updateSummary(value, paste) : value;
  });
  selectedCodes.value = result.failed.map((value) => value.code);
  batchSaving.value = false;

  if (result.failed.length > 0) {
    batchError.value = `已修改 ${result.succeeded.length} 项，${result.failed.length} 项失败。失败项仍保持选中，可以重试。`;
    announce(batchError.value, "warning");
    return;
  }
  batchEditOpen.value = false;
  announce(`已修改 ${result.succeeded.length} 个代码片段。`, "success");
}

async function removeSelected() {
  const targets = [...selectedValues.value];
  if (targets.length === 0) return;
  if (!window.confirm(`删除选中的 ${targets.length} 个代码片段？这些链接将永久失效。`)) return;

  batchDeleting.value = true;
  const result = await runBounded(targets, (value) => api.remove(value.code, value.revision));
  const removed = new Set(result.succeeded.map(({ item }) => item.code));
  values.value = values.value.filter((value) => !removed.has(value.code));
  selectedCodes.value = result.failed.map((value) => value.code);
  batchDeleting.value = false;

  if (result.failed.length > 0) {
    announce(`已删除 ${result.succeeded.length} 项，${result.failed.length} 项失败。失败项仍保持选中。`, "warning");
  } else {
    announce(`已删除 ${result.succeeded.length} 个代码片段。`, "success");
  }
}

function focusSearch(event: KeyboardEvent) {
  if (event.key !== "/" || event.metaKey || event.ctrlKey || event.altKey) return;
  const target = event.target as HTMLElement | null;
  if (target?.matches("input, textarea, [contenteditable='true']")) return;
  event.preventDefault();
  document.querySelector<HTMLInputElement>("#mine-search")?.focus();
}

function clearSearch() {
  query.value = "";
}

onMounted(() => {
  load();
  window.addEventListener("keydown", focusSearch);
});
onBeforeUnmount(() => window.removeEventListener("keydown", focusSearch));
</script>

<template>
  <section
    class="paste-mine-shell grid h-dvh min-h-[360px] w-full min-w-0 grid-rows-[40px_38px_minmax(0,1fr)_28px] overflow-hidden bg-[var(--paste-editor)] max-[760px]:min-h-[320px] max-[760px]:grid-rows-[44px_44px_minmax(0,1fr)_30px]"
    aria-labelledby="mine-title"
  >
    <header
      class="flex min-w-0 items-stretch border-b border-[var(--paste-line)] bg-[var(--paste-chrome)]"
      aria-label="代码片段工作区工具栏"
    >
      <NuxtLink
        to="/"
        class="flex shrink-0 items-center gap-[7px] border-r border-[var(--paste-line)] px-[11px] text-[var(--paste-ink)] no-underline max-[760px]:px-1.5"
        :aria-label="`${siteName} 首页`"
      >
        <span class="grid size-6 place-items-center rounded-md border border-[color-mix(in_srgb,var(--paste-blue)_28%,var(--paste-line))] bg-[var(--paste-blue-soft)] text-[var(--paste-blue)]" aria-hidden="true">
          <UIcon name="i-tabler-code-dots" class="size-4" />
        </span>
        <span class="text-[13px] font-[730] tracking-[-0.025em] max-[760px]:hidden">{{ siteName }}</span>
      </NuxtLink>

      <div class="flex min-w-[156px] max-w-[220px] basis-[190px] items-center gap-[7px] border-r border-[var(--paste-line)] bg-[var(--paste-editor)] px-3 text-[var(--paste-ink)] [box-shadow:inset_0_2px_var(--paste-blue)] max-[760px]:min-w-0 max-[760px]:max-w-none max-[760px]:flex-1 max-[760px]:px-[9px]" aria-current="page">
        <UIcon name="i-tabler-folders" class="size-4 shrink-0" />
        <h1 id="mine-title" class="truncate text-xs font-[620]">我的片段</h1>
      </div>
      <div class="min-w-2 flex-1 max-[760px]:hidden" aria-hidden="true" />

      <nav class="flex shrink-0 items-center gap-0.5 px-[5px] max-[760px]:px-1" aria-label="工作区操作">
        <UTooltip text="新建片段">
          <UButton
            to="/"
            icon="i-tabler-file-plus"
            color="primary"
            variant="ghost"
            size="sm"
            square
            aria-label="新建片段"
          />
        </UTooltip>
        <UTooltip text="切换颜色模式">
          <UColorModeButton
            color="neutral"
            variant="ghost"
            size="sm"
            aria-label="切换颜色模式"
          />
        </UTooltip>
        <div class="grid size-8 place-items-center max-[760px]:size-9 [&_button]:size-[30px] [&_button]:min-h-[30px] [&_button]:min-w-[30px] [&_button]:p-0 max-[760px]:[&_button]:size-9 max-[760px]:[&_button]:min-h-9 max-[760px]:[&_button]:min-w-9 [&_[data-slot=label]]:sr-only">
          <ConsumerAccountControl :context-actions="accountActions" trigger-mode="collapsed" />
        </div>
      </nav>
    </header>

    <div class="flex min-w-0 items-center gap-1.5 border-b border-[var(--paste-line)] bg-[var(--paste-editor)] py-[3px] pr-1.5 pl-2.5 max-[760px]:gap-[3px] max-[760px]:p-[4px_5px]" role="toolbar" aria-label="代码片段列表工具栏">
      <UTooltip :text="selectionState === true ? '取消选择筛选结果' : '选择全部筛选结果'">
        <UCheckbox
          :model-value="selectionState"
          size="sm"
          class="h-7 w-6 min-w-6 shrink-0 items-center justify-center max-[760px]:h-9 max-[760px]:w-[34px] max-[760px]:min-w-[34px]"
          aria-label="选择当前筛选结果"
          :disabled="loading || filtered.length === 0 || batchBusy"
          @update:model-value="toggleFilteredSelection"
        />
      </UTooltip>

      <template v-if="selectedValues.length > 0">
        <strong class="min-w-0 truncate text-xs font-[680] text-[var(--paste-ink)] max-[760px]:flex-1">已选择 {{ selectedValues.length }} 项</strong>
        <div class="ml-auto flex min-w-0 items-center gap-0.5 max-[760px]:ml-0 max-[760px]:shrink-0 max-[760px]:[&_button]:min-h-9">
          <UButton
            type="button"
            icon="i-tabler-adjustments"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="batchBusy"
            @click="openBatchEdit"
          >批量修改</UButton>
          <UButton
            type="button"
            icon="i-tabler-trash"
            color="neutral"
            variant="ghost"
            size="sm"
            class="text-[var(--paste-red)] hover:bg-[var(--paste-red-soft)] hover:text-[var(--paste-red)]"
            :loading="batchDeleting"
            :disabled="batchSaving"
            @click="removeSelected"
          >删除</UButton>
          <UTooltip text="取消选择">
            <UButton
              type="button"
              icon="i-tabler-x"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              aria-label="取消全部选择"
              :disabled="batchBusy"
              @click="clearSelection"
            />
          </UTooltip>
        </div>
      </template>

      <template v-else>
        <UInput
          id="mine-search"
          v-model="query"
          type="search"
          icon="i-tabler-search"
          placeholder="搜索标题、短码、语言或标签"
          aria-label="搜索代码片段"
          variant="none"
          size="sm"
          class="min-w-40 max-w-[520px] flex-1 max-[760px]:max-w-none"
          :ui="{ base: 'rounded-none ring-0' }"
          @keyup.esc="clearSearch"
        >
          <template #trailing>
            <kbd class="rounded-[3px] border border-[var(--paste-line)] px-[5px] text-[10px] leading-[18px] text-[var(--paste-ink-dim)] max-[760px]:hidden">/</kbd>
          </template>
        </UInput>
        <span class="ml-auto whitespace-nowrap font-mono text-[11px] text-[var(--paste-ink-dim)] max-[760px]:hidden">{{ filtered.length }} / {{ values.length }}</span>
        <UTooltip text="刷新列表">
          <UButton
            type="button"
            icon="i-tabler-refresh"
            color="neutral"
            variant="ghost"
            size="sm"
            square
            aria-label="刷新列表"
            :loading="loading"
            @click="load"
          />
        </UTooltip>
      </template>
    </div>

    <div class="min-h-0 min-w-0 overflow-auto bg-[var(--paste-editor)]" :aria-busy="loading">
      <div v-if="loading" class="paste-ledger-loading" role="status" aria-label="正在读取代码片段">
        <div class="sticky top-0 z-[2] grid min-h-[30px] grid-cols-[32px_minmax(240px,1fr)_132px_112px_120px] items-center border-b border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-3 text-[11px] font-[630] text-[var(--paste-ink-dim)] max-[760px]:hidden" aria-hidden="true">
          <span /><span>内容</span><span>访问</span><span>更新</span><span>操作</span>
        </div>
        <div v-for="index in 5" :key="index" class="grid min-h-[62px] grid-cols-[32px_minmax(240px,1fr)_132px_112px_120px] items-center border-b border-[var(--paste-line)] px-3 py-2 max-[760px]:grid-cols-[28px_minmax(0,1fr)_70px] max-[760px]:[&>:nth-child(4)]:hidden max-[760px]:[&>:nth-child(5)]:hidden">
          <USkeleton class="size-3.5" />
          <div><USkeleton class="h-3.5 w-52 max-w-full" /><USkeleton class="mt-2 h-2.5 w-72 max-w-full" /></div>
          <USkeleton class="h-3 w-16" />
          <USkeleton class="h-3 w-20" />
          <USkeleton class="ml-auto h-7 w-24" />
        </div>
      </div>

      <div v-else-if="error" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[54ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]" role="alert">
        <UIcon name="i-tabler-alert-circle" class="size-6 text-[var(--paste-red)]" />
        <strong>列表读取失败</strong>
        <p>{{ error }}</p>
        <UButton type="button" label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="load" />
      </div>

      <div v-else-if="values.length === 0" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[54ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]">
        <UIcon name="i-tabler-file-code" class="size-7" />
        <strong>还没有可管理的代码片段</strong>
        <p>登录状态下创建的内容会出现在这个工作区。</p>
        <UButton to="/" label="新建片段" icon="i-tabler-file-plus" color="primary" />
      </div>

      <div v-else-if="filtered.length === 0" class="grid min-h-full place-content-center justify-items-center gap-[9px] p-7 text-center text-xs text-[var(--paste-ink-soft)] [&_p]:mb-[3px] [&_p]:max-w-[54ch] [&_p]:leading-[1.6] [&_strong]:text-sm [&_strong]:text-[var(--paste-ink)]">
        <UIcon name="i-tabler-file-search" class="size-7" />
        <strong>没有匹配的代码片段</strong>
        <p>换一个关键词，或清除当前搜索。</p>
        <UButton type="button" label="清除搜索" icon="i-tabler-x" color="neutral" variant="outline" @click="clearSearch" />
      </div>

      <div v-else class="min-w-0">
        <div class="sticky top-0 z-[2] grid min-h-[30px] grid-cols-[32px_minmax(240px,1fr)_132px_112px_120px] items-center border-b border-[var(--paste-line)] bg-[var(--paste-surface-muted)] px-3 text-[11px] font-[630] text-[var(--paste-ink-dim)] max-[760px]:hidden" aria-hidden="true">
          <span /><span>内容</span><span>访问</span><span>更新</span><span>操作</span>
        </div>
        <div class="paste-ledger-rows" role="list" aria-label="我的片段">
          <div
            v-for="value in filtered"
            :key="value.code"
            class="paste-ledger-row group grid min-h-[62px] min-w-0 grid-cols-[32px_minmax(240px,1fr)_132px_112px_120px] items-center border-b border-[var(--paste-line)] px-3 py-2 transition-colors duration-100 hover:bg-[var(--paste-editor-active)] focus-within:bg-[var(--paste-editor-active)] data-[selected=true]:bg-[color-mix(in_srgb,var(--paste-blue)_9%,var(--paste-editor))] data-[selected=true]:hover:bg-[color-mix(in_srgb,var(--paste-blue)_13%,var(--paste-editor))] data-[selected=true]:focus-within:bg-[color-mix(in_srgb,var(--paste-blue)_13%,var(--paste-editor))] max-[760px]:grid-cols-[28px_minmax(0,1fr)_auto] max-[760px]:gap-x-2.5 max-[760px]:gap-y-2 max-[760px]:px-[9px] max-[760px]:py-[11px]"
            role="listitem"
            :data-selected="selectedCodeSet.has(value.code)"
            :aria-busy="deleting === value.code || batchBusy"
          >
            <UCheckbox
              :model-value="selectedCodeSet.has(value.code)"
              size="sm"
              class="h-7 w-6 items-center justify-start max-[760px]:col-start-1 max-[760px]:row-start-1 max-[760px]:row-end-4 max-[760px]:mt-[-8px] max-[760px]:ml-[-8px] max-[760px]:size-11 max-[760px]:self-start max-[760px]:justify-center"
              :aria-label="`选择 ${displayTitle(value.title)}`"
              :disabled="deleting === value.code || batchBusy"
              @update:model-value="toggleSelected(value.code, $event === true)"
            />
            <div class="min-w-0 pr-3.5 max-[760px]:col-start-2 max-[760px]:col-end-[-1] max-[760px]:pr-0">
              <NuxtLink class="block truncate text-[13px] font-[680] text-[var(--paste-ink)] no-underline hover:text-[var(--paste-blue)] hover:underline hover:underline-offset-3 max-[760px]:whitespace-normal max-[760px]:[overflow-wrap:anywhere]" :to="`/p/${value.code}`" :title="displayTitle(value.title)">{{ displayTitle(value.title) }}</NuxtLink>
              <div class="mt-[5px] flex min-w-0 items-center gap-x-[11px] gap-y-[5px] overflow-hidden whitespace-nowrap text-[11px] text-[var(--paste-ink-soft)] max-[760px]:gap-x-[9px]">
                <code class="shrink-0 font-[680] text-[var(--paste-blue)]">{{ value.code }}</code>
                <span>{{ value.fileCount }} 个文件</span>
                <span>{{ value.primaryLanguage }}</span>
                <span v-for="tag in value.tags.slice(0, 3)" :key="tag" class="max-w-[100px] truncate max-[760px]:hidden">#{{ tag }}</span>
              </div>
            </div>
            <div class="grid gap-0.5 text-xs text-[var(--paste-ink-soft)] max-[760px]:col-start-2 max-[760px]:row-start-2 [&_span]:flex [&_span]:items-center [&_span]:gap-[5px] [&_small]:text-[10px] [&_small]:text-[var(--paste-ink-dim)] max-[760px]:[&_small]:hidden">
              <span><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
              <small v-if="value.passwordProtected">另有密码</small>
            </div>
            <time class="font-mono text-[11px] text-[var(--paste-ink-soft)] max-[760px]:col-start-3 max-[760px]:row-start-2 max-[760px]:text-right" :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString("zh-CN") }}</time>
            <div class="flex justify-end gap-px opacity-70 transition-opacity duration-100 group-hover:opacity-100 group-focus-within:opacity-100 max-[760px]:col-start-2 max-[760px]:col-end-[-1] max-[760px]:row-start-3 max-[760px]:opacity-100 max-[760px]:[&_a]:size-11 max-[760px]:[&_a]:min-h-11 max-[760px]:[&_a]:min-w-11 max-[760px]:[&_button]:size-11 max-[760px]:[&_button]:min-h-11 max-[760px]:[&_button]:min-w-11">
              <UTooltip :text="copied === value.code ? '已复制' : '复制链接'">
                <UButton type="button" :icon="copied === value.code ? 'i-tabler-check' : 'i-tabler-copy'" color="neutral" variant="ghost" size="sm" square aria-label="复制链接" :disabled="deleting === value.code || batchBusy" @click="copyLink(value)" />
              </UTooltip>
              <UTooltip text="编辑">
                <UButton :to="`/?edit=${value.code}`" icon="i-tabler-edit" color="neutral" variant="ghost" size="sm" square aria-label="编辑代码片段" :disabled="deleting === value.code || batchBusy" />
              </UTooltip>
              <UTooltip text="删除">
                <UButton type="button" icon="i-tabler-trash" color="error" variant="ghost" size="sm" square aria-label="删除代码片段" :loading="deleting === value.code" :disabled="batchBusy" @click="remove(value)" />
              </UTooltip>
            </div>
          </div>
        </div>
      </div>
    </div>

    <footer class="paste-mine-statusbar flex min-w-0 items-center justify-between gap-4 border-t border-[var(--paste-status-border)] bg-[var(--paste-status)] px-2.5 pb-[env(safe-area-inset-bottom)] font-mono text-xs text-[var(--paste-status-text-muted)]" aria-label="工作区状态">
      <span class="flex items-center gap-[7px] whitespace-nowrap text-[var(--paste-status-text)]">
        <span class="size-[7px] rounded-full bg-[var(--paste-green)]" aria-hidden="true" />
        我的片段
      </span>
      <span aria-live="polite">{{ selectedValues.length ? `已选择 ${selectedValues.length} 项` : query ? `筛选 ${filtered.length} / ${values.length}` : operationMessage || `${values.length} 条记录` }}</span>
    </footer>

    <USlideover
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
          <p class="mt-[-4px] text-xs leading-[1.65] text-[var(--paste-ink-soft)]">“保持原设置”不会覆盖每个代码片段当前不同的值；已有密码和文件内容不会改变。</p>
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
            :disabled="batchSaving || !hasBatchChanges || selectedValues.length === 0"
            @click="applyBatchEdit"
          />
        </div>
      </template>
    </USlideover>
    <p class="sr-only" aria-live="polite">{{ copied ? "分享链接已复制" : operationMessage }}</p>
  </section>
</template>