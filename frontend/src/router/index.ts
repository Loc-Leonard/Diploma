import { createRouter, createWebHistory,  type RouteRecordRaw } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChangePasswordView from '../views/ChangePasswordView.vue'
import AdminUsersView from '../views/AdminUsersView.vue'
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
    path: '/:pathMatch(.*)*',
    redirect: '/login',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next({ name: 'login' })
    return
  }

  if (to.meta.adminOnly && auth.user?.role !== 'ADMIN') {
    next({ name: 'login' })
    return
  }

  next()
})

export default router