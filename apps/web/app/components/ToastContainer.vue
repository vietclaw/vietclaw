<script setup lang="ts">
import { CheckCircle2, AlertCircle, Info } from '@lucide/vue'

const { toasts, remove } = useToast()

function toastIcon(type: string) {
  if (type === 'success') return CheckCircle2
  if (type === 'error' || type === 'warning') return AlertCircle
  return Info
}

function iconClass(type: string) {
  if (type === 'success') return 'text-vc-success'
  if (type === 'error' || type === 'warning') return 'text-vc-error'
  return 'text-vc-text-muted'
}
</script>

<template>
  <div class="fixed bottom-5 right-5 z-50 flex max-w-sm flex-col gap-2">
    <TransitionGroup name="toast">
      <button
        v-for="t in toasts"
        :key="t.id"
        type="button"
        class="flex items-center gap-2.5 rounded-2xl border border-vc-border-subtle bg-vc-surface px-4 py-3 text-left text-sm text-vc-text shadow-[var(--vc-shadow-lg)]"
        @click="remove(t.id)"
      >
        <component :is="toastIcon(t.type)" :size="16" :stroke-width="1.75" class="shrink-0" :class="iconClass(t.type)" />
        {{ t.msg }}
      </button>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active {
  transition: opacity 0.35s var(--vc-ease), transform 0.35s var(--vc-ease);
}
.toast-leave-active {
  transition: opacity 0.2s var(--vc-ease), transform 0.2s var(--vc-ease);
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.97);
}
.toast-leave-to {
  opacity: 0;
  transform: scale(0.97);
}
</style>
