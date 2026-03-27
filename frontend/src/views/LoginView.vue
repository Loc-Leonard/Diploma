<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <div class="login-logo">АВТОРИЗАЦИЯ</div>
      </div>

      <form @submit.prevent="onSubmit" class="login-form">
        <div class="form-field">
          <label>Логин (email или телефон)</label>
          <input v-model="login" type="text" required />
        </div>

        <div class="form-field">
          <label>Пароль</label>
          <input v-model="password" type="password" required />
        </div>

        <button type="submit" :disabled="loading">
          {{ loading ? 'Вход...' : 'Войти' }}
        </button>

        <p v-if="error" class="error">{{ error }}</p>
      </form>
    </div>
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
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  padding: 16px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: #ffffff;
  border-radius: 16px;
  padding: 28px 30px 30px;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
  box-sizing: border-box;
  border: 1px solid #e5e7eb;
}

.login-header {
  margin-bottom: 20px;
  text-align: center;
}

.login-logo {
  font-size: 18px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #4b5563;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-field {
  display: flex;
  flex-direction: column;
  align-items: center;
}

label {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 6px;
  text-align: center;
}

input {
  width: 100%;
  max-width: 260px;
  padding: 8px 11px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  font-size: 14px;
  box-sizing: border-box;
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease;
  background-color: #f9fafb;
}

input:focus {
  border-color: #a5b4fc;        /* мягкий сиреневый */
  box-shadow: 0 0 0 1px rgba(129, 140, 248, 0.35);
  background-color: #ffffff;
}

button {
  width: 100%;
  max-width: 260px;
  align-self: center;
  margin-top: 4px;
  padding: 9px 0;
  border-radius: 999px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  background: #a5b4fc;          /* пастельный фиолетовый */
  color: #111827;
  transition: background-color 0.15s ease, transform 0.05s ease, box-shadow 0.15s ease;
  box-shadow: 0 8px 18px rgba(129, 140, 248, 0.35);
}

button:hover:not(:disabled) {
  background: #818cf8;
}

button:active:not(:disabled) {
  transform: translateY(1px);
  box-shadow: 0 4px 10px rgba(129, 140, 248, 0.4);
}

button:disabled {
  opacity: 0.6;
  cursor: default;
  box-shadow: none;
}

.error {
  margin-top: 8px;
  font-size: 13px;
  color: #dc2626;
  text-align: center;
}
</style>