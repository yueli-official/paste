<!--
THESIS: The editor owns the whole viewport; Paste refuses a separate website header above the work.
OWN-WORLD: Yueli mineral blue inside restrained, flat, VS Code-derived application chrome.
STORY: open a file, write code, set its boundary, and share without leaving the editor context.
FIRST VIEWPORT: one 40px title-and-tab bar, the code floor, and a 28px readable status bar; sharing stays transient.
FORM: a single integrated editor shell with product, files, commands, navigation, and identity on one line.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->
<script setup lang="ts">
import type { FailureFeedback } from "@yueli/http-runtime";
import { pasteFailureFeedback } from "../utils/pasteFailure";
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
const { isAdmin } = usePasteAdministration();
const { siteName, siteDescription } = useSiteSettings();
let nextFileID = 1;
const maxFileBytes = 1024 * 1024;
const maxContentBytes = 1024 * 1024;

const accountActions = computed<readonly AccountMenuAction[]>(() => [
  { label: "我的片段", icon: "i-tabler-folders", to: "/mine" },
  ...(isAdmin.value
    ? [{ label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" }]
    : []),
]);

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
const failure = ref<FailureFeedback | null>(null);
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
  title: computed(() => editCode.value ? `编辑片段 · ${siteName.value}` : `新建片段 · ${siteName.value}`),
  description: "粘贴一段或一组代码，设定访问边界，再分享一个稳定链接。",
  ogTitle: computed(() => siteName.value),
  ogDescription: computed(() => siteDescription.value),
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
  })) return "文件名需要使用片段内的相对路径，不能包含上级目录。";
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
  if (visibility.value === "private" && !loggedIn.value) return "私有片段需要先登录。";
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
  failure.value = null;
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
    failure.value = pasteFailureFeedback(caught, "分享未完成，请检查后重试。", {"/title":"title","/description":"description","/tags":"tags","/password":"password","/visibility":"visibility","/expiresAt":"expiry"});
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
  <section
    class="min-h-dvh bg-[var(--paste-editor)]"
    aria-labelledby="compose-title"
  >
    <form
      id="paste-compose-form"
      class="paste-workbench grid h-dvh min-h-[360px] grid-rows-[40px_minmax(0,1fr)_28px] overflow-hidden bg-[var(--paste-surface)] max-[720px]:min-h-[320px] max-[720px]:grid-rows-[44px_minmax(0,1fr)_30px]"
      :aria-busy="saving || loadingEdit"
      @submit.prevent="submit"
    >
      <header
        class="flex min-w-0 items-stretch border-b border-[var(--paste-line)] bg-[var(--paste-chrome)]"
        aria-label="代码片段编辑器工具栏"
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
            class="paste-editor-brand-label text-[13px] font-[730] tracking-[-0.025em] max-[720px]:hidden"
            >{{ siteName }}</span
          >
        </NuxtLink>

        <nav
          class="flex min-w-0 flex-1 overflow-x-auto overflow-y-hidden [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
          aria-label="片段文件"
        >
          <div
            v-for="(file, index) in files"
            :key="file.id"
            class="paste-file-tab group relative flex min-w-[108px] max-w-[220px] basis-44 items-center border-r border-[var(--paste-line)] text-[var(--paste-ink-soft)] opacity-100 data-[active=true]:bg-[var(--paste-editor)] data-[active=true]:text-[var(--paste-ink)] data-[active=true]:[box-shadow:inset_0_2px_var(--paste-blue)] data-[dragging=true]:opacity-50 data-[drop-position]:after:absolute data-[drop-position]:after:top-[3px] data-[drop-position]:after:bottom-[3px] data-[drop-position]:after:z-[2] data-[drop-position]:after:w-0.5 data-[drop-position]:after:rounded-[1px] data-[drop-position]:after:bg-[var(--paste-blue)] data-[drop-position]:after:content-[''] data-[drop-position=before]:after:-left-px data-[drop-position=after]:after:-right-px max-[720px]:min-w-[88px] max-[720px]:basis-[124px]"
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
              class="mx-1.5 h-[26px] min-w-0 flex-1 rounded-sm border border-[var(--paste-blue)] bg-[var(--paste-editor)] px-[5px] font-mono text-[11px] leading-6 text-[var(--paste-ink)] shadow-none outline-none selection:bg-[var(--paste-selection)]"
              @blur="commitRename(file)"
              @keyup.enter="commitRename(file)"
              @keyup.esc="cancelRename"
            />
            <button
              v-else
              type="button"
              class="paste-file-tab-button flex h-full min-w-0 flex-1 items-center gap-[7px] border-0 bg-transparent py-0 pr-2 pl-[11px] text-left font-mono text-[11px] text-inherit focus-visible:outline-offset-[-3px]"
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
              class="mr-1 opacity-0 transition-opacity duration-100 group-hover:opacity-100 group-focus-within:opacity-100 group-data-[active=true]:opacity-100"
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
              class="w-8 min-w-8 shrink-0 rounded-none"
              :disabled="files.length >= 20"
              @click="addFile"
            />
          </UTooltip>
        </nav>

        <div
          class="flex shrink-0 items-center gap-0.5 bg-[var(--paste-chrome)] px-[5px] max-[720px]:px-1"
        >
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
          <span
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
            <UColorModeButton
              color="neutral"
              variant="ghost"
              size="sm"
              aria-label="切换颜色模式"
              class="paste-editor-theme"
            />
          </UTooltip>
          <div
            class="grid size-8 place-items-center [&_button]:size-[30px] [&_button]:min-h-[30px] [&_button]:min-w-[30px] [&_button]:p-0 [&_[data-slot=label]]:sr-only"
          >
            <ConsumerAccountControl
              :context-actions="accountActions"
              trigger-mode="collapsed"
            />
          </div>
        </div>
      </header>

      <section
        class="min-h-0 min-w-0 bg-[var(--paste-editor)]"
        aria-label="当前文件"
      >
        <ClientOnly>
          <PasteCodeEditor
            v-model="activeFile.content"
            :language="activeFile.language"
            :label="`${activeFile.path} 代码编辑器`"
          />
          <template #fallback>
            <textarea
              v-model="activeFile.content"
              class="size-full resize-none border-0 bg-[var(--paste-editor)] px-5 py-[18px] font-mono leading-[1.72] text-[var(--paste-ink)] outline-0"
              aria-label="代码编辑器"
            />
          </template>
        </ClientOnly>
      </section>

      <footer
        class="paste-statusbar flex min-w-0 items-center justify-between gap-5 border-t border-[var(--paste-status-border)] bg-[var(--paste-status)] px-[9px] font-mono text-xs text-[var(--paste-status-text-muted)]"
        aria-label="编辑状态"
      >
        <span
          class="flex min-w-0 items-center gap-3 whitespace-nowrap text-[var(--paste-status-text)]"
        >
          <span class="size-1.5 rounded-full bg-[var(--paste-green)]" aria-hidden="true" />
          <h1 id="compose-title" class="font-[inherit] font-semibold">
            {{ editCode ? "编辑片段" : "新建片段" }}
          </h1>
        </span>
        <span
          class="flex min-w-0 items-center justify-end gap-3 whitespace-nowrap max-[720px]:[&>span:first-child]:hidden"
          aria-live="polite"
        >
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
            class="w-auto min-w-0"
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
            class="paste-status-size data-[over-limit=true]:font-bold data-[over-limit=true]:text-[var(--paste-status-error)]"
            :data-over-limit="activeFileBytes > maxFileBytes"
            :title="`当前文件 ${Math.ceil(activeFileBytes / 1024)}/1024 KiB；全部文件 ${Math.ceil(totalBytes / 1024)}/1024 KiB`"
          >{{ Math.ceil(activeFileBytes / 1024) }}/1024 KiB</span>
        </span>
      </footer>

      <USlideover
        v-model:open="shareOpen"
        title="分享代码片段"
        description="设置访问边界，然后生成链接。"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div class="grid gap-[18px] [&_label]:text-[11px] [&_label]:font-[680]">
            <UFormField :error="failure?.fieldErrors.title?.join(' ')" name="title" label="标题">
              <UInput
                v-model="title"
                form="paste-compose-form"
                name="title"
                maxlength="120"
                :placeholder="displayTitle('', activeFile.path)"
                class="w-full"
              />
            </UFormField>
            <UFormField :error="failure?.fieldErrors.description?.join(' ')" name="description" label="说明" hint="可选">
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
            <UFormField :error="failure?.fieldErrors.tags?.join(' ')" name="tags" label="标签" hint="最多 8 个">
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
            <div class="grid grid-cols-2 gap-2.5 max-[720px]:grid-cols-1">
              <UFormField :error="failure?.fieldErrors.visibility?.join(' ')" label="可见性">
                <USelect
                  v-model="visibility"
                  :items="visibilityItems"
                  value-key="value"
                  label-key="label"
                  class="w-full"
                  aria-label="可见性"
                />
              </UFormField>
              <UFormField :error="failure?.fieldErrors.expiry?.join(' ')" label="有效期">
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
              :error="failure?.fieldErrors.password?.join(' ')"
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
            <FailureNotice :feedback="failure" />
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
          <div class="flex w-full items-center justify-end gap-2">
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
              class="border-transparent bg-[var(--paste-blue)] text-white hover:bg-[var(--paste-blue-hover)]"
              :loading="saving || loadingEdit"
              :disabled="saving || loadingEdit"
            />
          </div>
        </template>
      </USlideover>
    </form>
  </section>
</template>
