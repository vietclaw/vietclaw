import { apiFetch } from '~/utils/api'

type StatusPayload = {
  version?: string
  commit?: string
  uptime?: string
  db_ok?: boolean
  mode?: string
}

type FrameworkPayload = {
  enabled?: boolean
  delegate_enabled?: boolean
  hooks_enabled?: boolean
  hooks_registered?: number
  agents?: unknown[]
}

export function useDaemon() {
  const status = useState<StatusPayload | null>('daemonStatus', () => null)
  const framework = useState<FrameworkPayload | null>('daemonFramework', () => null)
  const loading = useState('daemonLoading', () => false)
  const online = computed(() => status.value?.db_ok === true)

  async function refresh() {
    loading.value = true
    try {
      status.value = await apiFetch<StatusPayload>('/status')
      framework.value = await apiFetch<FrameworkPayload>('/api/framework')
    } catch {
      status.value = null
      framework.value = null
    } finally {
      loading.value = false
    }
  }

  if (import.meta.client && !status.value) {
    void refresh()
  }

  return { status, framework, loading, online, refresh }
}
