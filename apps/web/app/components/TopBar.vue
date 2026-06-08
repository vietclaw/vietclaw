<script setup lang="ts">
defineEmits<{ toggleMobile: [] }>()

const route = useRoute()

const breadcrumbs = computed(() => {
  const segments = route.path.split('/').filter(Boolean)
  const crumbs = [{ label: 'Home', to: '/' }]
  let path = ''
  for (const seg of segments) {
    path += `/${seg}`
    const label = seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, ' ')
    crumbs.push({ label, to: path })
  }
  return crumbs
})
</script>

<template>
  <header class="sticky top-0 z-30 flex h-14 items-center border-b border-[var(--border-0)] bg-[var(--bg-0)]/60 backdrop-blur-xl">
    <div class="flex w-full items-center gap-3 px-4 sm:px-6 lg:px-10">
      <!-- Mobile hamburger -->
      <button
        class="flex h-8 w-8 items-center justify-center rounded-lg text-[var(--fg-2)] vc-transition-fast hover:bg-[var(--bg-3)] hover:text-[var(--fg-0)] lg:hidden vc-focus"
        @click="$emit('toggleMobile')"
      >
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      <!-- Breadcrumbs -->
      <nav class="flex items-center gap-1.5 text-[12px]">
        <template v-for="(crumb, i) in breadcrumbs" :key="crumb.to">
          <NuxtLink
            v-if="i < breadcrumbs.length - 1"
            :to="crumb.to"
            class="text-[var(--fg-2)] vc-transition-fast hover:text-[var(--fg-0)] vc-focus"
          >
            {{ crumb.label }}
          </NuxtLink>
          <span v-else class="font-semibold text-[var(--fg-0)]">{{ crumb.label }}</span>
          <svg v-if="i < breadcrumbs.length - 1" class="h-3 w-3 text-[var(--fg-2)]/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </template>
      </nav>
    </div>
  </header>
</template>
