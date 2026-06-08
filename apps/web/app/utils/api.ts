export type ApiError = {
  error?: string
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...options,
    headers: {
      'content-type': 'application/json',
      ...(options.headers || {})
    }
  })

  if (!response.ok) {
    let message = response.statusText
    try {
      const payload = await response.json() as ApiError
      message = payload.error || message
    } catch {
      // keep status text
    }
    throw new Error(message)
  }

  return await response.json() as T
}

export function formatMoney(value?: number): string {
  if (!value) return '$0.0000'
  return `$${value.toFixed(4)}`
}

export function statusTone(ok?: boolean): string {
  return ok ? 'text-[var(--success)]' : 'text-[var(--warning)]'
}

