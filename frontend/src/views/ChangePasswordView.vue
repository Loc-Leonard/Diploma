<template>
  <div class="change-password-page">
    <div class="change-password-card">
      <div class="change-password-header">
        <div class="change-password-title">СМЕНА ПАРОЛЯ</div>
        <p class="change-password-subtitle">
          Придумайте новый надёжный пароль для аккаунта
        </p>
      </div>

      <form @submit.prevent="onSubmit" class="change-password-form">
        <div class="form-field">
          <label>Старый пароль</label>
          <input v-model="oldPassword" type="password" required />
        </div>

        <div class="form-field">
          <label>Новый пароль</label>
          <input v-model="newPassword" type="password" required />
        </div>

        <button type="submit" :disabled="loading">
          {{ loading ? 'Сохраняю...' : 'Сменить пароль' }}
        </button>

        <p v-if="error" class="error">{{ error }}</p>
        <p v-if="success" class="success">Пароль успешно изменён</p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const oldPassword = ref('')
const newPassword = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

const auth = useAuthStore()
const router = useRouter()

const API_BASE = import.meta.env.VITE_API_URL as string

async function onSubmit() {
  error.value = null
  success.value = false
  loading.value = true

  try {
    const res = await fetch(`${API_BASE}/auth/change-password`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({
        old_password: oldPassword.value,
        new_password: newPassword.value,
      }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка смены пароля')
    }

    success.value = true
    auth.mustChangePassword = false
    setTimeout(() => {
      router.push({ name: 'admin-users' })
    }, 1000)
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.change-password-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  padding: 16px;
}

.change-password-card {
  width: 100%;
  max-width: 420px;
  background: #ffffff;
  border-radius: 16px;
  padding: 28px 30px 30px;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
  box-sizing: border-box;
  border: 1px solid #e5e7eb;
}

.change-password-header {
  text-align: center;
  margin-bottom: 20px;
}

.change-password-title {
  font-size: 18px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #4b5563;
}

.change-password-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: #9ca3af;
}

.change-password-form {
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
  border-color: #a5b4fc;
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
  background: #a5b4fc;
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
  margin-top: 4px;
  font-size: 13px;
  color: #dc2626;
  text-align: center;
}

.success {
  margin-top: 4px;
  font-size: 13px;
  color: #16a34a;
  text-align: center;
}
</style>