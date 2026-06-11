<script setup lang="ts">
const route = useRoute()
const mobileOpen = useState('sidebarMobileOpen', () => false)
const isSettings = computed(() => route.path.startsWith('/settings'))

watch(() => route.path, () => {
  mobileOpen.value = false
})
</script>

<template>
  <div class="flex h-[100dvh] w-screen overflow-hidden bg-vc-bg">
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="mobileOpen"
          class="fixed inset-0 z-40 bg-vc-text/15 lg:hidden"
          @click="mobileOpen = false"
        />
      </Transition>
    </Teleport>

    <AppSidebar :open="mobileOpen" @close="mobileOpen = false" />

    <main class="flex h-full min-w-0 flex-1 flex-col">
      <TopBar v-if="!isSettings" @toggle-mobile="mobileOpen = !mobileOpen" />
      <div class="flex-1 overflow-hidden">
        <slot />
      </div>
    </main>

    <ToastContainer />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
