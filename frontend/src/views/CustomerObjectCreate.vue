<template>
  <div class="customer-layout">
    <CustomerLayout />

    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Новый объект</h1>
      </header>

      <!-- одна колонка, форма по центру -->
      <section class="create-object-wrapper">
        <div class="column column--objects">
          <div class="column-header">
            <h2>Данные объекта</h2>
          </div>

          <form class="object-form" @submit.prevent="submit">
            <div class="form-grid">
              <div class="form-field">
                <label>Название объекта</label>
                <input v-model="form.name" required />
              </div>

              <div class="form-field">
                <label>Город</label>
                <input v-model="form.city" required />
              </div>

              <div class="form-field form-field--full">
                <label>Адрес</label>
                <input v-model="form.address" required />
              </div>

              <div class="form-field form-field--full">
                <label>Описание</label>
                <textarea v-model="form.description" rows="3" />
              </div>

              <div class="form-field">
                <label>Плановая дата начала</label>
                <input type="date" v-model="form.plannedStartDate" />
              </div>

              <div class="form-field">
                <label>Плановая дата окончания</label>
                <input type="date" v-model="form.plannedEndDate" />
              </div>

              <div class="form-field">
                <label>Прораб</label>
                <select v-model.number="form.foremanUserId" required>
                  <option value="" disabled>Выберите прораба</option>
                  <option
                    v-for="u in foremen"
                    :key="u.id"
                    :value="u.id"
                  >
                    {{ u.full_name }}
                  </option>
                </select>
              </div>

              <div class="form-field">
                <label>Инспектор</label>
                <select v-model.number="form.inspectorUserId" required>
                  <option value="" disabled>Выберите инспектора</option>
                  <option
                    v-for="u in inspectors"
                    :key="u.id"
                    :value="u.id"
                  >
                    {{ u.full_name }}
                  </option>
                </select>
              </div>
            </div>

            <div class="form-actions">
              <button
                type="button"
                class="secondary-btn"
                @click="goToObjects"
              >
                Отмена
              </button>
              <button
                type="submit"
                class="primary-btn"
                :disabled="submitting"
              >
                Создать объект
              </button>
            </div>
          </form>
        </div>
      </section>
    </main>
  </div>
</template>


<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CustomerLayout from './CustomerLayout.vue'

const API_BASE = 'http://localhost:8080'

const router = useRouter()
const auth = useAuthStore()

type SimpleUser = {
  id: number
  full_name: string
}

const foremen = ref<SimpleUser[]>([])
const inspectors = ref<SimpleUser[]>([])
const submitting = ref(false)

const form = ref({
  name: '',
  city: '',
  address: '',
  description: '',
  plannedStartDate: '',
  plannedEndDate: '',
  foremanUserId: 0,
  inspectorUserId: 0,
})

async function loadUsers() {
  const headers = { Authorization: `Bearer ${auth.token}` }

  try {
    const [foremenRes, inspectorsRes] = await Promise.all([
      fetch(`${API_BASE}/customer/foremen-list`, { headers }),
      fetch(`${API_BASE}/customer/inspectors-list`, { headers }),
    ])

    if (foremenRes.ok) {
      foremen.value = await foremenRes.json()
    }
    if (inspectorsRes.ok) {
      inspectors.value = await inspectorsRes.json()
    }
  } catch (e) {
    console.error(e)
  }
}

async function submit() {
  submitting.value = true
  try {
    const body = {
      name: form.value.name,
      city: form.value.city,
      address: form.value.address,
      description: form.value.description,
      planned_start_date: form.value.plannedStartDate
        ? new Date(form.value.plannedStartDate).toISOString()
        : null,
      planned_end_date: form.value.plannedEndDate
        ? new Date(form.value.plannedEndDate).toISOString()
        : null,
      foreman_user_id: form.value.foremanUserId,
      inspector_user_id: form.value.inspectorUserId,
      lat: 0,
      lng: 0,
    }

    const res = await fetch(`${API_BASE}/customer/objects`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify(body),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка создания объекта')
    }

    goToObjects()
  } catch (e: any) {
    alert(e.message || 'Ошибка')
  } finally {
    submitting.value = false
  }
}

function goToObjects() {
  router.push({ name: 'customer-objects' })
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

onMounted(loadUsers)
</script>

<style scoped>
.customer-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

/* сайдбар — как раньше */
.sidebar {
  grid-column: 1;
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

.role-dot--customer {
  background: #34c924;
}

/* контент: симметричные отступы */
.customer-main {
  grid-column: 2;
  padding: 20px 32px; /* одинаковый внутренний отступ слева/справа */
  box-sizing: border-box;
}

.customer-header {
  margin-bottom: 16px;
}

.customer-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

/* враппер, который ограничивает ширину формы и центрирует её */
.create-object-wrapper {
  display: flex;
  justify-content: center;
}

/* ширина формы, чтобы отступ от правого края выглядел, как на главной */
.column {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
  border: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 960px;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.column-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

/* форма */
.object-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-field--full {
  grid-column: 1 / -1;
}

.form-field label {
  font-size: 13px;
  color: #6b7280;
}

.form-field input,
.form-field textarea,
.form-field select {
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  padding: 8px 10px;
  font-size: 14px;
}

.form-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>