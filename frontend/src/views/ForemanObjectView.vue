<template>
  <div class="foreman-layout">
    <aside class="sidebar">
      <!-- тот же sidebar -->
      <div class="sidebar-top">
        <div class="sidebar-logo">ЛОГОТИП</div>
        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active" @click="goBack">
            Объекты
          </button>
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
          <div class="object-city">
            {{ object.city }}, {{ object.address }}
          </div>
        </div>
      </header>

      <section v-if="loading" class="state">Загружаю объект...</section>
      <section v-else-if="error" class="state state--error">{{ error }}</section>

      <section v-else-if="object" class="work-section">
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
                  type="number"
                  min="0"
                  v-model.number="reportForm[item.id]"
                  placeholder="объем"
                />
              </td>
            </tr>
          </tbody>
        </table>

        <button
          class="primary-btn"
          @click="submitReports"
          :disabled="submitting"
        >
          Отправить выполненные работы
        </button>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
}

const object = ref<ForemanObjectDTO | null>(null)
const workItems = ref<WorkItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const submitting = ref(false)

// key: workItemId, value: qty
const reportForm = ref<Record<number, number>>({})

async function loadObject() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch(
      `${API_BASE}/foreman/objects/${route.params.id}`,
      {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      },
    )
    if (!res.ok) {
      throw new Error('Ошибка загрузки объекта')
    }
    const data: ForemanObjectDetailDTO = await res.json()
    object.value = data.object
    workItems.value = data.work_items
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
    if (!res.ok) {
      throw new Error('Ошибка отправки отчётов')
    }
    // очищаем форму
    reportForm.value = {}
  } catch (e: any) {
    alert(e.message || 'Ошибка')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push({ name: 'foreman-objects' })
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

onMounted(loadObject)
</script>