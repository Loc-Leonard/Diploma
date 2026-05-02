<template>
  <div>
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

    <section class="checks-layout">
      <aside class="checks-filters">
        <h2 class="filters-title">Фильтры</h2>

        <label class="filter-label">
          Статус
          <select v-model="statusFilter">
            <option value="">Все</option>
            <option value="PLANNED">Запланирована</option>
            <option value="IN_PROGRESS">Идёт</option>
            <option value="OVERDUE">Просрочена</option>
            <option value="FINISHED">Завершена</option>
          </select>
        </label>

        <label class="filter-label">
          Город
          <select v-model="cityFilter">
            <option value="">Все</option>
            <option
              v-for="city in uniqueCities"
              :key="city"
              :value="city"
            >
              {{ city }}
            </option>
          </select>
        </label>
      </aside>

      <section class="checks-list">
        <div class="list-header">
          <h2>Список проверок</h2>
          <span class="count-text">{{ filteredChecks.length }}</span>
        </div>

        <div v-if="checksLoading" class="state">Загружаю проверки...</div>
        <div v-else-if="checksError" class="state state--error">
          {{ checksError }}
        </div>

        <template v-else>
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
                Следующая проверка: {{ formatDate(check.planned_at) }}
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
        </template>
      </section>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()

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

const checks = ref<DashboardInspection[]>([])
const checksLoading = ref(false)
const checksError = ref<string | null>(null)

const search = ref('')
const statusFilter = ref<string>('')
const cityFilter = ref<string>('')

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

onMounted(() => {
  fetchChecks()
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
.customer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
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
  margin-left: auto;
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

.checks-layout {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 16px;
}

.checks-filters {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-self: start;
}

.filters-title {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.filter-label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: #4b5563;
}

.filter-label select {
  padding: 6px 8px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 12px;
}

.checks-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.list-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.list-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.count-text {
  font-size: 13px;
  color: #6b7280;
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
  padding: 12px 14px;
  background: #ffffff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}

.object-card-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.object-name {
  font-weight: 600;
  font-size: 14px;
  color: #111827;
}

.object-city {
  font-size: 12px;
  color: #6b7280;
}

.status-chip {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  white-space: nowrap;
}

.status-chip--planned {
  background: #e5e7eb;
  color: #374151;
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
  margin-top: 8px;
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
  margin-top: 10px;
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

@media (max-width: 900px) {
  .checks-layout {
    grid-template-columns: 1fr;
  }
}
</style>