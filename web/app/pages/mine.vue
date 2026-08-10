<script setup lang="ts">
import type { PasteSummary } from "../types/paste";
import { displayTitle, pasteErrorMessage } from "../utils/paste";

definePageMeta({ middleware: "auth" });

const api = usePasteApi();
const values = ref<PasteSummary[]>([]);
const query = ref("");
const loading = ref(true);
const deleting = ref("");
const error = ref("");
const copied = ref("");

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

useSeoMeta({
  title: "我的 Paste",
  description: "回看、编辑和删除你创建的 Paste。",
});

async function load() {
  loading.value = true;
  error.value = "";
  try {
    values.value = (await api.listMine()).pastes || [];
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
    error.value = "复制失败，请手动复制分享链接。";
  }
}

async function remove(value: PasteSummary) {
  if (!window.confirm(`删除“${displayTitle(value.title)}”？链接将永久返回已失效。`)) return;
  deleting.value = value.code;
  error.value = "";
  try {
    await api.remove(value.code, value.revision);
    values.value = values.value.filter((candidate) => candidate.code !== value.code);
  } catch (caught) {
    error.value = pasteErrorMessage(caught);
  } finally {
    deleting.value = "";
  }
}

onMounted(load);
</script>

<template>
  <section class="paste-mine paste-container" aria-labelledby="mine-title">
    <header class="paste-mine-heading">
      <div>
        <h1 id="mine-title">我的 Paste</h1>
        <p>这里仅显示登录后创建的内容；匿名 Paste 不会留下管理记录。</p>
      </div>
      <UButton to="/" label="新建 Paste" icon="i-tabler-plus" class="paste-button-primary min-h-11" />
    </header>

    <div class="paste-mine-toolbar">
      <label class="paste-search">
        <UIcon name="i-tabler-search" class="size-4" />
        <span class="sr-only">搜索 Paste</span>
        <input v-model="query" type="search" placeholder="搜索标题、短码、语言或标签">
        <kbd>/</kbd>
      </label>
      <span>{{ filtered.length }} 条记录</span>
    </div>

    <p v-if="error" class="paste-error paste-mine-error" role="alert">{{ error }}</p>

    <div class="paste-ledger" :aria-busy="loading">
      <div class="paste-ledger-head" aria-hidden="true">
        <span>内容</span><span>访问</span><span>更新</span><span>操作</span>
      </div>
      <div v-if="loading" class="paste-mine-state" role="status">
        <UIcon name="i-tabler-loader-2" class="size-5 animate-spin" />正在读取
      </div>
      <div v-else-if="filtered.length === 0" class="paste-mine-state">
        <span class="paste-empty-mark"><UIcon name="i-tabler-file-code" class="size-6" /></span>
        <strong>{{ query ? "没有匹配的 Paste" : "还没有可管理的 Paste" }}</strong>
        <p>{{ query ? "换一个关键词试试。" : "登录状态下创建的 Paste 会出现在这里。" }}</p>
      </div>
      <article v-for="value in filtered" v-else :key="value.code" class="paste-ledger-row">
        <div class="paste-ledger-primary">
          <NuxtLink :to="`/p/${value.code}`">{{ displayTitle(value.title) }}</NuxtLink>
          <div>
            <code>{{ value.code }}</code>
            <span>{{ value.fileCount }} 个文件</span>
            <span>{{ value.primaryLanguage }}</span>
            <span v-for="tag in value.tags.slice(0, 3)" :key="tag">#{{ tag }}</span>
          </div>
        </div>
        <div class="paste-ledger-access">
          <span><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
          <small v-if="value.passwordProtected">另有密码</small>
        </div>
        <time :datetime="value.updatedAt">{{ new Date(value.updatedAt).toLocaleDateString("zh-CN") }}</time>
        <div class="paste-row-actions">
          <UTooltip :text="copied === value.code ? '已复制' : '复制链接'">
            <UButton type="button" :icon="copied === value.code ? 'i-tabler-check' : 'i-tabler-copy'" color="neutral" variant="ghost" aria-label="复制链接" @click="copyLink(value)" />
          </UTooltip>
          <UTooltip text="编辑">
            <UButton :to="`/?edit=${value.code}`" icon="i-tabler-edit" color="neutral" variant="ghost" aria-label="编辑 Paste" />
          </UTooltip>
          <UTooltip text="删除">
            <UButton type="button" icon="i-tabler-trash" color="error" variant="ghost" aria-label="删除 Paste" :loading="deleting === value.code" @click="remove(value)" />
          </UTooltip>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.paste-mine { padding-block: 38px 56px; }
.paste-mine-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 28px; margin-bottom: 28px; }
.paste-mine-heading h1 { margin: 0; font-size: clamp(28px, 4vw, 42px); font-weight: 760; letter-spacing: -.04em; line-height: 1.05; }
.paste-mine-heading p { max-width: 65ch; margin: 9px 0 0; color: var(--paste-ink-soft); font-size: 13px; }
.paste-mine-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 12px; }
.paste-mine-toolbar > span { color: var(--paste-ink-soft); font-size: 12px; white-space: nowrap; }
.paste-search { display: flex; width: min(440px, 100%); min-height: 44px; align-items: center; gap: 9px; border: 1px solid var(--paste-line-strong); border-radius: 10px; padding: 0 11px; background: var(--paste-surface); color: var(--paste-ink-soft); }
.paste-search:focus-within { border-color: var(--paste-blue); box-shadow: 0 0 0 3px color-mix(in srgb, var(--paste-blue) 13%, transparent); }
.paste-search input { min-width: 0; flex: 1; border: 0; background: transparent; color: var(--paste-ink); font-size: 13px; outline: 0; }
.paste-search input::placeholder { color: var(--paste-ink-dim); }
.paste-search kbd { border: 1px solid var(--paste-line); border-radius: 5px; padding: 1px 6px; color: var(--paste-ink-dim); font-family: inherit; font-size: 10px; }
.paste-mine-error { margin: 0 0 12px; }
.paste-ledger { overflow: hidden; border: 1px solid var(--paste-line); border-radius: 14px; background: var(--paste-surface); }
.paste-ledger-head, .paste-ledger-row { display: grid; grid-template-columns: minmax(260px, 1fr) 150px 120px 138px; align-items: center; }
.paste-ledger-head { min-height: 42px; border-bottom: 1px solid var(--paste-line); padding: 0 14px; background: var(--paste-surface-muted); color: var(--paste-ink-soft); font-size: 11px; font-weight: 650; }
.paste-ledger-row { min-height: 78px; border-bottom: 1px solid var(--paste-line); padding: 10px 14px; transition: background 130ms ease; }
.paste-ledger-row:last-child { border-bottom: 0; }
.paste-ledger-row:hover { background: color-mix(in srgb, var(--paste-surface-muted) 62%, transparent); }
.paste-ledger-primary { min-width: 0; }
.paste-ledger-primary > a { color: var(--paste-ink); font-size: 14px; font-weight: 680; text-decoration: none; }
.paste-ledger-primary > a:hover { color: var(--paste-blue); text-decoration: underline; text-underline-offset: 3px; }
.paste-ledger-primary > div { display: flex; min-width: 0; flex-wrap: wrap; gap: 5px 12px; margin-top: 6px; color: var(--paste-ink-soft); font-size: 10px; }
.paste-ledger-primary code { color: var(--paste-blue); font-weight: 680; }
.paste-ledger-access { display: grid; gap: 3px; color: var(--paste-ink-soft); font-size: 11px; }
.paste-ledger-access span { display: flex; align-items: center; gap: 5px; }
.paste-ledger-access small { color: var(--paste-ink-dim); }
.paste-ledger-row time { color: var(--paste-ink-soft); font-size: 11px; }
.paste-row-actions { display: flex; justify-content: flex-end; gap: 2px; }
.paste-mine-state { display: grid; min-height: 270px; place-items: center; align-content: center; gap: 9px; padding: 30px; color: var(--paste-ink-soft); font-size: 13px; text-align: center; }
.paste-mine-state strong { color: var(--paste-ink); font-size: 15px; }
.paste-mine-state p { margin: 0; }
.paste-empty-mark { display: grid; width: 48px; height: 48px; place-items: center; border: 1px solid var(--paste-line); border-radius: 13px; background: var(--paste-surface-muted); }
@media (max-width: 780px) {
  .paste-mine { padding-block: 26px 40px; }
  .paste-mine-heading { align-items: stretch; flex-direction: column; }
  .paste-mine-heading > a { align-self: flex-start; }
  .paste-mine-toolbar { align-items: stretch; flex-direction: column; gap: 8px; }
  .paste-ledger { overflow: visible; border-right: 0; border-left: 0; border-radius: 0; }
  .paste-ledger-head { display: none; }
  .paste-ledger-row { grid-template-columns: 1fr auto; gap: 14px; padding: 15px 12px; }
  .paste-ledger-primary { grid-column: 1 / -1; }
  .paste-ledger-access { grid-column: 1; }
  .paste-ledger-row time { grid-column: 1; }
  .paste-row-actions { grid-column: 2; grid-row: 2 / span 2; align-self: center; }
  .paste-row-actions :deep(button), .paste-row-actions :deep(a) { min-width: 44px; min-height: 44px; }
}
</style>
