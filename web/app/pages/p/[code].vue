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
  <section
    class="min-h-dvh bg-[var(--paste-editor)]"
    :aria-label="shellTitle"
    :aria-busy="loading"
  >
    <div
      class="paste-reader-shell grid h-dvh min-h-[360px] w-full min-w-0 grid-rows-[40px_minmax(0,1fr)_28px] overflow-hidden bg-[var(--paste-editor)] max-[720px]:min-h-[320px] max-[720px]:grid-rows-[44px_minmax(0,1fr)_30px]"
    >
      <header
        class="flex min-w-0 items-stretch border-b border-[var(--paste-line)] bg-[var(--paste-chrome)]"
        aria-label="代码片段只读工具栏"
      >
        <NuxtLink
          to="/"
          class="flex shrink-0 items-center gap-[7px] border-r border-[var(--paste-line)] px-[11px] text-[var(--paste-ink)] no-underline max-[720px]:px-1.5"
          :aria-label="`${siteName} 首页`"
        >
          <span
            class="grid size-6 place-items-center rounded-md border border-[color-mix(in_srgb,var(--paste-blue)_28%,var(--paste-line))] bg-[var(--paste-blue-soft)] text-[var(--paste-blue)]"
            aria-hidden="true"
          >
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span
            class="text-[13px] font-[730] tracking-[-0.025em] max-[720px]:hidden"
            >{{ siteName }}</span
          >
        </NuxtLink>

        <nav
          v-if="value"
          class="flex min-w-0 flex-1 overflow-x-auto overflow-y-hidden [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
          role="tablist"
          aria-label="片段文件"
        >
          <button
            v-for="(file, index) in value.files"
            :id="`paste-reader-tab-${index}`"
            :key="`${file.path}-${index}`"
            type="button"
            role="tab"
            class="flex h-full min-w-[108px] max-w-[220px] basis-44 items-center gap-[7px] overflow-hidden border-0 border-r border-[var(--paste-line)] bg-transparent px-[11px] text-left text-xs text-[var(--paste-ink-soft)] hover:bg-[color-mix(in_srgb,var(--paste-editor-active)_48%,transparent)] hover:text-[var(--paste-ink)] data-[active=true]:bg-[var(--paste-editor)] data-[active=true]:text-[var(--paste-ink)] data-[active=true]:[box-shadow:inset_0_2px_var(--paste-blue)] max-[720px]:min-w-[88px] max-[720px]:basis-[124px] max-[720px]:px-[9px] [&>span]:truncate"
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
        <div
          v-else
          class="flex min-w-0 flex-1 overflow-x-auto overflow-y-hidden [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
          aria-hidden="true"
        >
          <div
            class="flex h-full min-w-[108px] max-w-[220px] basis-44 cursor-default items-center gap-[7px] overflow-hidden border-r border-[var(--paste-line)] bg-[var(--paste-editor)] px-[11px] text-xs text-[var(--paste-ink)] [box-shadow:inset_0_2px_var(--paste-blue)] max-[720px]:min-w-[88px] max-[720px]:basis-[124px] max-[720px]:px-[9px] [&>span]:truncate"
            data-active="true"
          >
            <UIcon :name="locked ? 'i-tabler-lock' : loadError ? 'i-tabler-file-off' : 'i-tabler-file-code'" class="size-4 shrink-0" />
            <span>{{ locked ? "受保护的片段" : loadError ? "无法打开" : code }}</span>
          </div>
        </div>

        <div
          class="flex shrink-0 items-center gap-0.5 bg-[var(--paste-chrome)] px-[5px] max-[720px]:px-1"
        >
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
              class="max-[720px]:hidden"
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
          <span
            v-if="value"
            class="mx-[3px] h-5 w-px bg-[var(--paste-line)] max-[720px]:hidden"
            aria-hidden="true"
          />
          <UTooltip text="我的片段">
            <UButton
              to="/mine"
              icon="i-tabler-folders"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              class="max-[720px]:hidden"
              aria-label="我的片段"
            />
          </UTooltip>
          <UTooltip text="切换颜色模式">
            <UColorModeButton color="neutral" variant="ghost" size="sm" aria-label="切换颜色模式" />
          </UTooltip>
          <div
            class="grid size-8 place-items-center [&_button]:size-[30px] [&_button]:min-h-[30px] [&_button]:min-w-[30px] [&_button]:p-0 [&_[data-slot=label]]:sr-only"
          >
            <ConsumerAccountControl :context-actions="accountActions" trigger-mode="collapsed" />
          </div>
        </div>
      </header>

      <div class="min-h-0 min-w-0 overflow-hidden bg-[var(--paste-editor)]">
        <div
          v-if="loading"
          class="grid size-full place-content-center justify-items-center gap-2.5 p-6 text-center text-xs text-[var(--paste-ink-soft)]"
          role="status"
        >
          <UIcon name="i-tabler-loader-2" class="size-6 animate-spin" />
          <span>正在打开片段</span>
        </div>

        <div
          v-else-if="locked"
          class="grid size-full place-content-center justify-items-center gap-2.5 p-6 text-center text-xs text-[var(--paste-ink-soft)] [&>h1]:mt-1 [&>h1]:text-xl [&>h1]:font-[730] [&>h1]:tracking-[-0.025em] [&>h1]:text-[var(--paste-ink)] [&>p]:max-w-[48ch] [&>p]:leading-[1.65]"
        >
          <span class="grid size-[38px] place-items-center rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface-muted)] text-[var(--paste-blue)]" aria-hidden="true"><UIcon name="i-tabler-lock" class="size-5" /></span>
          <h1 id="paste-title">需要访问密码</h1>
          <p>密码只用于这次打开，不会写入地址栏。</p>
          <form
            class="mt-2.5 grid w-[min(360px,calc(100vw-32px))] gap-2.5 text-left"
            @submit.prevent="unlock"
          >
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
            <p v-if="accessError" class="text-xs leading-[1.55] text-[var(--paste-red)]" role="alert">{{ accessError }}</p>
            <UButton
              type="submit"
              label="打开片段"
              icon="i-tabler-lock-open"
              block
              class="min-h-11 border-transparent bg-[var(--paste-blue)] text-white hover:bg-[var(--paste-blue-hover)]"
              :loading="unlocking"
              :disabled="unlocking || !password"
            />
          </form>
        </div>

        <div
          v-else-if="loadError || !value"
          class="grid size-full place-content-center justify-items-center gap-2.5 p-6 text-center text-xs text-[var(--paste-ink-soft)] [&>h1]:mt-1 [&>h1]:text-xl [&>h1]:font-[730] [&>h1]:tracking-[-0.025em] [&>h1]:text-[var(--paste-ink)] [&>p]:max-w-[48ch] [&>p]:leading-[1.65]"
        >
          <span class="grid size-[38px] place-items-center rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface-muted)] text-[var(--paste-red)]" aria-hidden="true"><UIcon name="i-tabler-file-off" class="size-5" /></span>
          <h1 id="paste-title">无法打开这个片段</h1>
          <p role="alert">{{ loadError || "内容不存在。" }}</p>
          <div class="mt-[5px] flex flex-wrap justify-center gap-[7px] max-[720px]:[&>*]:min-h-11">
            <UButton type="button" label="重新打开" icon="i-tabler-refresh" color="neutral" variant="outline" @click="load" />
            <UButton to="/" label="新建片段" icon="i-tabler-file-plus" color="primary" />
          </div>
        </div>

        <section
          v-else-if="activeFile"
          :id="`paste-reader-panel-${activeIndex}`"
          role="tabpanel"
          :aria-labelledby="`paste-reader-tab-${activeIndex}`"
          class="size-full min-h-0 min-w-0 [&_pre]:m-0 [&_pre]:h-full [&_pre]:overflow-auto [&_pre]:p-[18px] [&_pre]:font-mono [&_pre]:text-[13px] [&_pre]:leading-[1.72] [&_pre]:text-[var(--paste-ink)]"
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

      <footer
        class="paste-reader-statusbar flex min-w-0 items-center justify-between gap-4 border-t border-[var(--paste-status-border)] bg-[var(--paste-status)] px-2.5 pb-[env(safe-area-inset-bottom)] font-mono text-xs text-[var(--paste-status-text-muted)]"
        aria-label="片段状态"
      >
        <span class="flex min-w-0 items-center gap-2 text-[var(--paste-status-text)]">
          <span
            class="size-[7px] shrink-0 rounded-full bg-[var(--paste-green)] data-[state=loading]:bg-[var(--paste-status-text-muted)] data-[state=loading]:[animation:paste-reader-pulse_1.1s_ease-in-out_infinite_alternate] data-[state=locked]:bg-[var(--paste-status-error)] data-[state=error]:bg-[var(--paste-status-error)] motion-reduce:animate-none"
            :data-state="shellState"
            aria-hidden="true"
          />
          <h1 v-if="value" id="paste-title" class="truncate font-[inherit] font-[650]">{{ shellTitle }}</h1>
          <span v-else class="truncate font-[650]">{{ shellTitle }}</span>
        </span>
        <span v-if="value && activeFile" class="flex shrink-0 items-center justify-end gap-2 whitespace-nowrap max-[720px]:gap-[7px]">
          <span class="max-[720px]:hidden">{{ value.code }}</span>
          <span class="max-[720px]:hidden">{{ visibilityLabel }}</span>
          <span>{{ activeFile.language }}</span>
          <span>{{ activeIndex + 1 }}/{{ value.files.length }} 文件</span>
        </span>
        <span v-else class="flex shrink-0 items-center justify-end gap-2 whitespace-nowrap">{{ code }}</span>
      </footer>

      <USlideover
        v-if="value"
        v-model:open="infoOpen"
        title="片段信息"
        :description="value.code"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div class="grid gap-5">
            <div>
              <strong class="text-base font-[730] text-[var(--paste-ink)]">{{ displayTitle(value.title, value.files[0]?.path) }}</strong>
              <p v-if="value.description" class="mt-[7px] text-xs leading-[1.65] text-[var(--paste-ink-soft)]">{{ value.description }}</p>
            </div>
            <dl class="border-t border-[var(--paste-line)] [&>div]:grid [&>div]:min-h-[42px] [&>div]:grid-cols-[76px_minmax(0,1fr)] [&>div]:items-center [&>div]:gap-3 [&>div]:border-b [&>div]:border-[var(--paste-line)] [&_dd]:m-0 [&_dd]:overflow-wrap-anywhere [&_dd]:text-right [&_dd]:text-xs [&_dd]:text-[var(--paste-ink)] [&_dt]:text-[11px] [&_dt]:text-[var(--paste-ink-dim)]">
              <div><dt>访问</dt><dd>{{ visibilityLabel }}</dd></div>
              <div><dt>文件</dt><dd>{{ value.files.length }} 个</dd></div>
              <div><dt>创建</dt><dd>{{ new Date(value.createdAt).toLocaleString("zh-CN") }}</dd></div>
              <div v-if="value.expiresAt"><dt>到期</dt><dd>{{ new Date(value.expiresAt).toLocaleString("zh-CN") }}</dd></div>
            </dl>
            <div v-if="value.tags.length" class="flex flex-wrap gap-x-3 gap-y-1.5 text-xs text-[var(--paste-blue)]" aria-label="标签">
              <span v-for="tag in value.tags" :key="tag">#{{ tag }}</span>
            </div>
          </div>
        </template>
        <template #footer>
          <div class="flex w-full justify-end gap-[7px] max-[720px]:flex-col-reverse max-[720px]:[&>*]:min-h-11 max-[720px]:[&>*]:w-full max-[720px]:[&>*]:justify-center">
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
