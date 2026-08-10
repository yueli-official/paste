<!--
THESIS: Paste is an editor-first workbench where the document stays central and publishing controls stay at the edge.
OWN-WORLD: no — it inherits the established Yueli warm-paper, mineral-blue, flat-surface visual language.
STORY: assemble ordered files, shape access boundaries, publish once, then share a stable locator.
FIRST VIEWPORT: the complete three-part workbench is visible without marketing copy displacing the editor.
FORM: one continuous bordered work surface with a file rail, code floor, and publication rail.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import type { PasteFile, PasteVisibility } from "../types/paste";
import {
  displayTitle,
  expiryISO,
  languageFromPath,
  languageItems,
  pasteErrorMessage,
} from "../utils/paste";

interface LocalFile extends PasteFile { id: number }

const route = useRoute();
const api = usePasteApi();
const transfer = usePasteTransfer();
const { loggedIn, login } = useAuth();
let nextFileID = 1;

const files = ref<LocalFile[]>([
  { id: nextFileID++, path: "main.go", language: "go", content: "" },
]);
const activeFileID = ref(files.value[0]!.id);
const title = ref("");
const description = ref("");
const tags = ref("");
const visibility = ref<PasteVisibility>("unlisted");
const password = ref("");
const expiry = ref("7d");
const saving = ref(false);
const loadingEdit = ref(false);
const error = ref("");
const editRevision = ref<number>();
const editCode = computed(() =>
  typeof route.query.edit === "string" ? route.query.edit : "",
);
const activeFile = computed(() =>
  files.value.find((file) => file.id === activeFileID.value) || files.value[0]!,
);
const totalBytes = computed(() =>
  files.value.reduce((total, file) => total + new Blob([file.content]).size, 0),
);

useSeoMeta({
  title: editCode.value ? "编辑 Paste" : "新建 Paste",
  description: "粘贴一段或一组代码，设定访问边界，再分享一个稳定链接。",
  ogTitle: "月离 Paste",
  ogDescription: "多文件代码分享工作台。",
});

function addFile() {
  if (files.value.length >= 20) return;
  const index = files.value.length + 1;
  const file: LocalFile = {
    id: nextFileID++,
    path: `untitled-${index}.txt`,
    language: "text",
    content: "",
  };
  files.value.push(file);
  activeFileID.value = file.id;
}

function removeFile(file: LocalFile) {
  if (files.value.length === 1) return;
  const index = files.value.findIndex((candidate) => candidate.id === file.id);
  files.value.splice(index, 1);
  if (activeFileID.value === file.id) {
    activeFileID.value = files.value[Math.max(0, index - 1)]!.id;
  }
}

function moveFile(file: LocalFile, direction: -1 | 1) {
  const index = files.value.findIndex((candidate) => candidate.id === file.id);
  const target = index + direction;
  if (target < 0 || target >= files.value.length) return;
  const [moved] = files.value.splice(index, 1);
  files.value.splice(target, 0, moved!);
}

function inferLanguage(file: LocalFile) {
  file.language = languageFromPath(file.path);
}

function validate(): string {
  if (files.value.some((file) => !file.path.trim())) return "每个文件都需要一个文件名。";
  if (new Set(files.value.map((file) => file.path.trim())).size !== files.value.length) {
    return "文件名不能重复。";
  }
  if (files.value.every((file) => !file.content)) return "请先粘贴要分享的内容。";
  if (totalBytes.value > 1024 * 1024) return "全部文件合计不能超过 1 MiB。";
  if (visibility.value === "private" && !loggedIn.value) return "私有 Paste 需要先登录。";
  return "";
}

function fieldValue(form: HTMLFormElement, name: string): string {
  const field = form.elements.namedItem(name);
  return field instanceof HTMLInputElement || field instanceof HTMLTextAreaElement || field instanceof HTMLSelectElement
    ? field.value
    : "";
}

async function submit(event: SubmitEvent) {
  const form = event.currentTarget as HTMLFormElement;
  title.value = fieldValue(form, "title");
  description.value = fieldValue(form, "description");
  tags.value = fieldValue(form, "tags");
  visibility.value = fieldValue(form, "visibility") as PasteVisibility;
  expiry.value = fieldValue(form, "expiry");
  password.value = fieldValue(form, "password");
  error.value = validate();
  if (error.value) return;
  saving.value = true;
  const input = {
    title: title.value.trim(),
    description: description.value.trim(),
    tags: tags.value.split(",").map((tag) => tag.trim()).filter(Boolean),
    files: files.value.map(({ path, language, content }) => ({
      path: path.trim(), language, content,
    })),
    visibility: visibility.value,
    password: password.value,
    expiresAt: expiryISO(expiry.value),
    clearExpiry: Boolean(editCode.value) && expiry.value === "never",
  };
  try {
    const response = editCode.value && editRevision.value
      ? await api.update(editCode.value, editRevision.value, input)
      : await api.create(input);
    transfer.value[response.paste.code] = response.paste;
    await navigateTo(`/p/${response.paste.code}?${editCode.value ? "updated" : "created"}=1`);
  } catch (caught) {
    error.value = pasteErrorMessage(caught);
  } finally {
    saving.value = false;
  }
}

async function loadEdit() {
  if (!editCode.value) return;
  if (!loggedIn.value) {
    await login(`/?edit=${encodeURIComponent(editCode.value)}`);
    return;
  }
  loadingEdit.value = true;
  try {
    const response = await api.getMine(editCode.value);
    const value = response.paste;
    title.value = value.title;
    description.value = value.description || "";
    tags.value = value.tags.join(", ");
    visibility.value = value.visibility;
    password.value = "";
    expiry.value = value.expiresAt ? "keep" : "never";
    editRevision.value = value.revision;
    files.value = value.files.map((file) => ({ ...file, id: nextFileID++ }));
    activeFileID.value = files.value[0]!.id;
  } catch (caught) {
    error.value = pasteErrorMessage(caught);
  } finally {
    loadingEdit.value = false;
  }
}

onMounted(loadEdit);
</script>

<template>
  <section class="paste-compose paste-container" aria-labelledby="compose-title">
    <div class="paste-compose-intro">
      <div>
        <h1 id="compose-title">{{ editCode ? "继续编辑" : "把代码放好，再分享。" }}</h1>
        <p>{{ editCode ? `正在编辑 ${editCode}` : "匿名可直接创建；登录后可回看、修改和删除。" }}</p>
      </div>
      <span class="paste-capacity" aria-live="polite">
        {{ files.length }}/20 个文件 · {{ Math.ceil(totalBytes / 1024) }} KiB / 1024 KiB
      </span>
    </div>

    <form class="paste-workbench" :aria-busy="saving || loadingEdit" @submit.prevent="submit">
      <aside class="paste-file-rail" aria-label="Paste 文件">
        <div class="paste-rail-heading">
          <span>文件</span>
          <UTooltip text="添加文件">
            <UButton
              type="button"
              icon="i-tabler-plus"
              color="neutral"
              variant="ghost"
              size="sm"
              aria-label="添加文件"
              :disabled="files.length >= 20"
              @click="addFile"
            />
          </UTooltip>
        </div>
        <div class="paste-file-list">
          <button
            v-for="(file, index) in files"
            :key="file.id"
            type="button"
            class="paste-file-item"
            :data-active="activeFileID === file.id"
            :aria-current="activeFileID === file.id ? 'true' : undefined"
            @click="activeFileID = file.id"
          >
            <UIcon name="i-tabler-file-code" class="size-4 shrink-0" />
            <span class="truncate">{{ file.path || `文件 ${index + 1}` }}</span>
          </button>
        </div>
        <p class="paste-rail-note">按列表顺序展示；每个文件最多 256 KiB。</p>
      </aside>

      <section class="paste-editor-floor" aria-label="当前文件">
        <div class="paste-editor-toolbar">
          <div class="paste-path-field">
            <UIcon name="i-tabler-file" class="size-4" />
            <input
              v-model="activeFile.path"
              class="paste-path-input"
              aria-label="当前文件名"
              maxlength="180"
              spellcheck="false"
              @change="inferLanguage(activeFile)"
            >
          </div>
          <select v-model="activeFile.language" class="paste-language-select" aria-label="代码语言">
            <option v-for="item in languageItems" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
          <div class="paste-file-actions" aria-label="文件排序操作">
            <UTooltip text="上移">
              <UButton type="button" icon="i-tabler-arrow-up" color="neutral" variant="ghost" size="sm" aria-label="上移文件" :disabled="files[0]?.id === activeFile.id" @click="moveFile(activeFile, -1)" />
            </UTooltip>
            <UTooltip text="下移">
              <UButton type="button" icon="i-tabler-arrow-down" color="neutral" variant="ghost" size="sm" aria-label="下移文件" :disabled="files.at(-1)?.id === activeFile.id" @click="moveFile(activeFile, 1)" />
            </UTooltip>
            <UTooltip text="移除">
              <UButton type="button" icon="i-tabler-trash" color="error" variant="ghost" size="sm" aria-label="移除文件" :disabled="files.length === 1" @click="removeFile(activeFile)" />
            </UTooltip>
          </div>
        </div>
        <ClientOnly>
          <PasteCodeEditor
            v-model="activeFile.content"
            :language="activeFile.language"
            :label="`${activeFile.path} 代码编辑器`"
          />
          <template #fallback>
            <textarea v-model="activeFile.content" class="paste-editor-fallback" aria-label="代码编辑器" />
          </template>
        </ClientOnly>
      </section>

      <aside class="paste-publish-rail" aria-label="发布设置">
        <div class="paste-publish-heading">
          <div>
            <h2>发布设置</h2>
            <p>分享前确认内容与访问边界。</p>
          </div>
          <UIcon name="i-tabler-adjustments-horizontal" class="size-5" />
        </div>

        <div class="paste-setting-fields">
          <div>
            <label class="paste-field-label" for="paste-title">标题</label>
            <input id="paste-title" v-model="title" name="title" class="paste-input" maxlength="120" :placeholder="displayTitle('', activeFile.path)">
          </div>
          <div>
            <label class="paste-field-label" for="paste-description">说明（可选）</label>
            <textarea id="paste-description" v-model="description" name="description" class="paste-textarea" maxlength="2000" placeholder="告诉接收者这组文件是什么" />
          </div>
          <div>
            <label class="paste-field-label" for="paste-tags">标签（可选）</label>
            <input id="paste-tags" v-model="tags" name="tags" class="paste-input" placeholder="go, api, demo">
            <p class="paste-help">逗号分隔，最多 8 个。</p>
          </div>
          <div class="paste-setting-grid">
            <div>
              <label class="paste-field-label" for="paste-visibility">可见性</label>
              <select id="paste-visibility" v-model="visibility" name="visibility" class="paste-select">
                <option value="unlisted">知道链接即可访问</option>
                <option value="private" :disabled="!loggedIn">仅自己可访问</option>
              </select>
            </div>
            <div>
              <label class="paste-field-label" for="paste-expiry">有效期</label>
              <select id="paste-expiry" v-model="expiry" name="expiry" class="paste-select">
                <option v-if="editCode && expiry === 'keep'" value="keep">保持原到期时间</option>
                <option value="1h">1 小时</option>
                <option value="1d">1 天</option>
                <option value="7d">7 天</option>
                <option value="30d">30 天</option>
                <option value="never">不过期</option>
              </select>
            </div>
          </div>
          <div>
            <label class="paste-field-label" for="paste-password">访问密码（可选）</label>
            <input id="paste-password" v-model="password" name="password" class="paste-input" type="password" autocomplete="new-password" placeholder="不会出现在链接中">
            <p v-if="editCode" class="paste-help">留空会移除原有密码。</p>
          </div>
        </div>

        <div class="paste-publish-action">
          <p v-if="error" class="paste-error" role="alert">{{ error }}</p>
          <p v-else-if="!loggedIn && !editCode" class="paste-anonymous-note">
            <UIcon name="i-tabler-user-off" class="size-4" />匿名 Paste 创建后不能修改。
          </p>
          <UButton
            type="submit"
            :label="editCode ? '保存修改' : '生成分享链接'"
            :icon="editCode ? 'i-tabler-device-floppy' : 'i-tabler-send'"
            block
            size="lg"
            class="paste-button-primary min-h-12"
            :loading="saving || loadingEdit"
            :disabled="saving || loadingEdit"
          />
          <button v-if="!loggedIn && !editCode" type="button" class="paste-login-link" @click="login('/')">
            登录后创建，保留管理入口
          </button>
        </div>
      </aside>
    </form>
  </section>
</template>

<style scoped>
.paste-compose { padding-block: 30px 40px; }
.paste-compose-intro { display: flex; align-items: end; justify-content: space-between; gap: 24px; margin-bottom: 18px; }
.paste-compose-intro h1 { margin: 0; font-size: clamp(25px, 3vw, 38px); font-weight: 760; letter-spacing: -.035em; line-height: 1.08; text-wrap: balance; }
.paste-compose-intro p { margin: 8px 0 0; color: var(--paste-ink-soft); font-size: 13px; }
.paste-capacity { color: var(--paste-ink-soft); font-family: "SFMono-Regular", Consolas, monospace; font-size: 12px; white-space: nowrap; }
.paste-workbench { display: grid; min-height: calc(100vh - 174px); grid-template-columns: 190px minmax(420px, 1fr) 300px; overflow: hidden; border: 1px solid var(--paste-line-strong); border-radius: 15px; background: var(--paste-surface); box-shadow: var(--paste-shadow); }
.paste-file-rail { display: flex; min-width: 0; flex-direction: column; border-right: 1px solid var(--paste-line); background: var(--paste-surface-muted); }
.paste-rail-heading { display: flex; min-height: 54px; align-items: center; justify-content: space-between; padding: 0 10px 0 15px; border-bottom: 1px solid var(--paste-line); color: var(--paste-ink-soft); font-size: 12px; font-weight: 680; }
.paste-file-list { display: grid; gap: 3px; padding: 8px; }
.paste-file-item { display: flex; min-width: 0; min-height: 40px; align-items: center; gap: 8px; border: 0; border-radius: 8px; padding: 0 9px; background: transparent; color: var(--paste-ink-soft); font-size: 12px; text-align: left; }
.paste-file-item:hover { background: color-mix(in srgb, var(--paste-surface) 64%, transparent); color: var(--paste-ink); }
.paste-file-item[data-active="true"] { background: var(--paste-blue-soft); color: var(--paste-blue); box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--paste-blue) 20%, transparent); font-weight: 680; }
.paste-rail-note { margin: auto 14px 16px; color: var(--paste-ink-dim); font-size: 10px; line-height: 1.55; }
.paste-editor-floor { display: grid; min-width: 0; min-height: 0; grid-template-rows: 54px minmax(480px, 1fr); background: var(--paste-editor); }
.paste-editor-toolbar { display: flex; min-width: 0; align-items: center; gap: 8px; border-bottom: 1px solid var(--paste-line); padding: 0 10px 0 14px; background: var(--paste-surface); }
.paste-path-field { display: flex; min-width: 0; flex: 1; align-items: center; gap: 7px; color: var(--paste-ink-soft); }
.paste-path-input { width: 100%; min-width: 90px; border: 0; background: transparent; color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; font-size: 12px; font-weight: 650; outline: 0; }
.paste-path-input:focus { text-decoration: underline; text-decoration-color: var(--paste-blue); text-underline-offset: 4px; }
.paste-language-select { min-height: 34px; border: 1px solid var(--paste-line); border-radius: 8px; padding: 0 28px 0 9px; background: var(--paste-surface); color: var(--paste-ink-soft); font-size: 11px; }
.paste-file-actions { display: flex; align-items: center; }
.paste-editor-fallback { width: 100%; height: 100%; resize: none; border: 0; padding: 18px 20px; background: var(--paste-editor); color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; line-height: 1.72; outline: 0; }
.paste-publish-rail { display: flex; min-width: 0; flex-direction: column; border-left: 1px solid var(--paste-line); background: var(--paste-surface); }
.paste-publish-heading { display: flex; min-height: 76px; align-items: center; justify-content: space-between; gap: 16px; border-bottom: 1px solid var(--paste-line); padding: 14px 18px; }
.paste-publish-heading h2 { margin: 0; font-size: 15px; font-weight: 720; letter-spacing: -.015em; }
.paste-publish-heading p { margin: 4px 0 0; color: var(--paste-ink-soft); font-size: 10px; }
.paste-publish-heading > svg { color: var(--paste-ink-dim); }
.paste-setting-fields { display: grid; gap: 17px; padding: 18px; }
.paste-setting-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.paste-publish-action { display: grid; gap: 10px; margin-top: auto; border-top: 1px solid var(--paste-line); padding: 16px 18px 18px; background: var(--paste-surface-muted); }
.paste-anonymous-note { display: flex; align-items: center; gap: 6px; margin: 0; color: var(--paste-ink-soft); font-size: 11px; }
.paste-login-link { min-height: 38px; border: 0; background: transparent; color: var(--paste-blue); font-size: 11px; font-weight: 650; }
.paste-login-link:hover { text-decoration: underline; text-underline-offset: 3px; }
@media (max-width: 1060px) {
  .paste-workbench { grid-template-columns: 170px minmax(390px, 1fr); }
  .paste-publish-rail { grid-column: 1 / -1; display: grid; grid-template-columns: 170px minmax(0, 1fr) 260px; border-top: 1px solid var(--paste-line); border-left: 0; }
  .paste-publish-heading { align-items: flex-start; border-right: 1px solid var(--paste-line); border-bottom: 0; }
  .paste-setting-fields { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .paste-setting-fields > :nth-child(2) { grid-column: span 2; }
  .paste-publish-action { margin: 0; border-top: 0; border-left: 1px solid var(--paste-line); }
}
@media (max-width: 720px) {
  .paste-compose { width: 100%; padding-block: 20px 0; }
  .paste-compose-intro { width: calc(100% - 24px); margin-inline: auto; align-items: flex-start; flex-direction: column; gap: 10px; }
  .paste-compose-intro h1 { font-size: 28px; }
  .paste-workbench { min-height: 0; grid-template-columns: 1fr; overflow: visible; border-right: 0; border-left: 0; border-radius: 0; box-shadow: none; }
  .paste-file-rail { border-right: 0; border-bottom: 1px solid var(--paste-line); }
  .paste-file-list { display: flex; overflow-x: auto; padding: 7px 10px 10px; }
  .paste-file-item { min-width: 128px; min-height: 44px; }
  .paste-rail-note { display: none; }
  .paste-editor-floor { grid-template-rows: auto 58vh; }
  .paste-editor-toolbar { min-height: 58px; flex-wrap: wrap; padding-block: 7px; }
  .paste-path-field { flex-basis: calc(100% - 165px); }
  .paste-language-select { min-height: 44px; }
  .paste-publish-rail { grid-column: auto; display: flex; border-left: 0; }
  .paste-publish-heading { border-right: 0; border-bottom: 1px solid var(--paste-line); }
  .paste-setting-fields { grid-template-columns: 1fr; }
  .paste-setting-fields > :nth-child(2) { grid-column: auto; }
  .paste-input, .paste-select { min-height: 44px; }
  .paste-publish-action { border-top: 1px solid var(--paste-line); border-left: 0; padding-bottom: max(20px, env(safe-area-inset-bottom)); }
}
</style>
