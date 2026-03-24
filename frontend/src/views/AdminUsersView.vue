<template>
  <div class="admin-users-page">
    <header class="header">
      <h1>Пользователи</h1>
      <button @click="logout">Выйти</button>
    </header>

    <section class="create-user">
      <h2>Создать пользователя</h2>
      <form @submit.prevent="createUser">
        <div>
          <label>ФИО</label>
          <input v-model="fullName" type="text" required />
        </div>
        <div>
          <label>Email</label>
          <input v-model="email" type="email" />
        </div>
        <div>
          <label>Телефон</label>
          <input v-model="phone" type="text" />
        </div>
        <div>
          <label>Роль</label>
          <select v-model="role" required>
            <option value="EXECUTOR">Исполнитель</option>
            <option value="CUSTOMER">Заказчик</option>
            <option value="INSPECTOR">Инспектор</option>
            <option value="ADMIN">Админ</option>
          </select>
        </div>
        <div>
          <label>Временный пароль</label>
          <input v-model="password" type="text" required />
        </div>

        <button type="submit" :disabled="creating">
          {{ creating ? 'Создаю...' : 'Создать пользователя' }}
        </button>
        <p v-if="createError" class="error">{{ createError }}</p>
      </form>
    </section>

    <!-- позже тут можно добавить таблицу пользователей -->
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const API_BASE = 'http://localhost:8080'

const fullName = ref('')
const email = ref('')
const phone = ref('')
const role = ref('EXECUTOR')
const password = ref('temp123')

const creating = ref(false)
const createError = ref<string | null>(null)

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

async function createUser() {
  createError.value = null
  creating.value = true
  try {
    const res = await fetch(`${API_BASE}/admin/users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({
        full_name: fullName.value,
        email: email.value || null,
        phone: phone.value || null,
        role: role.value,
        password: password.value,
      }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка создания пользователя')
    }

    // Можно очистить форму или показать уведомление
    fullName.value = ''
    email.value = ''
    phone.value = ''
    role.value = 'EXECUTOR'
    password.value = 'temp123'
  } catch (e: any) {
    createError.value = e.message || 'Ошибка'
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.admin-users-page {
  max-width: 800px;
  margin: 40px auto;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.create-user form div {
  margin-bottom: 10px;
}
.error {
  color: red;
}
</style>