<template>
  <div class="customer-layout">
    <aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>
        <nav class="sidebar-nav">
          <button class="nav-item" :class="{ 'nav-item--active': isChecksActive }" @click="goToChecks">
            Проверки
          </button>
          <button class="nav-item" disabled>График</button>
          <button class="nav-item" disabled>Замечания</button>
          <button class="nav-item" :class="{ 'nav-item--active': isObjectsActive }" @click="goToObjects">
            <span>Объекты</span>
            <span v-if="pendingCount > 0" class="nav-badge">{{ pendingCount }}</span>
          </button>
          <button class="nav-item" disabled>Справочники</button>
        </nav>
      </div>

      <div class="sidebar-bottom">
        <div class="role-badge">
          <span class="role-dot role-dot--inspector"></span>
          <span>Инспектор</span>
        </div>
        <button class="logout-button" @click="logout">Выйти</button>
      </div>
    </aside>

    <main class="customer-main">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = 'http://localhost:8080'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const pendingCount = ref(0)
let pollId: number | null = null

const greeting = computed(() => {
  if (!auth.isAuthenticated) return 'Добрый день'
  return auth.user?.full_name
    ? `Добрый день, ${auth.user.full_name}`
    : 'Добрый день'
})

const isChecksActive = computed(() => route.name === 'inspector-checks')
const isObjectsActive = computed(() => route.name === 'inspector-objects')

async function fetchPendingCount() {
  try {
    const res = await fetch(`${API_BASE}/inspector/objects?status=WAITING_INSPECTOR_CONFIRMATION`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
      },
    })

    if (!res.ok) return

    const data = await res.json()
    pendingCount.value = Array.isArray(data) ? data.length : 0
  } catch {
    pendingCount.value = 0
  }
}

function goToChecks() {
  if (route.name !== 'inspector-checks') {
    router.push({ name: 'inspector-checks' })
  }
}

function goToObjects() {
  if (route.name !== 'inspector-objects') {
    router.push({ name: 'inspector-objects' })
  }
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

onMounted(() => {
  fetchPendingCount()
  pollId = window.setInterval(fetchPendingCount, 15000)
})

onBeforeUnmount(() => {
  if (pollId) window.clearInterval(pollId)
})
</script>

<style scoped>
.customer-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

.sidebar {
  grid-column: 1;
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
  display: flex;
  align-items: center;
  justify-content: space-between;
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

.nav-badge {
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  background: #dc2626;
  color: #fff;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
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

.role-dot--inspector {
  background: #9524c9;
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

.customer-main {
  grid-column: 2;
  min-width: 0;
  padding: 20px 24px;
  box-sizing: border-box;
}
</style>