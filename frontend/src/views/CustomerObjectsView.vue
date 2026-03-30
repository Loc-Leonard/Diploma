<template>
  <div class="customer-layout">
    <!-- Левое меню -->
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">ЛОГОТИП</div>

        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active">Объекты</button>
          <button class="nav-item" disabled>График</button>
          <button class="nav-item" disabled>Замечания</button>
          <button class="nav-item" disabled>Проверки</button>
          <button class="nav-item" disabled>Справочники</button>
        </nav>
      </div>

      <div class="sidebar-bottom">
        <div class="role-badge">
            <span class="role-dot role-dot--customer"></span>
            <span>Заказчик</span>
        </div>
        <button class="logout-button" @click="logout">Выйти</button>
      </div>
    </aside>

    <!-- Центральная часть: Объекты + Прорабы -->
    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Объекты</h1>

        <div class="customer-header-right">
          <!-- Поиск по названию объекта -->
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
        <!-- Колонка Объекты -->
        <div class="column column--objects">
          <div class="column-header">
            <h2>Объекты</h2>
            <div class="filters">
              <select v-model="statusFilter">
                <option value="">Статус</option>
                <option value="PLANNED">Запланирован</option>
                <option value="WAITING_INSPECTOR_CONFIRMATION">
                  Ожидает подтверждения
                </option>
                <option value="ACTIVE">Активен</option>
                <option value="FINISHED">Завершен</option>
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

          <div v-if="objectsLoading" class="state">Загружаю объекты...</div>
          <div v-else-if="objectsError" class="state state--error">
            {{ objectsError }}
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
                  <div class="object-city">{{ obj.city }}, {{ obj.address }}</div>
                </div>
                <span class="status-chip" :class="statusClass(obj.status)">
                  {{ statusLabel(obj.status) }}
                </span>
              </div>

              <div class="object-progress">
                <div class="progress-bar">
                  <div
                    class="progress-bar-fill"
                    :style="{ width: obj.progress + '%' }"
                  ></div>
                </div>
                <span class="progress-text">{{ obj.progress }}%</span>
              </div>

              <div class="object-people">
                <div v-if="obj.foreman">
                  <span class="label">Прораб:</span>
                  <span>{{ obj.foreman.full_name }}</span>
                </div>
              </div>

              <div class="object-actions">
                <button class="secondary-btn">Перейти</button>
              </div>
            </div>

            <div v-if="!filteredObjects.length" class="state">
              Объектов нет
            </div>
          </div>
        </div>

        <!-- Колонка Прорабы -->
        <div class="column column--foremen">
          <div class="column-header">
            <h2>Прорабы</h2>
          </div>

          <div v-if="foremenLoading" class="state">Загружаю прорабов...</div>
          <div v-else-if="foremenError" class="state state--error">
            {{ foremenError }}
          </div>
          <div v-else>
            <div
              v-for="f in foremen"
              :key="f.id"
              class="foreman-card"
            >
              <div class="foreman-name">{{ f.full_name }}</div>
              <div class="foreman-city">{{ f.city }}</div>
              <div class="foreman-object" v-if="f.current_object">
                Объект: {{ f.current_object.name }}
              </div>
              <button class="secondary-btn">Перейти</button>
            </div>

            <div v-if="!foremen.length" class="state">
              Прорабов нет
            </div>
          </div>
        </div>

        <!-- Колонка Карта -->
        <div class="column column--map">
          <div class="column-header">
            <h2>Карта</h2>
            <div class="filters">
              <!-- пока те же фильтры, что и над объектами -->
              <select v-model="statusFilter">
                <option value="">Статус</option>
                <option value="PLANNED">Запланирован</option>
                <option value="WAITING_INSPECTOR_CONFIRMATION">
                  Ожидает подтверждения
                </option>
                <option value="ACTIVE">Активен</option>
                <option value="FINISHED">Завершен</option>
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

          <!-- Здесь позже подключим Leaflet/OSM -->
          <div class="map-placeholder">
            Здесь будет карта с объектами
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'

const auth = useAuthStore()
const router = useRouter()

// ==== Типы данных (под API бэка) ====

type DashboardObjectStatus =
  | 'PLANNED'
  | 'WAITING_INSPECTOR_CONFIRMATION'
  | 'ACTIVE'
  | 'FINISHED'

type DashboardObject = {
  id: number
  name: string
  city: string
  address: string
  status: DashboardObjectStatus
  progress: number
  foreman?: {
    id: number
    full_name: string
  } | null
  planned_start_date?: string | null
  planned_end_date?: string | null
  lat: number
  lng: number
}

type DashboardForeman = {
  id: number
  full_name: string
  city: string
  current_object?: {
    id: number
    name: string
  } | null
}

// ==== Состояние ====

const objects = ref<DashboardObject[]>([])
const objectsLoading = ref(false)
const objectsError = ref<string | null>(null)

const foremen = ref<DashboardForeman[]>([])
const foremenLoading = ref(false)
const foremenError = ref<string | null>(null)

// фильтры и поиск
const search = ref('')
const statusFilter = ref<string>('')
const cityFilter = ref<string>('')

// ==== Навигация ====

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

// ==== Загрузка данных ====

async function fetchObjects() {
  objectsLoading.value = true
  objectsError.value = null

  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (cityFilter.value) params.set('city', cityFilter.value)

    const res = await fetch(
      `${API_BASE}/customer/dashboard/objects?${params.toString()}`,
      {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      },
    )

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка загрузки объектов')
    }

    const data = await res.json()
    objects.value = data
  } catch (e: any) {
    objectsError.value = e.message || 'Ошибка'
  } finally {
    objectsLoading.value = false
  }
}

async function fetchForemen() {
  foremenLoading.value = true
  foremenError.value = null

  try {
    const res = await fetch(`${API_BASE}/customer/dashboard/foremen`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка загрузки прорабов')
    }

    const data = await res.json()
    foremen.value = data
  } catch (e: any) {
    foremenError.value = e.message || 'Ошибка'
  } finally {
    foremenLoading.value = false
  }
}

// При первом входе грузим данные
onMounted(() => {
  fetchObjects()
  fetchForemen()
})

// Подгружать объекты при изменении фильтров (по-хорошему — с debounce)
watch([statusFilter, cityFilter], () => {
  fetchObjects()
})

// ==== Вычисления для отображения ====

const filteredObjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return objects.value
  return objects.value.filter((o) =>
    o.name.toLowerCase().includes(q),
  )
})

const uniqueCities = computed(() => {
  const set = new Set<string>()
  objects.value.forEach((o) => {
    if (o.city) set.add(o.city)
  })
  return Array.from(set)
})

function statusLabel(status: DashboardObjectStatus) {
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

function statusClass(status: DashboardObjectStatus) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--waiting':
      status === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': status === 'ACTIVE',
    'status-chip--finished': status === 'FINISHED',
  }
}
</script>

<style scoped>
.customer-layout {
  display: grid;
  /* меню 206, центр, карта 445 */
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

.role-dot--customer {
  background: #34c924;
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

/* Дашборд: только ширины колонок */
.dashboard {
  display: grid;
  grid-template-columns: 359px 292px 445px; /* как в макете */
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

/* Состояние */
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

.object-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: #e5e7eb;
  border-radius: 999px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: #4f46e5;
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

/* Карточка прораба */
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