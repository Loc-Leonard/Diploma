<template>
  <div class="change-password-page">
    <h1>Смена пароля</h1>
    <form @submit.prevent="onSubmit" class="change-password-form">
      <div>
        <label>Старый пароль</label>
        <input v-model="oldPassword" type="password" required />
      </div>
      <div>
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

const API_BASE = 'http://localhost:8080'

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
  max-width: 400px;
  margin: 80px auto;
}
.change-password-form div {
  margin-bottom: 12px;
}
.error {
  color: red;
}
.success {
  color: green;
}
</style>