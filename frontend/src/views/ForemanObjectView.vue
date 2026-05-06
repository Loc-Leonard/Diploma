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
      <!-- ХЕДЕР – гарантируем наличие объекта через computed‑obj -->
      <header class="foreman-header" v-if="object">
        <div>
          <h1>{{ obj.name }}</h1>
          <div class="object-city">
            {{ obj.city }}, {{ obj.address }}
          </div>
        </div>
      </header>

      <!-- Статусы загрузки/ошибки -->
      <section v-if="loading" class="state">Загружаю объект...</section>
      <section v-else-if="error" class="state state--error">{{ error }}</section>

      <!-- Основное содержимое (все внутри v‑if="object") -->
      <template v-else-if="object">
        <!-- Таблица работ -->
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

        <!-- Форма загрузки поставки -->
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

        <!-- Список распознанных поставок -->
        <section class="card">
          <h2>Распознанные поставки</h2>
          <div v-if="!deliveries.length" class="state">Пока нет загруженных поставок</div>

          <div v-for="delivery in deliveries" :key="delivery.id" class="delivery-card">
            <div class="delivery-card__title">
              <strong>
                {{ delivery.material
                  || delivery.documents?.[0]?.original_file_name
                  || 'Документ' }}
              </strong>
              <span class="delivery-card__badge">{{ delivery.source }}</span>
            </div>
            <div class="delivery-card__meta">
              <span>Кол-во: {{ delivery.qty }} {{ delivery.unit || '' }}</span>
              <span v-if="delivery.document_number">Документ: {{ delivery.document_number }}</span>

              <!-- *** ОТКОРРЕКТИРОВАНО *** -->
              <span v-if="delivery.documents?.length">
                {{ delivery.documents?.[0]?.original_file_name }}
              </span>

              <span>{{ formatDate(delivery.date) }}</span>
            </div>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = import.meta.env.VITE_API_URL as string

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

/* -------------------------------------------------------------
   Типы данных
------------------------------------------------------------- */
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

/* -------------------------------------------------------------
   Reactive‑state
------------------------------------------------------------- */
const object = ref<ForemanObjectDTO | null>(null)   // может быть null до загрузки
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

/* -------------------------------------------------------------
   Выводим объект через computed‑переменную, чтобы TypeScript
   точно «знал», что он уже загружен (non‑null assertion).
------------------------------------------------------------- */
const obj = computed(() => {
  if (!object.value) {
    // На практике эта ветка никогда не выполнится, т.к. шаблон
    // проверяет v‑if="object". Но TypeScript требует возвращаемый тип.
    throw new Error('Object not loaded')
  }
  return object.value
})

/* -------------------------------------------------------------
   Загрузка данных из API
------------------------------------------------------------- */
async function loadObject() {
  loading.value = true
  error.value = null

  try {
    const res = await fetch(`${API_BASE}/foreman/objects/${route.params.id}`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })
    if (!res.ok) throw new Error('Ошибка загрузки объекта')

    const data: ForemanObjectDetailDTO = await res.json()
    object.value = data.object
    workItems.value = data.work_items
    deliveries.value = data.deliveries ?? []
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
    .filter(([, qty]) => qty && qty > 0)
    .map(([workItemId, qty]) => ({
      work_item_id: Number(workItemId),
      qty,
      date: new Date().toISOString().slice(0, 10),
    }))

  if (!reports.length) return

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
    if (!res.ok) throw new Error('Ошибка отправки отчётов')
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
    if (!res.ok) throw new Error(data.error || 'Ошибка загрузки документа')

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

/* … остальные стили остаются без изменений … */

.delivery-card__meta {
  margin-top: 8px;
  color: #64748b;
  font-size: 13px;
}
</style>