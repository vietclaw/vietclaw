<script setup lang="ts">
import { ArrowLeft, RefreshCw, Save } from '@lucide/vue'

const { currentSessionId, sessions } = useChat()
const { loading, saving, dirty, load, save, reload, discard } = useSettings()

const chatHome = computed(() => {
  const id = currentSessionId.value || sessions.value[0]?.id
  return id ? `/p/${id}` : '/'
})

onMounted(() => {
  if (!useSettings().config.value) void load()
})
</script>

<template>
  <div class="flex h-full min-h-0 max-h-full flex-col overflow-hidden">
    <div class="flex shrink-0 items-center justify-between gap-4 border-b border-vc-border-subtle px-4 py-3 md:px-6 min-w-0">
      <NuxtLink :to="chatHome" class="vc-link flex items-center gap-1.5 text-sm">
        <ArrowLeft :size="16" :stroke-width="1.75" />
        Quay lại chat
      </NuxtLink>
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

    <div class="min-h-0 flex-1 overflow-y-auto vc-scrollbar">
      <div v-if="loading" class="mx-auto max-w-3xl space-y-3 p-6 md:p-8">
        <div v-for="i in 4" :key="i" class="h-12 rounded-lg bg-vc-bg-subtle animate-pulse" />
      </div>
      <div v-else class="mx-auto max-w-3xl p-4 md:p-8">
        <slot />
      </div>
    </div>
  </div>
</template>
