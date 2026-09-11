<script setup lang="ts">
import type {
  AdminNavigationItem,
  AdminSearchGroup,
  AdminShellMessages,
} from "@yueli/ui/admin";

const route = useRoute();
const { siteName } = useSiteSettings();

const section = computed<"pastes" | "users" | "settings">(() => {
  if (route.query.view === "users") return "users";
  if (route.query.view === "settings") return "settings";
  return "pastes";
});
const currentLabel = computed(() => ({
  pastes: "片段管理",
  users: "用户管理",
  settings: "站点设置",
})[section.value]);

const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索控制台",
  searchPlaceholder: "搜索页面与常用操作",
  currentLocation: "当前位置",
};

const navigation = computed<readonly AdminNavigationItem[]>(() => [
  {
    label: "片段管理",
    icon: "i-tabler-files",
    to: "/admin",
    active: section.value === "pastes",
  },
  {
    label: "用户管理",
    icon: "i-tabler-users",
    to: "/admin?view=users",
    active: section.value === "users",
  },
  {
    label: "站点设置",
    icon: "i-tabler-adjustments-horizontal",
    to: "/admin?view=settings",
    active: section.value === "settings",
  },
]);


const searchGroups = computed<readonly AdminSearchGroup[]>(() => [
  {
    id: "paste-admin-pages",
    label: "管理页面",
    items: navigation.value.map((item, index) => ({
      id: `paste-admin-page-${index}`,
      label: item.label,
      icon: item.icon,
      to: item.to,
    })),
  },
]);
</script>

<template>
  <YAdminConsoleLayout
    :navigation="navigation"
    :search-groups="searchGroups"
    :messages="messages"
    storage-key="paste-admin"
    main-id="paste-admin-main"
    :brand-label="siteName"
    brand-icon="i-tabler-code-dots"
    brand-to="/"
    :context-label="siteName"
    :current-label="currentLabel"
    back-to-top-label="返回顶部"
    data-paste-admin-shell
  >
    <template #account="{ collapsed }">
      <ConsumerManageAccountControl
        home-to=""
        show-appearance
        :trigger-mode="collapsed ? 'collapsed' : 'sidebar'"
      />
    </template>
    <slot />
  </YAdminConsoleLayout>
</template>
