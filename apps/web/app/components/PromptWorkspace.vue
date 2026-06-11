<script setup lang="ts">
const advancedOpen = useState('advancedConsoleOpen', () => false)
</script>

<template>
  <div class="flex h-full flex-col">
    <ChatPanel />
    <AdvancedConsole :open="advancedOpen" @close="advancedOpen = false" />
  </div>
</template>
