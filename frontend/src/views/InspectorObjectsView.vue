<template>
  <div>
      <header class="customer-header">
        <div class="customer-header-left">
          <h1 class="customer-title">Объекты</h1>
          <button class="map-btn" @click="showMap = true">
            🗺 Показать на карте
          </button>
        </div>
        <div class="customer-header-right">
          <div class="search-wrapper">
            <input v-model="search" type="text" placeholder="Поиск по названию, городу, адресу" />
          </div>
          <select v-model="statusFilter" class="filter-select">
            <option value="">Все статусы</option>
            <option value="WAITING_INSPECTOR_CONFIRMATION">Ожидают подтверждения</option>
            <option value="ACTIVE">Активные</option>
            <option value="PLANNED">Запланированные</option>
            <option value="FINISHED">Завершённые</option>
          </select>
        </div>
      </header>

      <!-- Секция: требуют подтверждения -->
      <section v-if="pendingObjects.length" class="pending-section">
        <div class="section-header">
          <h2>Требуют подтверждения</h2>
          <span class="pending-pill">{{ pendingObjects.length }}</span>
        </div>
        <div class="object-grid">
          <article
            v-for="obj in pendingObjects"
            :key="obj.id"
            class="object-card object-card--pending"
          >
            <div class="object-top">
              <div>
                <div class="object-name">{{ obj.name }}</div>
                <div class="object-address">{{ obj.city }}, {{ obj.address }}</div>
              </div>
              <span class="status-chip status-chip--pending">Ожидает подтверждения</span>
            </div>
            <div class="object-meta">
              <div><span class="label">Прораб:</span> {{ obj.foreman_name || '—' }}</div>
              <div><span class="label">Плановая дата:</span> {{ formatDate(obj.planned_start_date) }}</div>
            </div>
            <div class="object-actions">
              <button class="primary-btn" @click="openModal(obj)">Рассмотреть</button>
            </div>
          </article>
        </div>
      </section>

      <!-- Секция: все объекты -->
      <section class="list-section">
        <div class="section-header">
          <h2>Все объекты</h2>
          <span class="count-text">{{ filteredObjects.length }}</span>
        </div>
        <div v-if="loading" class="state">Загружаю объекты...</div>
        <div v-else-if="error" class="state state--error">{{ error }}</div>
        <div v-else-if="!filteredObjects.length" class="state">Объектов пока нет</div>
        <div v-else class="object-grid">
          <article v-for="obj in filteredObjects" :key="obj.id" class="object-card">
            <div class="object-top">
              <div>
                <div class="object-name">{{ obj.name }}</div>
                <div class="object-address">{{ obj.city }}, {{ obj.address }}</div>
              </div>
              <span class="status-chip" :class="statusClass(obj.status)">
                {{ statusLabel(obj.status) }}
              </span>
            </div>
            <div class="object-meta">
              <div><span class="label">Прораб:</span> {{ obj.foreman_name || '—' }}</div>
              <div><span class="label">Плановая дата:</span> {{ formatDate(obj.planned_start_date) }}</div>
            </div>
            <div class="object-actions">
              <button v-if="obj.has_pending_action" class="primary-btn" @click="openModal(obj)">
                Рассмотреть
              </button>
              <button v-else class="secondary-btn" @click="openModal(obj)">
                Открыть
              </button>
            </div>
          </article>
        </div>
      </section>

    <!-- Модалка -->
    <Teleport to="body">
      <div v-if="modal.open" class="modal-overlay" @click.self="closeModal">
        <div class="modal">

          <div class="modal-header">
            <h2>{{ modal.obj?.has_pending_action ? 'Подтверждение активации' : 'Объект' }}</h2>
            <button class="modal-close" @click="closeModal">✕</button>
          </div>

          <div class="modal-body">
            <div class="modal-object-name">{{ modal.obj?.name }}</div>
            <div class="modal-object-address">{{ modal.obj?.city }}, {{ modal.obj?.address }}</div>

            <div v-if="modal.loading" class="state">Загружаю данные...</div>
            <div v-else-if="modal.fetchError" class="state state--error">{{ modal.fetchError }}</div>

            <div v-else-if="modal.details" class="modal-details">
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
                <span class="label">Причина предыдущего отклонения:</span>
                <span>{{ modal.details.activation_reject_reason }}</span>
              </div>
            </div>

            <!-- Форма отклонения -->
            <div v-if="modal.mode === 'reject'" class="reject-form">
              <label class="reject-label">
                Причина отклонения <span class="required">*</span>
              </label>
              <textarea
                v-model="modal.rejectReason"
                class="reject-textarea"
                placeholder="Укажите причину отклонения активации..."
                rows="3"
              />
              <div v-if="modal.submitError" class="state state--error">{{ modal.submitError }}</div>
            </div>
          </div>

          <div class="modal-footer">
            <!-- Объект не требует действия — просто закрыть -->
            <template v-if="!modal.obj?.has_pending_action">
              <button class="secondary-btn" @click="closeModal">Закрыть</button>
            </template>

            <!-- Режим выбора действия -->
            <template v-else-if="modal.mode === 'idle'">
              <button class="secondary-btn" @click="closeModal">Закрыть</button>
              <button class="reject-btn" @click="modal.mode = 'reject'" :disabled="modal.submitting">
                Отклонить
              </button>
              <button class="primary-btn" @click="submitDecision('APPROVE')" :disabled="modal.submitting">
                {{ modal.submitting ? 'Подтверждаю...' : 'Подтвердить' }}
              </button>
            </template>

            <!-- Режим ввода причины отклонения -->
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
    <!--MAP-->
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
            <span class="map-hint">
              {{ filteredObjects.length }} Объектов для отображения
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useInspectorNotificationsStore } from '@/stores/inspectorNotifications'
const notifications = useInspectorNotificationsStore()
const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()
const showMap = ref(false)


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



const pendingObjects = notifications.pendingObjects

const filteredObjects = computed(() => {
  const q = search.value.trim().toLowerCase()
  return objects.value.filter(o => {
    const matchesSearch = !q ||
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

async function fetchObjectDetails(id: number) {
  modal.loading = true
  modal.fetchError = null
  modal.details = null
  try {
    const res = await fetch(`${API_BASE}/inspector/objects/${id}`, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'Ошибка загрузки объекта')
    modal.details = await res.json()
  } catch (e: any) {
    modal.fetchError = e.message || 'Ошибка'
  } finally {
    modal.loading = false
  }
}

function openModal(obj: InspectorObjectItem) {
  modal.open = true
  modal.obj = obj
  modal.details = null
  modal.mode = 'idle'
  modal.rejectReason = ''
  modal.submitError = null
  modal.submitting = false
  fetchObjectDetails(obj.id)
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
    'status-chip--pending': status === 'WAITING_INSPECTOR_CONFIRMATION',
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

.customer-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
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

.modal-card--map {
  width: 100%;
  max-width: 800px;
  padding: 20px 22px;
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

.modal-card {
  width: 100%;
  max-width: 560px;
  background: #ffffff;
  border-radius: 16px;
  padding: 22px 24px 20px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
  box-sizing: border-box;
}

.customer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.customer-title { margin: 0; font-size: 22px; font-weight: 600; color: #111827; }

.customer-header-right { display: flex; align-items: center; gap: 12px; }
.search-wrapper { width: 280px; }

.search-wrapper input,
.filter-select {
  width: 100%;
  padding: 8px 11px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #ffffff;
  font-size: 14px;
}

.pending-section, .list-section { margin-bottom: 20px; }

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}
.section-header h2 { margin: 0; font-size: 16px; font-weight: 600; }
.pending-pill, .count-text { font-size: 13px; color: #6b7280; }

.object-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
}

.object-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}
.object-card--pending {
  border-color: #fca5a5;
  box-shadow: 0 10px 24px rgba(220, 38, 38, 0.08);
}

.object-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.object-name { font-size: 16px; font-weight: 600; color: #111827; }
.object-address { margin-top: 4px; font-size: 13px; color: #6b7280; }

.object-meta { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #374151; }
.label { color: #6b7280; }

.object-actions { margin-top: 14px; display: flex; justify-content: flex-end; }

.primary-btn, .secondary-btn {
  padding: 8px 14px;
  border-radius: 999px;
  border: none;
  font-size: 14px;
  cursor: pointer;
}
.primary-btn { background: #c4b5fd; color: #111827; }
.primary-btn:disabled { opacity: 0.6; cursor: default; }
.secondary-btn { background: #ffffff; color: #374151; border: 1px solid #d1d5db; }

.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  white-space: nowrap;
}
.status-chip--planned  { background: #e5e7eb; color: #374151; }
.status-chip--pending  { background: #fee2e2; color: #b91c1c; }
.status-chip--active   { background: #dcfce7; color: #166534; }
.status-chip--finished { background: #dbeafe; color: #1d4ed8; }

.state { font-size: 14px; color: #6b7280; }
.state--error { color: #b91c1c; }

/* ===== Модалка ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.modal {
  background: #ffffff;
  border-radius: 16px;
  width: 520px;
  max-width: calc(100vw - 32px);
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 16px;
  border-bottom: 1px solid #e5e7eb;
  flex-shrink: 0;
}
.modal-header h2 { margin: 0; font-size: 17px; font-weight: 600; color: #111827; }

.modal-close {
  background: none;
  border: none;
  font-size: 18px;
  color: #9ca3af;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 6px;
  line-height: 1;
}
.modal-close:hover { background: #f3f4f6; color: #374151; }

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-object-name { font-size: 16px; font-weight: 600; color: #111827; margin-bottom: 4px; }
.modal-object-address { font-size: 13px; color: #6b7280; margin-bottom: 16px; }

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

.reject-notice {
  margin-top: 4px;
  padding: 10px 12px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.reject-form { margin-top: 16px; }
.reject-label { display: block; font-size: 14px; font-weight: 500; color: #374151; margin-bottom: 8px; }
.required { color: #dc2626; }

.reject-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 10px;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
}
.reject-textarea:focus { outline: none; border-color: #9524c9; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px 20px;
  border-top: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.reject-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  font-size: 14px;
  cursor: pointer;
  background: #fee2e2;
  color: #b91c1c;
}
.reject-btn:hover { background: #fecaca; }
.reject-btn:disabled { opacity: 0.6; cursor: default; }

</style>