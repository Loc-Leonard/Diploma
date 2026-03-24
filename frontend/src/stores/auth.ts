import { defineStore } from "pinia"
import { ref, computed } from 'vue'

interface User {
    id: number
    full_name: string
    role: string
}
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)
  const mustChangePassword = ref<boolean>(false)

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(data: { token: string; user: User; must_change_password: boolean }) {
    token.value = data.token
    user.value = data.user
    mustChangePassword.value = data.must_change_password
    localStorage.setItem('token', data.token)
  }

  function clearAuth() {
    token.value = null
    user.value = null
    mustChangePassword.value = false
    localStorage.removeItem('token')
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
