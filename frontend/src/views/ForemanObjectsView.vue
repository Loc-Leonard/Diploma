<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import ForemanLayout from './ForemanLayout.vue'

const API_BASE = 'http://localhost:8080'

const router = useRouter()
const auth = useAuthStore()

type ForemanObjectStatus =
  | 'PLANNED'
  | 'WAITING_INSPECTOR_CONFIRMATION'
  | 'ACTIVE'
  | 'FINISHED'

type ForemanObject = {
  id: number
  name: string
  city: string
  address: string
  status: ForemanObjectStatus
}

const objects = ref<ForemanObject[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const search = ref('')
const statusFilter = ref('')
const cityFilter = ref('')
const showMap = ref(false)

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

  return objects.value.filter((o) => {
    const matchesSearch =
      !q ||
      o.name.toLowerCase().includes(q) ||
      o.city.toLowerCase().includes(q) ||
      o.address.toLowerCase().includes(q)

    const matchesStatus =
      !statusFilter.value || o.status === statusFilter.value

    const matchesCity =
      !cityFilter.value || o.city === cityFilter.value

    return matchesSearch && matchesStatus && matchesCity
  })
})

const uniqueCities = computed(() => {
  const set = new Set<string>()
  objects.value.forEach((o) => {
    if (o.city) set.add(o.city)
  })
  return Array.from(set)
})

function statusLabel(status: ForemanObjectStatus) {
  switch (status) {
    case 'PLANNED':
      return 'Запланирован'
    case 'WAITING_INSPECTOR_CONFIRMATION':
      return 'Ожидает подтверждения'
    case 'ACTIVE':
      return 'Активен'
    case 'FINISHED':
      return 'Завершён'
    default:
      return status
  }
}

function statusClass(status: ForemanObjectStatus) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--waiting': status === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': status === 'ACTIVE',
    'status-chip--finished': status === 'FINISHED',
  }
}

function goToObject(id: number) {
  router.push({ name: 'foreman-object', params: { id } })
}

onMounted(loadObjects)
</script>

<template>
  <div class="foreman-layout-page">
    <ForemanLayout />

    <main class="foreman-main">
      <header class="foreman-header">
        <div class="foreman-header-left">
          <h1 class="foreman-title">Мои объекты</h1>
          <button class="map-btn" @click="showMap = true">
            Показать на карте
          </button>
        </div>

        <div class="foreman-header-right">
          <div class="search-wrapper">
            <input
              v-model="search"
              type="text"
              placeholder="Поиск по объекту"
            />
          </div>
        </div>
      </header>

      <div class="filters-row">
        <select v-model="statusFilter">
          <option value="">Все статусы</option>
          <option value="PLANNED">Запланирован</option>
          <option value="WAITING_INSPECTOR_CONFIRMATION">Ожидает подтверждения</option>
          <option value="ACTIVE">Активен</option>
          <option value="FINISHED">Завершён</option>
        </select>

        <select v-model="cityFilter">
          <option value="">Все города</option>
          <option v-for="city in uniqueCities" :key="city" :value="city">
            {{ city }}
          </option>
        </select>
      </div>

      <section class="objects-section">
        <div v-if="loading" class="state">Загружаю объекты...</div>
        <div v-else-if="error" class="state state--error">
          {{ error }}
        </div>

        <template v-else>
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

              <span class="status-chip" :class="statusClass(obj.status)">
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
        </template>
      </section>
    </main>

    <div v-if="showMap" class="modal-overlay" @click.self="showMap = false">
      <div class="modal-card modal-card--map">
        <div class="modal-map-header">
          <h2>Объекты на карте</h2>
          <button class="close-btn" @click="showMap = false">✕</button>
        </div>

        <div class="map-placeholder">
          <div class="map-placeholder-inner">
            <span class="map-icon">🗺</span>
            <span>Здесь будет карта с объектами</span>
            <span class="map-hint">{{ objects.length }} объект(ов) для отображения</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.foreman-layout-page {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

.foreman-main {
  padding: 24px 32px;
  box-sizing: border-box;
  min-width: 0;
}

.foreman-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.foreman-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.foreman-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

.foreman-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.search-wrapper input {
  width: 240px;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
  outline: none;
}

.map-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #374151;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.map-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.filters-row {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.filters-row select {
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 13px;
  min-width: 160px;
  cursor: pointer;
}

.objects-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.object-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 14px 16px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04);
  transition: box-shadow 0.15s;
}

.object-card:hover {
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.08);
}

.object-card-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.object-name {
  font-weight: 600;
  font-size: 15px;
  color: #111827;
}

.object-city {
  font-size: 12px;
  color: #6b7280;
  margin-top: 2px;
}

.status-chip {
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
}

.status-chip--planned { background: #e5e7eb; color: #374151; }
.status-chip--waiting { background: #fef3c7; color: #92400e; }
.status-chip--active { background: #dcfce7; color: #166534; }
.status-chip--finished { background: #e0f2fe; color: #1d4ed8; }

.object-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.secondary-btn {
  padding: 6px 14px;
  border-radius: 999px;
  border: none;
  background: #e5e7eb;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.15s;
}

.secondary-btn:hover {
  background: #d1d5db;
}

.state {
  font-size: 13px;
  color: #6b7280;
  padding: 8px 0;
}

.state--error {
  color: #b91c1c;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 16px;
  z-index: 50;
}

.modal-card {
  width: 100%;
  max-width: 800px;
  background: #ffffff;
  border-radius: 16px;
  padding: 20px 22px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
  box-sizing: border-box;
}

.modal-card--map {
  max-width: 800px;
}

.modal-map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.modal-map-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  color: #9ca3af;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.15s, color 0.15s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.map-placeholder {
  height: 420px;
  border-radius: 12px;
  border: 2px dashed #d1d5db;
  background: #f9fafb;
  display: flex;
  align-items: center;
  justify-content: center;
}

.map-placeholder-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #9ca3af;
  font-size: 14px;
}

.map-icon {
  font-size: 40px;
}

.map-hint {
  font-size: 12px;
  color: #d1d5db;
}

@media (max-width: 900px) {
  .foreman-main {
    padding: 16px 20px;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .foreman-layout-page {
    grid-template-columns: 1fr;
  }

  .foreman-main {
    padding: 16px;
  }

  .foreman-header-left {
    width: 100%;
  }

  .foreman-header-right {
    width: 100%;
    margin-left: 0;
  }

  .search-wrapper input {
    width: 100%;
  }

  .object-card-main {
    flex-direction: column;
    gap: 8px;
  }

  .modal-card--map {
    padding: 16px;
  }

  .map-placeholder {
    height: 280px;
  }
}
</style>