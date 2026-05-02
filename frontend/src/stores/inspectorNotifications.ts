import { defineStore } from "pinia";
import { ref, computed } from 'vue';
import { useAuthStore } from "./auth";

const API_BASE = 'http://localhost:8080'

type InspectorObjectStatus =
  | 'PLANNED'
  | 'WAITING_INSPECTOR_CONFIRMATION'
  | 'ACTIVE'
  | 'FINISHED'

  type InspectorObjectItem = {
    id: number
    name: string
    city: string
    address: string
    status: InspectorObjectStatus
    foreman_name: string
    planned_start_date?: string | null
    has_pending_action: boolean
  }

export const useInspectorNotificationsStore = defineStore('inspectorNotifications', () => {
  const auth = useAuthStore()

  const objects = ref<InspectorObjectItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const started = ref(false)
  let timer: number | null = null

  const pendingObjects = computed(() =>
    objects.value.filter(o => o.status === 'WAITING_INSPECTOR_CONFIRMATION')
  )
  const pendingCount = computed(() => pendingObjects.value.length)

  async function fetchPending() {
    if (!auth.token) return
    loading.value = true
    error.value = null

    try {
        const params = new URLSearchParams()
        params.set('status', 'WAITING_INSPECTOR_CONFIRMATION')

        const res = await fetch(`${API_BASE}/inspector/objects?${params.toString()}`, {
            headers: {
                Authorization: `Bearer ${auth.token}`,
            },
        })

        if (!res.ok) {
            throw new Error((await res.json().catch(() => ({}))).error || 'Ошибка загрузки уведомлений')
        }
        objects.value = await res.json()
    }   catch (e: any) {
        error.value = e.message || "ERROR"
    }   finally {
        loading.value = false
    }
  }

  function startPolling() {
    if (started.value) return
    started.value = true

    fetchPending()
    timer = window.setInterval(() => {
        fetchPending()
    }, 1000)
  }

  function stopPolling() {
    if (timer !== null) {
        clearInterval(timer)
        timer = null
    }
    started.value = false
  }

  return {
    objects,
    loading,
    error,
    pendingObjects,
    pendingCount,
    fetchPending,
    startPolling,
    stopPolling,
  }
})