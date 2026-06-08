<script setup lang="ts">
const { toasts, remove } = useToast()

function toastColor(type: string) {
  if (type === 'success') return 'bg-zinc-950 text-zinc-100 border-zinc-800'
  if (type === 'error' || type === 'warning') return 'bg-zinc-950 text-rose-400 border-rose-950/20'
  return 'bg-zinc-900 text-zinc-300 border-zinc-800'
}
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-1.5 max-w-xs">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="flex items-center gap-2 px-3 py-2 rounded border shadow-xl text-[11px] font-mono transition-all duration-200"
        :class="toastColor(t.type)"
      >
        <span>[{{ t.type.toUpperCase() }}] {{ t.msg }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active {
  transition: all 0.2s ease;
}
.toast-leave-active {
  transition: all 0.2s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
