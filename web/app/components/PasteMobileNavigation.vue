<script setup lang="ts">
import { MobileBottomNav, type MobileBottomNavItem } from "@yueli/ui/navigation/mobile-bottom-nav";

const route = useRoute();
const mobileNavigation = computed<MobileBottomNavItem[]>(() => [
  { id: "home", label: "首页", icon: "i-tabler-home", to: "/", active: route.path === "/" || route.path.startsWith("/p/") },
  { id: "browse", label: "内容库", icon: "i-tabler-folders", to: "/mine", active: route.path === "/mine" },
  { id: "account", label: "我的", icon: "i-tabler-user", to: "/me", active: route.path === "/me" },
]);
</script>
<template>
  <MobileBottomNav :items="mobileNavigation" label="移动主导航" />
</template>
