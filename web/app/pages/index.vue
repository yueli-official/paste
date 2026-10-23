<script setup lang="ts">
import type { FailureFeedback } from "@yueli/http-runtime";
import { pasteFailureFeedback } from "../utils/pasteFailure";

import type { PasteFile, PasteVisibility } from "../types/paste";
import {
  expiryISO,
  languageFromPath,
  languageItems,
  pasteErrorMessage,
} from "../utils/paste";

interface LocalFile extends PasteFile {
  id: number;
}

const route = useRoute();
const api = usePasteApi();
const libraryVersion = useState("paste-library-version", () => 0);
const transfer = usePasteTransfer();
const { loggedIn, login } = useAuth();
const { siteName, siteDescription } = useSiteSettings();
let nextFileID = 1;
const maxFileBytes = 1024 * 1024;
const maxContentBytes = 1024 * 1024;

const files = ref<LocalFile[]>([
  { id: nextFileID++, path: "untitled.txt", language: "text", content: "" },
]);
const activeFileID = ref(files.value[0]!.id);
const title = ref("");
const titleInput = useTemplateRef<HTMLInputElement>("titleInput");
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
const activeFile = computed(
  () =>
    files.value.find((file) => file.id === activeFileID.value) ||
    files.value[0]!,
);
const totalBytes = computed(() =>
  files.value.reduce((total, file) => total + new Blob([file.content]).size, 0),
);
const activeFileBytes = computed(
  () => new Blob([activeFile.value.content]).size,
);
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
function openShare() {
  shareOpen.value = true;
}

useSeoMeta({
  title: computed(() =>
    editCode.value
      ? `编辑片段 · ${siteName.value}`
      : `新建片段 · ${siteName.value}`,
  ),
  description: "粘贴一段或一组代码，设定访问边界，再分享一个稳定链接。",
  ogTitle: computed(() => siteName.value),
  ogDescription: computed(() => siteDescription.value),
});

function addFile() {
  if (files.value.length >= 20) return;
  let index = files.value.length + 1;
  while (files.value.some((file) => file.path === `untitled-${index}.txt`))
    index++;
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
  dragOverPosition.value =
    event.clientX < bounds.left + bounds.width / 2 ? "before" : "after";
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
}

function dropFile(targetFile: LocalFile, event: DragEvent) {
  const sourceID =
    draggedFileID.value || Number(event.dataTransfer?.getData("text/plain"));
  const sourceIndex = files.value.findIndex((file) => file.id === sourceID);
  let targetIndex = files.value.findIndex((file) => file.id === targetFile.id);
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) {
    clearFileDrag();
    return;
  }
  const [moved] = files.value.splice(sourceIndex, 1);
  if (sourceIndex < targetIndex) targetIndex -= 1;
  const insertIndex =
    targetIndex + (dragOverPosition.value === "after" ? 1 : 0);
  files.value.splice(insertIndex, 0, moved!);
  activeFileID.value = moved!.id;
  clearFileDrag();
}

function inferLanguage(file: LocalFile) {
  file.language = languageFromPath(file.path);
}

function changeLanguage(language: string) {
  const file = activeFile.value;
  if (file.language === language) return;
  const selected = languageItems.find((item) => item.value === language);
  if (!selected) return;
  const separator = Math.max(file.path.lastIndexOf("/"), file.path.lastIndexOf("\\"));
  const dot = file.path.lastIndexOf(".");
  const stem = dot > separator + 1 ? file.path.slice(0, dot) : file.path;
  let path = `${stem}.${selected.extension}`;
  let suffix = 2;
  while (files.value.some((other) => other.id !== file.id && other.path.toLowerCase() === path.toLowerCase())) {
    path = `${stem}-${suffix++}.${selected.extension}`;
  }
  file.path = path;
  file.language = language;
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
  if (files.value.some((file) => !file.path.trim()))
    return "每个文件都需要一个文件名。";
  if (
    files.value.some((file) => {
      const path = file.path.trim().replaceAll("\\", "/");
      return path.startsWith("/") || path.split("/").includes("..");
    })
  )
    return "文件名需要使用片段内的相对路径，不能包含上级目录。";
  if (
    new Set(files.value.map((file) => file.path.trim().toLowerCase())).size !==
    files.value.length
  ) {
    return "文件名不能重复。";
  }
  if (files.value.every((file) => !file.content))
    return "请先粘贴要分享的内容。";
  const oversized = files.value.find(
    (file) => new Blob([file.content]).size > maxFileBytes,
  );
  if (oversized)
    return `“${oversized.path}”超过文件限制；单个文件不能超过 1 MiB。`;
  if (totalBytes.value > maxContentBytes) return "全部文件合计不能超过 1 MiB。";
  if (tags.value.some((tag) => [...tag.trim()].length > 32))
    return "每个标签不能超过 32 个字符。";
  if (
    password.value &&
    ([...password.value].length < 8 || [...password.value].length > 128)
  ) {
    return "访问密码需要包含 8–128 个字符。";
  }
  if (visibility.value === "private" && !loggedIn.value)
    return "私有片段需要先登录。";
  return "";
}

async function requestShare() {
  if (editCode.value) return submit();
  error.value = validate();
  shareOpen.value = true;
}

async function submit() {
  if (
    saving.value ||
    loadingEdit.value ||
    (editCode.value && !editRevision.value)
  )
    return;
  failure.value = null;
  error.value = validate();
  if (error.value) return;
  saving.value = true;
  const input = {
    title: title.value.trim(),
    description: description.value.trim(),
    tags: tags.value.map((tag) => tag.trim()).filter(Boolean),
    files: files.value.map(({ path, language, content }) => ({
      path: path.trim(),
      language,
      content,
    })),
    visibility: visibility.value,
    password: password.value,
    expiresAt: expiryISO(expiry.value),
    clearExpiry: Boolean(editCode.value) && expiry.value === "never",
  };
  try {
    const response =
      editCode.value && editRevision.value
        ? await api.update(editCode.value, editRevision.value, input)
        : await api.create(input);
    transfer.value[response.paste.code] = response.paste;
    libraryVersion.value++;
    savedSnapshot.value = snapshot();
    await navigateTo(
      `/p/${response.paste.code}?${editCode.value ? "updated" : "created"}=1`,
    );
  } catch (caught) {
    failure.value = pasteFailureFeedback(caught, "分享未完成，请检查后重试。", {
      "/title": "title",
      "/description": "description",
      "/tags": "tags",
      "/password": "password",
      "/visibility": "visibility",
      "/expiresAt": "expiry",
    });
    if (failure.value.fieldErrors.title) {
      shareOpen.value = false;
      await nextTick();
      titleInput.value?.focus();
    }
  } finally {
    saving.value = false;
  }
}

async function loadEdit() {
  if (!editCode.value) return;
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

const markdownMode = ref<"source" | "rich">("source");
const viewMode = ref<"edit" | "split" | "preview">("split");
const wideScreen = ref(false);
const markdownView = computed(() => viewMode.value === "split" && !wideScreen.value ? "edit" : viewMode.value);
const preview = computed(() => activeFile.value.language === "markdown" && markdownView.value === "preview");
let splitMedia: MediaQueryList | undefined;
function updateSplitWidth() { wideScreen.value = splitMedia?.matches ?? false; }
onMounted(() => {
  splitMedia = window.matchMedia("(min-width: 1100px)");
  updateSplitWidth();
  splitMedia.addEventListener("change", updateSplitWidth);
});
onBeforeUnmount(() => splitMedia?.removeEventListener("change", updateSplitWidth));
const wrapping = ref(true);
const copied = ref(false);
const copyError = ref("");
const richConfirm = ref(false);
const hasContent = computed(() =>
  files.value.some((file) => file.content.length > 0),
);
const sharingAccess = computed(() => visibility.value === "private" ? "仅自己" : password.value ? "密码访问" : "链接访问");
const sharingExpiry = computed(() => expiry.value === "keep" ? "原有效期" : expiry.value === "never" ? "不过期" : expiryItems.value.find((item) => item.value === expiry.value)?.label);
const sharingLabel = computed(() => `${sharingAccess.value} · ${sharingExpiry.value}`);
const modeItems = computed(() => [
  {
    label: "排版编辑",
    icon: "i-tabler-typography",
    onSelect: () => chooseRich(),
  },
  {
    label: "源码编辑",
    icon: "i-tabler-code",
    onSelect: () => {
      markdownMode.value = "source";
      viewMode.value = "split";
    },
  },
]);
function chooseRich() {
  if (activeFile.value.content && markdownMode.value !== "rich")
    richConfirm.value = true;
  else {
    markdownMode.value = "rich";
    viewMode.value = "split";
  }
}
async function copySource() {
  try {
    await copyPasteText(activeFile.value.content);
    copied.value = true;
    copyError.value = "";
  } catch {
    copyError.value = "复制失败，请选择内容后手动复制。";
  }
}
watch([activeFileID, () => activeFile.value.content], () => {
  copied.value = false;
  copyError.value = "";
});
watch(activeFileID, () => {
  viewMode.value = "split";
  markdownMode.value = "source";
});
function snapshot() {
  return JSON.stringify({
    files: files.value,
    title: title.value,
    description: description.value,
    tags: tags.value,
    visibility: visibility.value,
    password: password.value,
    expiry: expiry.value,
  });
}
const savedSnapshot = ref(snapshot());
const dirty = computed(() => snapshot() !== savedSnapshot.value);
function resetDraft() {
  if (dirty.value && !window.confirm("当前内容尚未保存，确定放弃并新建？"))
    return;
  files.value = [
    { id: nextFileID++, path: "untitled.txt", language: "text", content: "" },
  ];
  activeFileID.value = files.value[0]!.id;
  title.value = "";
  description.value = "";
  tags.value = [];
  visibility.value = "unlisted";
  password.value = "";
  expiry.value = "7d";
  editRevision.value = undefined;
  error.value = "";
  failure.value = null;
  savedSnapshot.value = snapshot();
  if (editCode.value) navigateTo("/");
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value) {
    event.preventDefault();
    event.returnValue = "";
  }
}
onBeforeRouteLeave(
  () => !dirty.value || window.confirm("当前内容尚未保存，确定离开？"),
);
onMounted(async () => {
  await loadEdit();
  savedSnapshot.value = snapshot();
  window.addEventListener("beforeunload", beforeUnload);
});
onBeforeUnmount(() => window.removeEventListener("beforeunload", beforeUnload));
watch(editCode, async () => {
  if (editCode.value) {
    await loadEdit();
    savedSnapshot.value = snapshot();
  }
});
</script>

<template>
  <PasteWorkspace
    :current-title="title || '未命名内容'"
    :current-status="editCode ? (dirty ? '有未保存修改' : '已保存') : '未分享'"
    @create="resetDraft"
  >
    <form
      id="paste-compose-form"
      class="paste-compose paste-workbench h-full flex flex-col min-w-0 overflow-hidden"
      :aria-busy="saving || loadingEdit"
      @submit.prevent="requestShare"
    >
      <header
        class="paste-command-row min-h-[88px] [padding:20px_30px] flex items-center gap-5 [border-bottom:1px_solid_var(--paste-line)] max-[1100px]:flex-wrap max-[1100px]:gap-3 max-[1100px]:[padding:18px_22px] max-[760px]:p-4 max-[760px]:min-h-[76px]"
      >
        <h1 id="compose-title" class="sr-only">
          {{ editCode ? "编辑内容" : "新建内容" }}
        </h1>
        <input
          ref="titleInput"
          v-model="title"
          :aria-invalid="Boolean(failure?.fieldErrors.title)"
          :aria-describedby="failure?.fieldErrors.title ? 'title-error' : undefined"
          class="paste-title-input flex-1 [width:120px] min-w-0 [padding:5px_0] border-0 outline-none bg-transparent text-[21px] font-semibold [letter-spacing:-.035em] text-[var(--paste-ink)] placeholder:text-[var(--paste-ink-dim)] placeholder:font-medium focus:[box-shadow:0_1px_var(--paste-blue)] max-[1100px]:w-full max-[1100px]:basis-full max-[760px]:text-lg"
          name="title"
          maxlength="120"
          placeholder="添加标题（可选）"
          aria-label="标题"
        />
        <div
          class="paste-share-actions flex items-center gap-3 max-[1100px]:w-full max-[1100px]:justify-between max-[760px]:gap-2.5 max-[760px]:items-center"
        >
          <button
            type="button"
            class="paste-sharing-trigger inline-flex items-center gap-2.5 rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface)] px-3 min-h-10 text-xs text-[var(--paste-ink-soft)] transition-colors hover:border-[var(--paste-ink-dim)] hover:text-[var(--paste-ink)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--paste-blue)] max-[760px]:gap-2 max-[760px]:px-2.5"
            aria-label="打开分享设置"
            :aria-description="sharingLabel"
            :aria-expanded="shareOpen"
            aria-haspopup="dialog"
            @click="openShare"
          >
            <UIcon name="i-tabler-adjustments-horizontal" class="size-4 shrink-0" />
            <span>{{ sharingAccess }}</span>
            <span class="h-3 w-px bg-[var(--paste-line)]" aria-hidden="true" />
            <span class="text-[var(--paste-ink-dim)]">{{ sharingExpiry }}</span>
          </button>
          <UButton
            type="submit"
            :label="editCode ? '保存修改' : '生成链接'"
            :aria-label="editCode ? '保存修改' : '生成分享链接'"
            trailing-icon="i-tabler-arrow-up-right"
            class="paste-primary bg-[var(--paste-blue)] text-[#17200e] min-h-10 [padding:9px_18px] rounded-[7px] font-[650] whitespace-nowrap justify-center not-dark:text-white hover:bg-[var(--paste-blue-hover)] disabled:opacity-35 max-[760px]:min-h-10 max-[760px]:text-xs max-[760px]:[padding:8px_12px]"
            :loading="saving"
            :disabled="
              !hasContent || loadingEdit || (!!editCode && !editRevision)
            "
          />
        </div>
      </header>
      <div
        v-if="error || failure"
        class="paste-inline-feedback [padding:10px_28px] text-[13px] bg-[var(--paste-red-soft)] text-[var(--paste-red)]"
      >
        <FailureNotice :feedback="failure" />
        <p v-if="failure?.fieldErrors.title" id="title-error" role="alert">
          {{ failure.fieldErrors.title.join(" ") }}
        </p>
        <p v-if="error" role="alert">{{ error }}</p>
      </div>
      <div
        class="paste-editor-toolbar min-h-[58px] flex items-stretch justify-between gap-3 [border-bottom:1px_solid_var(--paste-line)] [padding:0_22px] max-[1100px]:flex-wrap max-[1100px]:gap-0 max-[760px]:[padding:0_10px]"
      >
        <nav
          class="paste-file-tabs flex flex-1 min-w-0 overflow-x-auto [scrollbar-width:thin] max-[1100px]:basis-full"
          aria-label="片段文件"
        >
          <div
            v-for="(file, index) in files"
            :key="file.id"
            class="paste-file-tab group relative flex items-center [min-width:132px] [max-width:230px] text-[var(--paste-ink-dim)] [&[data-active=true]]:text-[var(--paste-ink)] [&[data-active=true]]:[box-shadow:inset_0_-2px_var(--paste-blue)] [&[data-dragging=true]]:[opacity:.4] [&[data-drop-position=before]]:[border-left:2px_solid_var(--paste-blue)] [&[data-drop-position=after]]:[border-right:2px_solid_var(--paste-blue)] [&_.paste-file-tab-button]:text-xs [&_.paste-file-tab-button]:min-h-[54px]"
            :data-active="activeFileID === file.id"
            :data-dragging="draggedFileID === file.id"
            :data-drop-position="
              dragOverFileID === file.id ? dragOverPosition : undefined
            "
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
              @keydown.enter.prevent="commitRename(file)"
              @keyup.esc="cancelRename"
            />
            <button
              v-else
              type="button"
              class="paste-file-tab-button flex h-full min-w-0 flex-1 items-center gap-[7px] border-0 bg-transparent py-0 pr-2 pl-[11px] text-left font-mono text-[11px] text-inherit focus-visible:outline-offset-[-3px]"
              :aria-pressed="activeFileID === file.id"
              :aria-label="`${file.path || `文件 ${index + 1}`}，双击重命名，Alt 加方向键移动`"
              @click="() => { activeFileID = file.id; }"
              @dblclick.prevent="beginRename(file)"
              @keydown.alt.left.prevent="moveFile(file, -1)"
              @keydown.alt.right.prevent="moveFile(file, 1)"
            >
              <UIcon name="i-tabler-file-code" class="size-4 shrink-0" />
              <span class="truncate leading-6 py-px">{{
                file.path || `文件 ${index + 1}`
              }}</span>
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
          class="paste-editor-tools flex items-center gap-[5px] shrink-0 max-[1100px]:[padding:8px_0] max-[1100px]:[min-height:48px] max-[1100px]:w-full max-[1100px]:justify-end max-[1100px]:[border-top:1px_solid_var(--paste-line)] max-[760px]:justify-start max-[760px]:gap-0.5 max-[760px]:overflow-x-auto max-[760px]:[&_>_*]:shrink-0 max-[760px]:[&_button]:min-h-[38px]"
        >
          <template v-if="activeFile.language === 'markdown'">
            <UDropdownMenu :items="modeItems">
              <UButton :label="markdownMode === 'rich' ? '排版' : '源码'" trailing-icon="i-tabler-chevron-down" color="neutral" variant="ghost" size="sm" aria-label="选择 Markdown 编辑方式" />
            </UDropdownMenu>
            <div class="paste-view-controls flex items-center rounded-md bg-[var(--paste-surface)] p-0.5" role="group" aria-label="Markdown 布局">
              <UButton label="编辑" color="neutral" :variant="markdownView === 'edit' ? 'soft' : 'ghost'" size="xs" :aria-pressed="markdownView === 'edit'" @click="() => { viewMode = 'edit'; }" />
              <UButton v-if="wideScreen" label="双栏" icon="i-tabler-layout-columns" color="neutral" :variant="markdownView === 'split' ? 'soft' : 'ghost'" size="xs" :aria-pressed="markdownView === 'split'" @click="() => { viewMode = 'split'; }" />
              <UButton label="预览" color="neutral" :variant="markdownView === 'preview' ? 'soft' : 'ghost'" size="xs" :aria-pressed="markdownView === 'preview'" @click="() => { viewMode = 'preview'; }" />
            </div>
          </template>
          <USelectMenu
            :model-value="activeFile.language"
            @update:model-value="changeLanguage"
            :items="languageOptions"
            value-key="value"
            label-key="label"
            :search-input="{ placeholder: '搜索语言' }"
            aria-label="代码语言"
            variant="ghost"
            class="paste-language-select [min-width:110px] [width:124px] max-[760px]:[min-width:94px] max-[760px]:[width:125px]"
            size="sm"
          />
          <UButton
            v-if="
              !preview &&
              (activeFile.language !== 'markdown' || markdownMode === 'source')
            "
            icon="i-tabler-text-wrap"
            color="neutral"
            :variant="wrapping ? 'soft' : 'ghost'"
            size="sm"
            aria-label="自动换行"
            :aria-pressed="wrapping"
            @click="() => { wrapping = !wrapping; }"
          />
          <UButton
            :icon="copied ? 'i-tabler-check' : 'i-tabler-copy'"
            :aria-label="copied ? '已复制源码' : '复制源码'"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="!activeFile.content"
            @click="copySource"
            ><span class="hidden min-[761px]:inline">{{
              copied ? "已复制" : "复制源码"
            }}</span></UButton
          >
        </div>
      </div>
      <section
        class="paste-editor-area flex-1 min-h-0 min-w-0 overflow-hidden"
        aria-label="当前文件"
      >
        <p v-if="loadingEdit" class="p-8 text-muted" role="status">
          正在加载内容…
        </p>
        <div v-else class="paste-editing-panes h-full min-h-0" :class="{ 'paste-editing-panes-split grid grid-cols-2': activeFile.language === 'markdown' && markdownView === 'split' }">
          <div v-show="!preview" class="paste-source-pane flex flex-col min-h-0 min-w-0 h-full overflow-hidden" aria-label="编辑区">
            <div v-if="activeFile.language === 'markdown' && markdownView === 'split'" class="paste-pane-heading flex items-center shrink-0 h-9 px-6 border-b border-[var(--paste-line)] text-[var(--paste-ink-dim)] text-[11px]">{{ markdownMode === 'rich' ? '排版' : 'Markdown' }}</div>
            <div class="paste-pane-body min-h-0 flex-1 overflow-auto">
              <LazyContentEditor
                v-if="activeFile.language === 'markdown' && markdownMode === 'rich'"
                :key="activeFile.id"
                v-model="activeFile.content"
                class="paste-rich-editor max-w-[880px] px-6 py-5 m-auto [--content-editor-toolbar-top:-20px] max-[760px]:px-2.5 max-[760px]:py-3 max-[760px]:[--content-editor-toolbar-top:-12px]"
                :draft-enabled="false"
                placeholder="开始写作，或粘贴 Markdown…"
              />
              <ClientOnly v-else>
                <PasteCodeEditor :key="activeFile.id" v-model="activeFile.content" :language="activeFile.language" :line-wrapping="wrapping" autofocus placeholder="在这里粘贴或输入内容…" :label="`${activeFile.path} 代码编辑器`" />
                <template #fallback><textarea v-model="activeFile.content" class="paste-editor-fallback w-full h-full p-6 bg-[var(--paste-editor)] resize-none" aria-label="代码编辑器" /></template>
              </ClientOnly>
            </div>
          </div>
          <div v-if="activeFile.language === 'markdown' && markdownView !== 'edit'" :class="{ 'border-l border-[var(--paste-line)]': markdownView === 'split' }" class="paste-preview-pane flex flex-col min-h-0 min-w-0 h-full" aria-label="Markdown 预览">
            <div v-if="markdownView === 'split'" class="paste-pane-heading flex items-center shrink-0 h-9 px-6 border-b border-[var(--paste-line)] text-[var(--paste-ink-dim)] text-[11px]">预览</div>
            <div class="paste-document-scroll min-h-0 flex-1 overflow-auto">
              <div class="paste-document max-w-[840px] mx-auto px-8 py-8 text-[15px] leading-[1.85] [overflow-wrap:anywhere] max-[760px]:px-5 max-[760px]:py-6">
                <LazyContentProse v-if="activeFile.content" :content="activeFile.content" />
              </div>
            </div>
          </div>
        </div>
      </section>
      <footer
        class="paste-statusbar flex justify-between gap-3 [padding:10px_24px] min-h-[38px] [border-top:1px_solid_var(--paste-line)] text-[var(--paste-ink-dim)] text-[10px] bg-[var(--paste-editor)] max-[760px]:[padding:9px_14px] max-[760px]:[font-size:9px] max-[760px]:gap-2 max-[760px]:[&_>_span:last-child]:gap-2"
        aria-label="编辑状态"
      >
        <span v-if="copyError" role="alert">{{ copyError }}</span
        ><span v-else>{{
          loggedIn
            ? dirty
              ? "尚未保存"
              : "准备就绪"
            : "当前浏览器可管理"
        }}</span>
        <span class="flex items-center gap-4"
          ><span>{{ files.length }}/20 文件</span
          ><span
            class="paste-status-size [&[data-over-limit=true]]:text-[var(--paste-red)]"
            :data-over-limit="totalBytes > maxContentBytes"
            >{{ Math.ceil(totalBytes / 1024) }}/1024 KiB</span
          ><span>UTF-8</span></span
        >
      </footer>
      <USlideover
        v-model:open="shareOpen"
        :title="editCode ? '分享设置' : '生成分享链接'"
        :description="editCode ? undefined : '有效期从生成链接时开始。'"
        :ui="{ content: 'sm:max-w-sm' }"
      >
        <template #body>
          <div
            class="grid gap-[18px] [&_label]:text-[11px] [&_label]:font-[680]"
          >
            <div class="grid grid-cols-2 gap-2.5 max-[720px]:grid-cols-1">
              <UFormField
                :error="failure?.fieldErrors.visibility?.join(' ')"
                label="可见性"
              >
                <USelect
                  v-model="visibility"
                  :items="visibilityItems"
                  value-key="value"
                  label-key="label"
                  class="w-full"
                  aria-label="可见性"
                />
              </UFormField>
              <UFormField
                :error="failure?.fieldErrors.expiry?.join(' ')"
                label="有效期"
              >
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
                class="w-full"
              />
            </UFormField>
            <UFormField
              :error="failure?.fieldErrors.description?.join(' ')"
              name="description"
              label="说明"
              hint="可选"
            >
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
            <UFormField
              :error="failure?.fieldErrors.tags?.join(' ')"
              name="tags"
              label="标签"
              hint="最多 8 个"
            >
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
              description="可在此浏览器编辑、删除。Cookie 清除或访客会话到期后失效；登录可保存到账户。"
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
              type="button"
              @click="submit"
              :label="editCode ? '保存修改' : '生成分享链接'"
              :icon="editCode ? 'i-tabler-device-floppy' : 'i-tabler-send'"
              class="paste-primary bg-[var(--paste-blue)] text-[#17200e] min-h-10 [padding:9px_18px] rounded-[7px] font-[650] whitespace-nowrap justify-center not-dark:text-white hover:bg-[var(--paste-blue-hover)] disabled:opacity-35 max-[760px]:min-h-10 max-[760px]:text-xs max-[760px]:[padding:8px_12px]"
              :loading="saving || loadingEdit"
              :disabled="
                saving ||
                loadingEdit ||
                !hasContent ||
                (!!editCode && !editRevision)
              "
            />
          </div>
        </template>
      </USlideover>
      <UModal
        v-model:open="richConfirm"
        title="切换到排版编辑"
        description="排版编辑会在实际修改内容时重新生成 Markdown。HTML、自定义语法和原始空白格式可能改变；需要逐字保留时请继续使用源码编辑。"
      >
        <template #footer
          ><div class="flex justify-end gap-2 w-full">
            <UButton
              label="保留源码编辑"
              color="neutral"
              variant="ghost"
              @click="() => { richConfirm = false; }"
            /><UButton
              label="使用排版编辑"
              @click="
                markdownMode = 'rich';
                viewMode = 'split';
                richConfirm = false;
              "
            /></div
        ></template>
      </UModal>
    </form>
  </PasteWorkspace>
</template>
