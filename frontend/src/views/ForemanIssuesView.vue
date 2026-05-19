<template>
  <div class="customer-layout">
    <ForemanLayout />
    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Мои замечания</h1>
      </header>

      <section class="issues-section">
        <div class="section-header">
          <div class="issue-filters">
            <button v-for="item in filters" :key="item.value" type="button" class="filter-pill" :class="{ 'filter-pill--active': activeFilter === item.value }" @click="activeFilter = item.value">{{ item.label }}</button>
          </div>
        </div>

        <div v-if="loading" class="state">Загрузка...</div>
        <div v-else-if="error" class="state state--error">{{ error }}</div>
        <div v-else-if="groupedIssues.length === 0" class="state">Нет назначенных замечаний.</div>
        <div v-else class="grouped-list">
          <div v-for="group in groupedIssues" :key="group.id" class="object-group">
            <div class="object-group-header" @click="goToObject(group.id)">
              <span class="object-group-title">{{ group.name }}</span>
              <span class="object-group-count">{{ group.issues.length }}</span>
            </div>
            <div class="object-group-issues">
              <article v-for="issue in group.issues" :key="issue.id" class="issue-card" @click="openIssue(issue)">
                <div class="issue-card-title-row">
                  <span class="issue-type" :class="`issue-type--${issue.type}`">{{ issue.type === 'remark' ? 'Замечание' : 'Нарушение' }}</span>
                  <IssueStatusBadge :status="issue.display_status || issue.status" />
                </div>
                <div class="issue-title">{{ issue.title || issue.description }}</div>
                <div class="issue-description">{{ issue.description }}</div>
                <div class="issue-meta">
                  <span>Автор: {{ issue.author_name }}</span>
                  <span>Срок: {{ fmtDate(issue.due_date) }}</span>
                </div>
              </article>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import ForemanLayout from './ForemanLayout.vue'
import IssueStatusBadge from '@/components/issues/IssueStatusBadge.vue'

type IssueGroup = {
  id: number
  name: string
  issues: any[]
}

const API_BASE = 'http://localhost:8080'
const auth = useAuthStore()
const router = useRouter()

const objects = ref<{ id: number; name: string }[]>([])
const issues = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const activeFilter = ref('all')

const filters = [
  { value: 'all', label: 'Все' },
  { value: 'open', label: 'Открытые' },
  { value: 'in_progress', label: 'В работе' },
  { value: 'overdue', label: 'Просроченные' },
  { value: 'resolved_by_foreman', label: 'На проверке' },
  { value: 'closed', label: 'Закрытые' },
]

const filteredIssues = computed(() => {
  const items = issues.value ?? []
  switch (activeFilter.value) {
    case 'open': return items.filter((i: any) => i.status === 'open')
    case 'in_progress': return items.filter((i: any) => i.status === 'in_progress')
    case 'overdue': return items.filter((i: any) => (i.display_status || i.status) === 'overdue')
    case 'resolved_by_foreman': return items.filter((i: any) => i.status === 'resolved_by_foreman')
    case 'closed': return items.filter((i: any) => ['accepted', 'rejected'].includes(i.status))
    default: return items
  }
})

const groupedIssues = computed((): IssueGroup[] => {
  const map: Record<number, any[]> = {}
  
  for (const issue of filteredIssues.value) {
    const oid = issue.object_id as number
    if (!map[oid]) {
      map[oid] = []
    }
    map[oid]!.push(issue)
  }
  
  const result: IssueGroup[] = []
  for (const objectId in map) {
    const id = Number(objectId)
    const obj = objects.value.find(o => o.id === id)
    result.push({ 
      id, 
      name: obj?.name || `Объект #${id}`, 
      issues: map[id]!
    })
  }
  return result
})

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }
    const [objRes, issRes] = await Promise.all([
      fetch(`${API_BASE}/foreman/objects`, { headers }),
      fetch(`${API_BASE}/foreman/issues`, { headers })
    ])
    if (!objRes.ok || !issRes.ok) throw new Error('Ошибка загрузки данных')
    const objs = await objRes.json()
    objects.value = objs.map((o: any) => ({ id: o.id, name: o.name }))
    issues.value = await issRes.json()
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function openIssue(issue: any) {
  router.push({ name: 'foreman-object', params: { id: issue.object_id } })
}
function goToObject(id: number) {
  router.push({ name: 'foreman-object', params: { id } })
}
function fmtDate(v?: string | null) {
  if (!v) return '—'
  return new Date(v).toLocaleDateString('ru-RU')
}

onMounted(loadData)
</script>

<style scoped>
/* Стили аналогичны предыдущим */
.customer-layout { display: grid; grid-template-columns: 206px 1fr; min-height: 100vh; background: #f9fafb; }
.customer-main { grid-column: 2; padding: 24px 32px; }
.customer-header { margin-bottom: 16px; }
.customer-title { margin: 0; font-size: 22px; font-weight: 600; color: #111827; }
.issues-section { background: #fff; border: 1px solid #e5e7eb; border-radius: 16px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05); }
.section-header { display: flex; justify-content: flex-end; gap: 12px; align-items: flex-start; margin-bottom: 14px; flex-wrap: wrap; }
.issue-filters { display: flex; gap: 8px; flex-wrap: wrap; flex: 1; }
.filter-pill { border: 1px solid #d1d5db; background: #fff; color: #374151; border-radius: 999px; padding: 6px 12px; font-size: 12px; cursor: pointer; }
.filter-pill--active { background: #eef2ff; border-color: #a5b4fc; color: #4338ca; }
.state { font-size: 13px; color: #6b7280; padding: 16px 0; }
.state--error { color: #b91c1c; }
.grouped-list { display: flex; flex-direction: column; gap: 20px; }
.object-group { border: 1px solid #e5e7eb; border-radius: 12px; overflow: hidden; }
.object-group-header { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; background: #f9fafb; border-bottom: 1px solid #e5e7eb; cursor: pointer; transition: background 0.15s; }
.object-group-header:hover { background: #f3f4f6; }
.object-group-title { font-size: 15px; font-weight: 600; color: #111827; }
.object-group-count { font-size: 12px; color: #6b7280; background: #e5e7eb; padding: 2px 8px; border-radius: 999px; }
.object-group-issues { display: flex; flex-direction: column; gap: 8px; padding: 10px 12px; background: #fff; }
.issue-card { border: 1px solid #f3f4f6; border-radius: 10px; padding: 12px; background: #fff; cursor: pointer; transition: box-shadow 0.15s, border-color 0.15s; }
.issue-card:hover { border-color: #a5b4fc; box-shadow: 0 4px 12px rgba(15, 23, 42, 0.05); }
.issue-card-title-row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin-bottom: 6px; }
.issue-type { display: inline-flex; align-items: center; padding: 3px 10px; border-radius: 999px; font-size: 11px; font-weight: 600; }
.issue-type--remark { background: #e0f2fe; color: #0369a1; }
.issue-type--violation { background: #fef2f2; color: #b91c1c; }
.issue-title { font-size: 14px; font-weight: 600; color: #111827; }
.issue-description { margin-top: 4px; font-size: 13px; color: #4b5563; }
.issue-meta { display: flex; flex-wrap: wrap; gap: 8px 14px; margin-top: 10px; font-size: 12px; color: #6b7280; }
</style>