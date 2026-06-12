<script setup lang="ts">
import { Menu } from '@lucide/vue'

const route = useRoute()
const mobileOpen = useState('sidebarMobileOpen', () => false)
const isSettings = computed(() => route.path.startsWith('/settings'))
const { t } = useI18n()

watch(() => route.path, () => {
  mobileOpen.value = false
})
</script>

<template>
  <div class="vc-app-frame flex w-full max-w-full overflow-hidden bg-vc-bg">
    <div class="vc-grain" aria-hidden="true" />
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="mobileOpen"
          class="fixed inset-0 z-40 bg-vc-text/20 backdrop-blur-sm md:hidden"
          @click="mobileOpen = false"
        />
      </Transition>
    </Teleport>

    <AppSidebar :open="mobileOpen" @close="mobileOpen = false" />

    <main class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <TopBar
        v-if="!isSettings"
        @toggle-mobile="mobileOpen = !mobileOpen"
      />
      <div
        v-else
        class="flex h-12 shrink-0 items-center border-b border-vc-border-subtle px-4 md:hidden"
      >
        <button
          type="button"
          class="vc-btn-ghost rounded-full p-1.5"
          @click="mobileOpen = true"
        >
          <Menu :size="18" :stroke-width="1.5" />
        </button>
        <span class="ml-2 text-sm font-medium text-vc-text">{{ t('nav.settings') }}</span>
      </div>
      <div class="min-h-0 flex-1 overflow-hidden">
        <slot />
      </div>
    </main>

    <Teleport to="body">
      <ToastContainer />
    </Teleport>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s var(--vc-ease);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
