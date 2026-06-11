<script setup lang="ts">
const route = useRoute()
const mobileOpen = ref(false)

watch(() => route.path, () => {
  mobileOpen.value = false
})
</script>

<template>
  <div class="h-screen w-screen overflow-hidden flex flex-col relative grid-bg">
    <div class="flex-1 flex overflow-hidden z-10 relative">
      <!-- Mobile overlay -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="mobileOpen"
            class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm lg:hidden"
            @click="mobileOpen = false"
          />
        </Transition>
      </Teleport>

      <!-- Sidebar -->
      <AppSidebar :open="mobileOpen" @close="mobileOpen = false" />

      <!-- Main Workspace -->
      <main class="flex-1 flex flex-col h-full overflow-hidden bg-zinc-950/20 backdrop-blur-3xl">
        <TopBar @toggle-mobile="mobileOpen = !mobileOpen" />
        <div class="flex-1 overflow-hidden">
          <slot />
        </div>
      </main>
    </div>

    <ToastContainer />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
