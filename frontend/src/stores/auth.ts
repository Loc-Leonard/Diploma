import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface User {
  id: number
  full_name: string
  role: string
}

const STORAGE_KEY = 'auth' // один ключ вместо 'token'

export const useAuthStore = defineStore('auth', () => {
  // при создании стора один раз читаем из localStorage
  const saved = localStorage.getItem(STORAGE_KEY)
  let initialToken: string | null = null
  let initialUser: User | null = null

  if (saved) {
    try {
      const parsed = JSON.parse(saved) as { token: string | null; user: User | null }
      initialToken = parsed.token
      initialUser = parsed.user
    } catch {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  const token = ref<string | null>(initialToken)
  const user = ref<User | null>(initialUser)
  const mustChangePassword = ref<boolean>(false)

  const isAuthenticated = computed(() => !!token.value)

  function persist() {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ token: token.value, user: user.value }),
    )
  }

  function setAuth(data: { token: string; user: User; must_change_password: boolean }) {
    token.value = data.token
    user.value = data.user
    mustChangePassword.value = data.must_change_password
    persist()
  }

  function clearAuth() {
    token.value = null
    user.value = null
    mustChangePassword.value = false
    localStorage.removeItem(STORAGE_KEY)
  }

  return {
    token,
    user,
    mustChangePassword,
    isAuthenticated,
    setAuth,
    clearAuth,
  }
})