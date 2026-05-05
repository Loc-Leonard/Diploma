<template>
  <div class="foreman-layout">
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">ЛОГОТИП</div>
        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active" @click="goBack">Объекты</button>
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
      <header class="foreman-header" v-if="object">
        <div>
          <h1>{{ object.name }}</h1>
          <div class="object-city">{{ object.city }}, {{ object.address }}</div>
        </div>
      </header>

      <section v-if="loading" class="state">Загружаю объект...</section>
      <section v-else-if="error" class="state state--error">{{ error }}</section>

      <template v-else-if="object">
        <section class="card">
          <h2>Работы</h2>

          <table class="work-table">
            <thead>
              <tr>
                <th>Вид работ</th>
                <th>План</th>
                <th>Выполнено за сегодня</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in workItems" :key="item.id">
                <td>{{ item.name }}</td>
                <td>{{ item.plan_qty }}</td>
                <td>
                  <input
                    v-model.number="reportForm[item.id]"
                    type="number"
                    min="0"
                    placeholder="объем"
                  />
                </td>
              </tr>
            </tbody>
          </table>

          <button class="primary-btn" @click="submitReports" :disabled="submitting">
            Отправить выполненные работы
          </button>
        </section>

        <section class="card">
          <h2>Поставка материалов по CV</h2>

          <form class="delivery-form" @submit.prevent="submitDeliveryFile">
            <div class="delivery-grid">
              <div class="form-field">
                <label>Файл ТТН / скан</label>
                <input type="file" accept=".pdf,.jpg,.jpeg,.png" @change="onDeliveryFileChange" />
              </div>

              <div class="form-field">
                <label>Работа</label>
                <select v-model="deliveryForm.work_item_id">
                  <option value="">Без привязки</option>
                  <option v-for="item in workItems" :key="item.id" :value="String(item.id)">
                    {{ item.name }}
                  </option>
                </select>
              </div>

              <div class="form-field">
                <label>Дата поставки</label>
                <input v-model="deliveryForm.date" type="date" />
              </div>
            </div>

            <div v-if="deliveryError" class="state state--error">{{ deliveryError }}</div>
            <div v-if="deliverySuccess" class="state state--success">{{ deliverySuccess }}</div>

            <button class="primary-btn" type="submit" :disabled="deliverySubmitting">
              {{ deliverySubmitting ? 'Обрабатываем...' : 'Загрузить и распознать' }}
            </button>
          </form>
        </section>

        <section class="card">
          <h2>Распознанные поставки</h2>
          <div v-if="!deliveries.length" class="state">Пока нет загруженных поставок</div>

          <div v-for="delivery in deliveries" :key="delivery.id" class="delivery-card">
            <div class="delivery-card__title">
              <strong>{{ delivery.material || delivery.documents?.[0]?.original_file_name || 'Документ' }}</strong>
              <span class="delivery-card__badge">{{ delivery.source }}</span>
            </div>
            <div class="delivery-card__meta">
              <span>Кол-во: {{ delivery.qty }} {{ delivery.unit || '' }}</span>
              <span v-if="delivery.document_number">Документ: {{ delivery.document_number }}</span>
              <span v-if="delivery.documents?.length">{{ delivery.documents[0].original_file_name }}</span>
              <span>{{ formatDate(delivery.date) }}</span>
            </div>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

type WorkItem = {
  id: number
  name: string
  unit: string
  plan_qty: number
}

type MaterialDocument = {
  id: number
  original_file_name: string
  storage_path: string
  mime_type: string
}

type MaterialDelivery = {
  id: number
  date: string
  material: string
  qty: number
  unit: string
  document_number: string
  source: string
  documents?: MaterialDocument[]
}

type ForemanObjectDTO = {
  id: number
  name: string
  city: string
  address: string
  status: string
}

type ForemanObjectDetailDTO = {
  object: ForemanObjectDTO
  work_items: WorkItem[]
  deliveries: MaterialDelivery[]
}

const object = ref<ForemanObjectDTO | null>(null)
const workItems = ref<WorkItem[]>([])
const deliveries = ref<MaterialDelivery[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const submitting = ref(false)
const deliverySubmitting = ref(false)
const deliveryError = ref<string | null>(null)
const deliverySuccess = ref<string | null>(null)
const deliveryFile = ref<File | null>(null)
const reportForm = ref<Record<number, number>>({})
const deliveryForm = ref({
  work_item_id: '',
  date: new Date().toISOString().slice(0, 10),
})

async function loadObject() {
  loading.value = true
  error.value = null

  try {
    const res = await fetch(`${API_BASE}/foreman/objects/${route.params.id}`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })
    if (!res.ok) {
      throw new Error('Ошибка загрузки объекта')
    }

    const data: ForemanObjectDetailDTO = await res.json()
    object.value = data.object
    workItems.value = data.work_items
    deliveries.value = data.deliveries || []
  } catch (e: any) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}

async function submitReports() {
  const reports = Object.entries(reportForm.value)
    .filter(([, qty]) => qty && qty > 0)
    .map(([workItemId, qty]) => ({
      work_item_id: Number(workItemId),
      qty,
      date: new Date().toISOString().slice(0, 10),
    }))

  if (!reports.length) {
    return
  }

  submitting.value = true
  try {
    const res = await fetch(`${API_BASE}/foreman/objects/${route.params.id}/work-reports`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ reports }),
    })
    if (!res.ok) {
      throw new Error('Ошибка отправки отчётов')
    }
    reportForm.value = {}
  } catch (e: any) {
    alert(e.message || 'Ошибка')
  } finally {
    submitting.value = false
  }
}

function onDeliveryFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  deliveryFile.value = target.files?.[0] || null
}

async function submitDeliveryFile() {
  if (!deliveryFile.value) {
    deliveryError.value = 'Выберите файл'
    return
  }

  deliverySubmitting.value = true
  deliveryError.value = null
  deliverySuccess.value = null

  try {
    const formData = new FormData()
    formData.append('file', deliveryFile.value)
    if (deliveryForm.value.work_item_id) {
      formData.append('work_item_id', deliveryForm.value.work_item_id)
    }
    if (deliveryForm.value.date) {
      formData.append('date', deliveryForm.value.date)
    }

    const res = await fetch(`${API_BASE}/foreman/objects/${route.params.id}/deliveries/cv-upload`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
      body: formData,
    })

    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.error || 'Ошибка загрузки документа')
    }

    deliverySuccess.value = 'Файл обработан, поставка сохранена'
    deliveryFile.value = null
    await loadObject()
  } catch (e: any) {
    deliveryError.value = e.message || 'Ошибка'
  } finally {
    deliverySubmitting.value = false
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
  background: #f8fafc;
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

.sidebar-bottom {
  display: flex;
  flex-direction: column;
  gap: 10px;
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
  background: #2563eb;
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

.foreman-main {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.foreman-header h1,
.card h2 {
  margin: 0;
}

.object-city,
.state {
  color: #64748b;
}

.state--error {
  color: #b91c1c;
}

.state--success {
  color: #15803d;
}

.card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}

.work-table {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}

.work-table th,
.work-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}

.work-table input,
.delivery-form input,
.delivery-form select {
  width: 100%;
  border-radius: 10px;
  border: 1px solid #cbd5e1;
  padding: 8px 10px;
  background: #f8fafc;
  box-sizing: border-box;
}

.primary-btn {
  padding: 10px 16px;
  border-radius: 999px;
  border: none;
  background: #2563eb;
  color: #ffffff;
  cursor: pointer;
}

.delivery-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.delivery-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-field label {
  font-size: 13px;
  color: #475569;
}

.delivery-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px 14px;
  background: #f8fafc;
}

.delivery-card + .delivery-card {
  margin-top: 10px;
}

.delivery-card__title,
.delivery-card__meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.delivery-card__meta {
  margin-top: 8px;
  color: #64748b;
  font-size: 13px;
}

.delivery-card__badge {
  padding: 2px 8px;
  border-radius: 999px;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 12px;
}

@media (max-width: 960px) {
  .foreman-layout {
    grid-template-columns: 1fr;
  }

  .sidebar {
    width: auto;
    border-right: none;
    border-bottom: 1px solid #e5e7eb;
  }

  .delivery-grid {
    grid-template-columns: 1fr;
  }
}
</style>
