<script setup lang="ts">
import { zh_cn } from "@nuxt/ui/locale";
import { FeedbackToastRegion } from "@yueli/ui/feedback/pattern";

const { load, settings } = useSiteSettings();
await callOnce("paste-site-settings", () => load());
useHead(() => ({
  htmlAttrs: { lang: "zh-CN" },
  link: [{ rel: "icon", href: settings.value.faviconUrl || "/favicon.svg" }],
}));
</script>

<template>
  <UApp :locale="zh_cn" :toaster="null">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <FeedbackToastRegion />
  </UApp>
</template>
