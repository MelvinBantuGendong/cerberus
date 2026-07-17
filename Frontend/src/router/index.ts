import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import BuilderView from '../views/BuilderView.vue'
import AnalyticsView from '../views/AnalyticsView.vue'
import TestZoneView from '../views/TestZoneView.vue'

const routes = [
  {
    path: '/',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/builder',
    name: 'builder',
    component: BuilderView,
  },
  {
    path: '/test-zone',
    name: 'testzone',
    component: TestZoneView,
  },
  {
    path: '/analytics',
    name: 'analytics',
    component: AnalyticsView,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const isAuthenticated = localStorage.getItem('cerberus_auth') === 'true'
  if (to.name !== 'login' && !isAuthenticated) {
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
