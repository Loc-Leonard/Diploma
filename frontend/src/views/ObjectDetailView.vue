<template>
  <div class="layout">
    <CustomerLayout v-if="role === 'CUSTOMER'" />
    <InspectorLayout v-else-if="role === 'INSPECTOR'" />
    <aside v-else class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>
        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active" @click="goBack">← Объекты</button>
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

    <main class="main">
      <header class="page-header">
        <div class="page-header-left">
          <button class="back-btn" @click="goBack">← Назад</button>
          <h1 class="page-title">{{ detail?.object.name ?? '...' }}</h1>
          <span
            v-if="detail"
            class="status-chip"
            :class="statusClass(detail.object.status)"
          >
            {{ statusLabel(detail.object.status) }}
          </span>
        </div>
      </header>

      <div v-if="loading" class="state">Загружаю объект...</div>
      <div v-else-if="error" class="state state--error">{{ error }}</div>

      <div v-else-if="detail" class="detail-body">
        <aside class="detail-aside">
          <div class="mini-map">
            <div class="map-placeholder-box">🗺</div>
          </div>

          <section class="aside-section">
            <h2 class="aside-title">Ответственные лица</h2>
            <div class="person-block">
              <span class="person-role">Заказчик</span>
              <span class="person-name">{{ detail.object.customer?.full_name ?? '—' }}</span>
            </div>
            <div class="person-block">
              <span class="person-role">Прораб</span>
              <span class="person-name">{{ detail.object.foreman?.full_name ?? '—' }}</span>
            </div>
            <div class="person-block">
              <span class="person-role">Инспектор</span>
              <span class="person-name">{{ detail.object.inspector?.full_name ?? '—' }}</span>
            </div>
          </section>

          <section class="aside-section">
            <div class="date-row">
              <span class="date-label">Плановое начало</span>
              <span>{{ fmtDate(detail.object.planned_start_date) }}</span>
            </div>
            <div class="date-row">
              <span class="date-label">Плановое окончание</span>
              <span>{{ fmtDate(detail.object.planned_end_date) }}</span>
            </div>
            <div v-if="detail.object.actual_start_date" class="date-row">
              <span class="date-label">Фактическое начало</span>
              <span>{{ fmtDate(detail.object.actual_start_date) }}</span>
            </div>
          </section>

          <section v-if="detail.object.description" class="aside-section">
            <h2 class="aside-title">Описание</h2>
            <p class="aside-desc">{{ detail.object.description }}</p>
          </section>

          <template v-if="role === 'CUSTOMER'">
            <button
              v-if="detail.object.status === 'PLANNED'"
              class="action-btn action-btn--primary"
              @click="showActivateModal = true"
            >
              Активировать объект
            </button>
            <div
              v-if="detail.object.status === 'WAITING_INSPECTOR_CONFIRMATION'"
              class="info-notice"
            >
              Ожидает решения инспектора
            </div>
            <div
              v-if="detail.object.activation_reject_reason"
              class="reject-notice"
            >
              <span class="reject-label">Причина отклонения:</span>
              {{ detail.object.activation_reject_reason }}
            </div>
          </template>

          <template
            v-if="role === 'INSPECTOR' && detail.object.status === 'WAITING_INSPECTOR_CONFIRMATION'"
          >
            <div class="inspector-actions">
              <h2 class="aside-title">Решение по активации</h2>
              <button
                class="action-btn action-btn--primary"
                @click="approveActivation"
                :disabled="approveLoading"
              >
                {{ approveLoading ? 'Сохраняю...' : 'Подтвердить' }}
              </button>
              <button
                class="action-btn action-btn--danger"
                @click="showRejectModal = true"
              >
                Отклонить
              </button>
              <div v-if="decisionError" class="state state--error">
                {{ decisionError }}
              </div>
            </div>
          </template>
        </aside>

        <div class="detail-main">
          <section class="card">
            <div class="card-header">
              <h2>График работ</h2>
            </div>

            <div v-if="!ganttTasks.length" class="gantt-placeholder">
              Добавьте этапы с плановыми датами, чтобы увидеть график
            </div>

            <FrappeGantt
              v-else
              :tasks="ganttTasks"
              :view-mode="ganttViewMode"
              height="auto"
            />
          </section>

          <section class="card">
            <div class="card-header">
              <h2>Виды работ</h2>
              <button
                v-if="role === 'CUSTOMER'"
                class="action-btn action-btn--small"
                @click="showWorkItemForm = !showWorkItemForm"
              >
                {{ showWorkItemForm ? 'Отмена' : '+ Добавить этап' }}
              </button>
            </div>

            <div
              v-if="role === 'CUSTOMER' && showWorkItemForm"
              class="delivery-form"
            >
              <div class="form-row">
                <div class="form-field">
                  <label>Название *</label>
                  <input
                    v-model="workItemForm.name"
                    type="text"
                    placeholder="Земляные работы"
                  />
                </div>
                <div class="form-field">
                  <label>Единица измерения</label>
                  <input
                    v-model="workItemForm.unit"
                    type="text"
                    placeholder="м³, шт, м²"
                  />
                </div>
                <div class="form-field">
                  <label>Плановый объём</label>
                  <input
                    v-model.number="workItemForm.plan_qty"
                    type="number"
                    min="0"
                  />
                </div>
              </div>
              <div class="form-row">
                <div class="form-field">
                  <label>Дата начала</label>
                  <input v-model="workItemForm.planned_start_date" type="date" />
                </div>
                <div class="form-field">
                  <label>Дата окончания</label>
                  <input v-model="workItemForm.planned_end_date" type="date" />
                </div>
              </div>
              <div v-if="workItemError" class="state state--error">
                {{ workItemError }}
              </div>
              <div class="work-actions">
                <button
                  class="action-btn action-btn--primary"
                  @click="submitWorkItem"
                  :disabled="workItemLoading"
                >
                  {{ workItemLoading ? 'Сохраняю...' : 'Добавить' }}
                </button>
              </div>
            </div>

            <div v-if="!(detail.work_items?.length)" class="state">
              Работы не добавлены
            </div>

            <div v-else class="work-table-wrapper">
              <table class="work-table">
                <thead>
                  <tr>
                    <th>Наименование</th>
                    <th>Ед.</th>
                    <th>План</th>
                    <th v-if="role === 'CUSTOMER'">Действия</th>
                    <th v-if="role === 'FOREMAN'">Факт (сегодня)</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in (detail.work_items || [])"
                    :key="item.id"
                  >
                    <td>{{ item.name }}</td>
                    <td class="td-unit">{{ item.unit }}</td>
                    <td class="td-plan">{{ item.plan_qty }}</td>
                    <td v-if="role === 'CUSTOMER'" class="td-actions">
                      <button
                        class="icon-btn icon-btn--edit"
                        @click="openEditWorkItem(item)"
                        title="Редактировать"
                      >
                        ✏️
                      </button>
                      <button
                        class="icon-btn icon-btn--delete"
                        @click="deleteWorkItem(item.id)"
                        title="Удалить"
                      >
                        🗑️
                      </button>
                    </td>
                    <td v-if="role === 'FOREMAN'" class="td-input">
                      <input
                        type="number"
                        min="0"
                        v-model.number="reportForm[item.id]"
                        placeholder="0"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <template v-if="role === 'FOREMAN' && detail.work_items?.length">
              <div v-if="submitError" class="state state--error">
                {{ submitError }}
              </div>
              <div v-if="submitSuccess" class="state state--success">
                {{ submitSuccess }}
              </div>
              <div class="work-actions">
                <button
                  class="action-btn action-btn--primary"
                  @click="submitReports"
                  :disabled="submitting"
                >
                  {{ submitting ? 'Сохраняю...' : 'Сохранить отчёт' }}
                </button>
              </div>
            </template>
          </section>

          <section class="card">
            <div class="card-header">
              <h2>Поставки материалов</h2>
              <button
                v-if="role === 'FOREMAN'"
                class="action-btn action-btn--small"
                @click="showDeliveryForm = !showDeliveryForm"
              >
                {{ showDeliveryForm ? 'Отмена' : '+ Добавить' }}
              </button>
            </div>

            <div
              v-if="role === 'FOREMAN' && showDeliveryForm"
              class="delivery-form"
            >
              <div class="form-row">
                <div class="form-field">
                  <label>Материал</label>
                  <input
                    v-model="deliveryForm.material"
                    type="text"
                    placeholder="Кирпич, м³"
                  />
                </div>
                <div class="form-field">
                  <label>Количество</label>
                  <input
                    v-model.number="deliveryForm.qty"
                    type="number"
                    min="0"
                  />
                </div>
                <div class="form-field">
                  <label>Дата</label>
                  <input v-model="deliveryForm.date" type="date" />
                </div>
              </div>
              <div v-if="deliveryError" class="state state--error">
                {{ deliveryError }}
              </div>
              <div class="work-actions">
                <button
                  class="action-btn action-btn--primary"
                  @click="submitDelivery"
                  :disabled="deliveryLoading"
                >
                  {{ deliveryLoading ? 'Сохраняю...' : 'Добавить поставку' }}
                </button>
              </div>
            </div>

            <div
              v-if="!(detail.deliveries?.length) && !showDeliveryForm"
              class="state"
            >
              Поставок пока нет
            </div>

            <div
              v-else-if="detail.deliveries?.length"
              class="work-table-wrapper"
            >
              <table class="work-table">
                <thead>
                  <tr>
                    <th>Дата</th>
                    <th>Материал</th>
                    <th>Количество</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="d in (detail.deliveries || [])" :key="d.id">
                    <td>{{ fmtDate(d.date) }}</td>
                    <td>{{ d.material }}</td>
                    <td>{{ d.qty }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <!-- Секция документов с DocumentManager -->
          <section class="card">
            <DocumentManager
              v-if="detail?.object.id"
              :documents="documents"
              :loading="docLoading"
              :uploading="docUploading"
              :error="docError"
              :error-message="docErrorMessage"
              :deleting-id="docDeletingId"
              :downloading-id="docDownloadingId"
              :can-delete-doc="canDeleteDocument"
              @upload-click="triggerFileInput"
              @file-select="handleFileSelect"
              @retry="fetchDocuments"
              @download="downloadDocument"
              @delete="deleteDocument"
              @clear-error="docErrorMessage = null"
            />
            <input
              ref="fileInputRef"
              type="file"
              @change="handleFileSelect"
              accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx"
              class="hidden-file-input"
            />
          </section>
        </div>
      </div>

      <!-- Модалки (активация, редактирование, отклонение) -->
      <div
        v-if="showActivateModal"
        class="modal-overlay"
        @click.self="showActivateModal = false"
      >
        <div class="modal-card">
          <h2>Активация объекта</h2>
          <div class="form-field">
            <label>Чек-лист открытия (текст / JSON)</label>
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
          <div v-if="activateError" class="state state--error">
            {{ activateError }}
          </div>
          <div class="modal-actions">
            <button class="action-btn" @click="showActivateModal = false">
              Отмена
            </button>
            <button
              class="action-btn action-btn--primary"
              @click="submitActivate"
              :disabled="activateLoading"
            >
              {{ activateLoading ? 'Отправляю...' : 'Отправить на проверку' }}
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="showEditWorkItemModal"
        class="modal-overlay"
        @click.self="closeEditWorkItem"
      >
        <div class="modal-card">
          <h2>Редактирование этапа</h2>
          <div class="form-field">
            <label>Название *</label>
            <input
              v-model="editWorkItemForm.name"
              type="text"
              placeholder="Земляные работы"
            />
          </div>
          <div class="form-field">
            <label>Описание</label>
            <textarea
              v-model="editWorkItemForm.description"
              rows="3"
              placeholder="Описание этапа"
            />
          </div>
          <div class="form-row">
            <div class="form-field">
              <label>Плановый объём</label>
              <input
                v-model.number="editWorkItemForm.plan_qty"
                type="number"
                min="0"
              />
            </div>
          </div>
          <div class="form-row">
            <div class="form-field">
              <label>Дата начала</label>
              <input v-model="editWorkItemForm.planned_start_date" type="date" />
            </div>
            <div class="form-field">
              <label>Дата окончания</label>
              <input v-model="editWorkItemForm.planned_end_date" type="date" />
            </div>
          </div>
          <div v-if="editWorkItemError" class="state state--error">
            {{ editWorkItemError }}
          </div>
          <div class="modal-actions">
            <button class="action-btn" @click="closeEditWorkItem">
              Отмена
            </button>
            <button
              class="action-btn action-btn--primary"
              @click="submitEditWorkItem"
              :disabled="editWorkItemLoading"
            >
              {{ editWorkItemLoading ? 'Сохраняю...' : 'Сохранить' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Модалка отклонения (инспектор) -->
      <div
        v-if="showRejectModal"
        class="modal-overlay"
        @click.self="showRejectModal = false"
      >
        <div class="modal-card">
          <h2>Причина отклонения</h2>
          <div class="form-field">
            <label>Укажите причину</label>
            <textarea v-model="rejectReason" rows="3" />
          </div>
          <div v-if="decisionError" class="state state--error">
            {{ decisionError }}
          </div>
          <div class="modal-actions">
            <button class="action-btn" @click="showRejectModal = false">
              Отмена
            </button>
            <button
              class="action-btn action-btn--danger"
              @click="rejectActivation"
              :disabled="approveLoading"
            >
              {{ approveLoading ? 'Сохраняю...' : 'Отклонить' }}
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CustomerLayout from './CustomerLayout.vue'
import InspectorLayout from './InspectorLayout.vue'
import FrappeGantt from '@/components/FrappeGantt.vue'
import DocumentManager from '@/components/DocumentManager.vue'

const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const role = computed(() => auth.user?.role ?? '')
const greeting = computed(() => {
  const name = auth.user?.full_name
  return name ? `Добрый день, ${name}` : 'Добрый день'
})

// ─── Типы ────────────────────────────────────────────────────────────────────

type ObjStatus = 'PLANNED' | 'WAITING_INSPECTOR_CONFIRMATION' | 'ACTIVE' | 'FINISHED'

interface Person {
  id: number
  full_name: string
}

interface ObjectCore {
  id: number
  name: string
  city: string
  address: string
  description: string
  status: ObjStatus
  lat: number
  lng: number
  planned_start_date?: string | null
  planned_end_date?: string | null
  actual_start_date?: string | null
  init_act_file_path?: string
  init_checklist_json?: string
  activation_reject_reason?: string
  customer?: Person
  foreman?: Person
  inspector?: Person
}

interface WorkItem {
  id: number
  object_id: number
  name: string
  description: string
  unit: string
  plan_qty: number
  planned_start_date?: string | null
  planned_end_date?: string | null
  actual_start_date?: string | null
  actual_end_date?: string | null
  sort_order: number
  status: 'PLANNED' | 'IN_PROGRESS' | 'DONE' | 'DELAYED'
  depends_on_id?: number | null
  progress: number
}

interface Delivery {
  id: number
  date: string
  material: string
  qty: number
}

interface Document {
  id: number
  document_type: string
  original_file_name: string
  mime_type: string
  cv_status: string
  cv_confidence: number
  created_at: string
  uploaded_by: string
}

interface DetailResponse {
  object: ObjectCore
  work_items: WorkItem[]
  deliveries: Delivery[]
}

// ─── Состояние ───────────────────────────────────────────────────────────────

const detail = ref<DetailResponse | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const reportForm = ref<Record<number, number>>({})
const submitting = ref(false)
const submitError = ref<string | null>(null)
const submitSuccess = ref<string | null>(null)

const showDeliveryForm = ref(false)
const deliveryLoading = ref(false)
const deliveryError = ref<string | null>(null)
const deliveryForm = ref({
  material: '',
  qty: 0,
  date: new Date().toISOString().slice(0, 10),
})

// форма этапа (заказчик)
const showWorkItemForm = ref(false)
const workItemLoading = ref(false)
const workItemError = ref<string | null>(null)
const workItemForm = ref({
  name: '',
  unit: '',
  plan_qty: 0,
  planned_start_date: '',
  planned_end_date: '',
})

// редактирование этапа (заказчик)
const showEditWorkItemModal = ref(false)
const editingWorkItemId = ref<number | null>(null)
const editWorkItemLoading = ref(false)
const editWorkItemError = ref<string | null>(null)
const editWorkItemForm = ref({
  name: '',
  description: '',
  unit: '',
  plan_qty: 0,
  planned_start_date: '',
  planned_end_date: '',
  sort_order: 0,
  depends_on_id: null as number | null,
})

// активация
const showActivateModal = ref(false)
const activateLoading = ref(false)
const activateError = ref<string | null>(null)
const activateForm = ref({ checklist_json: '', act_file_path: '' })

// отклонение инспектором
const showRejectModal = ref(false)
const rejectReason = ref('')
const approveLoading = ref(false)
const decisionError = ref<string | null>(null)

// Документы
const documents = ref<Document[]>([])
const docLoading = ref(false)
const docUploading = ref(false)
const docError = ref<string | null>(null)
const docErrorMessage = ref<string | null>(null)
const docDeletingId = ref<number | null>(null)
const docDownloadingId = ref<number | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

// ─── Загрузка ───────────────────────────────────────────────────────────────

function endpointForRole() {
  const id = route.params.id
  switch (role.value) {
    case 'CUSTOMER':
      return `${API_BASE}/customer/objects/${id}`
    case 'FOREMAN':
      return `${API_BASE}/foreman/objects/${id}`
    case 'INSPECTOR':
      return `${API_BASE}/inspector/objects/${id}`
    default:
      return ''
  }
}

async function fetchDetail() {
  loading.value = true
  error.value = null
  try {
    const url = endpointForRole()
    if (!url) throw new Error('Неизвестная роль')

    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка загрузки',
      )

    const data = await res.json()
    const obj = data.object ?? data
    detail.value = {
      object: obj,
      work_items: Array.isArray(data.work_items) ? data.work_items : [],
      deliveries: Array.isArray(data.deliveries) ? data.deliveries : [],
    }
    
    // Загружаем документы после загрузки объекта
    if (obj.id) {
      fetchDocuments()
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// ─── Документы ───────────────────────────────────────────────────────────────

const documentsBaseUrl = computed(() => {
  if (!detail.value?.object.id) return ''
  return `${API_BASE}/${role.value.toLowerCase()}/objects/${detail.value.object.id}/documents`
})

async function fetchDocuments() {
  if (!documentsBaseUrl.value) return
  
  docLoading.value = true
  docError.value = null
  try {
    const res = await fetch(documentsBaseUrl.value, {
      headers: { 
        Authorization: `Bearer ${auth.token}`,
        'Content-Type': 'application/json'
      }
    })
    
    if (!res.ok) {
      const errorText = await res.text()
      let errorData = { error: 'Ошибка загрузки документов' }
      try {
        errorData = JSON.parse(errorText)
      } catch {}
      throw new Error(errorData.error || `Ошибка ${res.status}`)
    }
    
    const data = await res.json()
    documents.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    docError.value = e.message
    docErrorMessage.value = e.message
  } finally {
    docLoading.value = false
  }
}

function triggerFileInput() {
  if (fileInputRef.value) {
    fileInputRef.value.click()
  }
}

async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Проверка размера (10MB)
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    docErrorMessage.value = `Файл слишком большой (макс. 10MB). Размер: ${formatFileSize(file.size)}`
    return
  }

  // Проверка типа
  const allowedTypes = [
    'image/jpeg', 'image/png', 'image/jpg', 'image/gif',
    'application/pdf',
    'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  ]
  if (!allowedTypes.includes(file.type)) {
    docErrorMessage.value = 'Неподдерживаемый тип файла'
    return
  }

  await uploadFile(file)
  
  // Очищаем input
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

async function uploadFile(file: File) {
  if (!documentsBaseUrl.value) return
  
  docUploading.value = true
  docErrorMessage.value = null

  const formData = new FormData()
  formData.append('file', file)
  formData.append('document_type', 'OTHER')

  try {
    const uploadUrl = `${documentsBaseUrl.value}/upload`
    
    const res = await fetch(uploadUrl, {
      method: 'POST',
      headers: { 
        Authorization: `Bearer ${auth.token}`
      },
      body: formData
    })

    if (!res.ok) {
      const errorText = await res.text()
      let errorData = { error: 'Ошибка загрузки файла' }
      try {
        errorData = JSON.parse(errorText)
      } catch {}
      throw new Error(errorData.error || `Ошибка ${res.status}`)
    }

    // Обновляем список документов
    await fetchDocuments()
  } catch (e: any) {
    docErrorMessage.value = e.message
  } finally {
    docUploading.value = false
  }
}

async function downloadDocument(doc: Document) {
  docDownloadingId.value = doc.id
  docErrorMessage.value = null
  
  try {
    const downloadUrl = `${API_BASE}/${role.value}/objects/${detail.value?.object.id}/documents/${doc.id}/download`
    
    const res = await fetch(downloadUrl, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    
    if (!res.ok) {
      throw new Error(`Ошибка скачивания: ${res.status}`)
    }

    const blob = await res.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = doc.original_file_name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    docErrorMessage.value = `Ошибка скачивания: ${e.message}`
  } finally {
    docDownloadingId.value = null
  }
}

async function deleteDocument(doc: Document) {
  if (!confirm(`Удалить документ "${doc.original_file_name}"?`)) {
    return
  }

  docDeletingId.value = doc.id
  docErrorMessage.value = null

  try {
    const url = `${documentsBaseUrl.value}/${doc.id}`
    
    const res = await fetch(url, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${auth.token}` }
    })

    if (!res.ok) {
      const errorText = await res.text()
      let errorData = { error: 'Ошибка удаления' }
      try {
        errorData = JSON.parse(errorText)
      } catch {}
      throw new Error(errorData.error || `Ошибка ${res.status}`)
    }

    // Обновляем список
    documents.value = documents.value.filter(d => d.id !== doc.id)
  } catch (e: any) {
    docErrorMessage.value = e.message
  } finally {
    docDeletingId.value = null
  }
}

function canDeleteDocument(doc: Document) {
  const currentUserId = auth.user?.id
  const currentUserName = auth.user?.full_name
  const currentUserRole = auth.user?.role
  
  // Админ может удалять все документы
  if (currentUserRole === 'admin') {
    return true
  }
  
  // Пользователь может удалять свои документы
  return doc.uploaded_by === currentUserName || doc.uploaded_by === String(currentUserId)
}

function formatFileSize(bytes: number) {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// ─── Прораб ──────────────────────────────────────────────────────────────────

async function submitReports() {
  const reports = Object.entries(reportForm.value)
    .filter(([, qty]) => qty > 0)
    .map(([workItemId, qty]) => ({
      work_item_id: Number(workItemId),
      qty,
      date: new Date().toISOString().slice(0, 10),
    }))

  if (!reports.length) {
    submitError.value = 'Заполните хотя бы одну строку'
    return
  }

  submitting.value = true
  submitError.value = null
  submitSuccess.value = null
  try {
    const res = await fetch(
      `${API_BASE}/foreman/objects/${route.params.id}/work-reports`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({ reports }),
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )
    reportForm.value = {}
    submitSuccess.value = 'Отчёт сохранён'
    await fetchDetail()
    setTimeout(() => {
      submitSuccess.value = null
    }, 3000)
  } catch (e: any) {
    submitError.value = e.message
  } finally {
    submitting.value = false
  }
}

async function submitDelivery() {
  if (!deliveryForm.value.material.trim()) {
    deliveryError.value = 'Укажите материал'
    return
  }
  if (!deliveryForm.value.qty || deliveryForm.value.qty <= 0) {
    deliveryError.value = 'Укажите количество'
    return
  }

  deliveryLoading.value = true
  deliveryError.value = null
  try {
    const res = await fetch(
      `${API_BASE}/foreman/objects/${route.params.id}/deliveries`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(deliveryForm.value),
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )
    showDeliveryForm.value = false
    deliveryForm.value = {
      material: '',
      qty: 0,
      date: new Date().toISOString().slice(0, 10),
    }
    await fetchDetail()
  } catch (e: any) {
    deliveryError.value = e.message
  } finally {
    deliveryLoading.value = false
  }
}

// ─── Заказчик: этапы ─────────────────────────────────────────────────────────

async function submitWorkItem() {
  if (!workItemForm.value.name.trim()) {
    workItemError.value = 'Укажите название этапа'
    return
  }

  workItemLoading.value = true
  workItemError.value = null
  try {
    const body: Record<string, any> = {
      name: workItemForm.value.name,
      unit: workItemForm.value.unit,
      plan_qty: workItemForm.value.plan_qty,
    }
    if (workItemForm.value.planned_start_date)
      body.planned_start_date = new Date(
        workItemForm.value.planned_start_date,
      ).toISOString()
    if (workItemForm.value.planned_end_date)
      body.planned_end_date = new Date(
        workItemForm.value.planned_end_date,
      ).toISOString()

    const res = await fetch(
      `${API_BASE}/customer/objects/${route.params.id}/work-items`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(body),
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )

    showWorkItemForm.value = false
    workItemForm.value = {
      name: '',
      unit: '',
      plan_qty: 0,
      planned_start_date: '',
      planned_end_date: '',
    }
    await fetchDetail()
  } catch (e: any) {
    workItemError.value = e.message
  } finally {
    workItemLoading.value = false
  }
}

async function openEditWorkItem(item: WorkItem) {
  editingWorkItemId.value = item.id
  editWorkItemForm.value = {
    name: item.name,
    description: item.description || '',
    unit: item.unit || '',
    plan_qty: item.plan_qty,
    planned_start_date: item.planned_start_date ? item.planned_start_date.slice(0, 10) : '',
    planned_end_date: item.planned_end_date ? item.planned_end_date.slice(0, 10) : '',
    sort_order: item.sort_order,
    depends_on_id: item.depends_on_id || null,
  }
  editWorkItemError.value = null
  showEditWorkItemModal.value = true
}

async function submitEditWorkItem() {
  if (!editWorkItemForm.value.name.trim()) {
    editWorkItemError.value = 'Укажите название этапа'
    return
  }
  if (!editingWorkItemId.value) return

  editWorkItemLoading.value = true
  editWorkItemError.value = null
  try {
    const body: Record<string, any> = {
      name: editWorkItemForm.value.name,
      description: editWorkItemForm.value.description,
      unit: editWorkItemForm.value.unit,
      plan_qty: editWorkItemForm.value.plan_qty,
      sort_order: editWorkItemForm.value.sort_order,
      depends_on_id: editWorkItemForm.value.depends_on_id,
    }
    if (editWorkItemForm.value.planned_start_date)
      body.planned_start_date = new Date(
        editWorkItemForm.value.planned_start_date,
      ).toISOString()
    if (editWorkItemForm.value.planned_end_date)
      body.planned_end_date = new Date(
        editWorkItemForm.value.planned_end_date,
      ).toISOString()

    const res = await fetch(
      `${API_BASE}/customer/objects/${route.params.id}/work-items/${editingWorkItemId.value}`,
      {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(body),
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )

    showEditWorkItemModal.value = false
    editingWorkItemId.value = null
    await fetchDetail()
  } catch (e: any) {
    editWorkItemError.value = e.message
  } finally {
    editWorkItemLoading.value = false
  }
}

function closeEditWorkItem() {
  showEditWorkItemModal.value = false
  editingWorkItemId.value = null
  editWorkItemError.value = null
}

async function deleteWorkItem(itemId: number) {
  if (!confirm('Вы уверены, что хотите удалить этот этап?')) return

  try {
    const res = await fetch(
      `${API_BASE}/customer/objects/${route.params.id}/work-items/${itemId}`,
      {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )

    await fetchDetail()
  } catch (e: any) {
    alert(e.message || 'Ошибка при удалении')
  }
}

// ─── Заказчик: активация ─────────────────────────────────────────────────────

async function submitActivate() {
  if (!activateForm.value.checklist_json.trim()) {
    activateError.value = 'Заполните чек-лист'
    return
  }

  activateLoading.value = true
  activateError.value = null
  try {
    const res = await fetch(
      `${API_BASE}/customer/objects/${route.params.id}/activate`,
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
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )
    showActivateModal.value = false
    await fetchDetail()
  } catch (e: any) {
    activateError.value = e.message
  } finally {
    activateLoading.value = false
  }
}

// ─── Инспектор ───────────────────────────────────────────────────────────────

async function sendDecision(
  decision: 'APPROVE' | 'REJECT',
  rejection_reason = '',
) {
  approveLoading.value = true
  decisionError.value = null
  try {
    const res = await fetch(
      `${API_BASE}/inspector/objects/${route.params.id}/activation-decision`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({ decision, rejection_reason }),
      },
    )
    if (!res.ok)
      throw new Error(
        (await res.json().catch(() => ({}))).error ?? 'Ошибка',
      )
    showRejectModal.value = false
    rejectReason.value = ''
    await fetchDetail()
  } catch (e: any) {
    decisionError.value = e.message
  } finally {
    approveLoading.value = false
  }
}

async function approveActivation() {
  await sendDecision('APPROVE')
}

async function rejectActivation() {
  if (!rejectReason.value.trim()) {
    decisionError.value = 'Укажите причину'
    return
  }
  await sendDecision('REJECT', rejectReason.value.trim())
}

// ─── Хелперы ─────────────────────────────────────────────────────────────────

function fmtDate(v?: string | null) {
  if (!v) return '—'
  return new Date(v).toLocaleDateString('ru-RU')
}

function fmtChecklist(json: string) {
  try {
    return JSON.stringify(JSON.parse(json), null, 2)
  } catch {
    return json
  }
}

function statusLabel(s: ObjStatus) {
  const m: Record<ObjStatus, string> = {
    PLANNED: 'Запланирован',
    WAITING_INSPECTOR_CONFIRMATION: 'Ожидает подтверждения',
    ACTIVE: 'Активен',
    FINISHED: 'Завершён',
  }
  return m[s] ?? s
}

function statusClass(s: ObjStatus) {
  return {
    'status-chip--planned': s === 'PLANNED',
    'status-chip--waiting': s === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': s === 'ACTIVE',
    'status-chip--finished': s === 'FINISHED',
  }
}

function goBack() {
  switch (role.value) {
    case 'CUSTOMER':
      router.push({ name: 'customer-objects' })
      break
    case 'FOREMAN':
      router.push({ name: 'foreman-objects' })
      break
    case 'INSPECTOR':
      router.push({ name: 'inspector-objects' })
      break
  }
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

onMounted(fetchDetail)

// ─── Гантт ───────────────────────────────────────────────────────────────────

type GanttTask = {
  id: string
  name: string
  start: string
  end: string
  progress: number
}

function toYMD(dateStr?: string | null): string {
  if (!dateStr) return ''
  return new Date(dateStr).toISOString().slice(0, 10)
}

function getActualDateRange(item: WorkItem) {
  const start = item.actual_start_date ?? item.planned_start_date ?? null
  let end = item.actual_end_date ?? item.planned_end_date ?? null

  if (
    !item.actual_end_date &&
    item.planned_end_date &&
    (item.status === 'IN_PROGRESS' || item.status === 'DELAYED')
  ) {
    const today = new Date()
    const plannedEnd = new Date(item.planned_end_date)

    if (plannedEnd < today) {
      end = today.toISOString()
    }
  }

  return { start, end }
}

const ganttViewMode = computed<'Day' | 'Week' | 'Month'>(() => {
  const items = detail.value?.work_items ?? []
  if (!items.length) return 'Week'

  const ranges = items
    .map(getActualDateRange)
    .filter(
      (r): r is { start: string; end: string } => Boolean(r.start && r.end),
    )

  if (!ranges.length) return 'Week'

  const starts = ranges.map(r => new Date(r.start).getTime())
  const ends = ranges.map(r => new Date(r.end).getTime())

  const minStart = Math.min(...starts)
  const maxEnd = Math.max(...ends)
  const diffDays = Math.ceil((maxEnd - minStart) / (1000 * 60 * 60 * 24))

  if (diffDays <= 31) return 'Day'
  if (diffDays <= 90) return 'Week'
  return 'Month'
})

const ganttTasks = computed<GanttTask[]>(() =>
  (detail.value?.work_items ?? [])
    .map((i): GanttTask | null => {
      const { start, end } = getActualDateRange(i)
      if (!start || !end) return null

      return {
        id: String(i.id),
        name: i.name,
        start: toYMD(start),
        end: toYMD(end),
        progress: Math.round(i.progress ?? 0),
      }
    })
    .filter((task): task is GanttTask => task !== null)
)
</script>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}
.sidebar {
  width: 206px; display: flex; flex-direction: column;
  justify-content: space-between; padding: 20px 18px;
  background: #ffffff; border-right: 1px solid #e5e7eb;
}
.sidebar-logo { font-size: 15px; font-weight: 700; margin-bottom: 24px; color: #111827; }
.sidebar-nav { display: flex; flex-direction: column; gap: 6px; }
.nav-item {
  text-align: left; padding: 8px 10px; border-radius: 8px;
  border: none; background: transparent; font-size: 14px;
  color: #4b5563; cursor: pointer;
}
.nav-item--active { background: #eef2ff; color: #4338ca; }
.sidebar-bottom { display: flex; flex-direction: column; gap: 10px; }
.logout-button {
  padding: 7px 16px; border-radius: 999px; border: 1px solid #e5e7eb;
  background: #ffffff; font-size: 13px; color: #6b7280; cursor: pointer;
}
.role-badge { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; color: #6b7280; }
.role-dot { width: 10px; height: 10px; border-radius: 999px; }
.role-dot--foreman { background: #f59e0b; }

.main { padding: 24px 32px; box-sizing: border-box; min-width: 0; }

.page-header { display: flex; align-items: center; margin-bottom: 20px; }
.page-header-left { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.back-btn {
  padding: 6px 14px; border-radius: 999px; border: 1px solid #d1d5db;
  background: #fff; font-size: 14px; color: #374151; cursor: pointer;
}
.back-btn:hover { background: #f3f4f6; }
.page-title { margin: 0; font-size: 22px; font-weight: 600; color: #111827; }

.detail-body { display: grid; grid-template-columns: 220px 1fr; gap: 20px; align-items: start; }
.detail-aside { display: flex; flex-direction: column; gap: 14px; }

.mini-map .map-placeholder-box {
  width: 100%; height: 180px; border-radius: 12px;
  border: 1px solid #e5e7eb; background: #f3f4f6;
  display: flex; align-items: center; justify-content: center; font-size: 40px;
}

.aside-section { display: flex; flex-direction: column; gap: 8px; }
.aside-title { margin: 0 0 2px; font-size: 13px; font-weight: 600; color: #374151; text-transform: uppercase; letter-spacing: 0.04em; }
.aside-desc { margin: 0; font-size: 13px; color: #6b7280; line-height: 1.5; }
.person-block { display: flex; flex-direction: column; }
.person-role { font-size: 11px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.04em; }
.person-name { font-size: 14px; color: #4f46e5; font-weight: 500; }
.date-row { display: flex; justify-content: space-between; font-size: 13px; color: #374151; }
.date-label { color: #9ca3af; }

.action-btn {
  width: 100%; padding: 9px 14px; border-radius: 10px;
  border: 1px solid #e5e7eb; background: #fff;
  font-size: 14px; font-weight: 600; color: #111827;
  cursor: pointer; text-align: center;
}
.action-btn:hover:not(:disabled) { background: #f9fafb; }
.action-btn:disabled { opacity: 0.5; cursor: default; }
.action-btn--primary { background: #4f46e5; color: #fff; border-color: #4f46e5; }
.action-btn--primary:hover:not(:disabled) { background: #4338ca; border-color: #4338ca; }
.action-btn--danger { background: #dc2626; color: #fff; border-color: #dc2626; }
.action-btn--danger:hover:not(:disabled) { background: #b91c1c; border-color: #b91c1c; }
.action-btn--small { width: auto; padding: 5px 14px; font-size: 13px; font-weight: 500; border-radius: 999px; }
.inspector-actions { display: flex; flex-direction: column; gap: 8px; }

.info-notice {
  padding: 8px 12px; background: #fef3c7; border: 1px solid #fde68a;
  border-radius: 8px; font-size: 13px; color: #92400e;
}
.reject-notice {
  padding: 8px 12px; background: #fff7ed; border: 1px solid #fed7aa;
  border-radius: 8px; font-size: 13px; color: #92400e;
  display: flex; flex-direction: column; gap: 2px;
}
.reject-label { font-weight: 600; font-size: 12px; color: #78350f; }

.detail-main { display: flex; flex-direction: column; gap: 16px; min-width: 0; }

.card {
  background: #fff; border-radius: 16px; padding: 16px 18px;
  border: 1px solid #e5e7eb; box-shadow: 0 2px 8px rgba(15,23,42,0.04);
}
.card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
.card-header h2 { margin: 0; font-size: 16px; font-weight: 600; color: #111827; }

.work-table-wrapper { overflow-x: auto; }
.work-table { width: 100%; border-collapse: collapse; font-size: 14px; }
.work-table th {
  text-align: left; padding: 8px 10px; border-bottom: 1px solid #e5e7eb;
  font-size: 12px; color: #6b7280; font-weight: 500;
}
.work-table td { padding: 8px 10px; border-bottom: 1px solid #f3f4f6; color: #111827; }
.td-unit, .td-plan { color: #6b7280; }
.td-input input {
  width: 80px; padding: 4px 8px; border-radius: 6px;
  border: 1px solid #d1d5db; font-size: 14px; text-align: right;
}

.td-actions {
  display: flex;
  gap: 6px;
  white-space: nowrap;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  background: #ffffff;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, border-color 0.15s;
}

.icon-btn:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.icon-btn--edit:hover {
  background: #eff6ff;
  border-color: #3b82f6;
}

.icon-btn--delete:hover {
  background: #fef2f2;
  border-color: #ef4444;
}

.work-actions { display: flex; justify-content: flex-end; margin-top: 12px; }
.work-actions .action-btn { width: auto; padding: 8px 20px; border-radius: 999px; }

.delivery-form {
  margin-bottom: 14px; padding: 12px;
  background: #f9fafb; border-radius: 10px; border: 1px solid #e5e7eb;
}
.form-row { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; margin-bottom: 10px; }
.form-field { display: flex; flex-direction: column; gap: 4px; }
.form-field label { font-size: 12px; color: #6b7280; }
.form-field input, .form-field textarea {
  padding: 7px 10px; border-radius: 8px;
  border: 1px solid #d1d5db; font-size: 14px; background: #fff;
}

.hidden-file-input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.status-chip { padding: 3px 10px; border-radius: 999px; font-size: 11px; font-weight: 500; white-space: nowrap; }
.status-chip--planned  { background: #e5e7eb; color: #374151; }
.status-chip--waiting  { background: #fef3c7; color: #92400e; }
.status-chip--active   { background: #dcfce7; color: #166534; }
.status-chip--finished { background: #e0f2fe; color: #1d4ed8; }

.state { font-size: 13px; color: #6b7280; padding: 8px 0; }
.state--error   { color: #b91c1c; }
.state--success { color: #166534; }

.modal-overlay {
  position: fixed; inset: 0; background: rgba(15,23,42,0.55);
  display: flex; justify-content: center; align-items: center;
  padding: 16px; z-index: 50;
}

.modal-card {
  width: 100%; max-width: 520px; background: #fff;
  border-radius: 16px; padding: 22px 24px 20px;
  box-shadow: 0 20px 50px rgba(15,23,42,0.25); box-sizing: border-box;
}
.modal-card h2 { margin: 0 0 16px; font-size: 18px; font-weight: 600; color: #111827; }
.modal-card .form-field { margin-bottom: 12px; }
.modal-card textarea {
  width: 100%; padding: 8px 10px; border-radius: 8px;
  border: 1px solid #d1d5db; font-size: 14px;
  resize: vertical; min-height: 80px; box-sizing: border-box;
}
.modal-card input[type="text"] {
  width: 100%; padding: 8px 10px; border-radius: 8px;
  border: 1px solid #d1d5db; font-size: 14px; box-sizing: border-box;
}
.modal-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
.modal-actions .action-btn { width: auto; padding: 8px 20px; border-radius: 999px; }

.gantt-placeholder {
  height: 100px; border-radius: 10px;
  border: 2px dashed #e5e7eb; background: #f9fafb;
  display: flex; align-items: center;
  justify-content: center; color: #9ca3af; font-size: 14px;
}

@media (max-width: 900px) {
  .detail-body { grid-template-columns: 1fr; }
  .main { padding: 16px 20px; }
}
@media (max-width: 768px) {
  .layout { grid-template-columns: 1fr; }
  .sidebar { width: 100%; border-right: none; border-bottom: 1px solid #e5e7eb; }
  .form-row { grid-template-columns: 1fr; }
}
</style>