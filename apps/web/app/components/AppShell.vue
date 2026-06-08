<script setup lang="ts">
const route = useRoute()
const mobileOpen = ref(false)

watch(() => route.path, () => {
  mobileOpen.value = false
})
</script>

<template>
  <div class="min-h-[100dvh] bg-[var(--bg-0)] overflow-x-hidden">
    <!-- Global noise overlay -->
    <div class="pointer-events-none fixed inset-0 z-[60] opacity-[0.015]" style="background-image: url('data:image/svg+xml,%3Csvg viewBox=%270 0 512 512%27 xmlns=%27http://www.w3.org/2000/svg%27%3E%3Cfilter id=%27n%27%3E%3CfeTurbulence type=%27fractalNoise%27 baseFrequency=%270.65%27 numOctaves=%275%27 stitchTiles=%27stitch%27/%3E%3C/filter%3E%3Crect width=%27100%25%27 height=%27100%25%27 filter=%27url(%23n)%27/%3E%3C/svg%3E'); background-size: 256px 256px;" />

    <!-- Ambient background gradients -->
    <div class="pointer-events-none fixed inset-0 z-0">
      <div class="absolute -left-[20%] -top-[10%] h-[600px] w-[600px] rounded-full bg-[var(--accent)]/[0.03] blur-[120px]" />
      <div class="absolute -right-[10%] top-[40%] h-[400px] w-[400px] rounded-full bg-purple-500/[0.02] blur-[100px]" />
    </div>

    <!-- Mobile overlay -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="mobileOpen"
          class="fixed inset-0 z-40 bg-black/70 backdrop-blur-md lg:hidden"
          @click="mobileOpen = false"
        />
      </Transition>
    </Teleport>

    <!-- Sidebar -->
    <AppSidebar :open="mobileOpen" @close="mobileOpen = false" />

    <!-- Main -->
    <div class="relative z-10 lg:pl-[240px]">
      <TopBar @toggle-mobile="mobileOpen = !mobileOpen" />
      <main class="mx-auto max-w-[1280px] px-4 pb-20 pt-8 sm:px-6 lg:px-10">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s var(--ease-out);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
