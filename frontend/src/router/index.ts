import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChangePasswordView from '../views/ChangePasswordView.vue'
import AdminUsersView from '../views/AdminUsersView.vue'
import CustomerObjectsView from '../views/CustomerObjectsView.vue'
import { useAuthStore } from '../stores/auth'

import ForemanObjectView from '../views/ForemanObjectView.vue'
import ForemanObjectsView from '../views/ForemanObjectsView.vue'
import InspectorDashboardView from '../views/InspectorDashboardView.vue'
import CustomerObjectCreate from '../views/CustomerObjectCreate.vue'


const routes: RouteRecordRaw[] = [
  {
    path: '/customer/objects/new',
    name: 'customer-object-create',
    component: CustomerObjectCreate

  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/change-password',
    name: 'change-password',
    component: ChangePasswordView,
    meta: { requiresAuth: true },
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: AdminUsersView,
    meta: { requiresAuth: true, adminOnly: true },
  },
  {
    path: '/customer/objects',
    name: 'customer-objects',
    component: CustomerObjectsView,
    meta: { requiresAuth: true, customerOnly: true },
  },
  {
    path: '/foreman/objects',
    name: 'foreman-objects',
    component: ForemanObjectsView,
    meta: { requiresAuth: true, foremanOnly: true },
  },
  {
    path: '/foreman/objects/:id',
    name: 'foreman-object',
    component: ForemanObjectView,
    props: true,
    meta: { requiresAuth: true, foremanOnly: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login',
  },
  {
  path: '/inspector/checks',
  name: 'inspector-checks',
  component: InspectorDashboardView,
  meta: { requiresAuth: true, inspectorOnly: true },
},
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  // 1) Требуется авторизация — а её нет
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login' }
  }

  // 2) Ролевые ограничения
  if (to.meta.adminOnly && auth.user?.role !== 'ADMIN') {
    return { name: 'login' }
  }

  if (to.meta.customerOnly && auth.user?.role !== 'CUSTOMER') {
    return { name: 'login' }
  }

  if (to.meta.foremanOnly && auth.user?.role !== 'FOREMAN') {
    return { name: 'login' }
  }

  if (to.meta.inspectorOnly && auth.user?.role !== 'INSPECTOR') {
    return { name: 'login' }
  }

  // 3) Если уже залогинен и идём на /login — перекидываем по роли
  if (to.name === 'login' && auth.isAuthenticated) {
    switch (auth.user?.role) {
      case 'ADMIN':
        return { name: 'admin-users' }
      case 'CUSTOMER':
        return { name: 'customer-objects' }
      case 'FOREMAN':
        return { name: 'foreman-objects' }
      case 'INSPECTOR':
        return { name: 'inspector-checks' }
    }
  }

  // ничего не возвращаем — переход продолжается
})


export default router