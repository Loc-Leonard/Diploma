<template>
  <div class="customer-layout">
    <CustomerLayout />

    <main class="customer-main">

      <header class="customer-header">
        <div class="customer-header-left">
          <button class="back-btn" @click="router.push({ name: 'customer-objects' })">
            ← Назад
          </button>
          <h1 class="customer-title">{{ obj?.name ?? '...' }}</h1>
          <span v-if="obj" class="status-chip" :class="statusClass(obj.status)">
            {{ statusLabel(obj.status) }}
          </span>
        </div>
      </header>

      <div v-if="loading" class="state">Загружаю объект...</div>
      <div v-else-if="error" class="state state--error">{{ error }}</div>

      <div v-else-if="obj" class="detail-body">

        <aside class="detail-aside">
          <div class="mini-map">
            <div class="map-placeholder-box">🗺</div>
          </div>

          <section class="aside-section">
            <h2 class="aside-title">Ответственные лица</h2>
            <div class="person-block">
              <span class="person-role">Заказчик:</span>
              <span class="person-name">{{ obj.customer?.full_name ?? '—' }}</span>
            </div>
            <div class="person-block">
              <span class="person-role">Прораб:</span>
              <span class="person-name">{{ obj.foreman?.full_name ?? '—' }}</span>
            </div>
            <div class="person-block">
              <span class="person-role">Инспектор:</span>
              <span class="person-name">{{ obj.inspector?.full_name ?? '—' }}</span>
            </div>
          </section>

          <section class="aside-section">
            <div class="date-row">
              <span class="date-label">Плановое начало</span>
              <span>{{ fmtDate(obj.planned_start_date) }}</span>
            </div>
            <div class="date-row">
              <span class="date-label">Плановое окончание</span>
              <span>{{ fmtDate(obj.planned_end_date) }}</span>
            </div>
          </section>

          <button class="aside-action-btn">Редактировать ▾</button>
          <button class="aside-action-btn">История изменений ▾</button>
        </aside>

        <div class="detail-main">

          <section class="card">
            <div class="card-header">
              <h2>График</h2>
              <div class="card-filters">
                <span class="filter-pill">Задачи ▾</span>
                <span class="filter-pill">Статус ▾</span>
                <span class="filter-pill">Отв. лица ▾</span>
              </div>
            </div>
            <div class="gantt-placeholder">График работ появится здесь</div>
          </section>

          <div class="bottom-grid">
            <section class="card">
              <div class="card-header">
                <h2>Материалы</h2>
                <div class="card-filters">
                  <span class="filter-pill">Тип ▾</span>
                  <span class="filter-pill">Ответственные ▾</span>
                  <span class="filter-pill">Даты ▾</span>
                </div>
              </div>
              <div class="state">Материалов пока нет</div>
            </section>

            <section class="card">
              <div class="card-header">
                <h2>Документы</h2>
                <div class="card-filters">
                  <span class="filter-pill">Тип ▾</span>
                  <span class="filter-pill">Ответственные ▾</span>
                  <span class="filter-pill">Даты ▾</span>
                </div>
              </div>
              <div class="state">Документов пока нет</div>
            </section>
          </div>

        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CustomerLayout from './CustomerLayout.vue'

const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

type ObjStatus = 'PLANNED' | 'WAITING_INSPECTOR_CONFIRMATION' | 'ACTIVE' | 'FINISHED'
interface Person { id: number; full_name: string }
interface ObjectDetail {
  id: number
  name: string
  address: string
  city: string
  description: string
  status: ObjStatus
  lat: number
  lng: number
  planned_start_date?: string | null
  planned_end_date?: string | null
  customer?:  Person
  foreman?:   Person
  inspector?: Person
}

const obj     = ref<ObjectDetail | null>(null)
const loading = ref(true)
const error   = ref<string | null>(null)

async function fetchObject() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch(`${API_BASE}/customer/objects/${route.params.id}`, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error ?? 'Ошибка загрузки')
    obj.value = await res.json()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function fmtDate(v?: string | null) {
  if (!v) return '—'
  return new Date(v).toLocaleDateString('ru-RU')
}

function statusLabel(s: ObjStatus) {
  const m: Record<ObjStatus, string> = {
    PLANNED:                        'Запланирован',
    WAITING_INSPECTOR_CONFIRMATION: 'Ожидает подтверждения',
    ACTIVE:                         'Активен',
    FINISHED:                       'Завершён',
  }
  return m[s] ?? s
}

function statusClass(s: ObjStatus) {
  return {
    'status-chip--planned':  s === 'PLANNED',
    'status-chip--waiting':  s === 'WAITING_INSPECTOR_CONFIRMATION',
    'status-chip--active':   s === 'ACTIVE',
    'status-chip--finished': s === 'FINISHED',
  }
}

onMounted(fetchObject)
</script>

<style scoped>
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

.customer-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.customer-header-left { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }

.back-btn {
  padding: 6px 14px; border-radius: 999px;
  border: 1px solid #d1d5db; background: #fff;
  font-size: 14px; color: #374151; cursor: pointer;
  transition: background 0.15s;
}
.back-btn:hover { background: #f3f4f6; }

.customer-title { margin: 0; font-size: 22px; font-weight: 600; color: #111827; }

.detail-body {
  display: grid;
  grid-template-columns: 210px 1fr;
  gap: 20px;
  align-items: start;
}

.detail-aside { display: flex; flex-direction: column; gap: 14px; }

.mini-map .map-placeholder-box {
  width: 100%; height: 190px; border-radius: 12px;
  border: 1px solid #e5e7eb; background: #f3f4f6;
  display: flex; align-items: center; justify-content: center;
  font-size: 40px;
}

.aside-section { display: flex; flex-direction: column; gap: 10px; }
.aside-title { margin: 0 0 4px; font-size: 14px; font-weight: 600; color: #111827; }

.person-block { display: flex; flex-direction: column; }
.person-role { font-size: 11px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.04em; }
.person-name { font-size: 14px; color: #4f46e5; font-weight: 500; }

.date-row { display: flex; justify-content: space-between; font-size: 13px; color: #374151; }
.date-label { color: #9ca3af; }

.aside-action-btn {
  width: 100%; padding: 9px 12px; border-radius: 10px;
  border: 1px solid #e5e7eb; background: #fff;
  font-size: 14px; font-weight: 600; color: #111827;
  cursor: pointer; text-align: left; transition: background 0.15s;
}
.aside-action-btn:hover { background: #f9fafb; }

.detail-main { display: flex; flex-direction: column; gap: 16px; min-width: 0; }

.card {
  background: #fff; border-radius: 16px; padding: 16px 18px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 2px 8px rgba(15,23,42,0.04);
}

.card-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 14px;
}
.card-header h2 { margin: 0; font-size: 16px; font-weight: 600; color: #111827; }

.card-filters { display: flex; gap: 8px; }
.filter-pill {
  padding: 4px 10px; border-radius: 999px;
  border: 1px solid #d1d5db; font-size: 12px;
  color: #6b7280; cursor: pointer; background: #fff;
}

.gantt-placeholder {
  height: 260px; border-radius: 10px;
  border: 2px dashed #e5e7eb; background: #f9fafb;
  display: flex; align-items: center; justify-content: center;
  color: #9ca3af; font-size: 14px;
}

.bottom-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.status-chip {
  padding: 3px 10px; border-radius: 999px;
  font-size: 11px; font-weight: 500; white-space: nowrap;
}
.status-chip--planned  { background: #e5e7eb; color: #374151; }
.status-chip--waiting  { background: #fef3c7; color: #92400e; }
.status-chip--active   { background: #dcfce7; color: #166534; }
.status-chip--finished { background: #e0f2fe; color: #1d4ed8; }

.state { font-size: 13px; color: #6b7280; padding: 8px 0; }
.state--error { color: #b91c1c; }

@media (max-width: 900px) {
  .detail-body { grid-template-columns: 1fr; }
  .bottom-grid { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .customer-layout { grid-template-columns: 1fr; }
  .customer-main { padding: 16px; }
}
</style>