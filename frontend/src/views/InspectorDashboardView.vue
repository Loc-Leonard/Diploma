<template>
  <div class="customer-layout">
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>

        <nav class="sidebar-nav">
          <button
            class="nav-item"
            :class="{ 'nav-item--active': isChecksActive }"
            @click="goToChecks"
          >
            Проверки
          </button>

          <button class="nav-item" disabled>График</button>
          <button class="nav-item" disabled>Замечания</button>

          <button
            class="nav-item"
            :class="{ 'nav-item--active': isObjectsActive }"
            @click="goToObjects"
          >
            Объекты
          </button>

          <button class="nav-item" disabled>Справочники</button>
        </nav>
      </div>

      <div class="sidebar-bottom">
        <div class="role-badge">
          <span class="role-dot role-dot--inspector"></span>
          <span>Инспектор</span>
        </div>
        <button class="logout-button" @click="logout">Выйти</button>
      </div>
    </aside>

    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Проверки</h1>

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
            <h2>Проверки</h2>
            <div class="filters">
              <select v-model="statusFilter">
                <option value="">Статус</option>
                <option value="PLANNED">Запланирована</option>
                <option value="IN_PROGRESS">Идёт</option>
                <option value="OVERDUE">Просрочена</option>
                <option value="FINISHED">Завершена</option>
              </select>

              <select v-model="cityFilter">
                <option value="">Город</option>
                <option
                  v-for="city in uniqueCities"
                  :key="city"
                  :value="city"
                >
                  {{ city }}
                </option>
              </select>
            </div>
          </div>

          <div v-if="checksLoading" class="state">Загружаю проверки...</div>
          <div v-else-if="checksError" class="state state--error">
            {{ checksError }}
          </div>
          <div v-else>
            <div
              v-for="check in filteredChecks"
              :key="check.id"
              class="object-card"
            >
              <div class="object-card-main">
                <div>
                  <div class="object-name">{{ check.object_name }}</div>
                  <div class="object-city">
                    {{ check.city }}, {{ check.address }}
                  </div>
                </div>

                <span class="status-chip" :class="statusClass(check.status)">
                  {{ statusLabel(check.status) }}
                </span>
              </div>

              <div class="object-progress">
                <span class="progress-text">
                  Следующая проверка:
                  {{ formatDate(check.planned_at) }}
                </span>
              </div>

              <div class="object-people">
                <span class="label">Открытых замечаний:</span>
                <span>{{ check.issues_open }}</span>
              </div>

              <div class="object-actions">
                <button class="secondary-btn">Перейти</button>
              </div>
            </div>

            <div v-if="!filteredChecks.length" class="state">
              Проверок нет
            </div>
          </div>
        </div>

        <div class="column column--foremen">
          <div class="column-header">
            <h2>Объекты</h2>
          </div>

          <div v-if="objectsLoading" class="state">Загружаю объекты...</div>
          <div v-else-if="objectsError" class="state state--error">
            {{ objectsError }}
          </div>
          <div v-else>
            <div
              v-for="obj in objects"
              :key="obj.id"
              class="foreman-card"
            >
              <div class="foreman-name">{{ obj.name }}</div>
              <div class="foreman-city">
                {{ obj.city }}, {{ obj.address }}
              </div>
              <div class="foreman-object">
                Прораб: {{ obj.foreman_name || 'не назначен' }}
              </div>
              <div class="foreman-object">
                Активных проверок: {{ obj.active_checks }}
              </div>
              <div class="foreman-object">
                Просроченных: {{ obj.overdue_checks }}
              </div>
              <div class="foreman-object">
                Открытых замечаний: {{ obj.open_issues }}
              </div>
              <button class="secondary-btn" @click="goToObjects">
                Перейти
              </button>
            </div>

            <div v-if="!objects.length" class="state">
              Объектов нет
            </div>
          </div>
        </div>

        <div class="column column--map">
          <div class="column-header">
            <h2>Карта</h2>
            <div class="filters">
              <select v-model="statusFilter">
                <option value="">Статус</option>
                <option value="PLANNED">Запланирована</option>
                <option value="IN_PROGRESS">Идёт</option>
                <option value="OVERDUE">Просрочена</option>
                <option value="FINISHED">Завершена</option>
              </select>

              <select v-model="cityFilter">
                <option value="">Город</option>
                <option
                  v-for="city in uniqueCities"
                  :key="city"
                  :value="city"
                >
                  {{ city }}
                </option>
              </select>
            </div>
          </div>

          <div class="map-placeholder">
            Здесь будет карта с проверками
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const greeting = computed(() => {
  const u = auth.user
  return u?.full_name ? `Добрый день, ${u.full_name}` : 'Добрый день'
})

const isChecksActive = computed(() => route.name === 'inspector-checks')
const isObjectsActive = computed(() => route.name === 'inspector-objects')

type InspectionStatus =
  | 'PLANNED'
  | 'IN_PROGRESS'
  | 'FINISHED'
  | 'OVERDUE'

type DashboardInspection = {
  id: number
  object_id: number
  object_name: string
  city: string
  address: string
  status: InspectionStatus
  planned_at: string
  issues_open: number
}

type DashboardInspectorObject = {
  id: number
  name: string
  city: string
  address: string
  foreman_name: string | null
  active_checks: number
  overdue_checks: number
  open_issues: number
}

const checks = ref<DashboardInspection[]>([])
const checksLoading = ref(false)
const checksError = ref<string | null>(null)

const objects = ref<DashboardInspectorObject[]>([])
const objectsLoading = ref(false)
const objectsError = ref<string | null>(null)

const search = ref('')
const statusFilter = ref<string>('')
const cityFilter = ref<string>('')

function goToChecks() {
  if (route.name !== 'inspector-checks') {
    router.push({ name: 'inspector-checks' })
  }
}

function goToObjects() {
  if (route.name !== 'inspector-objects') {
    router.push({ name: 'inspector-objects' })
  }
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

async function fetchChecks() {
  checksLoading.value = true
  checksError.value = null

  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (cityFilter.value) params.set('city', cityFilter.value)

    const res = await fetch(
      `${API_BASE}/inspector/dashboard/checks?${params.toString()}`,
      {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      },
    )

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка загрузки проверок')
    }

    checks.value = await res.json()
  } catch (e: any) {
    checksError.value = e.message || 'Ошибка'
  } finally {
    checksLoading.value = false
  }
}

async function fetchObjects() {
  objectsLoading.value = true
  objectsError.value = null

  try {
    const res = await fetch(`${API_BASE}/inspector/dashboard/objects`, {
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
    objectsError.value = e.message || 'Ошибка'
  } finally {
    objectsLoading.value = false
  }
}

onMounted(() => {
  fetchChecks()
  fetchObjects()
})

watch([statusFilter, cityFilter], () => {
  fetchChecks()
})

const filteredChecks = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return checks.value

  return checks.value.filter((c) =>
    c.object_name.toLowerCase().includes(q),
  )
})

const uniqueCities = computed(() => {
  const set = new Set<string>()
  checks.value.forEach((c) => {
    if (c.city) set.add(c.city)
  })
  return Array.from(set)
})

function statusLabel(status: InspectionStatus) {
  switch (status) {
    case 'PLANNED':
      return 'Запланирована'
    case 'IN_PROGRESS':
      return 'Идёт'
    case 'FINISHED':
      return 'Завершена'
    case 'OVERDUE':
      return 'Просрочена'
    default:
      return status
  }
}

function statusClass(status: InspectionStatus) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--active': status === 'IN_PROGRESS',
    'status-chip--finished': status === 'FINISHED',
    'status-chip--overdue': status === 'OVERDUE',
  }
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return d.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.customer-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.primary-btn {
  padding: 8px 14px;
  border-radius: 999px;
  border: none;
  background: #4f46e5;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
}

.customer-layout {
  display: grid;
  grid-template-columns: 206px auto 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

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

.role-dot--inspector {
  background: #9524c9;
}

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

.dashboard {
  display: grid;
  grid-template-columns: 359px 292px 445px;
  gap: 16px;
}

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

.filters {
  display: flex;
  gap: 6px;
}

.filters select {
  padding: 6px 8px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 12px;
}

.state {
  font-size: 13px;
  color: #6b7280;
}

.state--error {
  color: #b91c1c;
}

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

.status-chip--overdue {
  background: #fee2e2;
  color: #b91c1c;
}

.object-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.progress-text {
  font-size: 12px;
  color: #4b5563;
}

.object-people {
  margin-top: 6px;
  font-size: 12px;
  color: #374151;
}

.object-people .label {
  color: #6b7280;
  margin-right: 4px;
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

.foreman-card {
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  padding: 10px 12px;
  margin-bottom: 8px;
  background: #f9fafb;
}

.foreman-name {
  font-weight: 600;
  font-size: 14px;
}

.foreman-city,
.foreman-object {
  font-size: 12px;
  color: #6b7280;
}

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