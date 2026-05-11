import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ChangePasswordView from '../views/ChangePasswordView.vue'
import AdminUsersView from '../views/AdminUsersView.vue'
import CustomerObjectsView from '../views/CustomerObjectsView.vue'
import CustomerObjectCreate from '../views/CustomerObjectCreate.vue'
import CustomerObjectView from '../views/CustomerObjectView.vue'
import ForemanObjectsView from '../views/ForemanObjectsView.vue'
import ForemanObjectView from '../views/ForemanObjectView.vue'
import InspectorLayout from '@/views/InspectorLayout.vue'
import InspectorDashboardView from '../views/InspectorDashboardView.vue'
import InspectorObjectsView from '../views/InspectorObjectsView.vue'
import InspectorObjectView from '../views/InspectorObjectView.vue'
import ObjectDetailView from '../views/ObjectDetailView.vue'
import { useAuthStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: LoginView },
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
  
  // ===== CUSTOMER =====
  {
    path: '/customer/objects/new',
    name: 'customer-object-create',
    component: CustomerObjectCreate,
    meta: { requiresAuth: true, customerOnly: true },
  },
  {
    path: '/customer/objects',
    name: 'customer-objects',
    component: CustomerObjectsView,
    meta: { requiresAuth: true, customerOnly: true },
  },
  {
    path: '/customer/objects/:id',
    name: 'customer-object',
    component: CustomerObjectView,
    meta: { requiresAuth: true, customerOnly: true },
  },
  {
    path: '/customer/objects/:id',
    name: 'customer-object-details',
    component: ObjectDetailView,
    meta: { requiresAuth: true, customerOnly: true },
  },
  
  // ===== FOREMAN =====
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
    meta: { requiresAuth: true, foremanOnly: true },
  },
  {
    path: '/foreman/objects/:id',
    name: 'foreman-object',
    component: ObjectDetailView,
    meta: { requiresAuth: true, foremanOnly: true },
  },
  
  // ===== INSPECTOR =====
  {
    path: '/inspector',
    component: InspectorLayout,
    meta: { requiresAuth: true, inspectorOnly: true },
    children: [
      { path: '', redirect: { name: 'inspector-checks' } },
      { path: 'checks', name: 'inspector-checks', component: InspectorDashboardView },
      {
        path: 'objects',
        name: 'inspector-objects',
        component: InspectorObjectsView,
      },
      {
        path: 'objects/:id',
        name: 'inspector-object',
        component: InspectorObjectView,
      },
    ],
  },
  {
    path: '/inspector/objects/:id',
    name: 'inspector-object-details',
    component: ObjectDetailView,
    meta: { requiresAuth: true, inspectorOnly: true },
  },
  
  // ===== CATCH ALL =====
  { path: '/:pathMatch(.*)*', redirect: '/login' },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) return { name: 'login' }
  if (to.meta.adminOnly     && auth.user?.role !== 'ADMIN')     return { name: 'login' }
  if (to.meta.customerOnly  && auth.user?.role !== 'CUSTOMER')  return { name: 'login' }
  if (to.meta.foremanOnly   && auth.user?.role !== 'FOREMAN')   return { name: 'login' }
  if (to.meta.inspectorOnly && auth.user?.role !== 'INSPECTOR') return { name: 'login' }

  if (to.name === 'login' && auth.isAuthenticated) {
    switch (auth.user?.role) {
      case 'ADMIN':     return { name: 'admin-users' }
      case 'CUSTOMER':  return { name: 'customer-objects' }
      case 'FOREMAN':   return { name: 'foreman-objects' }
      case 'INSPECTOR': return { name: 'inspector-checks' }
    }
  }
})

export default router