<template>
  <div class="customer-layout">
    <InspectorLayout />

    <main class="customer-main">
      <header class="customer-header">
        <div class="customer-header-left">
          <h1 class="customer-title">Объекты</h1>
          <button class="map-btn" @click="showMap = true">
            🗺 Показать на карте
          </button>
        </div>

        <div class="customer-header-right">
          <div class="search-wrapper">
            <input
              v-model="search"
              type="text"
              placeholder="Поиск по названию, городу, адресу"
            />
          </div>
        </div>
      </header>

      <!-- Фильтры -->
      <div class="filters-row">
        <select v-model="statusFilter">
          <option value="">Все статусы</option>
          <option value="WAITING_INSPECTOR_CONFIRMATION">Ожидает подтверждения</option>
          <option value="ACTIVE">Активен</option>
          <option value="PLANNED">Запланирован</option>
          <option value="FINISHED">Завершён</option>
        </select>
      </div>

      <!-- Секция: требуют подтверждения -->
      <section v-if="pendingObjects.length" class="objects-section pending-section">
        <div class="section-label">Требуют подтверждения</div>
        <div
          v-for="obj in pendingObjects"
          :key="obj.id"
          class="object-card object-card--pending"
        >
          <div class="object-card-main">
            <div>
              <div class="object-name">{{ obj.name }}</div>
              <div class="object-city">{{ obj.city }}, {{ obj.address }}</div>
            </div>
            <span class="status-chip status-chip--pending">
              Ожидает подтверждения
            </span>
          </div>

          <div class="object-people">
            <div>
              <span class="label">Прораб:</span>
              <span>{{ obj.foreman_name || '—' }}</span>
            </div>
            <div v-if="obj.planned_start_date">
              <span class="label">Плановая дата:</span>
              <span>{{ formatDate(obj.planned_start_date) }}</span>
            </div>
          </div>

          <div class="object-actions">
            <button class="primary-btn" @click="openApprovalModal(obj)">
              Рассмотреть
            </button>
          </div>
        </div>
      </section>

      <!-- Секция: все объекты -->
      <section class="objects-section">
        <div class="section-label">Все объекты</div>
        
        <div v-if="loading" class="state">Загружаю объекты...</div>
        <div v-else-if="error" class="state state--error">{{ error }}</div>

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
                  :style="{ width: `${normalizedProgress(obj.progress)}%` }"
                ></div>
              </div>
                <span class="progress-text">{{ normalizedProgress(obj.progress) }}%</span>
            </div>

            <div class="object-people">
              <div>
                <span class="label">Прораб:</span>
                <span>{{ obj.foreman_name || '—' }}</span>
              </div>
              <div v-if="obj.planned_start_date">
                <span class="label">Плановая дата начала:</span>
                <span>{{ formatDate(obj.planned_start_date) }}</span>
              </div>
            </div>

            <div v-if="obj.status === 'WAITING_INSPECTOR_CONFIRMATION' && obj.activation_reject_reason" class="reject-notice">
              <span class="reject-notice-label">Причина предыдущего отклонения:</span>
              <span>{{ obj.activation_reject_reason }}</span>
            </div>

            <div class="object-actions">
              <button
                v-if="obj.has_pending_action || obj.status === 'WAITING_INSPECTOR_CONFIRMATION'"
                class="primary-btn"
                @click="openApprovalModal(obj)"
              >
                Рассмотреть
              </button>
              <button
                v-else
                class="secondary-btn"
                @click="openDetailsPage(obj)"
              >
                Перейти
              </button>
            </div>
          </div>

          <div v-if="!filteredObjects.length && !pendingObjects.length" class="state">Объектов нет</div>
        </template>
      </section>
    </main>

    <!-- Модалка подтверждения/отклонения -->
    <Teleport to="body">
      <div v-if="modal.open" class="modal-overlay" @click.self="closeModal">
        <div class="modal-card">
          <h2>{{ modal.obj?.has_pending_action ? 'Подтверждение активации' : 'Объект' }}</h2>
          
          <div v-if="modal.loading" class="state">Загружаю данные...</div>
          <div v-else-if="modal.fetchError" class="state state--error">{{ modal.fetchError }}</div>

          <template v-else-if="modal.details">
            <div class="modal-details">
              <div class="modal-detail-row">
                <span class="label">Прораб</span>
                <span>{{ modal.details.foreman_name || '—' }}</span>
              </div>
              <div class="modal-detail-row">
                <span class="label">Статус</span>
                <span class="status-chip" :class="statusClass(modal.details.status)" style="font-size:12px">
                  {{ statusLabel(modal.details.status) }}
                </span>
              </div>
              <div class="modal-detail-row">
                <span class="label">Плановое начало</span>
                <span>{{ formatDate(modal.details.planned_start_date) }}</span>
              </div>
              <div class="modal-detail-row">
                <span class="label">Плановое окончание</span>
                <span>{{ formatDate(modal.details.planned_end_date) }}</span>
              </div>
              <div v-if="modal.details.description" class="modal-detail-row">
                <span class="label">Описание</span>
                <span>{{ modal.details.description }}</span>
              </div>
              <div v-if="modal.details.init_checklist_json" class="modal-detail-row modal-detail-col">
                <span class="label">Чеклист</span>
                <pre class="checklist-pre">{{ formatChecklist(modal.details.init_checklist_json) }}</pre>
              </div>
              <div v-if="modal.details.init_act_file_path" class="modal-detail-row">
                <span class="label">Акт</span>
                <span>{{ modal.details.init_act_file_path }}</span>
              </div>
              <div v-if="modal.details.activation_reject_reason" class="reject-notice">
                <span class="reject-notice-label">Причина предыдущего отклонения:</span>
                <span>{{ modal.details.activation_reject_reason }}</span>
              </div>
            </div>

            <div v-if="modal.mode === 'reject'" class="form-field">
              <label>Причина отклонения <span class="required">*</span></label>
              <textarea
                v-model="modal.rejectReason"
                rows="3"
                placeholder="Укажите причину отклонения активации..."
              />
              <div v-if="modal.submitError" class="state state--error">{{ modal.submitError }}</div>
            </div>
          </template>

          <div class="modal-actions">
            <template v-if="!modal.obj?.has_pending_action">
              <button class="secondary-btn" @click="closeModal">Закрыть</button>
            </template>

            <template v-else-if="modal.mode === 'idle'">
              <button class="secondary-btn" @click="closeModal">Отмена</button>
              <button class="reject-btn" @click="modal.mode = 'reject'" :disabled="modal.submitting">
                Отклонить
              </button>
              <button class="primary-btn" @click="submitDecision('APPROVE')" :disabled="modal.submitting">
                {{ modal.submitting ? 'Подтверждаю...' : 'Подтвердить' }}
              </button>
            </template>

            <template v-else-if="modal.mode === 'reject'">
              <button class="secondary-btn" @click="modal.mode = 'idle'" :disabled="modal.submitting">
                Назад
              </button>
              <button class="reject-btn" @click="submitDecision('REJECT')" :disabled="modal.submitting">
                {{ modal.submitting ? 'Отклоняю...' : 'Подтвердить отклонение' }}
              </button>
            </template>
          </div>
        </div>
      </div>
    </Teleport>

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
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useInspectorNotificationsStore } from '@/stores/inspectorNotifications'
import CustomerLayout from './CustomerLayout.vue'
import AppMap from '@/components/AppMap.vue'
import InspectorLayout from './InspectorLayout.vue'

const notifications = useInspectorNotificationsStore()
const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()
const showMap = ref(false)
const router = useRouter()

type InspectorObjectStatus = 'PLANNED' | 'WAITING_INSPECTOR_CONFIRMATION' | 'ACTIVE' | 'FINISHED'

type InspectorObjectItem = {
  id: number
  name: string
  city: string
  address: string
  status: InspectorObjectStatus
  foreman_name: string
  planned_start_date?: string | null
  has_pending_action: boolean
  lat: number
  lng: number
  activation_reject_reason?: string | null
  progress?: number
}

type ObjectDetails = {
  id: number
  name: string
  city: string
  address: string
  description: string
  status: InspectorObjectStatus
  foreman_name: string
  planned_start_date?: string | null
  planned_end_date?: string | null
  actual_start_date?: string | null
  init_checklist_json: string
  init_act_file_path: string
  activation_reject_reason: string
}

const objects = ref<InspectorObjectItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const search = ref('')
const statusFilter = ref('')

const modal = reactive({
  open: false,
  obj: null as InspectorObjectItem | null,
  details: null as ObjectDetails | null,
  loading: false,
  fetchError: null as string | null,
  mode: 'idle' as 'idle' | 'reject',
  rejectReason: '',
  submitError: null as string | null,
  submitting: false,
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

const pendingObjects = notifications.pendingObjects

const filteredObjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  return objects.value.filter(o => {
    const matchesSearch =
      !q ||
      o.name.toLowerCase().includes(q) ||
      o.city.toLowerCase().includes(q) ||
      o.address.toLowerCase().includes(q) ||
      (o.foreman_name || '').toLowerCase().includes(q)
    const matchesStatus = !statusFilter.value || o.status === statusFilter.value
    return matchesSearch && matchesStatus
  })
})

async function fetchObjects() {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    const res = await fetch(`${API_BASE}/inspector/objects?${params}`, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'Ошибка загрузки')
    objects.value = await res.json()
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}
function normalizedProgress(value?: number | null){
  if (typeof value !== 'number' || Number.isNaN(value)) return 0
  return Math.min(100, Math.max(0, Math.round(value)))
}
async function fetchObjectDetails(id: number) {
  modal.loading = true
  modal.fetchError = null
  modal.details = null

  try {
    const res = await fetch(`${API_BASE}/inspector/objects/${id}`, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })

    if (!res.ok) {
      throw new Error((await res.json().catch(() => ({}))).error ?? 'Ошибка загрузки')
    }

    const data = await res.json()
    const core = data.object

    modal.details = {
      id: core.id,
      name: core.name ?? '',
      city: core.city ?? '',
      address: core.address ?? '',
      description: core.description ?? '',
      status: core.status ?? 'PLANNED',
      foreman_name: core.foreman?.full_name ?? '',
      planned_start_date: core.planned_start_date ?? null,
      planned_end_date: core.planned_end_date ?? null,
      actual_start_date: core.actual_start_date ?? null,
      init_checklist_json: core.init_checklist_json ?? '',
      init_act_file_path: core.init_act_file_path ?? '',
      activation_reject_reason: core.activation_reject_reason ?? '',
    }
  } catch (e: any) {
    modal.fetchError = e.message
  } finally {
    modal.loading = false
  }
}

function openApprovalModal(obj: InspectorObjectItem) {
  modal.open = true
  modal.obj = obj
  modal.details = null
  modal.mode = 'idle'
  modal.rejectReason = ''
  modal.submitError = null
  modal.submitting = false
  fetchObjectDetails(obj.id)
}

function openDetailsPage(obj: InspectorObjectItem) {
  router.push({ name: 'inspector-object-details', params: { id: obj.id } })
}

function closeModal() {
  modal.open = false
  modal.obj = null
  modal.details = null
  modal.mode = 'idle'
  modal.rejectReason = ''
  modal.submitError = null
}

async function submitDecision(decision: 'APPROVE' | 'REJECT') {
  if (!modal.obj) return

  if (decision === 'REJECT' && !modal.rejectReason.trim()) {
    modal.submitError = 'Укажите причину отклонения'
    return
  }

  modal.submitting = true
  modal.submitError = null
  try {
    const body: Record<string, string> = { decision }
    if (decision === 'REJECT') body.rejection_reason = modal.rejectReason.trim()

    const res = await fetch(`${API_BASE}/inspector/objects/${modal.obj.id}/activation-decision`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${auth.token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    if (!res.ok) {
      throw new Error((await res.json().catch(() => ({}))).error || 'Ошибка отправки')
    }

    closeModal()
    await Promise.all([
      fetchObjects(),
      notifications.fetchPending(),
    ])
  } catch (e: any) {
    modal.submitError = e.message || 'Ошибка'
  } finally {
    modal.submitting = false
  }
}

function formatChecklist(json: string) {
  try { return JSON.stringify(JSON.parse(json), null, 2) }
  catch { return json }
}

function statusLabel(status: InspectorObjectStatus) {
  const map: Record<InspectorObjectStatus, string> = {
    PLANNED: 'Запланирован',
    WAITING_INSPECTOR_CONFIRMATION: 'Ожидает подтверждения',
    ACTIVE: 'Активен',
    FINISHED: 'Завершён',
  }
  return map[status] ?? status
}

function statusClass(status: InspectorObjectStatus) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--waiting': status === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': status === 'ACTIVE',
    'status-chip--finished': status === 'FINISHED',
  }
}

function formatDate(value?: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleDateString('ru-RU')
}

watch(statusFilter, fetchObjects)
onMounted(fetchObjects)
</script>

<style scoped>
/* === Базовый лейаут (как у клиента) === */
.customer-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  grid-template-rows: 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

.customer-main {
  grid-column: 2;
  grid-row: 1;
  padding: 24px 32px;
  box-sizing: border-box;
  min-width: 0;
}

/* === Хедер === */
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

/* === Кнопки === */
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

.reject-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  font-size: 14px;
  cursor: pointer;
  background: #fee2e2;
  color: #b91c1c;
  transition: background 0.15s;
}
.reject-btn:hover:not(:disabled) { background: #fecaca; }
.reject-btn:disabled { opacity: 0.6; cursor: default; }

/* === Фильтры === */
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

/* === Список объектов === */
.objects-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-label {
  font-size: 14px;
  font-weight: 600;
  color: #6b7280;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
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
.object-card--pending {
  border-color: #fca5a5;
  box-shadow: 0 2px 8px rgba(220, 38, 38, 0.08);
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

/* === Статусы === */
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
.status-chip--pending   { background: #fee2e2; color: #b91c1c; }

/*=== Progress Bar === */
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

/* === Мета-информация === */
.object-people {
  margin-top: 8px;
  font-size: 12px;
  color: #374151;
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
.object-people .label { color: #9ca3af; margin-right: 4px; }

/* === Уведомление об отклонении === */
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

/* === Действия === */
.object-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

/* === Состояния === */
.state { font-size: 13px; color: #6b7280; padding: 8px 0; }
.state--error { color: #b91c1c; }

/* === Модалки === */
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

/* === Карта === */
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

/* === Форма в модалке === */
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

.required { color: #dc2626; }

/* === Детали объекта в модалке === */
.modal-details { display: flex; flex-direction: column; gap: 10px; margin-bottom: 8px; }
.modal-detail-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 14px;
  color: #374151;
}
.modal-detail-row .label { color: #6b7280; min-width: 160px; flex-shrink: 0; }
.modal-detail-col { flex-direction: column; gap: 6px; }
.modal-detail-col .label { min-width: unset; }

.checklist-pre {
  font-size: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 10px;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

/* === Адаптив === */
@media (max-width: 900px) {
  .customer-main { padding: 16px 20px; max-width: 100%; }
}

@media (max-width: 768px) {
  .customer-layout { grid-template-columns: 1fr; }
  .customer-main { padding: 16px; }
  .customer-header-left { width: 100%; }
  .customer-header-right { width: 100%; margin-left: 0; }
  .search-wrapper input { width: 100%; }
  .object-card-main { flex-direction: column; gap: 8px; }
  .modal-card--map { padding: 16px; }
  .map-placeholder { height: 280px; }
}
</style>