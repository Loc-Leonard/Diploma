import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChangePasswordView from '../views/ChangePasswordView.vue'
import AdminUsersView from '../views/AdminUsersView.vue'
import CustomerObjectsView from '../views/CustomerObjectsView.vue'
import { useAuthStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
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
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// переписываем guard без next() (Vue Router 4)
router.beforeEach((to, from) => {
  const auth = useAuthStore()

  // если маршрут требует авторизации, а пользователь не залогинен
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login' }
  }

  // если маршрут только для админа
  if (to.meta.adminOnly && auth.user?.role !== 'ADMIN') {
    return { name: 'login' }
  }

  // если уже залогинен и идём на /login — сразу перекидываем по роли
  if (to.name === 'login' && auth.isAuthenticated) {
    if (auth.user?.role === 'ADMIN') {
      return { name: 'admin-users' }
    }
    if (auth.user?.role === 'CUSTOMER') {
      return { name: 'customer-objects' }
    }
  }

  // ничего не возвращаем — переход продолжается
})

export default router