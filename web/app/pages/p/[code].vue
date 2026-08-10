<script setup lang="ts">
import type { Paste, PasteFile } from "../../types/paste";
import { displayTitle, pasteErrorMessage, pasteProblemCode } from "../../utils/paste";

const route = useRoute();
const api = usePasteApi();
const transfer = usePasteTransfer();
const code = computed(() => String(route.params.code || ""));
const value = ref<Paste>();
const activeIndex = ref(0);
const password = ref("");
const loading = ref(true);
const unlocking = ref(false);
const locked = ref(false);
const error = ref("");
const copied = ref("");

const activeFile = computed<PasteFile | undefined>(() => value.value?.files[activeIndex.value]);

useSeoMeta({
  title: computed(() => value.value ? displayTitle(value.value.title, activeFile.value?.path) : `Paste ${code.value}`),
  description: computed(() => value.value?.description || "月离 Paste 代码分享"),
});

async function load() {
  loading.value = true;
  error.value = "";
  const carried = transfer.value[code.value];
  if (carried) {
    value.value = carried;
    loading.value = false;
    return;
  }
  try {
    value.value = (await api.get(code.value)).paste;
  } catch (caught) {
    if (pasteProblemCode(caught) === "paste.password_required") locked.value = true;
    else error.value = pasteErrorMessage(caught);
  } finally {
    loading.value = false;
  }
}

async function unlock() {
  if (!password.value) return;
  unlocking.value = true;
  error.value = "";
  try {
    value.value = (await api.access(code.value, password.value)).paste;
    locked.value = false;
    password.value = "";
  } catch (caught) {
    error.value = pasteErrorMessage(caught);
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
    error.value = "复制失败，请选中内容后手动复制。";
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

onMounted(load);
</script>

<template>
  <section class="paste-view paste-container" aria-labelledby="paste-title">
    <div v-if="loading" class="paste-view-state" role="status">
      <UIcon name="i-tabler-loader-2" class="size-6 animate-spin" />
      <span>正在打开 Paste</span>
    </div>

    <div v-else-if="locked" class="paste-access-gate">
      <div class="paste-gate-mark"><UIcon name="i-tabler-lock" class="size-6" /></div>
      <h1 id="paste-title">需要访问密码</h1>
      <p>密码只在本次请求中发送，不会写入地址栏。</p>
      <form @submit.prevent="unlock">
        <label class="paste-field-label" for="paste-access-password">访问密码</label>
        <input id="paste-access-password" v-model="password" class="paste-input" type="password" autocomplete="current-password" autofocus>
        <p v-if="error" class="paste-error" role="alert">{{ error }}</p>
        <UButton type="submit" label="打开 Paste" icon="i-tabler-lock-open" block size="lg" class="paste-button-primary min-h-12" :loading="unlocking" />
      </form>
    </div>

    <div v-else-if="error || !value" class="paste-view-state paste-view-error">
      <UIcon name="i-tabler-file-off" class="size-7" />
      <h1 id="paste-title">无法打开这个 Paste</h1>
      <p role="alert">{{ error || "内容不存在。" }}</p>
      <UButton to="/" label="新建 Paste" icon="i-tabler-plus" color="neutral" variant="outline" />
    </div>

    <template v-else>
      <header class="paste-view-heading">
        <div>
          <div class="paste-view-title-line">
            <h1 id="paste-title">{{ displayTitle(value.title, value.files[0]?.path) }}</h1>
            <span class="paste-code-label">{{ value.code }}</span>
          </div>
          <p v-if="value.description">{{ value.description }}</p>
          <div class="paste-view-meta">
            <span><UIcon name="i-tabler-files" class="size-4" />{{ value.files.length }} 个文件</span>
            <span><UIcon :name="value.visibility === 'private' ? 'i-tabler-lock' : 'i-tabler-link'" class="size-4" />{{ value.visibility === "private" ? "仅自己" : "持链访问" }}</span>
            <span v-if="value.expiresAt"><UIcon name="i-tabler-clock" class="size-4" />{{ new Date(value.expiresAt).toLocaleString("zh-CN") }} 到期</span>
          </div>
        </div>
        <div class="paste-view-actions">
          <UButton :label="copied === 'link' ? '已复制链接' : '复制链接'" icon="i-tabler-link" color="neutral" variant="outline" @click="copyShareLink" />
          <UButton :label="copied === 'all' ? '已复制全部' : '复制全部'" icon="i-tabler-copy" class="paste-button-primary" @click="copyAll" />
        </div>
      </header>

      <div v-if="route.query.created === '1'" class="paste-created-note" role="status">
        <UIcon name="i-tabler-circle-check" class="size-5" />
        <span>Paste 已创建。现在可以安全地复制分享链接。</span>
      </div>

      <div class="paste-view-workbench">
        <aside class="paste-view-files" aria-label="文件列表">
          <button
            v-for="(file, index) in value.files"
            :key="file.path"
            type="button"
            :data-active="activeIndex === index"
            @click="activeIndex = index"
          >
            <UIcon name="i-tabler-file-code" class="size-4 shrink-0" />
            <span class="truncate">{{ file.path }}</span>
            <small>{{ file.content.split("\n").length }}</small>
          </button>
        </aside>
        <section class="paste-view-code" :aria-label="activeFile?.path">
          <div class="paste-view-codebar">
            <span><UIcon name="i-tabler-file" class="size-4" />{{ activeFile?.path }}</span>
            <div>
              <span>{{ activeFile?.language }}</span>
              <UButton
                type="button"
                :icon="copied === activeFile?.path ? 'i-tabler-check' : 'i-tabler-copy'"
                color="neutral"
                variant="ghost"
                size="sm"
                :aria-label="`复制 ${activeFile?.path}`"
                @click="activeFile && copy(activeFile.content, activeFile.path)"
              />
            </div>
          </div>
          <ClientOnly>
            <PasteCodeEditor
              v-if="activeFile"
              :key="activeFile.path"
              v-model="activeFile.content"
              :language="activeFile.language"
              :label="`${activeFile.path} 只读代码`"
              read-only
            />
            <template #fallback><pre>{{ activeFile?.content }}</pre></template>
          </ClientOnly>
        </section>
      </div>
    </template>
  </section>
</template>

<style scoped>
.paste-view { padding-block: 34px 50px; }
.paste-view-state { display: grid; min-height: 55vh; place-items: center; align-content: center; gap: 12px; color: var(--paste-ink-soft); text-align: center; }
.paste-view-state h1 { margin: 6px 0 0; color: var(--paste-ink); font-size: 24px; }
.paste-view-state p { margin: 0; max-width: 48ch; }
.paste-view-error > svg { color: var(--paste-red); }
.paste-access-gate { width: min(420px, 100%); margin: 12vh auto; text-align: center; }
.paste-gate-mark { display: grid; width: 52px; height: 52px; margin: 0 auto 18px; place-items: center; border: 1px solid var(--paste-line); border-radius: 14px; background: var(--paste-surface-muted); color: var(--paste-blue); }
.paste-access-gate h1 { margin: 0; font-size: 28px; letter-spacing: -.03em; }
.paste-access-gate > p { margin: 10px auto 24px; color: var(--paste-ink-soft); font-size: 13px; }
.paste-access-gate form { display: grid; gap: 12px; text-align: left; }
.paste-view-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 32px; margin-bottom: 22px; }
.paste-view-title-line { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
.paste-view-heading h1 { margin: 0; font-size: clamp(25px, 3vw, 38px); font-weight: 750; letter-spacing: -.035em; line-height: 1.12; }
.paste-code-label { border: 1px solid var(--paste-line); border-radius: 7px; padding: 4px 7px; color: var(--paste-blue); font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; font-weight: 680; }
.paste-view-heading > div > p { max-width: 70ch; margin: 9px 0 0; color: var(--paste-ink-soft); font-size: 13px; line-height: 1.65; }
.paste-view-meta { display: flex; flex-wrap: wrap; gap: 8px 18px; margin-top: 13px; color: var(--paste-ink-soft); font-size: 11px; }
.paste-view-meta span { display: inline-flex; align-items: center; gap: 5px; }
.paste-view-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.paste-created-note { display: flex; align-items: center; gap: 8px; margin: -4px 0 16px; color: var(--paste-green); font-size: 12px; }
.paste-view-workbench { display: grid; min-height: calc(100vh - 230px); grid-template-columns: 220px minmax(0, 1fr); overflow: hidden; border: 1px solid var(--paste-line-strong); border-radius: 15px; background: var(--paste-surface); box-shadow: var(--paste-shadow); }
.paste-view-files { display: grid; align-content: start; gap: 3px; border-right: 1px solid var(--paste-line); padding: 9px; background: var(--paste-surface-muted); }
.paste-view-files button { display: grid; min-width: 0; min-height: 42px; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 8px; border: 0; border-radius: 8px; padding: 0 9px; background: transparent; color: var(--paste-ink-soft); font-size: 12px; text-align: left; }
.paste-view-files button:hover { background: var(--paste-surface); color: var(--paste-ink); }
.paste-view-files button[data-active="true"] { background: var(--paste-blue-soft); color: var(--paste-blue); box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--paste-blue) 20%, transparent); font-weight: 680; }
.paste-view-files small { color: var(--paste-ink-dim); font-family: "SFMono-Regular", Consolas, monospace; font-size: 9px; }
.paste-view-code { display: grid; min-width: 0; min-height: 0; grid-template-rows: 52px minmax(480px, 1fr); background: var(--paste-editor); }
.paste-view-codebar { display: flex; align-items: center; justify-content: space-between; gap: 14px; border-bottom: 1px solid var(--paste-line); padding: 0 10px 0 15px; background: var(--paste-surface); color: var(--paste-ink-soft); font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; }
.paste-view-codebar > span, .paste-view-codebar > div { display: flex; align-items: center; gap: 8px; }
.paste-view-code pre { overflow: auto; margin: 0; padding: 18px; color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; font-size: 13px; line-height: 1.7; }
@media (max-width: 720px) {
  .paste-view { width: 100%; padding-block: 22px 0; }
  .paste-view-heading, .paste-created-note { width: calc(100% - 24px); margin-inline: auto; }
  .paste-view-heading { flex-direction: column; gap: 18px; }
  .paste-view-actions { width: 100%; }
  .paste-view-actions > * { flex: 1; min-height: 44px; }
  .paste-view-workbench { min-height: 0; grid-template-columns: 1fr; overflow: visible; border-right: 0; border-left: 0; border-radius: 0; box-shadow: none; }
  .paste-view-files { display: flex; overflow-x: auto; border-right: 0; border-bottom: 1px solid var(--paste-line); }
  .paste-view-files button { min-width: 150px; min-height: 44px; }
  .paste-view-code { grid-template-rows: 52px 65vh; }
}
</style>
