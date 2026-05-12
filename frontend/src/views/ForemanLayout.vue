<template>
<aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>
        <nav class="sidebar-nav">
          <button class="nav-item nav-item--active" @click="goBack">
            Объекты
          </button>
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
</template>
<script setup lang="ts">

import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

// Приветствие в левом верхнем углу
const greeting = computed(() => {
  if (!auth.isAuthenticated) {
    return 'Добрый день'
  }
  const u = auth.user
  return u?.full_name ? `Добрый день, ${u.full_name}` : 'Добрый день'
})
function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}
</script>
<style scoped>
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

.role-dot--foreman {
  background: #f1ce06;
}
</style>