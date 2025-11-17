import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('../views/init/Init.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/init/callback',
    name: 'InitCallback',
    component: () => import('../views/init/InitCallback.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/auth/wechat/callback',
    name: 'WeChatCallback',
    component: () => import('../views/auth/WeChatCallback.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/dashboard/Dashboard.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  
  // 初始化页面不需要检查初始化状态
  if (to.name === 'Init') {
    next()
    return
  }
  
  // 检查系统是否已初始化
  try {
    const { checkInitStatus } = await import('@/api/init')
    const status = await checkInitStatus()
    if (!status.initialized) {
      // 如果未初始化，跳转到初始化页面
      next({ name: 'Init' })
      return
    }
  } catch (error) {
    // 如果检查失败，允许继续（可能是网络问题）
    console.error('检查初始化状态失败:', error)
  }
  
  // 如果访问登录页且已登录，重定向到工作台
  if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
    return
  }
  
  // 如果需要认证但未登录，重定向到登录页
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }
  
  next()
})

export default router

