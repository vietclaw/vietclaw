<script setup lang="ts">
import { X } from '@lucide/vue'
import { apiFetch } from '~/utils/api'

const isOpen = useState('imageGenOpen', () => false)
const toast = useToast()

const prompt = ref('')
const loading = ref(false)
const resultUrl = ref('')

async function generate() {
  if (!prompt.value.trim() || loading.value) return
  loading.value = true
  resultUrl.value = ''
  try {
    const data = await apiFetch<{ predictions?: { bytesBase64Encoded: string }[] }>('/api/image', {
      method: 'POST',
      body: JSON.stringify({ prompt: prompt.value, sampleCount: 1 })
    })
    if (data.predictions?.[0]?.bytesBase64Encoded) {
      resultUrl.value = `data:image/png;base64,${data.predictions[0].bytesBase64Encoded}`
    } else {
      toast.add('No image returned', 'error')
    }
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Image generation failed', 'error')
  } finally {
    loading.value = false
  }
}

function download() {
  if (!resultUrl.value) return
  const a = document.createElement('a')
  a.href = resultUrl.value
  a.download = `vietclaw_${Date.now()}.png`
  document.body.appendChild(a)
  a.click()
  a.remove()
}

function close() {
  isOpen.value = false
  prompt.value = ''
  resultUrl.value = ''
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center"
        @click.self="close"
      >
        <div class="w-full max-w-lg bg-zinc-950 border border-zinc-800 rounded-lg p-5 shadow-2xl relative">
          <button class="absolute top-4 right-4 p-1 rounded hover:bg-zinc-900 text-zinc-500" @click="close">
            <X :size="16" />
          </button>
          <h3 class="text-sm font-semibold text-zinc-100 mb-1">Image Generator</h3>
          <p class="text-[11px] text-zinc-500 mb-4">Generate images using the AI pipeline.</p>

          <div class="space-y-4">
            <div>
              <label class="text-[10px] text-zinc-400 uppercase tracking-wider block mb-1">Prompt</label>
              <textarea
                v-model="prompt"
                rows="3"
                placeholder="Describe the image you want to generate..."
                class="w-full bg-zinc-900 border border-zinc-800 rounded p-2.5 text-xs text-zinc-200 placeholder-zinc-700 focus:outline-none focus:border-zinc-700 resize-none font-mono"
              />
            </div>

            <!-- Loading -->
            <div
              v-if="loading"
              class="aspect-video w-full rounded border border-dashed border-zinc-800 bg-zinc-900/10 flex flex-col items-center justify-center space-y-2"
            >
              <div class="w-6 h-6 border-2 border-zinc-300 border-t-transparent rounded-full animate-spin" />
              <p class="text-[10px] text-zinc-500 animate-pulse">Generating...</p>
            </div>

            <!-- Result -->
            <div
              v-if="resultUrl && !loading"
              class="aspect-video w-full rounded border border-zinc-800 bg-zinc-950 flex items-center justify-center overflow-hidden relative group"
            >
              <img :src="resultUrl" alt="Generated image" class="w-full h-full object-cover">
              <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                <button
                  class="p-2 rounded bg-zinc-900 border border-zinc-800 text-zinc-300 hover:text-white transition-colors"
                  @click="download"
                >
                  Download
                </button>
              </div>
            </div>
          </div>

          <div class="mt-6 flex justify-end gap-2.5">
            <button
              class="px-3 py-1.5 rounded border border-zinc-800 hover:bg-zinc-900 text-zinc-400 text-xs transition-colors"
              @click="close"
            >
              Close
            </button>
            <button
              class="px-4 py-1.5 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 text-xs font-semibold transition-colors disabled:opacity-30"
              :disabled="loading || !prompt.trim()"
              @click="generate"
            >
              Generate
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
