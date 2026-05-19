<template>
<aside class="sidebar">
      <div class="sidebar-top">
        <div class="sidebar-logo">{{ greeting }}</div>
        <nav class="sidebar-nav">
           <button 
            class="nav-item" 
            :class="{'nav-item--active' : route.name === 'foreman-objects' || route.name === 'foreman-object'}"
            @click="router.push({ name: 'foreman-objects' })"
            >
            Объекты
          </button>
          <button class="nav-item" :class="{ 'nav-item--active': route.name === 'foreman-issues' }"
          @click="router.push({ name: 'foreman-issues' })"
          >
          <span>Замечания</span>
          <span v-if="notifications.unreadCount > 0" class="nav-badge">
            {{ notifications.unreadCount }}
          </span>
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
</template>
<script setup lang="ts">

import { computed, onBeforeMount, onBeforeUnmount, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useForemanNotificationsStore } from '@/stores/foremanNotifications'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const notifications = useForemanNotificationsStore()

// Приветствие в левом верхнем углу
const greeting = computed(() => {
  if (!auth.isAuthenticated) {
    return 'Добрый день'
  }
  const u = auth.user
  return u?.full_name ? `Добрый день, ${u.full_name}` : 'Добрый день'
})

function onIssueClick() {
  notifications.markAsRead()
  router.push({ name: 'foreman-issues' })
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
  notifications.stopPolling()
}

onMounted(()=> {
  notifications.startPolling()
})

onBeforeUnmount(()=> {
  notifications.stopPolling()
})
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
  transition: default;
}

.nav-item--active {
  background: #eef2ff;
  color: #4338ca;
}

.nav-badge {
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  background: #dc2626; /* красный */
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: 6px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.85; transform: scale(1.05); }
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