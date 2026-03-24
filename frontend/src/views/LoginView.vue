<template>
  <div class="login-page">
    <h1>Вход</h1>
    <form @submit.prevent="onSubmit" class="login-form">
      <div>
        <label>Логин (email или телефон)</label>
        <input v-model="login" type="text" required />
      </div>
      <div>
        <label>Пароль</label>
        <input v-model="password" type="password" required />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Вход...' : 'Войти' }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const login = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const router = useRouter()
const auth = useAuthStore()

const API_BASE = 'http://localhost:8080'

async function onSubmit() {
  error.value = null
  loading.value = true
  try {
    const res = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        login: login.value,
        password: password.value,
      }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка авторизации')
    }

    const data = await res.json()
    auth.setAuth(data)

    if (data.must_change_password) {
      router.push({ name: 'change-password' })
    } else if (data.user.role === 'ADMIN') {
      router.push({ name: 'admin-users' })
    } else {
      // позже: другие роли
      router.push({ name: 'admin-users' })
    }
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  max-width: 400px;
  margin: 80px auto;
}
.login-form div {
  margin-bottom: 12px;
}
.error {
  color: red;
  margin-top: 8px;
}
</style>