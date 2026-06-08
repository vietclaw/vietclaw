<script setup lang="ts">
const route = useRoute()
</script>

<template>
  <div class="min-h-screen">
    <div class="mx-auto grid min-h-screen w-full max-w-[1520px] grid-cols-1 lg:grid-cols-[252px_1fr]">
      <AppSidebar />
      <main class="min-w-0 border-l border-white/0 lg:border-white/[0.07]">
        <TopBar :title="String(route.meta.title || 'VietClaw')" />
        <div class="px-4 pb-8 pt-4 sm:px-6 lg:px-8">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

