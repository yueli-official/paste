<script setup lang="ts">
useSeoMeta({ title: "我的" });
const { loggedIn } = useAuth();
const colorMode = useColorMode();
const appearanceOptions = [
  { label: "跟随系统", value: "system" },
  { label: "浅色", value: "light" },
  { label: "深色", value: "dark" },
];

const { siteName } = useSiteSettings();
const { isAdmin } = usePasteAdministration();
const actions = computed(() => isAdmin.value ? [
  { label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" },
] : []);
</script>

<template>
  <div class="mx-auto my-5 w-full max-w-xl space-y-4 px-4">
    <ConsumerAccountPanel :site-name="siteName" :can-manage="isAdmin" :context-actions="actions" />
    <UFormField v-if="!loggedIn" label="外观" class="rounded-xl border border-default bg-default p-4">
      <USelect v-model="colorMode.preference" :items="appearanceOptions" class="w-full" />
    </UFormField>
  </div>
</template>
