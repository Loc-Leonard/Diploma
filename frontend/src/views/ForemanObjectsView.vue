<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'

const router = useRouter()
const auth = useAuthStore()

// Приветствие в левом верхнем углу
const greeting = computed(() => {
  if (!auth.isAuthenticated) {
    return 'Добрый день'
  }
  const u = auth.user
  return u?.full_name ? `Добрый день, ${u.full_name}` : 'Добрый день'
})

type ForemanObject = {
  id: number
  name: string
  city: string
  address: string
  status: 'PLANNED' | 'WAITING_INSPECTOR_CONFIRMATION' | 'ACTIVE' | 'FINISHED'
}

const objects = ref<ForemanObject[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const search = ref('')

async function loadObjects() {
  loading.value = true
  error.value = null

  try {
    const res = await fetch(`${API_BASE}/foreman/objects`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка загрузки объектов')
    }
    objects.value = await res.json()
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}

const filteredObjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return objects.value
  return objects.value.filter((o) =>
    o.name.toLowerCase().includes(q),
  )
})

function statusLabel(status: ForemanObject['status']) {
  switch (status) {
    case 'PLANNED':
      return 'Запланирован'
    case 'WAITING_INSPECTOR_CONFIRMATION':
      return 'Ожидает подтверждения'
    case 'ACTIVE':
      return 'Активен'
    case 'FINISHED':
      return 'Завершен'
    default:
      return status
  }
}

function goToObject(id: number) {
  router.push({ name: 'foreman-object', params: { id } })
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' }) // если у логина другое имя маршрута — поправь здесь
}

onMounted(loadObjects)
</script>

<template>
  <div class="customer-layout">
    <!-- Левое меню -->
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>

        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active">Объекты</button>
          <button class="nav-item" disabled>Поставки</button>
          <button class="nav-item" disabled>Замечания</button>
        </nav>
      </div>

      <div class="sidebar-bottom">
        <div class="role-badge">
          <span class="role-dot role-dot--foreman"></span>
          <span>Прораб</span>
        </div>
        <button class="logout-button" @click="logout">Выйти</button>
      </div>
    </aside>

    <!-- Центральная часть: список объектов -->
    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Мои объекты</h1>

        <div class="customer-header-right">
          <div class="search-wrapper">
            <input
              v-model="search"
              type="text"
              placeholder="Поиск по объекту"
            />
          </div>
        </div>
      </header>

      <section class="dashboard">
        <div class="column column--objects">
          <div class="column-header">
            <h2>Объекты</h2>
          </div>

          <div v-if="loading" class="state">Загружаю объекты...</div>
          <div v-else-if="error" class="state state--error">
            {{ error }}
          </div>
          <div v-else>
            <div
              v-for="obj in filteredObjects"
              :key="obj.id"
              class="object-card"
            >
              <div class="object-card-main">
                <div>
                  <div class="object-name">{{ obj.name }}</div>
                  <div class="object-city">
                    {{ obj.city }}, {{ obj.address }}
                  </div>
                </div>
                <span class="status-chip">
                  {{ statusLabel(obj.status) }}
                </span>
              </div>

              <div class="object-actions">
                <button class="secondary-btn" @click="goToObject(obj.id)">
                  Перейти
                </button>
              </div>
            </div>

            <div v-if="!filteredObjects.length" class="state">
              Объектов нет
            </div>
          </div>
        </div>

        <!-- Правая колонка-заглушка -->
        <div class="column column--foremen">
          <div class="column-header">
            <h2>Информация</h2>
          </div>
          <div class="state">
            Здесь позже появятся сводки по поставкам и работам
          </div>
        </div>
      </section>
    </main>

    <!-- Правая колонка: карта -->
    <aside class="map-aside">
      <div class="column column--map">
        <div class="column-header">
          <h2>Карта</h2>
        </div>

        <div class="map-placeholder">
          Здесь будет карта ваших объектов
        </div>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.customer-layout {
  display: grid;
  grid-template-columns: 206px auto 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

/* Сайдбар */
.sidebar {
  grid-column: 1;
  width: 206px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px 18px;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
}

.sidebar-logo {
  font-size: 20px;
  font-weight: 700;
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

.nav-item[disabled] {
  opacity: 0.5;
  cursor: default;
}

.sidebar-bottom {
  display: flex;
  flex-direction: column;
  gap: 10px;
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

.role-dot--foreman {
  background: #f1ce06;
}

/* Центральная часть */
.customer-main {
  grid-column: 2;
  padding: 20px 24px;
  box-sizing: border-box;
  margin-left: 35px;
}

.customer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.customer-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

.customer-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-wrapper {
  max-width: 280px;
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

/* Дашборд */
.dashboard {
  display: grid;
  grid-template-columns: 359px 292px 445px;
  gap: 16px;
}

/* Колонки */
.column {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
  border: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.column-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.state {
  font-size: 13px;
  color: #6b7280;
}

.state--error {
  color: #b91c1c;
}

/* Карточка объекта */
.object-card {
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  padding: 10px 12px;
  margin-bottom: 8px;
  background: #f9fafb;
}

.object-card-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.object-name {
  font-weight: 600;
  font-size: 14px;
}

.object-city {
  font-size: 12px;
  color: #6b7280;
}

.status-chip {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
}

.status-chip--planned {
  background: #e5e7eb;
  color: #374151;
}

.status-chip--waiting {
  background: #fef3c7;
  color: #92400e;
}

.status-chip--active {
  background: #dcfce7;
  color: #166534;
}

.status-chip--finished {
  background: #e0f2fe;
  color: #1d4ed8;
}

.object-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.secondary-btn {
  padding: 6px 12px;
  border-radius: 999px;
  border: none;
  background: #e5e7eb;
  font-size: 12px;
  cursor: pointer;
}

/* Карта */
.map-placeholder {
  flex: 1;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
  background: #f9fafb;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 13px;
  margin-top: 8px;
}
</style>