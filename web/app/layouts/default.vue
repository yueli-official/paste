<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const { siteName } = useSiteSettings();
const { isAdmin } = useAuth();
const accountActions = computed<readonly AccountMenuAction[]>(() => [
  { label: "我的片段", icon: "i-tabler-folders", to: "/mine" },
  ...(isAdmin.value
    ? [{ label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" }]
    : []),
]);
const route = useRoute();
const applicationShell = computed(() =>
  route.path === "/" || route.path === "/mine" || route.path === "/admin" || route.path.startsWith("/p/"),
);
</script>

<template>
  <div class="paste-page-frame">
    <header v-if="!applicationShell" class="paste-public-header">
      <nav class="paste-public-nav" aria-label="主导航">
        <NuxtLink to="/" class="paste-wordmark" :aria-label="`${siteName} 首页`">
          <span class="paste-wordmark-mark" aria-hidden="true">
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span>{{ siteName }}</span>
        </NuxtLink>
        <div class="paste-nav-actions">
          <UButton
            to="/mine"
            label="我的片段"
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
