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
    path: '/auth/wechat/add-user/callback',
    name: 'WeChatAddUserCallback',
    component: () => import('../views/user/WeChatAddUserCallback.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/auth/change-password',
    name: 'ChangePassword',
    component: () => import('../views/auth/ChangePassword.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/dashboard/Dashboard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/user',
    name: 'User',
    component: () => import('../views/user/User.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/permission',
    name: 'Permission',
    component: () => import('../views/permission/Permission.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/department',
    name: 'Department',
    component: () => import('../views/department/Department.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/project/:id/boards',
    name: 'BoardList',
    component: () => import('../views/board/BoardList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/project/:id/gantt',
    name: 'Gantt',
    component: () => import('../views/gantt/Gantt.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/project/:id/progress',
    name: 'Progress',
    component: () => import('../views/progress/Progress.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/project/:id',
    name: 'ProjectDetail',
    component: () => import('../views/project/ProjectDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/project',
    name: 'Project',
    component: () => import('../views/project/Project.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/requirement/:id',
    name: 'RequirementDetail',
    component: () => import('../views/requirement/RequirementDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/requirement',
    name: 'Requirement',
    component: () => import('../views/requirement/Requirement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/bug/:id',
    name: 'BugDetail',
    component: () => import('../views/bug/BugDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/bug',
    name: 'Bug',
    component: () => import('../views/bug/Bug.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/task/:id',
    name: 'TaskDetail',
    component: () => import('../views/task/TaskDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/task',
    name: 'Task',
    component: () => import('../views/task/Task.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/version/:id',
    name: 'VersionDetail',
    component: () => import('../views/version/VersionDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/version',
    name: 'Version',
    component: () => import('../views/version/Version.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/test-case',
    name: 'TestCase',
    component: () => import('../views/test/TestCase.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/product',
    name: 'Product',
    component: () => import('../views/product/Product.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/board/:id',
    name: 'Board',
    component: () => import('../views/board/Board.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource/statistics',
    name: 'ResourceStatistics',
    component: () => import('../views/resource/ResourceStatistics.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/reports',
    name: 'Report',
    component: () => import('../views/report/Report.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/system/wechat-settings',
    name: 'WeChatSettings',
    component: () => import('../views/system/WeChatSettings.vue'),
    meta: { requiresAuth: false } // 初始化阶段也需要访问
  },
  {
    path: '/system/backup-settings',
    name: 'BackupSettings',
    component: () => import('../views/system/BackupSettings.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/system/log-settings',
    name: 'LogSettings',
    component: () => import('../views/system/LogSettings.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/reports/daily/create',
    name: 'CreateDailyReport',
    component: () => import('../views/report/Report.vue'),
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
  
  // 检查系统是否已初始化
  try {
    const { checkInitStatus } = await import('@/api/init')
    const status = await checkInitStatus()
    
    // 如果访问初始化页面但系统已初始化，跳转到登录页
    if (to.name === 'Init' && status.initialized) {
      next({ name: 'Login' })
      return
    }
    
    // 如果访问其他页面但系统未初始化，跳转到初始化页面
    // 排除初始化相关路由和微信回调路由
    // WeChatSettings只有在from=init时才允许访问（从初始化页面跳转过来）
    if (to.name !== 'Init' && 
        to.name !== 'InitCallback' && 
        to.name !== 'WeChatCallback' && 
        to.name !== 'WeChatAddUserCallback' &&
        !status.initialized) {
      // 如果是微信设置页面，但不是从初始化页面跳转过来的，跳转到初始化页面
      if (to.name === 'WeChatSettings' && to.query.from !== 'init') {
        next({ name: 'Init' })
        return
      }
      // 其他页面直接跳转到初始化页面
      if (to.name !== 'WeChatSettings') {
        next({ name: 'Init' })
        return
      }
    }
  } catch (error) {
    // 如果检查失败，允许继续（可能是网络问题）
    console.error('检查初始化状态失败:', error)
    // 如果访问Init页面或微信回调页面且检查失败，允许继续（可能是网络问题）
    if (to.name === 'Init' || 
        to.name === 'InitCallback' || 
        to.name === 'WeChatCallback' || 
        to.name === 'WeChatAddUserCallback' ||
        to.name === 'WeChatSettings') { // 允许访问微信设置页面
      next()
      return
    }
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
  
  // 如果用户已登录但用户信息未加载，先加载用户信息（刷新页面时的情况）
  if (authStore.isAuthenticated && !authStore.user && to.meta.requiresAuth) {
    try {
      await authStore.loadUserInfo()
    } catch (error) {
      // 如果加载失败，可能是token过期，跳转到登录页
      console.error('加载用户信息失败:', error)
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }
  }
  
  // 如果用户已登录且需要修改密码，且访问的不是修改密码页面，强制跳转到修改密码页面
  if (authStore.isAuthenticated && authStore.isFirstLogin && to.name !== 'ChangePassword') {
    next({ name: 'ChangePassword' })
    return
  }
  
  next()
})

export default router

