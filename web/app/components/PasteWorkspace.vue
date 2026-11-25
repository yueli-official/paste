<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";
import type { PasteSummary } from "../types/paste";
import { displayTitle, pasteProblemCode } from "../utils/paste";

withDefaults(defineProps<{ currentTitle?: string; currentStatus?: string }>(), {
  currentTitle: "",
  currentStatus: "未分享",
});
const emit = defineEmits<{ create: [] }>();
const route = useRoute();
const { loggedIn, login } = useAuth();
const { settings } = useSiteSettings();
const { isAdmin } = usePasteAdministration();
const api = usePasteApi();
const libraryVersion = useState("paste-library-version", () => 0);
const navigationOpen = ref(false);
const query = ref("");
const recent = ref<PasteSummary[]>([]);
const loading = ref(false);
const loadError = ref(false);
const accountActions = computed<readonly AccountMenuAction[]>(() => [
  { label: "内容库", icon: "i-tabler-folders", to: "/mine" },
  ...(isAdmin.value
    ? [{ label: "管理后台", icon: "i-tabler-shield-cog", to: "/admin" }]
    : []),
]);
async function loadRecent() {
  recent.value = [];
  loading.value = true;
  loadError.value = false;
  try {
    recent.value = (await api.listMine({ size: 8 })).items;
  } catch (caught) {
    loadError.value = loggedIn.value || pasteProblemCode(caught) !== "paste.not_authenticated";
  } finally {
    loading.value = false;
  }
}
function create() {
  navigationOpen.value = false;
  if (route.path === "/") emit("create");
  else navigateTo("/");
}
function search() {
  navigationOpen.value = false;
  navigateTo({
    path: "/mine",
    query: query.value.trim() ? { q: query.value.trim() } : {},
  });
}
onMounted(loadRecent);
watch(loggedIn, loadRecent);
watch(libraryVersion, loadRecent);
watch(
  () => route.fullPath,
  () => {
    navigationOpen.value = false;
  },
);
</script>

<template>
  <div
    class="paste-app flex h-dvh flex-col min-h-[420px] overflow-hidden bg-[var(--paste-page)] max-[760px]:min-h-[420px]"
  >
    <header
      class="paste-global-header h-16 shrink-0 [padding:0_28px] flex items-center justify-between [border-bottom:1px_solid_var(--paste-line)] bg-[var(--paste-page)] max-[760px]:h-[58px] max-[760px]:[padding:0_16px_0_8px]"
    >
      <div class="flex min-w-0 items-center gap-3">
        <UButton
          class="paste-mobile-menu hidden max-[760px]:inline-flex"
          icon="i-tabler-menu-2"
          color="neutral"
          variant="ghost"
          aria-label="打开内容导航"
          @click="() => { navigationOpen = true; }"
        />
        <NuxtLink
          to="/"
          class="paste-wordmark [font:600_36px/.9_Georgia,_Times_New_Roman,_serif] [letter-spacing:-1.8px] text-[var(--paste-ink)] no-underline [&_span]:text-[var(--paste-blue)] max-[760px]:text-[31px]"
          aria-label="Paste 首页"
        >
          <img v-if="settings.logoUrl" :src="settings.logoUrl" alt="" width="32" height="32" class="mr-2 inline-block size-8 object-contain align-middle" />
          paste<span>.</span>
        </NuxtLink>
        <span
          v-if="settings.description"
          class="paste-editor-brand-label paste-brand-caption ml-4 pl-5 [border-left:1px_solid_var(--paste-line)] truncate max-w-[min(52vw,640px)] text-xs text-[var(--paste-ink-dim)] max-[760px]:hidden"
          >{{ settings.description }}</span
        >
      </div>
      <div class="flex items-center gap-3">
        <UColorModeButton
          class="hidden md:inline-flex"
          color="neutral"
          variant="ghost"
          aria-label="切换颜色模式"
        />
        <YSiteLauncher />
        <ConsumerAccountControl
          class="hidden md:block"
          :context-actions="accountActions"
          trigger-mode="collapsed"
        />
      </div>
    </header>
    <div
      class="paste-app-body grid [grid-template-columns:248px_minmax(0,1fr)] min-h-0 flex-1 max-[1100px]:[grid-template-columns:218px_minmax(0,1fr)] max-[760px]:block"
    >
      <button
        v-if="navigationOpen"
        class="paste-sidebar-backdrop max-[760px]:fixed max-[760px]:[inset:58px_0_0] max-[760px]:[background:#0008] max-[760px]:z-[39]"
        aria-label="关闭内容导航"
        @click="() => { navigationOpen = false; }"
      />
      <aside
        class="paste-sidebar flex flex-col gap-3.5 min-h-0 overflow-y-auto [padding:26px_18px_18px] [border-right:1px_solid_var(--paste-line)] bg-[var(--paste-page)] max-[760px]:hidden max-[760px]:data-[open=true]:flex max-[760px]:data-[open=true]:fixed max-[760px]:data-[open=true]:top-[58px] max-[760px]:data-[open=true]:bottom-0 max-[760px]:data-[open=true]:left-0 max-[760px]:data-[open=true]:[width:min(290px,85vw)] max-[760px]:data-[open=true]:z-40 max-[760px]:data-[open=true]:[box-shadow:16px_0_60px_#0008]"
        :data-open="navigationOpen"
        aria-label="内容导航"
      >
        <div class="flex items-center gap-2">
          <UButton
            label="新建"
            icon="i-tabler-plus"
            color="neutral"
            variant="outline"
            class="paste-new-button justify-center w-full min-h-[42px] text-sm bg-[var(--paste-surface)] [&_>_span:first-child]:text-[var(--paste-blue)]"
            @click="create"
          />
          <UButton
            class="paste-mobile-menu hidden max-[760px]:inline-flex"
            icon="i-tabler-x"
            color="neutral"
            variant="ghost"
            aria-label="关闭内容导航"
            @click="() => { navigationOpen = false; }"
          />
        </div>
        <form @submit.prevent="search">
          <UInput
            v-model="query"
            icon="i-tabler-search"
            placeholder="搜索内容库"
            aria-label="搜索内容库"
            class="w-full"
          />
        </form>
        <div
          v-if="currentTitle"
          class="paste-current-section [margin:14px_-8px_0]"
        >
          <div
            class="paste-section-label block text-[var(--paste-ink-dim)] text-xs font-medium [padding:0_10px]"
          >
            当前{{ route.path === "/" ? "编辑" : "阅读" }}
          </div>
          <div
            class="paste-current-item flex items-start gap-3 [padding:16px_14px] mt-3 rounded-[7px] [background:linear-gradient(100deg,var(--paste-blue-soft),transparent)] [box-shadow:inset_3px_0_var(--paste-blue)] text-sm [&_>_span]:mt-0.5"
          >
            <UIcon name="i-tabler-file-text" class="size-5 shrink-0" />
            <div class="min-w-0">
              <div class="truncate font-medium">{{ currentTitle }}</div>
              <div
                class="paste-item-meta flex gap-1.5 items-center text-[var(--paste-ink-dim)] mt-1.5 text-[11px]"
              >
                <span
                  class="paste-status-dot [width:5px] [height:5px] bg-[var(--paste-ink-dim)] rounded-full inline-block shrink-0"
                />{{ currentStatus }}
              </div>
            </div>
          </div>
        </div>
        <div
          class="paste-library-heading flex items-center justify-between [margin:16px_0_0] text-[var(--paste-ink-dim)] [&_.paste-section-label]:[padding:0]"
        >
          <NuxtLink
            to="/mine"
            class="paste-section-label block text-[var(--paste-ink-dim)] text-xs font-medium [padding:0_10px]"
            >内容库</NuxtLink
          ><UIcon name="i-tabler-folders" class="size-4" />
        </div>
        <p
          v-if="loading"
          class="paste-sidebar-empty text-[var(--paste-ink-dim)] text-[13px] [line-height:1.8] [padding:10px_0]"
          role="status"
        >
          正在读取内容库…
        </p>
        <div
          v-else-if="loadError"
          class="paste-sidebar-empty text-[var(--paste-ink-dim)] text-[13px] [line-height:1.8] [padding:10px_0]"
        >
          <p>内容库暂时无法加载</p>
          <UButton label="重试" variant="link" @click="loadRecent" />
        </div>
        <p
          v-else-if="!recent.length"
          class="paste-sidebar-empty text-[var(--paste-ink-dim)] text-[13px] [line-height:1.8] [padding:10px_0]"
        >
          你创建的内容会显示在这里。
        </p>
        <nav
          v-else
          class="paste-recent-list min-w-0 [margin:0_-8px] grid grid-cols-[minmax(0,1fr)] gap-1"
          aria-label="最近内容"
        >
          <NuxtLink
            v-for="item in recent"
            :key="item.code"
            :to="`/p/${item.code}`"
            class="paste-recent-item min-w-0 flex gap-3 items-center [padding:12px_10px] rounded-md text-[13px] text-[var(--paste-ink-soft)] hover:text-[var(--paste-ink)] hover:bg-[var(--paste-surface)] [&[aria-current=page]]:text-[var(--paste-ink)] [&[aria-current=page]]:bg-[var(--paste-surface)]"
            :aria-current="route.params.code === item.code ? 'page' : undefined"
          >
            <UIcon
              :name="
                item.primaryLanguage === 'markdown'
                  ? 'i-tabler-file-text'
                  : 'i-tabler-code'
              "
              class="size-5 shrink-0"
            />
            <div class="min-w-0">
              <div class="truncate">{{ displayTitle(item.title) }}</div>
              <div
                class="paste-item-meta flex gap-1.5 items-center text-[var(--paste-ink-dim)] mt-1.5 text-[11px]"
              >
                {{ item.fileCount }} 个文件 ·
                {{
                  new Date(item.updatedAt).toLocaleDateString("zh-CN", {
                    month: "short",
                    day: "numeric",
                  })
                }}
              </div>
            </div>
          </NuxtLink>
        </nav>
        <NuxtLink
          v-if="recent.length"
          to="/mine"
          class="paste-view-all flex gap-2 items-center text-xs text-[var(--paste-ink-soft)] pt-1"
          >查看全部 <UIcon name="i-tabler-arrow-right" class="size-4"
        /></NuxtLink>
      </aside>
      <main
        class="paste-main min-w-0 min-h-0 bg-[var(--paste-editor)] max-[760px]:h-full"
      >
        <slot />
      </main>
    </div>
    <PasteMobileNavigation />
  </div>
</template>
