<script setup lang="ts">
import type { Paste, PasteFile } from "../../types/paste";
import {
  displayTitle,
  pasteErrorMessage,
  pasteProblemCode,
} from "../../utils/paste";

const route = useRoute();
const api = usePasteApi();
const libraryVersion = useState("paste-library-version", () => 0);
const { loggedIn } = useAuth();
const ownedRevision = ref<number>();
const deleting = ref(false);
const deleteError = ref("");
const deleteOpen = ref(false);
const transfer = usePasteTransfer();
const copyError = ref("");
const showSource = ref(false);
const wrapping = ref(true);
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

const activeFile = computed<PasteFile | undefined>(
  () => value.value?.files[activeIndex.value],
);
const created = computed(() => route.query.created === "1");
const visibilityLabel = computed(() =>
  value.value?.visibility === "private" ? "仅自己" : "链接访问",
);
const shellTitle = computed(() => {
  if (loading.value) return "正在打开片段";
  if (locked.value) return "需要访问密码";
  if (loadError.value || !value.value) return "无法打开这个片段";
  return displayTitle(value.value.title, value.value.files[0]?.path);
});

useSeoMeta({
  title: computed(() =>
    value.value
      ? `${displayTitle(value.value.title, activeFile.value?.path)} · ${siteName.value}`
      : `${code.value} · ${siteName.value}`,
  ),
  description: computed(
    () => value.value?.description || siteDescription.value,
  ),
});

async function load() {
  loading.value = true;
  locked.value = false;
  loadError.value = "";
  accessError.value = "";
  activeIndex.value = 0;
  value.value = undefined;
  ownedRevision.value = undefined;
  const carried = transfer.value[code.value];
  if (carried) {
    value.value = carried;
    delete transfer.value[code.value];
    loading.value = false;

    return;
  }
  try {
    try {
      value.value = (await api.getMine(code.value)).paste;
    } catch (caught) {
      if (!["paste.not_authenticated", "paste.forbidden", "paste.not_found"].includes(pasteProblemCode(caught))) throw caught;
      value.value = (await api.get(code.value)).paste;
    }
  } catch (caught) {
    if (pasteProblemCode(caught) === "paste.password_required")
      locked.value = true;
    else loadError.value = pasteErrorMessage(caught);
  } finally {
    loading.value = false;
  }
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
    copyError.value = "";
    await copyPasteText(text);
    copied.value = key;
    window.setTimeout(() => {
      if (copied.value === key) copied.value = "";
    }, 1800);
  } catch {
    copyError.value = "复制失败，请选择内容后手动复制。";
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
  const fallback = import.meta.client
    ? window.location.href
    : value.value.shareUrl;
  return copy(value.value.shareUrl || fallback, "link");
}

function openInfo() {
  infoOpen.value = true;
}

async function checkOwnership() {
  ownedRevision.value = undefined;
  deleteError.value = "";
  if (!value.value) return;
  const requestedCode = code.value;
  try {
    const owned = (await api.getMine(requestedCode)).paste;
    if (code.value === requestedCode)
      ownedRevision.value = owned.revision;
  } catch {
    // Only the owner endpoint grants management controls.
  }
}

async function removeOwned() {
  if (ownedRevision.value === undefined || deleting.value) return;
  deleting.value = true;
  deleteError.value = "";
  try {
    await api.remove(code.value, ownedRevision.value);
    libraryVersion.value++;
    deleteOpen.value = false;
    await navigateTo("/mine");
  } catch (caught) {
    deleteError.value = pasteErrorMessage(caught);
  } finally {
    deleting.value = false;
  }
}

watch([loggedIn, value], checkOwnership);

function selectAdjacentFile(index: number, direction: -1 | 1) {
  if (!value.value) return;
  const next =
    (index + direction + value.value.files.length) % value.value.files.length;
  activeIndex.value = next;
  nextTick(() =>
    document
      .querySelector<HTMLButtonElement>(`#paste-reader-tab-${next}`)
      ?.focus(),
  );
}

onMounted(load);
watch(code, (next, previous) => {
  if (next !== previous) load();
});

watch(activeIndex, () => {
  showSource.value = false;
  copied.value = "";
});
</script>
<template>
  <PasteWorkspace
    :current-title="value ? shellTitle : ''"
    :current-status="
      value?.visibility === 'private' ? '仅自己可访问' : '链接分享'
    "
  >
    <div
      class="paste-reader-shell h-full flex flex-col min-w-0 overflow-hidden"
      :aria-busy="loading"
    >
      <header
        class="paste-command-row min-h-[88px] [padding:20px_30px] flex items-center gap-5 [border-bottom:1px_solid_var(--paste-line)] max-[1100px]:flex-wrap max-[1100px]:gap-3 max-[1100px]:[padding:18px_22px] max-[760px]:p-4 max-[760px]:min-h-[76px]"
      >
        <h1
          v-if="value"
          id="paste-title"
          class="paste-reader-title text-[21px] font-semibold flex-1 min-w-0 [letter-spacing:-.025em] overflow-hidden text-ellipsis whitespace-nowrap max-[760px]:text-lg"
        >
          {{ shellTitle }}
        </h1>
        <span
          v-else
          class="paste-reader-title text-[21px] font-semibold flex-1 min-w-0 [letter-spacing:-.025em] overflow-hidden text-ellipsis whitespace-nowrap max-[760px]:text-lg"
          >{{ loading ? "正在打开…" : "分享内容" }}</span
        >
        <div v-if="value" class="flex flex-wrap items-center gap-2 max-[760px]:w-full">
          <UButton v-if="ownedRevision !== undefined" :to="`/?edit=${code}`" icon="i-tabler-edit" label="编辑" aria-label="编辑此内容" color="neutral" variant="ghost" :disabled="deleting" />
          <UButton v-if="ownedRevision !== undefined" icon="i-tabler-trash" label="删除" aria-label="删除此链接" color="neutral" variant="ghost" class="hover:text-[var(--paste-red)]" :loading="deleting" @click="deleteError = ''; deleteOpen = true" />
          <UButton
            icon="i-tabler-info-circle"
            label="详情"
            aria-label="查看片段信息"
            color="neutral"
            variant="ghost"
            @click="openInfo"
          />
          <UButton
            :label="copied === 'link' ? '链接已复制' : '复制链接'"
            aria-label="复制分享链接"
            :icon="copied === 'link' ? 'i-tabler-check' : 'i-tabler-link'"
            class="paste-primary bg-[var(--paste-blue)] text-[#17200e] min-h-10 [padding:9px_18px] rounded-[7px] font-[650] whitespace-nowrap justify-center not-dark:text-white hover:bg-[var(--paste-blue-hover)] disabled:[opacity:.35] max-[760px]:min-h-10 max-[760px]:text-xs max-[760px]:[padding:8px_12px]"
            @click="copyShareLink"
          />
        </div>
      </header>
      <div
        v-if="value && (created || route.query.updated === '1')"
        class="paste-reader-success flex items-center gap-3 [padding:12px_30px] bg-[var(--paste-blue-soft)] [border-bottom:1px_solid_var(--paste-line)] text-[var(--paste-ink-soft)] text-[13px] [&_>_span:first-child]:text-[var(--paste-blue)] max-[760px]:[padding:10px_16px] max-[760px]:flex-wrap"
        role="status"
      >
        <UIcon name="i-tabler-circle-check" class="size-5" /><span>{{
          created ? "链接已生成，可以分享了" : "修改已保存，原链接继续有效"
        }}</span
        ><input
          :value="value.shareUrl"
          readonly
          aria-label="分享链接"
          class="paste-reader-link min-w-0 flex-1 border-0 outline-0 bg-transparent text-xs text-[var(--paste-ink-soft)] max-[760px]:basis-full"
          @focus="($event.target as HTMLInputElement).select()"
        />
      </div>
      <p
        v-if="copyError"
        role="alert"
        class="paste-inline-feedback [padding:10px_28px] text-[13px] bg-[var(--paste-red-soft)] text-[var(--paste-red)]"
      >
        {{ copyError }}
      </p>
      <div
        v-if="value"
        class="paste-editor-toolbar min-h-[58px] flex items-stretch justify-between gap-3 [border-bottom:1px_solid_var(--paste-line)] [padding:0_22px] max-[1100px]:flex-wrap max-[1100px]:gap-0 max-[760px]:[padding:0_10px]"
      >
        <nav
          class="paste-file-tabs min-h-[54px] flex flex-1 min-w-0 overflow-x-auto [scrollbar-width:thin] max-[1100px]:basis-full"
          role="tablist"
          aria-label="片段文件"
        >
          <button
            v-for="(file, index) in value.files"
            :id="`paste-reader-tab-${index}`"
            :key="file.path"
            type="button"
            role="tab"
            class="paste-file-tab gap-2 px-3 text-xs relative flex items-center [min-width:132px] [max-width:230px] text-[var(--paste-ink-dim)] [&[data-active=true]]:text-[var(--paste-ink)] [&[data-active=true]]:[box-shadow:inset_0_-2px_var(--paste-blue)] [&[data-dragging=true]]:[opacity:.4] [&[data-drop-position=before]]:[border-left:2px_solid_var(--paste-blue)] [&[data-drop-position=after]]:[border-right:2px_solid_var(--paste-blue)] [&_.paste-file-tab-button]:text-xs [&_.paste-file-tab-button]:min-h-[54px]"
            :data-active="activeIndex === index"
            :aria-selected="activeIndex === index"
            :aria-controls="`paste-reader-panel-${index}`"
            :tabindex="activeIndex === index ? 0 : -1"
            :aria-label="file.path"
            @click="() => { activeIndex = index; }"
            @keydown.left.prevent="selectAdjacentFile(index, -1)"
            @keydown.right.prevent="selectAdjacentFile(index, 1)"
          >
            <UIcon name="i-tabler-file-code" class="size-4 shrink-0" /><span
              class="truncate leading-6 py-px"
              >{{ file.path }}</span
            >
          </button>
        </nav>
        <div
          v-if="activeFile"
          class="paste-editor-tools flex items-center gap-[5px] shrink-0 max-[1100px]:[padding:8px_0] max-[1100px]:[min-height:48px] max-[1100px]:w-full max-[1100px]:justify-end max-[1100px]:[border-top:1px_solid_var(--paste-line)] max-[760px]:justify-start max-[760px]:gap-0.5 max-[760px]:overflow-x-auto max-[760px]:[&_>_*]:shrink-0 max-[760px]:[&_button]:min-h-[38px]"
        >
          <UButton
            v-if="activeFile.language === 'markdown'"
            :label="showSource ? '阅读' : '查看源码'"
            :icon="showSource ? 'i-tabler-book' : 'i-tabler-code'"
            color="neutral"
            variant="ghost"
            @click="() => { showSource = !showSource; }"
          />
          <span class="px-2 text-xs text-muted">{{ activeFile.language }}</span>
          <UButton
            v-if="activeFile.language !== 'markdown' || showSource"
            icon="i-tabler-text-wrap"
            aria-label="自动换行"
            :aria-pressed="wrapping"
            color="neutral"
            :variant="wrapping ? 'soft' : 'ghost'"
            @click="() => { wrapping = !wrapping; }"
          />
          <UButton
            :label="copied === activeFile.path ? '已复制' : '复制源码'"
            :aria-label="`复制 ${activeFile.path}`"
            :icon="
              copied === activeFile.path ? 'i-tabler-check' : 'i-tabler-copy'
            "
            color="neutral"
            variant="ghost"
            @click="copy(activeFile.content, activeFile.path)"
          />
        </div>
      </div>
      <div class="paste-editor-area flex-1 min-h-0 min-w-0 overflow-hidden">
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
          <span
            class="grid size-[38px] place-items-center rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface-muted)] text-[var(--paste-blue)]"
            aria-hidden="true"
            ><UIcon name="i-tabler-lock" class="size-5"
          /></span>
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
            <p
              v-if="accessError"
              class="text-xs leading-[1.55] text-[var(--paste-red)]"
              role="alert"
            >
              {{ accessError }}
            </p>
            <UButton
              type="submit"
              label="打开片段"
              icon="i-tabler-lock-open"
              block
              class="paste-primary bg-[var(--paste-blue)] text-[#17200e] min-h-10 [padding:9px_18px] rounded-[7px] font-[650] whitespace-nowrap justify-center not-dark:text-white hover:bg-[var(--paste-blue-hover)] disabled:[opacity:.35] max-[760px]:min-h-10 max-[760px]:text-xs max-[760px]:[padding:8px_12px]"
              :loading="unlocking"
              :disabled="unlocking || !password"
            />
          </form>
        </div>

        <div
          v-else-if="loadError || !value"
          class="grid size-full place-content-center justify-items-center gap-2.5 p-6 text-center text-xs text-[var(--paste-ink-soft)] [&>h1]:mt-1 [&>h1]:text-xl [&>h1]:font-[730] [&>h1]:tracking-[-0.025em] [&>h1]:text-[var(--paste-ink)] [&>p]:max-w-[48ch] [&>p]:leading-[1.65]"
        >
          <span
            class="grid size-[38px] place-items-center rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface-muted)] text-[var(--paste-red)]"
            aria-hidden="true"
            ><UIcon name="i-tabler-file-off" class="size-5"
          /></span>
          <h1 id="paste-title">无法打开这个片段</h1>
          <p role="alert">{{ loadError || "内容不存在。" }}</p>
          <div
            class="mt-[5px] flex flex-wrap justify-center gap-[7px] max-[720px]:[&>*]:min-h-11"
          >
            <UButton
              type="button"
              label="重新打开"
              icon="i-tabler-refresh"
              color="neutral"
              variant="outline"
              @click="load"
            />
            <UButton
              to="/"
              label="新建内容"
              icon="i-tabler-file-plus"
              color="primary"
            />
          </div>
        </div>

        <section
          v-else-if="activeFile"
          :id="`paste-reader-panel-${activeIndex}`"
          role="tabpanel"
          :aria-labelledby="`paste-reader-tab-${activeIndex}`"
          class="size-full min-h-0 min-w-0 [&_pre]:m-0 [&_pre]:h-full [&_pre]:overflow-auto [&_pre]:p-[18px] [&_pre]:font-mono [&_pre]:text-[13px] [&_pre]:leading-[1.72] [&_pre]:text-[var(--paste-ink)]"
        >
          <div
            v-if="activeFile.language === 'markdown' && !showSource"
            class="paste-document-scroll h-full overflow-auto"
          >
            <div
              class="paste-document [max-width:840px] m-auto [padding:48px_44px_80px] text-[15px] [line-height:1.85] [overflow-wrap:anywhere] max-[760px]:[padding:28px_20px_60px]"
            >
              <LazyContentProse :content="activeFile.content" />
            </div>
          </div>
          <ClientOnly v-else>
            <PasteCodeEditor
              :key="activeFile.path"
              :model-value="activeFile.content"
              :language="activeFile.language"
              :label="`${activeFile.path} 只读代码`"
              :line-wrapping="wrapping"
              read-only
            />
            <template #fallback>
              <pre>{{ activeFile.content }}</pre>
            </template>
          </ClientOnly>
        </section>
      </div>

      <footer
        class="paste-statusbar flex justify-between gap-3 [padding:10px_24px] min-h-[38px] [border-top:1px_solid_var(--paste-line)] text-[var(--paste-ink-dim)] text-[10px] bg-[var(--paste-editor)] max-[760px]:[padding:9px_14px] max-[760px]:[font-size:9px] max-[760px]:gap-2 max-[760px]:[&_>_span:last-child]:gap-2"
        aria-label="片段状态"
      >
        <span
          >{{ value ? visibilityLabel : shellTitle
          }}{{ value?.passwordProtected ? " · 密码保护" : "" }}</span
        ><span v-if="value"
          >只读 · {{ value.files.length }} 个文件 · UTF-8</span
        >
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
              <strong class="text-base font-[730] text-[var(--paste-ink)]">{{
                displayTitle(value.title, value.files[0]?.path)
              }}</strong>
              <p
                v-if="value.description"
                class="mt-[7px] text-xs leading-[1.65] text-[var(--paste-ink-soft)]"
              >
                {{ value.description }}
              </p>
            </div>
            <dl
              class="border-t border-[var(--paste-line)] [&>div]:grid [&>div]:min-h-[42px] [&>div]:grid-cols-[76px_minmax(0,1fr)] [&>div]:items-center [&>div]:gap-3 [&>div]:border-b [&>div]:border-[var(--paste-line)] [&_dd]:m-0 [&_dd]:overflow-wrap-anywhere [&_dd]:text-right [&_dd]:text-xs [&_dd]:text-[var(--paste-ink)] [&_dt]:text-[11px] [&_dt]:text-[var(--paste-ink-dim)]"
            >
              <div>
                <dt>访问</dt>
                <dd>{{ visibilityLabel }}</dd>
              </div>
              <div>
                <dt>文件</dt>
                <dd>{{ value.files.length }} 个</dd>
              </div>
              <div>
                <dt>创建</dt>
                <dd>{{ new Date(value.createdAt).toLocaleString("zh-CN") }}</dd>
              </div>
              <div v-if="value.expiresAt">
                <dt>到期</dt>
                <dd>{{ new Date(value.expiresAt).toLocaleString("zh-CN") }}</dd>
              </div>
            </dl>
            <div
              v-if="value.tags.length"
              class="flex flex-wrap gap-x-3 gap-y-1.5 text-xs text-[var(--paste-blue)]"
              aria-label="标签"
            >
              <span v-for="tag in value.tags" :key="tag">#{{ tag }}</span>
            </div>
          </div>
        </template>
        <template #footer>
          <div
            class="flex w-full justify-end gap-[7px] max-[720px]:flex-col-reverse max-[720px]:[&>*]:min-h-11 max-[720px]:[&>*]:w-full max-[720px]:[&>*]:justify-center"
          >
            <UButton
              v-if="activeFile"
              type="button"
              :label="
                copied === activeFile.path ? '已复制当前' : '复制当前文件'
              "
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

      <PasteDeleteDialog v-model:open="deleteOpen" :titles="[shellTitle]" :busy="deleting" :error="deleteError" @confirm="removeOwned" />
      <p class="sr-only" aria-live="polite">{{ copied ? "内容已复制" : "" }}</p>
    </div>
  </PasteWorkspace>
</template>
