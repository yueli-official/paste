<script setup lang="ts">
const open = defineModel<boolean>("open", { default: false });
defineProps<{
  titles: string[];
  busy: boolean;
  error?: string;
}>();
defineEmits<{ confirm: [] }>();
const cancelButton = useTemplateRef<{ $el: HTMLElement }>("cancelButton");
function cancel() { open.value = false; }
function focusCancel(event: Event) {
  event.preventDefault();
  nextTick(() => cancelButton.value?.$el.focus());
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="titles.length > 1 ? `删除 ${titles.length} 项内容？` : '删除此内容？'"
    description="删除后无法恢复，分享链接将立即失效。"
    :dismissible="!busy"
    :close="!busy"
    :content="{ onOpenAutoFocus: focusCancel }"
    :ui="{ content: 'sm:max-w-md', footer: 'justify-end' }"
  >
    <template #body>
      <ul class="max-h-48 space-y-2 overflow-y-auto rounded-lg border border-[var(--paste-line)] bg-[var(--paste-surface)] p-3 text-sm text-[var(--paste-ink)]">
        <li v-for="(title, index) in titles" :key="index" class="flex items-start gap-2">
          <UIcon name="i-tabler-file-text" class="mt-0.5 size-4 shrink-0 text-[var(--paste-ink-dim)]" />
          <span class="min-w-0 break-words [overflow-wrap:anywhere]">{{ title }}</span>
        </li>
      </ul>
      <UAlert v-if="error" class="mt-3" color="error" variant="subtle" icon="i-tabler-alert-circle" title="删除未完成" :description="error" />
    </template>
    <template #footer>
      <UButton ref="cancelButton" label="取消" color="neutral" variant="outline" :disabled="busy" @click="cancel" />
      <UButton :label="error ? '重试删除' : '确认删除'" color="error" icon="i-tabler-trash" :loading="busy" :disabled="busy || titles.length === 0" @click="$emit('confirm')" />
    </template>
  </UModal>
</template>
