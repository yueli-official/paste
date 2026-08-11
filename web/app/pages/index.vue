<!--
THESIS: The editor owns the whole viewport; Paste refuses a separate website header above the work.
OWN-WORLD: Yueli mineral blue inside restrained, flat, VS Code-derived application chrome.
STORY: open a file, write code, set its boundary, and share without leaving the editor context.
FIRST VIEWPORT: one 40px title-and-tab bar, the code floor, and a 28px readable status bar; sharing stays transient.
FORM: a single integrated editor shell with product, files, commands, navigation, and identity on one line.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";
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
const maxFileBytes = 1024 * 1024;
const maxContentBytes = 1024 * 1024;

const accountActions: readonly AccountMenuAction[] = [
  {
    label: "我的 Paste",
    icon: "i-tabler-folders",
    to: "/mine",
  },
];

const files = ref<LocalFile[]>([
  { id: nextFileID++, path: "main.go", language: "go", content: "" },
]);
const activeFileID = ref(files.value[0]!.id);
const title = ref("");
const description = ref("");
const tags = ref<string[]>([]);
const visibility = ref<PasteVisibility>("unlisted");
const password = ref("");
const expiry = ref("7d");
const shareOpen = ref(false);
const renamingFileID = ref<number>();
const renameDraft = ref("");
const draggedFileID = ref<number>();
const dragOverFileID = ref<number>();
const dragOverPosition = ref<"before" | "after">();
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
const activeFileBytes = computed(() => new Blob([activeFile.value.content]).size);
const languageOptions: { label: string; value: string }[] = languageItems.map(
  ({ label, value }) => ({ label, value }),
);
const visibilityItems = computed(() => [
  {
    label: "知道链接即可访问",
    value: "unlisted" as const,
    icon: "i-tabler-link",
  },
  {
    label: "仅自己可访问",
    value: "private" as const,
    icon: "i-tabler-lock",
    disabled: !loggedIn.value,
  },
]);
const expiryItems = computed(() => [
  ...(editCode.value && expiry.value === "keep"
    ? [{ label: "保持原到期时间", value: "keep" }]
    : []),
  { label: "1 小时", value: "1h" },
  { label: "1 天", value: "1d" },
  { label: "7 天", value: "7d" },
  { label: "30 天", value: "30d" },
  { label: "不过期", value: "never" },
]);
const visibilityLabel = computed(() =>
  visibility.value === "private" ? "仅自己" : "链接访问",
);

function openShare() {
  shareOpen.value = true;
}

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
  if (renamingFileID.value === file.id) cancelRename();
}

function moveFile(file: LocalFile, direction: -1 | 1) {
  const index = files.value.findIndex((candidate) => candidate.id === file.id);
  const target = index + direction;
  if (target < 0 || target >= files.value.length) return;
  const [moved] = files.value.splice(index, 1);
  files.value.splice(target, 0, moved!);
}

function clearFileDrag() {
  draggedFileID.value = undefined;
  dragOverFileID.value = undefined;
  dragOverPosition.value = undefined;
}

function startFileDrag(file: LocalFile, event: DragEvent) {
  if (renamingFileID.value === file.id) {
    event.preventDefault();
    return;
  }
  draggedFileID.value = file.id;
  event.dataTransfer?.setData("text/plain", String(file.id));
  if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
}

function updateFileDrop(file: LocalFile, event: DragEvent) {
  if (!draggedFileID.value || draggedFileID.value === file.id) {
    dragOverFileID.value = undefined;
    dragOverPosition.value = undefined;
    return;
  }
  const bounds = (event.currentTarget as HTMLElement).getBoundingClientRect();
  dragOverFileID.value = file.id;
  dragOverPosition.value = event.clientX < bounds.left + bounds.width / 2 ? "before" : "after";
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
}

function dropFile(targetFile: LocalFile, event: DragEvent) {
  const sourceID = draggedFileID.value || Number(event.dataTransfer?.getData("text/plain"));
  const sourceIndex = files.value.findIndex((file) => file.id === sourceID);
  let targetIndex = files.value.findIndex((file) => file.id === targetFile.id);
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) {
    clearFileDrag();
    return;
  }
  const [moved] = files.value.splice(sourceIndex, 1);
  if (sourceIndex < targetIndex) targetIndex -= 1;
  const insertIndex = targetIndex + (dragOverPosition.value === "after" ? 1 : 0);
  files.value.splice(insertIndex, 0, moved!);
  activeFileID.value = moved!.id;
  clearFileDrag();
}

function inferLanguage(file: LocalFile) {
  file.language = languageFromPath(file.path);
}

function beginRename(file: LocalFile) {
  activeFileID.value = file.id;
  renameDraft.value = file.path;
  renamingFileID.value = file.id;
}

function commitRename(file: LocalFile) {
  const nextPath = renameDraft.value.trim();
  if (nextPath) {
    file.path = nextPath;
    inferLanguage(file);
  }
  cancelRename();
}

function cancelRename() {
  renamingFileID.value = undefined;
  renameDraft.value = "";
}

async function loginForManagement() {
  await login("/");
}

function validate(): string {
  if (files.value.some((file) => !file.path.trim())) return "每个文件都需要一个文件名。";
  if (files.value.some((file) => {
    const path = file.path.trim().replaceAll("\\", "/");
    return path.startsWith("/") || path.split("/").includes("..");
  })) return "文件名需要使用 Paste 内的相对路径，不能包含上级目录。";
  if (new Set(files.value.map((file) => file.path.trim().toLowerCase())).size !== files.value.length) {
    return "文件名不能重复。";
  }
  if (files.value.every((file) => !file.content)) return "请先粘贴要分享的内容。";
  const oversized = files.value.find((file) => new Blob([file.content]).size > maxFileBytes);
  if (oversized) return `“${oversized.path}”超过文件限制；单个文件不能超过 1 MiB。`;
  if (totalBytes.value > maxContentBytes) return "全部文件合计不能超过 1 MiB。";
  if (tags.value.some((tag) => [...tag.trim()].length > 32)) return "每个标签不能超过 32 个字符。";
  if (password.value && ([...password.value].length < 8 || [...password.value].length > 128)) {
    return "访问密码需要包含 8–128 个字符。";
  }
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
  password.value = fieldValue(form, "password");
  error.value = validate();
  if (error.value) return;
  saving.value = true;
  const input = {
    title: title.value.trim(),
    description: description.value.trim(),
    tags: tags.value.map((tag) => tag.trim()).filter(Boolean),
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
    tags.value = [...value.tags];
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
  <section class="paste-compose" aria-labelledby="compose-title">
    <form
      id="paste-compose-form"
      class="paste-workbench"
      :aria-busy="saving || loadingEdit"
      @submit.prevent="submit"
    >
      <header class="paste-editor-chrome" aria-label="Paste 编辑器工具栏">
        <NuxtLink to="/" class="paste-editor-brand" aria-label="Paste 首页">
          <span class="paste-editor-brand-mark" aria-hidden="true">
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span class="paste-editor-brand-label">Paste</span>
        </NuxtLink>

        <nav class="paste-file-tabs" aria-label="Paste 文件">
          <div
            v-for="(file, index) in files"
            :key="file.id"
            class="paste-file-tab"
            :data-active="activeFileID === file.id"
            :data-dragging="draggedFileID === file.id"
            :data-drop-position="dragOverFileID === file.id ? dragOverPosition : undefined"
            :draggable="renamingFileID !== file.id"
            @dragstart="startFileDrag(file, $event)"
            @dragover.prevent="updateFileDrop(file, $event)"
            @drop.prevent="dropFile(file, $event)"
            @dragend="clearFileDrag"
          >
            <input
              v-if="renamingFileID === file.id"
              v-model="renameDraft"
              autofocus
              type="text"
              maxlength="180"
              spellcheck="false"
              aria-label="重命名文件"
              class="paste-rename-input"
              @blur="commitRename(file)"
              @keyup.enter="commitRename(file)"
              @keyup.esc="cancelRename"
            />
            <button
              v-else
              type="button"
              class="paste-file-tab-button"
              :aria-pressed="activeFileID === file.id"
              :aria-label="`${file.path || `文件 ${index + 1}`}，双击重命名，Alt 加方向键移动`"
              @click="activeFileID = file.id"
              @dblclick.prevent="beginRename(file)"
              @keydown.alt.left.prevent="moveFile(file, -1)"
              @keydown.alt.right.prevent="moveFile(file, 1)"
            >
              <UIcon name="i-tabler-file-code" class="size-4 shrink-0" />
              <span class="truncate">{{ file.path || `文件 ${index + 1}` }}</span>
            </button>
            <UButton
              v-if="files.length > 1 && renamingFileID !== file.id"
              type="button"
              icon="i-tabler-x"
              color="neutral"
              variant="ghost"
              size="xs"
              square
              :aria-label="`删除 ${file.path}`"
              class="paste-tab-close"
              @click.stop="removeFile(file)"
            />
          </div>
          <UTooltip text="添加文件">
            <UButton
              type="button"
              icon="i-tabler-plus"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              aria-label="添加文件"
              class="paste-add-file"
              :disabled="files.length >= 20"
              @click="addFile"
            />
          </UTooltip>
        </nav>

        <div class="paste-editor-actions">
          <UTooltip text="分享">
            <UButton
              type="button"
              icon="i-tabler-share-3"
              color="primary"
              variant="ghost"
              size="sm"
              square
              aria-label="打开分享设置"
              @click="openShare"
            />
          </UTooltip>
          <span class="paste-editor-action-divider" aria-hidden="true" />
          <UTooltip text="我的 Paste">
            <UButton
              to="/mine"
              icon="i-tabler-folders"
              color="neutral"
              variant="ghost"
              size="sm"
              square
              class="paste-editor-history"
              aria-label="我的 Paste"
            />
          </UTooltip>
          <UTooltip text="切换颜色模式">
            <UColorModeButton
              color="neutral"
              variant="ghost"
              size="sm"
              aria-label="切换颜色模式"
              class="paste-editor-theme"
            />
          </UTooltip>
          <div class="paste-editor-account">
            <ConsumerAccountControl
              :context-actions="accountActions"
              trigger-mode="collapsed"
            />
          </div>
        </div>
      </header>

      <section class="paste-editor-floor" aria-label="当前文件">
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

      <footer class="paste-statusbar" aria-label="编辑状态">
        <span class="paste-status-primary">
          <span class="paste-status-dot" aria-hidden="true" />
          <h1 id="compose-title">{{ editCode ? "编辑 Paste" : "新建 Paste" }}</h1>
        </span>
        <span class="paste-status-meta" aria-live="polite">
          <span>{{ visibilityLabel }}</span>
          <USelectMenu
            v-model="activeFile.language"
            :items="languageOptions"
            value-key="value"
            label-key="label"
            :search-input="false"
            :content="{ align: 'end', side: 'top', sideOffset: 6, collisionPadding: 8 }"
            variant="none"
            size="xs"
            class="paste-status-language"
            :ui="{
              base: 'min-h-0 rounded-none px-1.5 py-0 text-[12px] text-[var(--paste-status-text)] ring-0 hover:bg-[var(--paste-status-hover)]',
              content: 'w-52 min-w-52 rounded-md',
              viewport: 'max-h-[33rem] py-1',
              item: 'min-h-8 text-xs',
              trailingIcon: 'size-3',
            }"
            aria-label="代码语言"
          />
          <span>{{ files.length }}/20 文件</span>
          <span
            class="paste-status-size"
            :data-over-limit="activeFileBytes > maxFileBytes"
            :title="`当前文件 ${Math.ceil(activeFileBytes / 1024)}/1024 KiB；全部文件 ${Math.ceil(totalBytes / 1024)}/1024 KiB`"
          >{{ Math.ceil(activeFileBytes / 1024) }}/1024 KiB</span>
        </span>
      </footer>

      <USlideover
        v-model:open="shareOpen"
        title="分享 Paste"
        description="设置访问边界，然后生成链接。"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div class="paste-share-fields">
            <UFormField name="title" label="标题">
              <UInput
                v-model="title"
                form="paste-compose-form"
                name="title"
                maxlength="120"
                :placeholder="displayTitle('', activeFile.path)"
                class="w-full"
              />
            </UFormField>
            <UFormField name="description" label="说明" hint="可选">
              <UTextarea
                v-model="description"
                form="paste-compose-form"
                name="description"
                :rows="3"
                autoresize
                maxlength="2000"
                placeholder="这组文件用于什么？"
                class="w-full"
              />
            </UFormField>
            <UFormField name="tags" label="标签" hint="最多 8 个">
              <UInputTags
                v-model="tags"
                name="tags"
                :max="8"
                add-on-paste
                add-on-blur
                delimiter=","
                placeholder="输入后回车"
                class="w-full"
              />
            </UFormField>
            <div class="paste-setting-grid">
              <UFormField label="可见性">
                <USelect
                  v-model="visibility"
                  :items="visibilityItems"
                  value-key="value"
                  label-key="label"
                  class="w-full"
                  aria-label="可见性"
                />
              </UFormField>
              <UFormField label="有效期">
                <USelect
                  v-model="expiry"
                  :items="expiryItems"
                  value-key="value"
                  label-key="label"
                  class="w-full"
                  aria-label="有效期"
                />
              </UFormField>
            </div>
            <UFormField
              name="password"
              label="访问密码"
              :hint="editCode ? '留空即移除' : '可选'"
            >
              <UInput
                v-model="password"
                form="paste-compose-form"
                name="password"
                type="password"
                autocomplete="new-password"
                icon="i-tabler-key"
                placeholder="不会出现在链接中"
                class="w-full"
              />
            </UFormField>
            <UAlert
              v-if="error"
              color="error"
              variant="subtle"
              icon="i-tabler-alert-circle"
              title="无法完成分享"
              :description="error"
              role="alert"
            />
            <UAlert
              v-else-if="!loggedIn && !editCode"
              color="neutral"
              variant="subtle"
              icon="i-tabler-user-off"
              description="匿名创建后不能修改。"
            />
          </div>
        </template>
        <template #footer>
          <div class="paste-share-footer">
            <UButton
              v-if="!loggedIn && !editCode"
              type="button"
              label="登录后创建"
              color="neutral"
              variant="ghost"
              @click="loginForManagement"
            />
            <UButton
              form="paste-compose-form"
              type="submit"
              :label="editCode ? '保存修改' : '生成分享链接'"
              :icon="editCode ? 'i-tabler-device-floppy' : 'i-tabler-send'"
              class="paste-button-primary"
              :loading="saving || loadingEdit"
              :disabled="saving || loadingEdit"
            />
          </div>
        </template>
      </USlideover>
    </form>
  </section>
</template>

<style scoped>
.paste-compose { min-height: 100dvh; background: var(--paste-editor); }
.paste-workbench { display: grid; height: 100dvh; min-height: 360px; grid-template-rows: 40px minmax(0, 1fr) 28px; overflow: hidden; background: var(--paste-surface); }
.paste-editor-chrome { display: flex; min-width: 0; align-items: stretch; border-bottom: 1px solid var(--paste-line); background: var(--paste-chrome); }
.paste-editor-brand { display: flex; flex: none; align-items: center; gap: 7px; border-right: 1px solid var(--paste-line); padding: 0 11px; color: var(--paste-ink); text-decoration: none; }
.paste-editor-brand-mark { display: grid; width: 24px; height: 24px; place-items: center; border: 1px solid color-mix(in srgb, var(--paste-blue) 28%, var(--paste-line)); border-radius: 6px; background: var(--paste-blue-soft); color: var(--paste-blue); }
.paste-editor-brand-label { font-size: 13px; font-weight: 730; letter-spacing: -0.025em; }
.paste-file-tabs { display: flex; min-width: 0; flex: 1; overflow-x: auto; overflow-y: hidden; scrollbar-width: none; }
.paste-file-tabs::-webkit-scrollbar { display: none; }
.paste-file-tab { position: relative; display: flex; min-width: 108px; max-width: 220px; flex: 0 1 176px; align-items: center; border-right: 1px solid var(--paste-line); color: var(--paste-ink-soft); }
.paste-file-tab[data-active="true"] { background: var(--paste-editor); box-shadow: inset 0 2px var(--paste-blue); color: var(--paste-ink); }
.paste-file-tab[data-dragging="true"] { opacity: 0.48; }
.paste-file-tab[data-drop-position]::after { position: absolute; z-index: 2; top: 3px; bottom: 3px; width: 2px; border-radius: 1px; background: var(--paste-blue); content: ""; }
.paste-file-tab[data-drop-position="before"]::after { left: -1px; }
.paste-file-tab[data-drop-position="after"]::after { right: -1px; }
.paste-file-tab-button { display: flex; min-width: 0; height: 100%; flex: 1; align-items: center; gap: 7px; border: 0; padding: 0 8px 0 11px; background: transparent; color: inherit; font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; text-align: left; }
.paste-file-tab-button:focus-visible { outline-offset: -3px; }
.paste-rename-input { min-width: 0; height: 26px; flex: 1; margin: 0 6px; border: 1px solid var(--paste-blue); border-radius: 2px; padding: 0 5px; background: var(--paste-editor); color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; font-size: 11px; line-height: 24px; outline: none; box-shadow: none; }
.paste-rename-input::selection { background: var(--paste-selection); }
.paste-tab-close { margin-right: 4px; opacity: 0; transition: opacity 120ms ease; }
.paste-file-tab[data-active="true"] .paste-tab-close,
.paste-file-tab:hover .paste-tab-close,
.paste-file-tab:focus-within .paste-tab-close { opacity: 1; }
.paste-add-file { width: 32px; min-width: 32px; flex: none; border-radius: 0; }
.paste-editor-actions { display: flex; flex: none; align-items: center; gap: 2px; padding: 0 5px; background: var(--paste-chrome); }
.paste-editor-action-divider { width: 1px; height: 20px; margin-inline: 3px; background: var(--paste-line); }
.paste-editor-account { display: grid; width: 32px; height: 32px; place-items: center; }
.paste-editor-account :deep(button) { width: 30px; min-width: 30px; height: 30px; min-height: 30px; padding: 0; }
.paste-editor-account :deep([data-slot="label"]) { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.paste-editor-floor { min-width: 0; min-height: 0; background: var(--paste-editor); }
.paste-editor-fallback { width: 100%; height: 100%; resize: none; border: 0; padding: 18px 20px; background: var(--paste-editor); color: var(--paste-ink); font-family: "SFMono-Regular", Consolas, monospace; line-height: 1.72; outline: 0; }
.paste-share-fields { display: grid; gap: 18px; }
.paste-setting-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.paste-share-fields :deep(label) { font-size: 11px; font-weight: 680; }
.paste-share-footer { display: flex; width: 100%; align-items: center; justify-content: flex-end; gap: 8px; }
.paste-statusbar { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: 20px; border-top: 1px solid var(--paste-status-border); padding: 0 9px; background: var(--paste-status); color: var(--paste-status-text-muted); font-family: "SFMono-Regular", Consolas, monospace; font-size: 12px; }
.paste-status-primary, .paste-status-meta { display: flex; min-width: 0; align-items: center; gap: 12px; }
.paste-status-primary { color: var(--paste-status-text); white-space: nowrap; }
.paste-status-primary h1 { margin: 0; font: inherit; font-weight: 600; }
.paste-status-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--paste-green); }
.paste-status-meta { justify-content: flex-end; white-space: nowrap; }
.paste-status-language { width: auto; min-width: 0; }
.paste-status-size[data-over-limit="true"] { color: var(--paste-status-error); font-weight: 700; }
@media (max-width: 720px) {
  .paste-compose, .paste-workbench { height: 100dvh; min-height: 320px; }
  .paste-workbench { grid-template-rows: 44px minmax(0, 1fr) 30px; }
  .paste-editor-brand { padding-inline: 6px; }
  .paste-editor-brand-label { display: none; }
  .paste-file-tab { min-width: 88px; flex-basis: 124px; }
  .paste-editor-actions { gap: 2px; padding-inline: 4px; }
  .paste-editor-action-divider { display: none; }
  .paste-editor-history { display: none; }
  .paste-status-meta > span:first-child { display: none; }
  .paste-setting-grid { grid-template-columns: 1fr; }
}
</style>
