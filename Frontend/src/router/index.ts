import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import BuilderView from '../views/BuilderView.vue'
import TestZoneView from '../views/TestZoneView.vue'
import GuideView from '../views/GuideView.vue'

const routes = [
  {
    path: '/',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/guide',
    name: 'guide',
    component: GuideView,
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
