<template>
  <div class="admin-layout">
    <!-- Левое меню -->
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">ЛОГО</div>

        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active">Пользователи</button>
          <button class="nav-item">Справочники</button>
        </nav>
      </div>

      <div class="sidebar-bottom">
        <div class="role-badge">
          <span class="role-dot role-dot--admin"></span>
          <span>Администратор</span>
        </div>
        <button class="logout-button" @click="logout">Выйти</button>
      </div>
    </aside>

    <!-- Центральный блок -->
    <main class="admin-main">
      <header class="admin-header">
        <h1 class="admin-title">Пользователи</h1>

        <div class="admin-header-right">
          <button class="primary-btn" @click="toggleCreate">
            {{ showCreate ? 'Скрыть форму' : 'Добавить пользователя' }}
          </button>

          <div class="search-wrapper">
            <input
              v-model="search"
              type="text"
              placeholder="Поиск по ФИО, телефону, email"
            />
          </div>
        </div>
      </header>

      <section class="table-card">
        <div v-if="usersLoading" class="table-state">
          Загружаю пользователей...
        </div>
        <div v-else-if="usersError" class="table-state table-state--error">
          {{ usersError }}
        </div>
        <template v-else>
          <table class="users-table" v-if="filteredUsers.length">
            <thead>
              <tr>
                <th>ФИО</th>
                <th>Роль</th>
                <th>Контакт</th>
                <th>Статус</th>
                <th>Последний вход</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in filteredUsers" :key="user.id">
                <td>{{ user.full_name }}</td>
                <td>
                  <span class="role-chip" :class="roleClass(user.role)">
                    {{ roleLabel(user.role) }}
                  </span>
                </td>
                <td>
                  <div class="contact-cell">
                    <div v-if="user.phone">{{ user.phone }}</div>
                    <div v-if="user.email">{{ user.email }}</div>
                    <div v-if="!user.phone && !user.email">—</div>
                  </div>
                </td>
                <td>
                  <span
                    class="status-chip"
                    :class="{
                      'status-chip--active': user.status === 'ACTIVE',
                      'status-chip--blocked': user.status === 'BLOCKED'
                    }"
                  >
                    {{ user.status === 'ACTIVE' ? 'Активен' : 'Заблокирован' }}
                  </span>
                </td>
                <td>{{ user.last_login || '—' }}</td>
              </tr>
            </tbody>
          </table>

          <div v-else class="table-state">Пользователей пока нет</div>

          <div class="table-footer">
            <span>Показано {{ filteredUsers.length }} из {{ users.length }}</span>
          </div>
        </template>
      </section>
    </main>

    <!-- Правая панель "Создание пользователя" -->
    <aside class="create-user-panel" v-if="showCreate">
      <section class="create-user-card">
        <div class="create-user-header">
          <h2>Создание пользователя</h2>
          <p>Заполните данные нового пользователя и задайте временный пароль</p>
        </div>

        <form @submit.prevent="createUser" class="create-user-form">
          <div class="form-row">
            <label>ФИО</label>
            <input v-model="fullName" type="text" required />
          </div>

          <div class="form-row">
            <label>Email</label>
            <input
              v-model="email"
              type="email"
              placeholder="example@mail.ru"
            />
          </div>

          <div class="form-row">
            <label>Телефон</label>
            <input
              v-model="phone"
              type="text"
              placeholder="+7..."
            />
          </div>

          <div class="form-row">
            <label>Роль</label>
            <select v-model="role" required>
              <option value="CUSTOMER">Заказчик</option>
              <option value="FOREMAN">Прораб</option>
              <option value="INSPECTOR">Инспектор</option>
              <option value="ADMIN">Админ</option>
            </select>
          </div>

          <div class="form-row" v-if="tempPassword">
            <label>Временный пароль</label>
            <input :value="tempPassword" type="text" readonly />
          </div>

          <button
            type="submit"
            :disabled="creating"
            class="primary-btn primary-btn--full"
          >
            {{ creating ? 'Создаю...' : 'Сохранить' }}
          </button>

          <p v-if="createError" class="error">{{ createError }}</p>
        </form>
      </section>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

const API_BASE = 'http://localhost:8080'

type UserItem = {
  id: number
  full_name: string
  email?: string | null
  phone?: string | null
  role: string
  status: 'ACTIVE' | 'BLOCKED'
  last_login?: string | null
}

// состояние списка пользователей
const users = ref<UserItem[]>([])
const usersLoading = ref(false)
const usersError = ref<string | null>(null)

// поиск
const search = ref('')

// создание пользователя
const fullName = ref('')
const email = ref('')
const phone = ref('')
const role = ref('EXECUTOR')
const creating = ref(false)
const createError = ref<string | null>(null)
const tempPassword = ref<string | null>(null)

// правая панель
const showCreate = ref(false)

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

function toggleCreate() {
  showCreate.value = !showCreate.value
}

// загрузка пользователей
async function fetchUsers() {
  usersLoading.value = true
  usersError.value = null
  try {
    const res = await fetch(`${API_BASE}/admin/users`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка загрузки пользователей')
    }

    const data = await res.json()
    users.value = data
  } catch (e: any) {
    usersError.value = e.message || 'Ошибка'
  } finally {
    usersLoading.value = false
  }
}

onMounted(fetchUsers)

// фильтрация по поиску
const filteredUsers = computed(() =>
  users.value.filter((u) => {
    const q = search.value.trim().toLowerCase()
    if (!q) return true
    return (
      u.full_name.toLowerCase().includes(q) ||
      (u.phone || '').toLowerCase().includes(q) ||
      (u.email || '').toLowerCase().includes(q)
    )
  }),
)

// подписи и цвета ролей
function roleLabel(r: string) {
  switch (r) {
    case 'CUSTOMER':
      return 'Заказчик'
    case 'FOREMAN':
      return 'Прораб'
    case 'EXECUTOR':
      return 'Исполнитель'
    case 'ADMIN':
      return 'Админ'
    default:
      return r
  }
}

function roleClass(r: string) {
  return {
    'role-chip--customer': r === 'CUSTOMER',
    'role-chip--foreman': r === 'FOREMAN',
    'role-chip--executor': r === 'EXECUTOR',
    'role-chip--admin': r === 'ADMIN',
  }
}

// создание пользователя
async function createUser() {
  createError.value = null

  const emailVal = email.value.trim() || null
  const phoneVal = phone.value.trim() || null

  if (!emailVal && !phoneVal) {
    createError.value = 'Укажите email или телефон'
    return
  }

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
        email: emailVal,
        phone: phoneVal,
        role: role.value,
      }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка создания пользователя')
    }
    const data = await res.json()
    tempPassword.value = data.temp_password || null

    fullName.value = ''
    email.value = ''
    phone.value = ''
    role.value = 'EXECUTOR'

    await fetchUsers()
  } catch (e: any) {
    createError.value = e.message || 'Ошибка'
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.admin-layout {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr) 360px;
  min-height: 100vh;
  background: #f9fafb;
}

/* Сайдбар */
.sidebar {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px 18px;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
}

.sidebar-logo {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 24px;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nav-item {
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  border: none;
  background: transparent;
  font-size: 14px;
  color: #4b5563;
  cursor: pointer;
}

.nav-item--active {
  background: #eef2ff;
  color: #4338ca;
}

.sidebar-bottom {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #6b7280;
}

.role-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
}

.role-dot--admin {
  background: #3b82f6;
}

/* Центральная часть */
.admin-main {
  padding: 20px 24px;
  box-sizing: border-box;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.admin-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

.admin-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.primary-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  background: #a5b4fc;
  color: #111827;
  font-size: 14px;
  cursor: pointer;
}

.primary-btn--full {
  width: 100%;
}

.search-wrapper {
  max-width: 260px;
  width: 100%;
}

.search-wrapper input {
  width: 100%;
  padding: 8px 11px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
}

.logout-button {
  padding: 7px 16px;
  border-radius: 999px;
  border: 1px solid #e5e7eb;
  background: #ffffff;
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
}

/* Таблица */
.table-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px 18px 14px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.08);
  border: 1px solid #e5e7eb;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.users-table th,
.users-table td {
  padding: 8px 10px;
  text-align: left;
}

.users-table thead {
  background: #f9fafb;
}

.users-table tbody tr:nth-child(odd) {
  background: #f9fafb;
}

.table-footer {
  margin-top: 8px;
  font-size: 12px;
  color: #6b7280;
}

.table-state {
  font-size: 13px;
  color: #6b7280;
}

.table-state--error {
  color: #b91c1c;
}

/* Контакт в таблице */
.contact-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
}

/* Роли */
.role-chip {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.role-chip--customer {
  background: #dcfce7;
  color: #166534;
}

.role-chip--foreman {
  background: #fef9c3;
  color: #854d0e;
}

.role-chip--executor {
  background: #ede9fe;
  color: #5b21b6;
}

.role-chip--admin {
  background: #dbeafe;
  color: #1d4ed8;
}

/* Статусы */
.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.status-chip--active {
  background: #dcfce7;
  color: #166534;
}

.status-chip--blocked {
  background: #fee2e2;
  color: #b91c1c;
}

/* Правая панель */
.create-user-panel {
  padding: 20px 20px;
  border-left: 1px solid #e5e7eb;
  background: #ffffff;
  box-sizing: border-box;
}

.create-user-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 24px 20px 24px;
  box-sizing: border-box;
}

.create-user-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  color: #111827;
}

.create-user-header p {
  margin: 6px 0 0;
  font-size: 13px;
  color: #9ca3af;
}

.create-user-form {
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-row label {
  font-size: 13px;
  color: #6b7280;
}

.form-row input,
.form-row select {
  padding: 8px 11px;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  font-size: 14px;
  box-sizing: border-box;
  outline: none;
  background-color: #f9fafb;
}

.error {
  margin-top: 4px;
  font-size: 13px;
  color: #dc2626;
}
</style>