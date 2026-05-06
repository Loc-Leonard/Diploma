<template>
  <aside class="sidebar">
    <div class="sidebar-top">
      <div class="sidebar-logo">{{ greeting }}</div>
      <nav class="sidebar-nav">
        <button
          class="nav-item"
          :class="{ 'nav-item--active': route.name === 'customer-objects' || route.name === 'customer-object-details' }"
          @click="router.push({ name: 'customer-objects' })"
        >
          Объекты
        </button>
        <button class="nav-item" disabled>График</button>
        <button class="nav-item" disabled>Замечания</button>
        <button class="nav-item" disabled>Проверки</button>
        <button class="nav-item" disabled>Справочники</button>
      </nav>
    </div>
    <div class="sidebar-bottom">
      <div class="role-badge">
        <span class="role-dot role-dot--customer"></span>
        <span>Заказчик</span>
      </div>
      <button class="logout-button" @click="logout">Выйти</button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const greeting = computed(() => {
  const name = auth.user?.full_name
  return name ? `Добрый день, ${name}` : 'Добрый день'
})

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px 18px;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
}

.sidebar-logo { font-size: 15px; font-weight: 700; margin-bottom: 24px; color: #111827; }

.sidebar-nav { display: flex; flex-direction: column; gap: 6px; }

.nav-item {
  text-align: left; padding: 8px 10px; border-radius: 8px;
  border: none; background: transparent; font-size: 14px;
  color: #4b5563; cursor: pointer; transition: background 0.15s;
}
.nav-item--active { background: #eef2ff; color: #4338ca; }
.nav-item[disabled] { opacity: 0.5; cursor: default; }

.sidebar-bottom { display: flex; flex-direction: column; gap: 10px; }

.logout-button {
  padding: 7px 16px; border-radius: 999px;
  border: 1px solid #e5e7eb; background: #ffffff;
  font-size: 13px; color: #6b7280; cursor: pointer;
}

.role-badge { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; color: #6b7280; }
.role-dot { width: 10px; height: 10px; border-radius: 999px; }
.role-dot--customer { background: #34c924; }
</style>