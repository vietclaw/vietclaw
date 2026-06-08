<script setup lang="ts">
import { X } from '@lucide/vue'

const settingsOpen = useState('settingsOpen', () => false)
const { loadConfig, saveConfig } = useChat()
const toast = useToast()

const config = ref({
  apiKey: '',
  model: 'gemini-2.5-flash-preview-09-2025',
  temperature: 0.7,
  persona: 'general',
  voice: 'Zephyr'
})

watch(settingsOpen, (v) => {
  if (v) config.value = loadConfig()
})

function save() {
  saveConfig({ ...config.value })
  settingsOpen.value = false
  toast.add('Preferences saved', 'success')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="settingsOpen"
        class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center"
        @click.self="settingsOpen = false"
      >
        <div class="w-full max-w-md bg-zinc-950 border border-zinc-800 rounded-lg p-5 shadow-2xl relative">
          <button class="absolute top-4 right-4 p-1 rounded hover:bg-zinc-900 text-zinc-500" @click="settingsOpen = false">
            <X :size="16" />
          </button>
          <h3 class="text-sm font-semibold text-zinc-100 flex items-center gap-2 mb-4">Preferences</h3>

          <div class="space-y-4">
            <div>
              <label class="text-[10px] text-zinc-400 uppercase tracking-wider block mb-1">API Key (optional)</label>
              <input
                v-model="config.apiKey"
                type="password"
                placeholder="Leave empty for server default"
                class="w-full bg-zinc-900 border border-zinc-800 rounded px-3 py-2 text-xs text-zinc-200 focus:outline-none focus:border-zinc-700 font-mono"
              >
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="text-[10px] text-zinc-400 uppercase tracking-wider block mb-1">Temperature</label>
                <input
                  v-model.number="config.temperature"
                  type="range"
                  min="0"
                  max="2"
                  step="0.1"
                  class="w-full mt-2 h-1 bg-zinc-800 rounded appearance-none cursor-pointer accent-zinc-200"
                >
                <div class="flex justify-between text-[9px] text-zinc-500 mt-1">
                  <span>Precise (0)</span>
                  <span class="text-zinc-200 font-bold">{{ config.temperature }}</span>
                </div>
              </div>
              <div>
                <label class="text-[10px] text-zinc-400 uppercase tracking-wider block mb-1">Persona</label>
                <select
                  v-model="config.persona"
                  class="w-full bg-zinc-900 border border-zinc-800 rounded px-2 py-2 text-xs text-zinc-200 focus:outline-none focus:border-zinc-700 font-mono"
                >
                  <option value="general">Default Assistant</option>
                  <option value="programmer">Software Engineer</option>
                  <option value="creator">Creative Writer</option>
                  <option value="analyst">Data Analyst</option>
                  <option value="psychologist">Clinical Psychologist</option>
                </select>
              </div>
            </div>
          </div>

          <div class="mt-6 flex justify-end gap-2.5">
            <button
              class="px-3 py-1.5 rounded border border-zinc-800 hover:bg-zinc-900 text-zinc-400 text-xs transition-colors"
              @click="settingsOpen = false"
            >
              Close
            </button>
            <button
              class="px-4 py-1.5 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 text-xs font-semibold transition-colors"
              @click="save"
            >
              Save Config
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.15s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
