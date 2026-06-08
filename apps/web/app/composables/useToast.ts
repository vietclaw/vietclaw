export type ToastType = 'info' | 'success' | 'error' | 'warning'

export type Toast = {
  id: number
  msg: string
  type: ToastType
}

const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
  function add(msg: string, type: ToastType = 'info') {
    const id = nextId++
    toasts.value.push({ id, msg, type })
    setTimeout(() => remove(id), 3000)
  }

  function remove(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return { toasts, add, remove }
}
