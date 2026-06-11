<script setup lang="ts">
import { ArrowLeft, Menu, RefreshCw, Save } from '@lucide/vue'

const route = useRoute()
const mobileOpen = useState('sidebarMobileOpen', () => false)
const { loading, saving, dirty, load, save, reload, discard } = useSettings()

const nav = [
  { to: '/settings', label: 'Tổng quan', exact: true },
  { to: '/settings/providers', label: 'Providers' },
  { to: '/settings/budget', label: 'Budget' },
  { to: '/settings/channels', label: 'Kênh' },
  { to: '/settings/memory', label: 'Memory' },
  { to: '/settings/logs', label: 'Logs' },
]

onMounted(() => {
  if (!useSettings().config.value) void load()
})

function isActive(path: string, exact?: boolean) {
  if (exact) return route.path === path
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div class="flex shrink-0 items-center justify-between gap-4 border-b border-vc-border-subtle px-4 py-3 md:px-6">
      <div class="flex items-center gap-2 min-w-0">
        <button
          type="button"
          class="vc-btn-ghost rounded-md p-1.5 md:hidden"
          @click="mobileOpen = true"
        >
          <Menu :size="18" :stroke-width="1.75" />
        </button>
        <NuxtLink to="/" class="vc-link flex items-center gap-1.5 text-sm">
          <ArrowLeft :size="16" :stroke-width="1.75" />
          Quay lại chat
        </NuxtLink>
      </div>
      <div class="flex items-center gap-2">
        <span v-if="dirty" class="text-xs text-vc-accent">Chưa lưu</span>
        <button
          type="button"
          class="vc-btn vc-btn-ghost text-xs disabled:opacity-40"
          :disabled="saving"
          @click="discard"
        >
          Hủy
        </button>
        <button
          type="button"
          class="vc-btn vc-btn-ghost text-xs disabled:opacity-40"
          :disabled="saving"
          @click="reload"
        >
          <RefreshCw :size="14" :stroke-width="1.75" />
          Tải lại
        </button>
        <button
          type="button"
          class="vc-btn vc-btn-primary text-xs disabled:opacity-40"
          :disabled="saving || !dirty"
          @click="save"
        >
          <Save :size="14" :stroke-width="1.75" />
          Lưu
        </button>
      </div>
    </div>

    <div class="flex min-h-0 flex-1">
      <nav class="hidden w-44 shrink-0 border-r border-vc-border-subtle bg-vc-surface px-2 py-4 md:block">
        <p class="mb-3 px-3 text-xs font-semibold text-vc-text">Cài đặt</p>
        <ul class="space-y-0.5">
          <li v-for="item in nav" :key="item.to">
            <NuxtLink
              :to="item.to"
              class="block rounded-md px-3 py-2 text-sm transition-colors"
              :class="isActive(item.to, item.exact)
                ? 'bg-vc-bg-subtle font-medium text-vc-text'
                : 'text-vc-text-secondary hover:bg-vc-bg-subtle hover:text-vc-text'"
            >
              {{ item.label }}
            </NuxtLink>
          </li>
        </ul>
      </nav>

      <div class="min-h-0 flex-1 overflow-y-auto vc-scrollbar">
        <div v-if="loading" class="mx-auto max-w-3xl space-y-3 p-6 md:p-8">
          <div v-for="i in 4" :key="i" class="h-12 rounded-lg bg-vc-bg-subtle animate-pulse" />
        </div>
        <div v-else class="mx-auto max-w-3xl p-4 md:p-8">
          <nav class="mb-6 flex gap-2 overflow-x-auto md:hidden vc-scrollbar">
            <NuxtLink
              v-for="item in nav"
              :key="item.to"
              :to="item.to"
              class="shrink-0 rounded-full px-3 py-1.5 text-xs transition-colors"
              :class="isActive(item.to, item.exact)
                ? 'bg-vc-bg-subtle font-medium text-vc-text'
                : 'text-vc-text-muted'"
            >
              {{ item.label }}
            </NuxtLink>
          </nav>
          <slot />
        </div>
      </div>
    </div>
  </div>
</template>
