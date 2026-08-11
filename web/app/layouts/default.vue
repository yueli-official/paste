<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const accountActions: readonly AccountMenuAction[] = [
  {
    label: "我的 Paste",
    icon: "i-tabler-folders",
    to: "/mine",
  },
];
const route = useRoute();
const applicationShell = computed(() =>
  route.path === "/" || route.path === "/mine" || route.path.startsWith("/p/"),
);
</script>

<template>
  <div class="paste-page-frame">
    <header v-if="!applicationShell" class="paste-public-header">
      <nav class="paste-public-nav" aria-label="主导航">
        <NuxtLink to="/" class="paste-wordmark" aria-label="Paste 首页">
          <span class="paste-wordmark-mark" aria-hidden="true">
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span>Paste</span>
        </NuxtLink>
        <div class="paste-nav-actions">
          <UButton
            to="/mine"
            label="我的 Paste"
            icon="i-tabler-folders"
            color="neutral"
            variant="ghost"
            class="hidden min-h-9 sm:inline-flex"
          />
          <UTooltip text="切换颜色模式">
            <UColorModeButton
              color="neutral"
              variant="ghost"
              aria-label="切换颜色模式"
            />
          </UTooltip>
          <ConsumerAccountControl :context-actions="accountActions" />
        </div>
      </nav>
    </header>
    <main><slot /></main>
  </div>
</template>
