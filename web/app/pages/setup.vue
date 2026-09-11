<script setup lang="ts">
definePageMeta({ layout: false, middleware: "auth" });
useSeoMeta({ title: "初始化站点" });
const api = usePasteApi();
const { siteName } = useSiteSettings();
const { data: setup, error, refresh } = await useAsyncData("paste-setup", () => api.getSetup());
const saving = ref(false);
const failed = ref(false);
async function claim() {
  if (saving.value || !setup.value?.canClaim) return;
  saving.value = true; failed.value = false;
  try { await api.claimAdministrator(); await refresh(); await navigateTo("/admin"); }
  catch { failed.value = true; await refresh(); }
  finally { saving.value = false; }
}
</script>
<template>
  <main class="min-h-svh bg-default p-4 sm:p-8">
    <header class="mx-auto flex max-w-lg items-center justify-between"><NuxtLink to="/" class="font-semibold">{{ siteName }}</NuxtLink><ConsumerManageAccountControl home-to="/" /></header>
    <section class="mx-auto mt-12 max-w-lg rounded-lg border border-default p-6 space-y-5">
      <h1 class="text-xl font-semibold">初始化站点</h1>
      <UAlert v-if="error || failed" color="error" title="初始化未完成" description="请重试；若已由其他用户完成，请联系本站管理员。"><template #actions><UButton label="重试加载" color="neutral" variant="outline" @click="refresh()" /></template></UAlert>
      <template v-if="setup?.claimed"><p class="text-sm text-muted">本站已完成初始化。</p><UButton to="/admin" label="进入后台" /></template>
      <template v-else><p class="text-sm text-muted">确认后，当前账号将成为本站管理员。</p><UButton :loading="saving" :disabled="!setup?.canClaim" :label="saving ? '初始化中' : '初始化站点并成为管理员'" icon="i-tabler-shield-check" @click="claim" /></template>
    </section>
  </main>
</template>
