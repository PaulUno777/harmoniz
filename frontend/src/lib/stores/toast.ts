import { writable } from 'svelte/store'

export type ToastType = 'success' | 'error' | 'warning'

export interface Toast {
  id: string
  type: ToastType
  message: string
}

const { subscribe, update } = writable<Toast[]>([])

function add(type: ToastType, message: string, duration = 3500) {
  const id = crypto.randomUUID()
  update(list => [...list, { id, type, message }])
  setTimeout(() => remove(id), duration)
}

function remove(id: string) {
  update(list => list.filter(t => t.id !== id))
}

export const toast = {
  subscribe,
  success: (message: string) => add('success', message),
  error: (message: string) => add('error', message),
  warning: (message: string) => add('warning', message),
  remove,
}
