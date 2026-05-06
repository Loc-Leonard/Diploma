<template>
  <div class="foreman-layout">
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>
        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active" @click="goBack">
            Объекты
          </button>
          <button class="nav-item" disabled>График</button>
          <button class="nav-item" disabled>Замечания</button>
          <button class="nav-item" disabled>Справочники</button>
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

    <main class="foreman-main">
      <section v-if="loading" class="state">Загружаю объект...</section>
      <section v-else-if="error" class="state state--error">{{ error }}</section>

      <template v-else-if="object">
        <!-- Шапка объекта -->
        <header class="object-header">
          <div class="object-header-left">
            <div class="object-header-meta">
              <button class="back-btn" @click="goBack">← Объекты</button>
              <span
                class="status-chip"
                :class="statusClass(object.status)"
              >{{ statusLabel(object.status) }}</span>
            </div>
            <h1 class="object-title">{{ object.name }}</h1>
            <div class="object-city">{{ object.city }}, {{ object.address }}</div>
          </div>

          <div class="object-header-people">
            <div v-if="object.customer" class="person-badge">
              <span class="person-role">Заказчик</span>
              <span class="person-name">{{ object.customer.full_name }}</span>
            </div>
            <div v-if="object.inspector" class="person-badge">
              <span class="person-role">Инспектор</span>
              <span class="person-name">{{ object.inspector.full_name }}</span>
            </div>
          </div>
        </header>

        <!-- Даты -->
        <div class="dates-row" v-if="object.planned_start_date || object.planned_end_date">
          <div class="date-item" v-if="object.planned_start_date">
            <span class="date-label">Начало</span>
            <span class="date-value">{{ formatDate(object.planned_start_date) }}</span>
          </div>
          <div class="date-divider" v-if="object.planned_start_date && object.planned_end_date">→</div>
          <div class="date-item" v-if="object.planned_end_date">
            <span class="date-label">Окончание</span>
            <span class="date-value">{{ formatDate(object.planned_end_date) }}</span>
          </div>
        </div>

        <!-- Блок работ -->
        <section class="work-section">
          <div class="section-header">
            <h2>Работы</h2>
            <span class="section-hint">Введите объём выполненных работ за сегодня</span>
          </div>

          <div v-if="!workItems.length" class="state">Работы не назначены</div>

          <div v-else class="work-table-wrapper">
            <table class="work-table">
              <thead>
                <tr>
                  <th>Вид работ</th>
                  <th>Ед.</th>
                  <th>План</th>
                  <th>За сегодня</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in workItems" :key="item.id">
                  <td>{{ item.name }}</td>
                  <td class="td-unit">{{ item.unit }}</td>
                  <td class="td-plan">{{ item.plan_qty }}</td>
                  <td class="td-input">
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

          <div v-if="submitError" class="state state--error">{{ submitError }}</div>
          <div v-if="submitSuccess" class="state state--success">{{ submitSuccess }}</div>

          <div class="work-actions">
            <button
              class="primary-btn"
              @click="submitReports"
              :disabled="submitting || !workItems.length"
            >
              {{ submitting ? 'Отправляем...' : 'Отправить отчёт' }}
            </button>
          </div>
        </section>

        <!-- Блок поставок -->
        <section class="delivery-section">
          <div class="section-header">
            <h2>Поставка материалов</h2>
            <button class="secondary-btn" @click="showDeliveryForm = !showDeliveryForm">
              {{ showDeliveryForm ? 'Отмена' : '+ Добавить' }}
            </button>
          </div>

          <div v-if="showDeliveryForm" class="delivery-form">
            <div class="form-row">
              <div class="form-field">
                <label>Материал</label>
                <input v-model="deliveryForm.material" type="text" placeholder="Бетон М300" />
              </div>
              <div class="form-field">
                <label>Количество</label>
                <input v-model.number="deliveryForm.qty" type="number" min="0" placeholder="0" />
              </div>
              <div class="form-field">
                <label>Дата</label>
                <input v-model="deliveryForm.date" type="date" />
              </div>
            </div>
            <div v-if="deliveryError" class="state state--error">{{ deliveryError }}</div>
            <div class="work-actions">
              <button class="primary-btn" @click="submitDelivery" :disabled="deliveryLoading">
                {{ deliveryLoading ? 'Сохраняем...' : 'Сохранить поставку' }}
              </button>
            </div>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = import.meta.env.VITE_API_URL as string

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const greeting = computed(() => {
  const name = auth.user?.full_name
  return name ? `Добрый день, ${name}` : 'Добрый день'
})

type ObjectPerson = {
  id: number
  full_name: string
}

type ObjectCore = {
  id: number
  name: string
  city: string
  address: string
  description: string
  status: string
  lat: number
  lng: number
  planned_start_date: string | null
  planned_end_date: string | null
  customer: ObjectPerson | null
  foreman: ObjectPerson | null
  inspector: ObjectPerson | null
}

type WorkItem = {
  id: number
  name: string
  unit: string
  plan_qty: number
}

const object = ref<ObjectCore | null>(null)
const workItems = ref<WorkItem[]>([])
const deliveries = ref<MaterialDelivery[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const submitting = ref(false)
const submitError = ref<string | null>(null)
const submitSuccess = ref<string | null>(null)
const reportForm = ref<Record<number, number>>({})
const deliveryForm = ref({
  work_item_id: '',
  date: new Date().toISOString().slice(0, 10),
})

const showDeliveryForm = ref(false)
const deliveryLoading = ref(false)
const deliveryError = ref<string | null>(null)
const deliveryForm = ref({
  material: '',
  qty: 0,
  date: new Date().toISOString().slice(0, 10),
})

async function loadObject() {
  loading.value = true
  error.value = null

  try {
    const res = await fetch(
      `${API_BASE}/foreman/objects/${route.params.id}`,
      { headers: { Authorization: `Bearer ${auth.token}` } },
    )
    if (!res.ok) throw new Error('Ошибка загрузки объекта')
    const data = await res.json()
    object.value = data.object
    workItems.value = data.work_items ?? []
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}

/* -------------------------------------------------------------
   Остальные функции (не меняются)
------------------------------------------------------------- */
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
    const res = await fetch(`${API_BASE}/foreman/objects/${route.params.id}/work-reports`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
    )
    if (!res.ok) throw new Error('Ошибка отправки отчёта')
    reportForm.value = {}
    submitSuccess.value = 'Отчёт успешно отправлен'
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
  } catch (e: any) {
    deliveryError.value = e.message || 'Ошибка'
  } finally {
    deliveryLoading.value = false
  }
}

function formatDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

function statusLabel(status: string) {
  switch (status) {
    case 'PLANNED': return 'Запланирован'
    case 'WAITING_INSPECTOR_CONFIRMATION': return 'Ожидает подтверждения'
    case 'ACTIVE': return 'Активен'
    case 'FINISHED': return 'Завершён'
    default: return status
  }
}

function statusClass(status: string) {
  return {
    'status-chip--planned': status === 'PLANNED',
    'status-chip--waiting': status === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active': status === 'ACTIVE',
    'status-chip--finished': status === 'FINISHED',
  }
}

function goBack() {
  router.push({ name: 'foreman-objects' })
}
function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}
function formatDate(value: string) {
  return new Date(value).toLocaleDateString('ru-RU')
}
onMounted(loadObject)
</script>

<style scoped>
.foreman-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

/* Сайдбар — один в один с customer */
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
  background: #f59e0b;
}

/* Основная область */
.foreman-main {
  padding: 24px 32px;
  box-sizing: border-box;
  max-width: 900px;
}

/* Шапка объекта */
.object-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.object-header-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.back-btn {
  background: none;
  border: none;
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
}

.back-btn:hover {
  color: #4338ca;
}

.object-title {
  margin: 0 0 4px;
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

/* Статус-чипы */
.status-chip {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
}

.status-chip--planned   { background: #e5e7eb; color: #374151; }
.status-chip--waiting   { background: #fef3c7; color: #92400e; }
.status-chip--active    { background: #dcfce7; color: #166534; }
.status-chip--finished  { background: #e0f2fe; color: #1d4ed8; }

/* Даты */
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

/* Секции */
.work-section,
.delivery-section {
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

.section-hint {
  font-size: 12px;
  color: #9ca3af;
}

/* Таблица работ */
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

.td-input input {
  width: 90px;
  padding: 5px 8px;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
  transition: border-color 0.15s, box-shadow 0.15s;
  outline: none;
}

.td-input input:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 1px rgba(129, 140, 248, 0.35);
  background: #ffffff;
}

/* Кнопки */
.primary-btn {
  padding: 8px 18px;
  border-radius: 999px;
  border: none;
  background: #4f46e5;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.primary-btn:hover:not(:disabled) {
  background: #4338ca;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: default;
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

.work-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

/* Форма поставки */
.delivery-form {
  padding-top: 4px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 160px 160px;
  gap: 12px;
  margin-bottom: 12px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-field label {
  font-size: 12px;
  color: #6b7280;
}

.form-field input {
  padding: 7px 10px;
  border-radius: 10px;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  font-size: 14px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.form-field input:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 1px rgba(129, 140, 248, 0.35);
  background: #ffffff;
}

/* Состояния */
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

/* Адаптив */
@media (max-width: 900px) {
  .foreman-main {
    padding: 16px;
  }

  .object-header {
    flex-direction: column;
    gap: 12px;
  }

  .object-header-people {
    align-items: flex-start;
  }

  .form-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .foreman-layout {
    grid-template-columns: 1fr;
  }

  .sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid #e5e7eb;
    padding: 16px;
  }
}
</style>