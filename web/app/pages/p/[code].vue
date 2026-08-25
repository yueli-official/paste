<!--
THESIS: A shared Paste opens as the editor's read-only mode, never as an article wrapped around a code card.
OWN-WORLD: Cool porcelain and ink-navy code floors inside flat mineral-blue, VS Code-derived application chrome.
STORY: choose a file, read the code, copy what matters, and inspect access facts without leaving the editor context.
FIRST VIEWPORT: one 40px product-and-file bar, a viewport-filling code surface, and one 28px document status bar.
FORM: the established editor shell in recipient mode; metadata and secondary copy actions stay in a transient information rail.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";
import type { Paste, PasteFile } from "../../types/paste";
import { displayTitle, pasteErrorMessage, pasteProblemCode } from "../../utils/paste";

const route = useRoute();
const api = usePasteApi();
const transfer = usePasteTransfer();
const toast = useToast();
const { isAdmin } = useAuth();
const { siteName, siteDescription } = useSiteSettings();
const code = computed(() => String(route.params.code || ""));
const value = ref<Paste>();
const activeIndex = ref(0);
const password = ref("");
const loading = ref(true);
const unlocking = ref(false);
const locked = ref(false);
const loadError = ref("");
const accessError = ref("");
const copied = ref("");
const infoOpen = ref(false);
const createdNotified = ref(false);

const accountActions = computed<readonly AccountMenuAction[]>(() => [
  { label: "我的片段", icon: "i-tabler-folders", to: "/mine" },
  ...(isAdmin.value
    ? [{ label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" }]
    : []),
]);
const activeFile = computed<PasteFile | undefined>(() => value.value?.files[activeIndex.value]);
const created = computed(() => route.query.created === "1");
const visibilityLabel = computed(() => value.value?.visibility === "private" ? "仅自己" : "链接访问");
const shellState = computed(() => {
  if (loading.value) return "loading";
  if (locked.value) return "locked";
  if (loadError.value || !value.value) return "error";
  return created.value ? "created" : "ready";
});
const shellTitle = computed(() => {
  if (loading.value) return "正在打开片段";
  if (locked.value) return "需要访问密码";
  if (loadError.value || !value.value) return "无法打开这个片段";
  return displayTitle(value.value.title, value.value.files[0]?.path);
});

useSeoMeta({
  title: computed(() => value.value ? `${displayTitle(value.value.title, activeFile.value?.path)} · ${siteName.value}` : `${code.value} · ${siteName.value}`),
  description: computed(() => value.value?.description || siteDescription.value),
});

async function load() {
  loading.value = true;
  locked.value = false;
  loadError.value = "";
  accessError.value = "";
  activeIndex.value = 0;
  value.value = undefined;
  const carried = transfer.value[code.value];
  if (carried) {
    value.value = carried;
    loading.value = false;
    notifyCreated();
    return;
  }
  try {
    value.value = (await api.get(code.value)).paste;
    notifyCreated();
  } catch (caught) {
    if (pasteProblemCode(caught) === "paste.password_required") locked.value = true;
    else loadError.value = pasteErrorMessage(caught);
  } finally {
    loading.value = false;
  }
}

function notifyCreated() {
  if (!created.value || createdNotified.value) return;
  createdNotified.value = true;
  toast.add({
    title: "代码片段已创建",
    description: "链接已经可以复制和分享。",
    color: "success",
  });
}

async function unlock() {
  if (!password.value) return;
  unlocking.value = true;
  accessError.value = "";
  try {
    value.value = (await api.access(code.value, password.value)).paste;
    locked.value = false;
    activeIndex.value = 0;
    password.value = "";
  } catch (caught) {
    accessError.value = pasteErrorMessage(caught);
  } finally {
    unlocking.value = false;
  }
}

async function copy(text: string, key: string) {
  try {
    await navigator.clipboard.writeText(text);
    copied.value = key;
    window.setTimeout(() => {
      if (copied.value === key) copied.value = "";
    }, 1800);
  } catch {
    toast.add({
      title: "复制失败",
      description: "请选中内容后手动复制。",
      color: "error",
    });
  }
}

function copyAll() {
  if (!value.value) return;
  const content = value.value.files
    .map((file) => `===== ${file.path} =====\n${file.content}`)
    .join("\n\n");
  return copy(content, "all");
}

function copyShareLink() {
  if (!value.value) return;
  const fallback = import.meta.client ? window.location.href : value.value.shareUrl;
  return copy(value.value.shareUrl || fallback, "link");
}

function openInfo() {
  infoOpen.value = true;
}

function selectAdjacentFile(index: number, direction: -1 | 1) {
  if (!value.value) return;
  const next = (index + direction + value.value.files.length) % value.value.files.length;
  activeIndex.value = next;
  nextTick(() => document.querySelector<HTMLButtonElement>(`#paste-reader-tab-${next}`)?.focus());
}

onMounted(load);
watch(code, (next, previous) => {
  if (next !== previous) load();
});
</script>

<template>
  <section class="paste-reader" :aria-label="shellTitle" :aria-busy="loading">
    <div class="paste-reader-shell">
      <header class="paste-reader-chrome" aria-label="代码片段只读工具栏">
        <NuxtLink to="/" class="paste-reader-brand" :aria-label="`${siteName} 首页`">
          <span class="paste-reader-brand-mark" aria-hidden="true">
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span class="paste-reader-brand-label">{{ siteName }}</span>
        </NuxtLink>

        <nav v-if="value" class="paste-reader-tabs" role="tablist" aria-label="片段文件">
          <button
            v-for="(file, index) in value.files"
            :id="`paste-reader-tab-${index}`"
            :key="`${file.path}-${index}`"
            type="button"
            role="tab"
            class="paste-reader-tab"
            :data-active="activeIndex === index"
            :aria-selected="activeIndex === index"
            :aria-controls="`paste-reader-panel-${index}`"
            :tabindex="activeIndex === index ? 0 : -1"
            :aria-label="file.path"
            @click="activeIndex = index"
            @keydown.left.prevent="selectAdjacentFile(index, -1)"
            @keydown.right.prevent="selectAdjacentFile(index, 1)"
          >
            <UIcon name="i-tabler-file-code" class="size-4 shrink-0" />
            <span>{{ file.path }}</span>
          </button>
        </nav>
        <div v-else class="paste-reader-tabs" aria-hidden="true">
          <div class="paste-reader-tab paste-reader-tab-placeholder" data-active="true">
            <UIcon :name="locked ? 'i-tabler-lock' : loadError ? 'i-tabler-file-off' : 'i-tabler-file-code'" class="size-4 shrink-0" />
            <span>{{ locked ? "受保护的片段" : loadError ? "无法打开" : code }}</span>
          </div>
        </div>

        <div class="paste-reader-actions">
          <UTooltip v-if="value" text="片段信息">
            <UButton
              type="button"
              icon="i-tabler-info-circle"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              aria-label="查看片段信息"
              @click="openInfo"
            />
          </UTooltip>
          <UTooltip v-if="activeFile" :text="copied === activeFile.path ? '已复制当前文件' : '复制当前文件'">
            <UButton
              type="button"
              :icon="copied === activeFile.path ? 'i-tabler-check' : 'i-tabler-copy'"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              class="paste-reader-copy-current"
              :aria-label="`复制 ${activeFile.path}`"
              @click="copy(activeFile.content, activeFile.path)"
            />
          </UTooltip>
          <UTooltip v-if="value" :text="copied === 'link' ? '链接已复制' : '复制分享链接'">
            <UButton
              type="button"
              :icon="copied === 'link' ? 'i-tabler-check' : 'i-tabler-share-3'"
              color="primary"
              variant="ghost"
              size="sm"
              square
              aria-label="复制分享链接"
              @click="copyShareLink"
            />
          </UTooltip>
          <span v-if="value" class="paste-reader-action-divider" aria-hidden="true" />
          <UTooltip text="我的片段">
            <UButton
              to="/mine"
              icon="i-tabler-folders"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              class="paste-reader-history"
              aria-label="我的片段"
            />
          </UTooltip>
          <UTooltip text="切换颜色模式">
            <UColorModeButton color="neutral" variant="ghost" size="sm" aria-label="切换颜色模式" />
          </UTooltip>
          <div class="paste-reader-account">
            <ConsumerAccountControl :context-actions="accountActions" trigger-mode="collapsed" />
          </div>
        </div>
      </header>

      <div class="paste-reader-floor">
        <div v-if="loading" class="paste-reader-state" role="status">
          <UIcon name="i-tabler-loader-2" class="size-6 animate-spin" />
          <span>正在打开片段</span>
        </div>

        <div v-else-if="locked" class="paste-reader-state paste-reader-gate">
          <span class="paste-reader-state-mark" aria-hidden="true"><UIcon name="i-tabler-lock" class="size-5" /></span>
          <h1 id="paste-title">需要访问密码</h1>
          <p>密码只用于这次打开，不会写入地址栏。</p>
          <form class="paste-reader-gate-form" @submit.prevent="unlock">
            <UFormField label="访问密码" class="w-full">
              <UInput
                id="paste-access-password"
                v-model="password"
                type="password"
                autocomplete="current-password"
                autofocus
                icon="i-tabler-key"
                aria-label="访问密码"
                class="w-full"
              />
            </UFormField>
            <p v-if="accessError" class="paste-reader-error" role="alert">{{ accessError }}</p>
            <UButton
              type="submit"
              label="打开片段"
              icon="i-tabler-lock-open"
              block
              class="paste-button-primary min-h-11"
              :loading="unlocking"
              :disabled="unlocking || !password"
            />
          </form>
        </div>

        <div v-else-if="loadError || !value" class="paste-reader-state paste-reader-failure">
          <span class="paste-reader-state-mark" aria-hidden="true"><UIcon name="i-tabler-file-off" class="size-5" /></span>
          <h1 id="paste-title">无法打开这个片段</h1>
          <p role="alert">{{ loadError || "内容不存在。" }}</p>
          <div class="paste-reader-state-actions">
            <UButton type="button" label="重新打开" icon="i-tabler-refresh" color="neutral" variant="outline" @click="load" />
            <UButton to="/" label="新建片段" icon="i-tabler-file-plus" color="primary" />
          </div>
        </div>

        <section
          v-else-if="activeFile"
          :id="`paste-reader-panel-${activeIndex}`"
          role="tabpanel"
          :aria-labelledby="`paste-reader-tab-${activeIndex}`"
          class="paste-reader-code"
        >
          <ClientOnly>
            <PasteCodeEditor
              :key="activeFile.path"
              :model-value="activeFile.content"
              :language="activeFile.language"
              :label="`${activeFile.path} 只读代码`"
              read-only
            />
            <template #fallback><pre>{{ activeFile.content }}</pre></template>
          </ClientOnly>
        </section>
      </div>

      <footer class="paste-reader-statusbar" aria-label="片段状态">
        <span class="paste-reader-status-primary">
          <span class="paste-reader-status-dot" :data-state="shellState" aria-hidden="true" />
          <h1 v-if="value" id="paste-title">{{ shellTitle }}</h1>
          <span v-else>{{ shellTitle }}</span>
        </span>
        <span v-if="value && activeFile" class="paste-reader-status-meta">
          <span class="paste-reader-status-code">{{ value.code }}</span>
          <span class="paste-reader-status-access">{{ visibilityLabel }}</span>
          <span>{{ activeFile.language }}</span>
          <span>{{ activeIndex + 1 }}/{{ value.files.length }} 文件</span>
        </span>
        <span v-else class="paste-reader-status-meta">{{ code }}</span>
      </footer>

      <USlideover
        v-if="value"
        v-model:open="infoOpen"
        title="片段信息"
        :description="value.code"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div class="paste-reader-info">
            <div>
              <strong>{{ displayTitle(value.title, value.files[0]?.path) }}</strong>
              <p v-if="value.description">{{ value.description }}</p>
            </div>
            <dl>
              <div><dt>访问</dt><dd>{{ visibilityLabel }}</dd></div>
              <div><dt>文件</dt><dd>{{ value.files.length }} 个</dd></div>
              <div><dt>创建</dt><dd>{{ new Date(value.createdAt).toLocaleString("zh-CN") }}</dd></div>
              <div v-if="value.expiresAt"><dt>到期</dt><dd>{{ new Date(value.expiresAt).toLocaleString("zh-CN") }}</dd></div>
            </dl>
            <div v-if="value.tags.length" class="paste-reader-tags" aria-label="标签">
              <span v-for="tag in value.tags" :key="tag">#{{ tag }}</span>
            </div>
          </div>
        </template>
        <template #footer>
          <div class="paste-reader-info-actions">
            <UButton
              v-if="activeFile"
              type="button"
              :label="copied === activeFile.path ? '已复制当前' : '复制当前文件'"
              icon="i-tabler-copy"
              color="neutral"
              variant="outline"
              @click="copy(activeFile.content, activeFile.path)"
            />
            <UButton
              v-if="value.files.length > 1"
              type="button"
              :label="copied === 'all' ? '已复制全部' : '复制全部文件'"
              icon="i-tabler-copy"
              color="neutral"
              variant="outline"
              @click="copyAll"
            />
            <UButton
              type="button"
              :label="copied === 'link' ? '链接已复制' : '复制分享链接'"
              icon="i-tabler-share-3"
              color="primary"
              @click="copyShareLink"
            />
          </div>
        </template>
      </USlideover>
      <p class="sr-only" aria-live="polite">{{ copied ? "内容已复制" : "" }}</p>
    </div>
  </section>
</template>

<style scoped>
.paste-reader { min-height: 100dvh; background: var(--paste-editor); }
.paste-reader-shell { display: grid; width: 100%; height: 100dvh; min-width: 0; min-height: 360px; grid-template-rows: 40px minmax(0, 1fr) 28px; overflow: hidden; background: var(--paste-editor); }
.paste-reader-chrome { display: flex; min-width: 0; align-items: stretch; border-bottom: 1px solid var(--paste-line); background: var(--paste-chrome); }
.paste-reader-brand { display: flex; flex: none; align-items: center; gap: 7px; border-right: 1px solid var(--paste-line); padding: 0 11px; color: var(--paste-ink); text-decoration: none; }
.paste-reader-brand-mark { display: grid; width: 24px; height: 24px; place-items: center; border: 1px solid color-mix(in srgb, var(--paste-blue) 28%, var(--paste-line)); border-radius: 6px; background: var(--paste-blue-soft); color: var(--paste-blue); }
.paste-reader-brand-label { font-size: 13px; font-weight: 730; letter-spacing: -.025em; }
.paste-reader-tabs { display: flex; min-width: 0; flex: 1; overflow-x: auto; overflow-y: hidden; scrollbar-width: none; }
.paste-reader-tabs::-webkit-scrollbar { display: none; }
.paste-reader-tab { display: flex; min-width: 108px; max-width: 220px; height: 100%; flex: 0 1 176px; align-items: center; gap: 7px; overflow: hidden; border: 0; border-right: 1px solid var(--paste-line); padding: 0 11px; background: transparent; color: var(--paste-ink-soft); font: inherit; font-size: 12px; text-align: left; }
.paste-reader-tab > span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.paste-reader-tab:hover { background: color-mix(in srgb, var(--paste-editor-active) 48%, transparent); color: var(--paste-ink); }
.paste-reader-tab[data-active="true"] { background: var(--paste-editor); box-shadow: inset 0 2px var(--paste-blue); color: var(--paste-ink); }
.paste-reader-tab-placeholder { cursor: default; }
.paste-reader-actions { display: flex; flex: none; align-items: center; gap: 2px; padding: 0 5px; background: var(--paste-chrome); }
.paste-reader-action-divider { width: 1px; height: 20px; margin-inline: 3px; background: var(--paste-line); }
.paste-reader-account { display: grid; width: 32px; height: 32px; place-items: center; }
.paste-reader-account :deep(button) { width: 30px; min-width: 30px; height: 30px; min-height: 30px; padding: 0; }
.paste-reader-account :deep([data-slot="label"]) { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.paste-reader-floor { min-width: 0; min-height: 0; overflow: hidden; background: var(--paste-editor); }
.paste-reader-code { width: 100%; height: 100%; min-width: 0; min-height: 0; }
.paste-reader-code pre { height: 100%; overflow: auto; margin: 0; padding: 18px; color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; font-size: 13px; line-height: 1.72; }
.paste-reader-state { display: grid; width: 100%; height: 100%; place-content: center; justify-items: center; gap: 10px; padding: 24px; color: var(--paste-ink-soft); font-size: 12px; text-align: center; }
.paste-reader-state-mark { display: grid; width: 38px; height: 38px; place-items: center; border: 1px solid var(--paste-line); border-radius: 8px; background: var(--paste-surface-muted); color: var(--paste-blue); }
.paste-reader-state h1 { margin: 4px 0 0; color: var(--paste-ink); font-size: 20px; font-weight: 730; letter-spacing: -.025em; }
.paste-reader-state > p { max-width: 48ch; margin: 0; line-height: 1.65; }
.paste-reader-gate-form { display: grid; width: min(360px, calc(100vw - 32px)); gap: 10px; margin-top: 10px; text-align: left; }
.paste-reader-error { margin: 0; color: var(--paste-red); font-size: 12px; line-height: 1.55; }
.paste-reader-failure .paste-reader-state-mark { color: var(--paste-red); }
.paste-reader-state-actions { display: flex; flex-wrap: wrap; justify-content: center; gap: 7px; margin-top: 5px; }
.paste-reader-statusbar { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 16px; border-top: 1px solid var(--paste-status-border); padding: 0 10px; padding-bottom: env(safe-area-inset-bottom); background: var(--paste-status); color: var(--paste-status-text-muted); font-family: "SFMono-Regular", Consolas, monospace; font-size: 12px; }
.paste-reader-status-primary, .paste-reader-status-meta { display: flex; min-width: 0; align-items: center; gap: 8px; }
.paste-reader-status-primary { color: var(--paste-status-text); }
.paste-reader-status-primary h1, .paste-reader-status-primary > span:last-child { overflow: hidden; margin: 0; font: inherit; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.paste-reader-status-dot { width: 7px; height: 7px; flex: none; border-radius: 50%; background: var(--paste-green); }
.paste-reader-status-dot[data-state="loading"] { background: var(--paste-status-text-muted); animation: paste-reader-pulse 1.1s ease-in-out infinite alternate; }
.paste-reader-status-dot[data-state="locked"] { background: var(--paste-status-error); }
.paste-reader-status-dot[data-state="error"] { background: var(--paste-status-error); }
.paste-reader-status-meta { flex: none; justify-content: flex-end; white-space: nowrap; }
.paste-reader-info { display: grid; gap: 20px; }
.paste-reader-info strong { color: var(--paste-ink); font-size: 16px; font-weight: 730; }
.paste-reader-info p { margin: 7px 0 0; color: var(--paste-ink-soft); font-size: 12px; line-height: 1.65; }
.paste-reader-info dl { margin: 0; border-top: 1px solid var(--paste-line); }
.paste-reader-info dl > div { display: grid; min-height: 42px; grid-template-columns: 76px minmax(0, 1fr); align-items: center; gap: 12px; border-bottom: 1px solid var(--paste-line); }
.paste-reader-info dt { color: var(--paste-ink-dim); font-size: 11px; }
.paste-reader-info dd { overflow-wrap: anywhere; margin: 0; color: var(--paste-ink); font-size: 12px; text-align: right; }
.paste-reader-tags { display: flex; flex-wrap: wrap; gap: 6px 12px; color: var(--paste-blue); font-size: 12px; }
.paste-reader-info-actions { display: flex; width: 100%; justify-content: flex-end; gap: 7px; }
@keyframes paste-reader-pulse { to { opacity: .45; } }
@media (prefers-reduced-motion: reduce) { .paste-reader-status-dot[data-state="loading"] { animation: none; } }
@media (max-width: 720px) {
  .paste-reader, .paste-reader-shell { height: 100dvh; min-height: 320px; }
  .paste-reader-shell { grid-template-rows: 44px minmax(0, 1fr) 30px; }
  .paste-reader-brand { padding-inline: 6px; }
  .paste-reader-brand-label { display: none; }
  .paste-reader-tab { min-width: 88px; flex-basis: 124px; padding-inline: 9px; }
  .paste-reader-actions { padding-inline: 4px; }
  .paste-reader-action-divider, .paste-reader-copy-current, .paste-reader-history { display: none; }
  .paste-reader-account, .paste-reader-account :deep(button) { width: 36px; min-width: 36px; height: 36px; min-height: 36px; }
  .paste-reader-status-code, .paste-reader-status-access { display: none; }
  .paste-reader-status-meta { gap: 7px; }
  .paste-reader-state-actions > * { min-height: 44px; }
  .paste-reader-info-actions { flex-direction: column-reverse; }
  .paste-reader-info-actions > * { width: 100%; min-height: 44px; justify-content: center; }
}
</style>
