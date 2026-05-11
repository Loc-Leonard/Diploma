<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import DocumentManager from '../components/DocumentManager.vue'

const API_BASE = 'http://localhost:8080'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const object = ref<any>(null)
const workItems = ref<any[]>([])
const deliveries = ref<any[]>([])
const loading = ref(false)
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

const showRejectModal = ref(false)
const rejectReason = ref('')
const approveLoading = ref(false)
const decisionError = ref<string | null>(null)

async function loadObject() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch(
      `${API_BASE}/inspector/objects/${route.params.id}`,
      { headers: { Authorization: `Bearer ${auth.token}` } },
    )
    if (!res.ok) throw new Error('Ошибка загрузки объекта')
    const data = await res.json()
    object.value = data.object
    workItems.value = data.work_items ?? []
    deliveries.value = data.deliveries ?? []
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}

async function submitReports() {
  const reports = Object.entries(reportForm.value)
    .filter(([, qty]) => qty > 0)
    .map(([workItemId, qty]) => ({
      work_item_id: Number(workItemId),
      qty,
      date: new Date().toISOString().slice(0, 10),
    }))

  if (!reports.length) {
    submitError.value = 'Введите объём хотя бы по одной работе'
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
    if (!res.ok) throw new Error('Ошибка отправки отчёта')
    reportForm.value = {}
    submitSuccess.value = 'Отчёт успешно сохранён'
    setTimeout(() => (submitSuccess.value = null), 3000)
  } catch (e: any) {
    submitError.value = e.message || 'Ошибка'
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
    if (!res.ok) throw new Error('Ошибка сохранения поставки')
    showDeliveryForm.value = false
    deliveryForm.value = {
      material: '',
      qty: 0,
      date: new Date().toISOString().slice(0, 10),
    }
    await loadObject()
  } catch (e: any) {
    deliveryError.value = e.message || 'Ошибка'
  } finally {
    deliveryLoading.value = false
  }
}

async function approveActivation() {
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
        body: JSON.stringify({ decision: 'APPROVE' }),
      },
    )
    if (!res.ok) throw new Error('Ошибка')
    await loadObject()
  } catch (e: any) {
    decisionError.value = e.message || 'Ошибка'
  } finally {
    approveLoading.value = false
  }
}

async function rejectActivation() {
  if (!rejectReason.value.trim()) {
    decisionError.value = 'Укажите причину'
    return
  }
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
        body: JSON.stringify({ decision: 'REJECT', rejection_reason: rejectReason.value.trim() }),
      },
    )
    if (!res.ok) throw new Error('Ошибка')
    showRejectModal.value = false
    rejectReason.value = ''
    await loadObject()
  } catch (e: any) {
    decisionError.value = e.message || 'Ошибка'
  } finally {
    approveLoading.value = false
  }
}

function formatDate(v?: string | null) {
  if (!v) return '—'
  return new Date(v).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

function statusLabel(s: string) {
  const m: Record<string, string> = {
    PLANNED: 'Запланирован',
    WAITING_INSPECTOR_CONFIRMATION: 'Ожидает подтверждения',
    ACTIVE: 'Активен',
    FINISHED: 'Завершён',
  }
  return m[s] ?? s
}

function statusClass(s: string) {
  return {
    'status-chip--planned': s === 'PLANNED',
    'status-chip--waiting': s === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': s === 'ACTIVE',
    'status-chip--finished': s === 'FINISHED',
  }
}

function goBack() {
  router.push({ name: 'inspector-objects' })
}

onMounted(loadObject)
</script>

<style scoped>
.inspector-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

.sidebar {
  width: 206px;
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
  transition: background 0.15s;
}

.nav-item:hover:not(:disabled) {
  background: #f3f4f6;
}

.nav-item--active {
  background: #fef3c7;
  color: #92400e;
  font-weight: 500;
}

.nav-item:disabled {
  opacity: 0.6;
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
  transition: background 0.15s;
}

.logout-button:hover {
  background: #f3f4f6;
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
  background: #8b5cf6;
}

/* ===== Main Content ===== */
.inspector-main {
  padding: 24px 32px;
  box-sizing: border-box;
  max-width: 900px;
}

/* ===== Object Header ===== */
.object-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.object-header-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.object-header-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.back-btn {
  background: none;
  border: none;
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  transition: color 0.15s;
}

.back-btn:hover {
  color: #8b5cf6;
}

.object-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

.object-city {
  font-size: 13px;
  color: #6b7280;
}

.object-header-people {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: flex-end;
}

.person-badge {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.person-role {
  font-size: 11px;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.person-name {
  font-size: 13px;
  color: #374151;
  font-weight: 500;
}

/* ===== Status Chips ===== */
.status-chip {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
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

/* ===== Dates Row ===== */
.dates-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 10px 14px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  width: fit-content;
}

.date-item {
  display: flex;
  flex-direction: column;
}

.date-label {
  font-size: 11px;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.date-value {
  font-size: 14px;
  font-weight: 500;
  color: #111827;
}

.date-divider {
  color: #d1d5db;
  font-size: 16px;
}

/* ===== Sections ===== */
.work-section,
.delivery-section,
.documents-section,
.activation-section {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 16px 20px;
  margin-bottom: 16px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 12px;
  flex-wrap: wrap;
}

.section-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

/* ===== Work Table ===== */
.work-table-wrapper {
  overflow-x: auto;
}

.work-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.work-table th {
  text-align: left;
  padding: 8px 10px;
  font-size: 12px;
  font-weight: 500;
  color: #6b7280;
  border-bottom: 1px solid #e5e7eb;
  white-space: nowrap;
}

.work-table td {
  padding: 8px 10px;
  border-bottom: 1px solid #f3f4f6;
  color: #111827;
}

.work-table tbody tr:last-child td {
  border-bottom: none;
}

.work-table tbody tr:hover {
  background: #f9fafb;
}

.td-unit {
  color: #9ca3af;
  font-size: 12px;
}

.td-plan {
  font-weight: 500;
  color: #374151;
}

/* ===== Deliveries List ===== */
.deliveries-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.delivery-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.delivery-info {
  display: flex;
  gap: 12px;
  align-items: center;
}

.delivery-material {
  font-weight: 500;
  color: #111827;
}

.delivery-qty {
  font-size: 13px;
  color: #6b7280;
}

.delivery-date {
  font-size: 12px;
  color: #9ca3af;
}

/* ===== Checklist & Act Blocks ===== */
.checklist-block {
  margin-bottom: 16px;
}

.checklist-block h3 {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  margin-bottom: 8px;
}

.checklist-pre {
  font-size: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 10px;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  color: #374151;
}

.act-block {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 13px;
  color: #374151;
  margin-bottom: 16px;
}

.act-label {
  font-weight: 500;
}

.reject-notice {
  padding: 12px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
  margin-bottom: 16px;
}

.reject-notice .reject-label {
  font-weight: 600;
  margin-bottom: 4px;
}

.reject-notice p {
  margin: 0;
}

/* ===== Activation Actions ===== */
.activation-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.primary-btn {
  padding: 8px 18px;
  border-radius: 999px;
  border: none;
  background: #8b5cf6;
  color: #ffffff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.primary-btn:hover:not(:disabled) {
  background: #7c3aed;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.secondary-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #374151;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.secondary-btn:hover:not(:disabled) {
  background: #f3f4f6;
}

.reject-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  background: #fee2e2;
  color: #b91c1c;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.reject-btn:hover:not(:disabled) {
  background: #fecaca;
}

.reject-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

/* ===== State Messages ===== */
.state {
  font-size: 13px;
  color: #6b7280;
  padding: 8px 0;
}

.state--error {
  color: #b91c1c;
}

.state--success {
  color: #16a34a;
}

/* ===== Modal ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 16px;
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

.modal-header h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #111827;
}

.modal-close {
  background: none;
  border: none;
  font-size: 18px;
  color: #9ca3af;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 6px;
  line-height: 1;
  transition: background 0.15s, color 0.15s;
}

.modal-close:hover {
  background: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
}

.reject-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 8px;
}

.required {
  color: #dc2626;
}

.reject-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 10px;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.reject-textarea:focus {
  outline: none;
  border-color: #8b5cf6;
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px 20px;
  border-top: 1px solid #e5e7eb;
  flex-shrink: 0;
}

/* ===== Responsive ===== */
@media (max-width: 900px) {
  .inspector-main {
    padding: 16px;
  }

  .object-header {
    flex-direction: column;
    gap: 12px;
  }

  .object-header-people {
    align-items: flex-start;
  }

  .activation-actions {
    flex-direction: column;
  }

  .primary-btn,
  .reject-btn {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .modal-card--map {
    width: 100%;
    max-width: calc(100vw - 32px);
    padding: 16px;
  }
  .map-placeholder {
    height: 280px;
  }
}
</style>