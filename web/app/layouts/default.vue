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
  <div class="min-h-dvh [&>main]:min-w-0">
    <header
      v-if="!applicationShell"
      class="relative z-30 border-b border-[color-mix(in_srgb,var(--paste-line)_76%,transparent)] bg-[var(--paste-surface)]"
    >
      <nav
        class="flex min-h-12 w-full items-center justify-between gap-4 px-3 max-[720px]:px-2 max-[720px]:[&_a]:min-h-11 max-[720px]:[&_a]:min-w-11 max-[720px]:[&_button]:min-h-11 max-[720px]:[&_button]:min-w-11"
        aria-label="主导航"
      >
        <NuxtLink
          to="/"
          class="inline-flex items-center gap-[7px] text-[15px] font-[730] tracking-[-0.025em] text-[var(--paste-ink)] no-underline"
          :aria-label="`${siteName} 首页`"
        >
          <span
            class="grid size-6 place-items-center rounded-md border border-[color-mix(in_srgb,var(--paste-blue)_28%,var(--paste-line))] bg-[var(--paste-blue-soft)] text-[var(--paste-blue)]"
            aria-hidden="true"
          >
            <UIcon name="i-tabler-code-dots" class="size-4" />
          </span>
          <span>{{ siteName }}</span>
        </NuxtLink>
        <div class="flex items-center gap-0.5">
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
