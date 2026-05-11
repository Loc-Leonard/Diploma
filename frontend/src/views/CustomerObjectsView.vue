<template>
  <div class="customer-layout">
    <CustomerLayout />

    <main class="customer-main">
      <header class="customer-header">
        <div class="customer-header-left">
          <h1 class="customer-title">Объекты</h1>
          <button class="primary-btn" @click="goCreateObject">
            Создать объект
          </button>
          <button class="map-btn" @click="showMap = true">
            🗺 Показать на карте
          </button>
        </div>

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

      <!-- Фильтры -->
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

      <!-- Список объектов -->
      <section class="objects-section">
        <div v-if="objectsLoading" class="state">Загружаю объекты...</div>
        <div v-else-if="objectsError" class="state state--error">{{ objectsError }}</div>

        <template v-else>
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

            <div v-if="obj.activation_reject_reason" class="reject-notice">
              <span class="reject-notice-label">Причина отклонения:</span>
              <span>{{ obj.activation_reject_reason }}</span>
            </div>

            <div class="object-actions">
              <button
                v-if="obj.status === 'PLANNED'"
                class="primary-btn"
                @click="openActivateForm(obj.id)"
              >
                Активировать
              </button>
              <button
                v-else
                class="secondary-btn"
                @click="goToObject(obj.id)"
              >
                Перейти
              </button>
            </div>
          </div>

          <div v-if="!filteredObjects.length" class="state">Объектов нет</div>
        </template>
      </section>
    </main>

    <!-- Модалка активации -->
    <div v-if="activatingObjectId !== null" class="modal-overlay" @click.self="cancelActivate">
      <div class="modal-card">
        <h2>Активация объекта #{{ activatingObjectId }}</h2>
        <div class="form-field">
          <label>Чек-лист открытия (текст/JSON)</label>
          <textarea
            v-model="activateForm.checklist_json"
            rows="4"
            :disabled="activateLoading"
          />
        </div>
        <div class="form-field">
          <label>Путь к файлу акта</label>
          <input
            v-model="activateForm.act_file_path"
            type="text"
            placeholder="/files/acts/act-1.pdf"
            :disabled="activateLoading"
          />
        </div>
        <div v-if="activateError" class="state state--error">{{ activateError }}</div>
        <div class="modal-actions">
          <button class="secondary-btn" @click="cancelActivate">Отмена</button>
          <button class="primary-btn" @click="submitActivate" :disabled="activateLoading">
            {{ activateLoading ? 'Активируем...' : 'Активировать' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Модалка карты -->
    <div v-if="showMap" class="modal-overlay" @click.self="showMap = false">
      <div class="modal-card modal-card--map">
        <div class="modal-map-header">
          <h2>Объекты на карте</h2>
          <button class="close-btn" @click="showMap = false">✕</button>
        </div>
        <AppMap
          v-if="mapMarkers.length"
          :markers="mapMarkers"
          height="420px"
        />
          <div v-else class="map-placeholder">
            <div class="map-placeholder-inner">
            <span class="map-icon">🗺</span>
            <span>Нет объектов для отображения на карте</span>
            <span class="map-hint">{{ objects.length }} объект(ов) для отображения</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CustomerLayout from './CustomerLayout.vue'
import AppMap from '@/components/AppMap.vue'

const API_BASE = 'http://localhost:8080'

const auth = useAuthStore()
const router = useRouter()

function goCreateObject() {
  router.push({ name: 'customer-object-create' })
}

function goToObject(id: number) {
  router.push({ name: 'customer-object-details', params: { id } })
}

const greeting = computed(() => {
  const name = auth.user?.full_name
  return name ? `Добрый день, ${name}` : 'Добрый день'
})

const mapMarkers = computed(() =>
  objects.value
    .filter(
      (o) =>
        Number.isFinite(o.lat) &&
        Number.isFinite(o.lng) &&
        !(o.lat === 0 && o.lng === 0),
    )
    .map((o) => ({
      lat: o.lat,
      lng: o.lng,
      title: o.name,
      subtitle: `${o.city}, ${o.address}`,
    })),
)

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
  foreman?: { id: number; full_name: string } | null
  planned_start_date?: string | null
  planned_end_date?: string | null
  lat: number
  lng: number
  activation_reject_reason?: string | null
}

type ActivateForm = {
  checklist_json: string
  act_file_path: string
}

// Состояние
const objects = ref<DashboardObject[]>([])
const objectsLoading = ref(false)
const objectsError = ref<string | null>(null)

const search = ref('')
const statusFilter = ref('')
const cityFilter = ref('')

const showMap = ref(false)

const activatingObjectId = ref<number | null>(null)
const activateForm = ref<ActivateForm>({ checklist_json: '', act_file_path: '' })
const activateLoading = ref(false)
const activateError = ref<string | null>(null)

// Загрузка
async function fetchObjects() {
  objectsLoading.value = true
  objectsError.value = null
  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (cityFilter.value) params.set('city', cityFilter.value)

    const res = await fetch(
      `${API_BASE}/customer/dashboard/objects?${params.toString()}`,
      { headers: { Authorization: `Bearer ${auth.token}` } },
    )
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

onMounted(fetchObjects)
watch([statusFilter, cityFilter], fetchObjects)

// Вычисления
const filteredObjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return objects.value
  return objects.value.filter((o) => o.name.toLowerCase().includes(q))
})

const uniqueCities = computed(() => {
  const set = new Set<string>()
  objects.value.forEach((o) => { if (o.city) set.add(o.city) })
  return Array.from(set)
})

function statusLabel(status: DashboardObjectStatus) {
  switch (status) {
    case 'PLANNED': return 'Запланирован'
    case 'WAITING_INSPECTOR_CONFIRMATION': return 'Ожидает подтверждения'
    case 'ACTIVE': return 'Активен'
    case 'FINISHED': return 'Завершён'
    default: return status
  }
}

function statusClass(status: DashboardObjectStatus) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--waiting': status === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': status === 'ACTIVE',
    'status-chip--finished': status === 'FINISHED',
  }
}

function openActivateForm(objectId: number) {
  activatingObjectId.value = objectId
  activateError.value = null
  activateForm.value = { checklist_json: '', act_file_path: '' }
}

function cancelActivate() {
  activatingObjectId.value = null
}

async function submitActivate() {
  if (!activatingObjectId.value) return
  if (!activateForm.value.checklist_json.trim()) {
    activateError.value = 'Заполните чек-лист'
    return
  }

  activateLoading.value = true
  activateError.value = null
  try {
    const res = await fetch(
      `${API_BASE}/customer/objects/${activatingObjectId.value}/activate`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({
          checklist_json: activateForm.value.checklist_json,
          act_file_path: activateForm.value.act_file_path || undefined,
        }),
      },
    )
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка активации')
    }
    await fetchObjects()
    activatingObjectId.value = null
  } catch (e: any) {
    activateError.value = e.message || 'Ошибка'
  } finally {
    activateLoading.value = false
  }
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.customer-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  grid-template-rows: 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

/* Сайдбар */
.sidebar {
  grid-column: 1;
  grid-row: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px 18px;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
}

.sidebar-logo {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 24px;
  color: #111827;
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

.role-dot { width: 10px; height: 10px; border-radius: 999px; }
.role-dot--customer { background: #34c924; }

/* Основная область */
.customer-main {
  grid-column: 2;
  grid-row: 1;
  padding: 24px 32px;
  box-sizing: border-box;
  min-width: 0;
}

.customer-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.customer-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
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

.search-wrapper input {
  width: 240px;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
  outline: none;
}

/* Кнопки */
.primary-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  background: #4f46e5;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.primary-btn:hover:not(:disabled) { background: #4338ca; }
.primary-btn:disabled { opacity: 0.5; cursor: default; }

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

.secondary-btn:hover { background: #d1d5db; }

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

/* Фильтры */
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

/* Список объектов */
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

.status-chip--planned   { background: #e5e7eb; color: #374151; }
.status-chip--waiting   { background: #fef3c7; color: #92400e; }
.status-chip--active    { background: #dcfce7; color: #166534; }
.status-chip--finished  { background: #e0f2fe; color: #1d4ed8; }

.object-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.progress-bar {
  flex: 1;
  height: 5px;
  background: #e5e7eb;
  border-radius: 999px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: #4f46e5;
  border-radius: 999px;
}

.progress-text {
  font-size: 12px;
  color: #6b7280;
  min-width: 32px;
  text-align: right;
}

.object-people {
  margin-top: 8px;
  font-size: 12px;
  color: #374151;
}

.object-people .label { color: #9ca3af; margin-right: 4px; }

.reject-notice {
  margin-top: 8px;
  padding: 8px 10px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 8px;
  font-size: 12px;
  color: #92400e;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.reject-notice-label {
  font-weight: 600;
  color: #78350f;
}

.object-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

/* Состояния */
.state { font-size: 13px; color: #6b7280; padding: 8px 0; }
.state--error { color: #b91c1c; }

/* Модалки */
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
  max-width: 560px;
  background: #ffffff;
  border-radius: 16px;
  padding: 22px 24px 20px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
  box-sizing: border-box;
}

.modal-card h2 {
  margin: 0 0 16px;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.modal-card--map {
  max-width: 800px;
  padding: 20px 22px;
}

.modal-map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.modal-map-header h2 { margin: 0; }

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

.close-btn:hover { background: #f3f4f6; color: #374151; }

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

.map-icon { font-size: 40px; }

.map-hint {
  font-size: 12px;
  color: #d1d5db;
}

/* Форма активации */
.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.form-field label {
  font-size: 12px;
  color: #6b7280;
}

.form-field textarea,
.form-field input {
  border-radius: 10px;
  border: 1px solid #d1d5db;
  padding: 7px 10px;
  font-size: 14px;
  outline: none;
  background: #f9fafb;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.form-field textarea:focus,
.form-field input:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 1px rgba(129, 140, 248, 0.35);
  background: #ffffff;
}

.form-field textarea { resize: vertical; min-height: 80px; }

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

/* Адаптив */
@media (max-width: 900px) {
  .customer-main { padding: 16px 20px; max-width: 100%; }
}

@media (max-width: 768px) {
  .customer-layout { grid-template-columns: 1fr; }
  .sidebar { width: 100%; border-right: none; border-bottom: 1px solid #e5e7eb; padding: 16px; }
  .customer-main { padding: 16px; }
  .customer-header-left { width: 100%; }
  .customer-header-right { width: 100%; margin-left: 0; }
  .search-wrapper input { width: 100%; }
  .object-card-main { flex-direction: column; gap: 8px; }
  .modal-card--map { padding: 16px; }
  .map-placeholder { height: 280px; }
}
</style>