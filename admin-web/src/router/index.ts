import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const isJwtExpired = (token: string): boolean => {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return true
    const payload = JSON.parse(atob(parts[1]))
    if (!payload?.exp) return true
    return Date.now() >= payload.exp * 1000
  } catch {
    return true
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/index.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'announcements',
        name: 'Announcements',
        component: () => import('@/views/announcements/index.vue'),
        meta: { title: '公告管理', icon: 'Bell' },
      },
      {
        path: 'announcements/create',
        name: 'AnnouncementCreate',
        component: () => import('@/views/announcements/edit.vue'),
        meta: { title: '发布公告', hidden: true },
      },
      {
        path: 'announcements/:id/edit',
        name: 'AnnouncementEdit',
        component: () => import('@/views/announcements/edit.vue'),
        meta: { title: '编辑公告', hidden: true },
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/logs/index.vue'),
        meta: { title: '系统日志', icon: 'Document' },
      },
      {
        path: 'email-broadcast',
        name: 'EmailBroadcast',
        component: () => import('@/views/email-broadcast/index.vue'),
        meta: { title: '邮件群发', icon: 'Message' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/index.vue'),
        meta: { title: '系统设置', icon: 'Setting', superAdminOnly: true },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || '管理后台'} - 知外助手`

  if (to.meta.public) {
    next()
    return
  }

  const authStore = useAuthStore()
  if (!authStore.isLoggedIn) {
    next('/login')
    return
  }

  if (!authStore.token || isJwtExpired(authStore.token)) {
    authStore.logout()
    next('/login')
    return
  }

  if (to.meta.superAdminOnly && !authStore.isSuperAdmin) {
    next('/dashboard')
    return
  }

  next()
})

export default router
