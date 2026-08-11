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
  title: "我的 Paste",
  description: "回看、编辑和删除你创建的 Paste。",
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
  toast.add({
    title: color === "success" ? "操作完成" : color === "warning" ? "部分操作未完成" : "操作失败",
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
  announce(`已修改 ${result.succeeded.length} 个 Paste。`, "success");
}

async function removeSelected() {
  const targets = [...selectedValues.value];
  if (targets.length === 0) return;
  if (!window.confirm(`删除选中的 ${targets.length} 个 Paste？这些链接将永久失效。`)) return;

  batchDeleting.value = true;
  const result = await runBounded(targets, (value) => api.remove(value.code, value.revision));
  const removed = new Set(result.succeeded.map(({ item }) => item.code));
  values.value = values.value.filter((value) => !removed.has(value.code));
  selectedCodes.value = result.failed.map((value) => value.code);
  batchDeleting.value = false;

  if (result.failed.length > 0) {
    announce(`已删除 ${result.succeeded.length} 项，${result.failed.length} 项失败。失败项仍保持选中。`, "warning");
  } else {
    announce(`已删除 ${result.succeeded.length} 个 Paste。`, "success");
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
  <section class="paste-mine-shell" aria-labelledby="mine-title">
    <header class="paste-mine-chrome" aria-label="Paste 工作区工具栏">
      <NuxtLink to="/" class="paste-mine-brand" aria-label="Paste 首页">
        <span class="paste-mine-brand-mark" aria-hidden="true">
          <UIcon name="i-tabler-code-dots" class="size-4" />
        </span>
        <span class="paste-mine-brand-label">Paste</span>
      </NuxtLink>

      <div class="paste-mine-tab" aria-current="page">
        <UIcon name="i-tabler-folders" class="size-4 shrink-0" />
        <h1 id="mine-title">我的 Paste</h1>
      </div>
      <div class="paste-mine-chrome-fill" aria-hidden="true" />

      <nav class="paste-mine-actions" aria-label="工作区操作">
        <UTooltip text="新建 Paste">
          <UButton
            to="/"
            icon="i-tabler-file-plus"
            color="primary"
            variant="ghost"
            size="sm"
            square
            aria-label="新建 Paste"
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
        <div class="paste-mine-account">
          <ConsumerAccountControl trigger-mode="collapsed" />
        </div>
      </nav>
    </header>

    <div class="paste-mine-toolbar" role="toolbar" aria-label="Paste 列表工具栏">
      <UTooltip :text="selectionState === true ? '取消选择筛选结果' : '选择全部筛选结果'">
        <UCheckbox
          :model-value="selectionState"
          size="sm"
          class="paste-select-all"
          aria-label="选择当前筛选结果"
          :disabled="loading || filtered.length === 0 || batchBusy"
          @update:model-value="toggleFilteredSelection"
        />
      </UTooltip>

      <template v-if="selectedValues.length > 0">
        <strong class="paste-selection-count">已选择 {{ selectedValues.length }} 项</strong>
        <div class="paste-selection-actions">
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
            class="paste-bulk-delete"
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
          aria-label="搜索 Paste"
          variant="none"
          size="sm"
          class="paste-mine-search"
          :ui="{ base: 'rounded-none ring-0' }"
          @keyup.esc="clearSearch"
        >
          <template #trailing>
            <kbd class="paste-search-key">/</kbd>
          </template>
        </UInput>
        <span class="paste-toolbar-count">{{ filtered.length }} / {{ values.length }}</span>
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

    <div class="paste-mine-viewport" :aria-busy="loading">
      <div v-if="loading" class="paste-ledger-loading" role="status" aria-label="正在读取 Paste">
        <div class="paste-ledger-head" aria-hidden="true">
          <span /><span>内容</span><span>访问</span><span>更新</span><span>操作</span>
        </div>
        <div v-for="index in 5" :key="index" class="paste-ledger-skeleton">
          <USkeleton class="size-3.5" />
          <div><USkeleton class="h-3.5 w-52 max-w-full" /><USkeleton class="mt-2 h-2.5 w-72 max-w-full" /></div>
          <USkeleton class="h-3 w-16" />
          <USkeleton class="h-3 w-20" />
          <USkeleton class="ml-auto h-7 w-24" />
        </div>
      </div>

      <div v-else-if="error" class="paste-mine-state" role="alert">
        <UIcon name="i-tabler-alert-circle" class="size-6 text-[var(--paste-red)]" />
        <strong>列表读取失败</strong>
        <p>{{ error }}</p>
        <UButton type="button" label="重新读取" icon="i-tabler-refresh" color="neutral" variant="outline" @click="load" />
      </div>

      <div v-else-if="values.length === 0" class="paste-mine-state">
        <UIcon name="i-tabler-file-code" class="size-7" />
        <strong>还没有可管理的 Paste</strong>
        <p>登录状态下创建的内容会出现在这个工作区。</p>
        <UButton to="/" label="新建 Paste" icon="i-tabler-file-plus" color="primary" />
      </div>

      <div v-else-if="filtered.length === 0" class="paste-mine-state">
        <UIcon name="i-tabler-file-search" class="size-7" />
        <strong>没有匹配的 Paste</strong>
        <p>换一个关键词，或清除当前搜索。</p>
        <UButton type="button" label="清除搜索" icon="i-tabler-x" color="neutral" variant="outline" @click="clearSearch" />
      </div>

      <div v-else class="paste-ledger">
        <div class="paste-ledger-head" aria-hidden="true">
          <span /><span>内容</span><span>访问</span><span>更新</span><span>操作</span>
        </div>
        <div class="paste-ledger-rows" role="list" aria-label="我的 Paste">
          <div
            v-for="value in filtered"
            :key="value.code"
            class="paste-ledger-row"
            role="listitem"
            :data-selected="selectedCodeSet.has(value.code)"
            :aria-busy="deleting === value.code || batchBusy"
          >
            <UCheckbox
              :model-value="selectedCodeSet.has(value.code)"
              size="sm"
              class="paste-row-select"
              :aria-label="`选择 ${displayTitle(value.title)}`"
              :disabled="deleting === value.code || batchBusy"
              @update:model-value="toggleSelected(value.code, $event === true)"
            />
            <div class="paste-ledger-primary">
              <NuxtLink :to="`/p/${value.code}`" :title="displayTitle(value.title)">{{ displayTitle(value.title) }}</NuxtLink>
              <div>
                <code>{{ value.code }}</code>
                <span>{{ value.fileCount }} 个文件</span>
                <span>{{ value.primaryLanguage }}</span>
                <span v-for="tag in value.tags.slice(0, 3)" :key="tag" class="paste-ledger-tag">#{{ tag }}</span>
              </div>
            </div>
            <div class="paste-ledger-access">
              <span><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
              <small v-if="value.passwordProtected">另有密码</small>
            </div>
            <time :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString("zh-CN") }}</time>
            <div class="paste-row-actions">
              <UTooltip :text="copied === value.code ? '已复制' : '复制链接'">
                <UButton type="button" :icon="copied === value.code ? 'i-tabler-check' : 'i-tabler-copy'" color="neutral" variant="ghost" size="sm" square aria-label="复制链接" :disabled="deleting === value.code || batchBusy" @click="copyLink(value)" />
              </UTooltip>
              <UTooltip text="编辑">
                <UButton :to="`/?edit=${value.code}`" icon="i-tabler-edit" color="neutral" variant="ghost" size="sm" square aria-label="编辑 Paste" :disabled="deleting === value.code || batchBusy" />
              </UTooltip>
              <UTooltip text="删除">
                <UButton type="button" icon="i-tabler-trash" color="error" variant="ghost" size="sm" square aria-label="删除 Paste" :loading="deleting === value.code" :disabled="batchBusy" @click="remove(value)" />
              </UTooltip>
            </div>
          </div>
        </div>
      </div>
    </div>

    <footer class="paste-mine-statusbar" aria-label="工作区状态">
      <span class="paste-mine-status-primary">
        <span class="paste-mine-status-dot" aria-hidden="true" />
        我的 Paste
      </span>
      <span aria-live="polite">{{ selectedValues.length ? `已选择 ${selectedValues.length} 项` : query ? `筛选 ${filtered.length} / ${values.length}` : operationMessage || `${values.length} 条记录` }}</span>
    </footer>

    <USlideover
      v-model:open="batchEditOpen"
      title="批量修改"
      :description="`为选中的 ${selectedValues.length} 个 Paste 统一修改访问设置。`"
      :dismissible="!batchSaving"
      :close="!batchSaving"
      :ui="{ content: 'sm:max-w-sm' }"
    >
      <template #body>
        <div class="paste-batch-fields">
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
          <p class="paste-batch-note">“保持原设置”不会覆盖每个 Paste 当前不同的值；已有密码和文件内容不会改变。</p>
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
        <div class="paste-batch-footer">
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

<style scoped>
.paste-mine-shell { display: grid; width: 100%; height: 100dvh; min-width: 0; min-height: 360px; grid-template-rows: 40px 38px minmax(0, 1fr) 28px; overflow: hidden; background: var(--paste-editor); }
.paste-mine-chrome { display: flex; min-width: 0; align-items: stretch; border-bottom: 1px solid var(--paste-line); background: var(--paste-chrome); }
.paste-mine-brand { display: flex; flex: none; align-items: center; gap: 7px; border-right: 1px solid var(--paste-line); padding: 0 11px; color: var(--paste-ink); text-decoration: none; }
.paste-mine-brand-mark { display: grid; width: 24px; height: 24px; place-items: center; border: 1px solid color-mix(in srgb, var(--paste-blue) 28%, var(--paste-line)); border-radius: 6px; background: var(--paste-blue-soft); color: var(--paste-blue); }
.paste-mine-brand-label { font-size: 13px; font-weight: 730; letter-spacing: -.025em; }
.paste-mine-tab { display: flex; min-width: 156px; max-width: 220px; flex: 0 1 190px; align-items: center; gap: 7px; border-right: 1px solid var(--paste-line); padding: 0 12px; background: var(--paste-editor); box-shadow: inset 0 2px var(--paste-blue); color: var(--paste-ink); }
.paste-mine-tab h1 { overflow: hidden; margin: 0; font-size: 12px; font-weight: 620; text-overflow: ellipsis; white-space: nowrap; }
.paste-mine-chrome-fill { min-width: 8px; flex: 1; }
.paste-mine-actions { display: flex; flex: none; align-items: center; gap: 2px; padding: 0 5px; }
.paste-mine-account { display: grid; width: 32px; height: 32px; place-items: center; }
.paste-mine-account :deep(button) { width: 30px; min-width: 30px; height: 30px; min-height: 30px; padding: 0; }
.paste-mine-account :deep([data-slot="label"]) { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.paste-mine-toolbar { display: flex; min-width: 0; align-items: center; gap: 6px; border-bottom: 1px solid var(--paste-line); padding: 3px 6px 3px 10px; background: var(--paste-editor); }
.paste-select-all { width: 24px; min-width: 24px; height: 28px; flex: none; align-items: center; justify-content: center; }
.paste-mine-search { min-width: 160px; max-width: 520px; flex: 1; }
.paste-mine-search :deep(input) { font-size: 12px; }
.paste-search-key { border: 1px solid var(--paste-line); border-radius: 3px; padding: 0 5px; color: var(--paste-ink-dim); font-family: inherit; font-size: 10px; line-height: 18px; }
.paste-toolbar-count { margin-left: auto; color: var(--paste-ink-dim); font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; white-space: nowrap; }
.paste-selection-count { overflow: hidden; color: var(--paste-ink); font-size: 12px; font-weight: 680; text-overflow: ellipsis; white-space: nowrap; }
.paste-selection-actions { display: flex; min-width: 0; align-items: center; gap: 2px; margin-left: auto; }
.paste-bulk-delete { color: var(--paste-red); }
.paste-bulk-delete:hover { background: var(--paste-red-soft); color: var(--paste-red); }
.paste-mine-viewport { min-width: 0; min-height: 0; overflow: auto; background: var(--paste-editor); }
.paste-ledger { min-width: 0; }
.paste-ledger-head, .paste-ledger-row, .paste-ledger-skeleton { display: grid; grid-template-columns: 32px minmax(240px, 1fr) 132px 112px 120px; align-items: center; }
.paste-ledger-head { position: sticky; z-index: 2; top: 0; min-height: 30px; border-bottom: 1px solid var(--paste-line); padding: 0 12px; background: var(--paste-surface-muted); color: var(--paste-ink-dim); font-size: 11px; font-weight: 630; }
.paste-ledger-row { min-width: 0; min-height: 62px; border-bottom: 1px solid var(--paste-line); padding: 8px 12px; transition: background 120ms ease; }
.paste-ledger-row:hover, .paste-ledger-row:focus-within { background: var(--paste-editor-active); }
.paste-ledger-row[data-selected="true"] { background: color-mix(in srgb, var(--paste-blue) 9%, var(--paste-editor)); }
.paste-ledger-row[data-selected="true"]:hover, .paste-ledger-row[data-selected="true"]:focus-within { background: color-mix(in srgb, var(--paste-blue) 13%, var(--paste-editor)); }
.paste-row-select { width: 24px; height: 28px; align-items: center; justify-content: flex-start; }
.paste-ledger-primary { min-width: 0; padding-right: 14px; }
.paste-ledger-primary > a { display: block; overflow: hidden; color: var(--paste-ink); font-size: 13px; font-weight: 680; text-decoration: none; text-overflow: ellipsis; white-space: nowrap; }
.paste-ledger-primary > a:hover { color: var(--paste-blue); text-decoration: underline; text-underline-offset: 3px; }
.paste-ledger-primary > div { display: flex; min-width: 0; align-items: center; gap: 5px 11px; margin-top: 5px; overflow: hidden; color: var(--paste-ink-soft); font-size: 11px; white-space: nowrap; }
.paste-ledger-primary code { flex: none; color: var(--paste-blue); font-weight: 680; }
.paste-ledger-tag { overflow: hidden; max-width: 100px; flex: 0 1 auto; text-overflow: ellipsis; }
.paste-ledger-access { display: grid; gap: 2px; color: var(--paste-ink-soft); font-size: 12px; }
.paste-ledger-access span { display: flex; align-items: center; gap: 5px; }
.paste-ledger-access small { color: var(--paste-ink-dim); font-size: 10px; }
.paste-ledger-row time { color: var(--paste-ink-soft); font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; }
.paste-row-actions { display: flex; justify-content: flex-end; gap: 1px; opacity: .72; transition: opacity 120ms ease; }
.paste-ledger-row:hover .paste-row-actions, .paste-ledger-row:focus-within .paste-row-actions { opacity: 1; }
.paste-ledger-skeleton { min-height: 62px; border-bottom: 1px solid var(--paste-line); padding: 8px 12px; }
.paste-mine-state { display: grid; min-height: 100%; place-content: center; justify-items: center; gap: 9px; padding: 28px; color: var(--paste-ink-soft); font-size: 12px; text-align: center; }
.paste-mine-state strong { color: var(--paste-ink); font-size: 14px; }
.paste-mine-state p { max-width: 54ch; margin: 0 0 3px; line-height: 1.6; }
.paste-mine-statusbar { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 16px; border-top: 1px solid var(--paste-status-border); padding: 0 10px; padding-bottom: env(safe-area-inset-bottom); background: var(--paste-status); color: var(--paste-status-text-muted); font-family: "SFMono-Regular", Consolas, monospace; font-size: 12px; }
.paste-mine-status-primary { display: flex; align-items: center; gap: 7px; color: var(--paste-status-text); white-space: nowrap; }
.paste-mine-status-dot { width: 7px; height: 7px; border-radius: 50%; background: var(--paste-green); }
.paste-batch-fields { display: grid; gap: 18px; }
.paste-batch-note { margin: -4px 0 0; color: var(--paste-ink-soft); font-size: 12px; line-height: 1.65; }
.paste-batch-footer { display: flex; width: 100%; align-items: center; justify-content: flex-end; gap: 6px; }
@media (max-width: 760px) {
  .paste-mine-shell { grid-template-rows: 44px 44px minmax(0, 1fr) 30px; min-height: 320px; }
  .paste-mine-brand { padding-inline: 6px; }
  .paste-mine-brand-label { display: none; }
  .paste-mine-tab { min-width: 0; max-width: none; flex: 1; padding-inline: 9px; }
  .paste-mine-chrome-fill { display: none; }
  .paste-mine-actions { padding-inline: 4px; }
  .paste-mine-account, .paste-mine-account :deep(button) { width: 36px; min-width: 36px; height: 36px; min-height: 36px; }
  .paste-mine-toolbar { gap: 3px; padding: 4px 5px; }
  .paste-select-all { width: 34px; min-width: 34px; height: 36px; }
  .paste-mine-search { max-width: none; }
  .paste-search-key, .paste-toolbar-count { display: none; }
  .paste-selection-count { min-width: 0; flex: 1; }
  .paste-selection-actions { flex: none; margin-left: 0; }
  .paste-selection-actions :deep(button) { min-height: 36px; }
  .paste-ledger-head { display: none; }
  .paste-ledger-row { grid-template-columns: 28px minmax(0, 1fr) auto; gap: 8px 10px; padding: 11px 9px; }
  .paste-row-select { grid-column: 1; grid-row: 1 / 4; width: 44px; height: 44px; align-self: start; justify-content: center; margin: -8px 0 0 -8px; }
  .paste-ledger-primary { grid-column: 2 / -1; padding-right: 0; }
  .paste-ledger-primary > a { overflow-wrap: anywhere; white-space: normal; }
  .paste-ledger-primary > div { gap: 5px 9px; }
  .paste-ledger-tag { display: none; }
  .paste-ledger-access { grid-column: 2; grid-row: 2; }
  .paste-ledger-access small { display: none; }
  .paste-ledger-row time { grid-column: 3; grid-row: 2; text-align: right; }
  .paste-row-actions { grid-column: 2 / -1; grid-row: 3; justify-content: flex-end; opacity: 1; }
  .paste-row-actions :deep(button), .paste-row-actions :deep(a) { width: 44px; min-width: 44px; height: 44px; min-height: 44px; }
  .paste-ledger-skeleton { grid-template-columns: 28px minmax(0, 1fr) 70px; }
  .paste-ledger-skeleton > :nth-child(4), .paste-ledger-skeleton > :nth-child(5) { display: none; }
}
</style>
