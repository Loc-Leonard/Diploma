import { defineStore } from "pinia";
import { ref, computed } from 'vue';
import { useAuthStore } from "./auth";

const API_BASE = 'http://localhost:8080'

export const useForemanNotificationsStore = defineStore('foremanNotifications', () => {
  const auth = useAuthStore()

  const unreadCount = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const started = ref(false)
  let timer: number | null = null

  async function fetchUnreadCount() {
    if (!auth.token) return
    loading.value = true
    error.value = null

    try {
      const res = await fetch(`${API_BASE}/foreman/issues/unread-count`, {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      })
      console.log('Notifications response status:', res.status)

      if (!res.ok) {
        throw new Error((await res.json().catch(() => ({}))).error || 'Ошибка загрузки уведомлений')
      }
      const data = await res.json()
      console.log('Unread count:', data.count)
      unreadCount.value = data.count ?? 0
    } catch (e: any) {
      console.log('Notification fetch error', e)
      error.value = e.message || "ERROR"
    } finally {
      loading.value = false
    }
  }

  function startPolling() {
    if (started.value) return
    started.value = true
    fetchUnreadCount()
    timer = window.setInterval(() => {
      fetchUnreadCount()
    }, 15000) // опрос каждые 15 секунд
  }

  function stopPolling() {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
    started.value = false
  }

  function markAsRead() {
    // Опционально: можно сбрасывать счётчик при переходе на страницу замечаний
    unreadCount.value = 0
  }

  return {
    unreadCount,
    loading,
    error,
    fetchUnreadCount,
    startPolling,
    stopPolling,
    markAsRead,
  }
})