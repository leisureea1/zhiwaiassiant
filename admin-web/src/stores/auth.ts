import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, usersApi } from '@/api'
import router from '@/router'

export interface AdminUser {
  id: string
  username: string
  realName?: string
  email?: string
  role: string
  avatar?: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(sessionStorage.getItem('admin_token') || '')
  const refreshToken = ref(sessionStorage.getItem('admin_refresh_token') || '')
  const user = ref<AdminUser | null>(
    JSON.parse(sessionStorage.getItem('admin_user') || 'null')
  )

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(
    () => user.value?.role === 'ADMIN' || user.value?.role === 'SUPER_ADMIN'
  )
  const isSuperAdmin = computed(() => user.value?.role === 'SUPER_ADMIN')

  async function login(studentId: string, password: string) {
    const res: any = await authApi.login({ studentId, password })
    // 响应拦截器已自动解包 { code, data } -> data
    // res = { accessToken, refreshToken, user }
    if (res.user?.role !== 'ADMIN' && res.user?.role !== 'SUPER_ADMIN') {
      throw new Error('无管理员权限')
    }
    token.value = res.accessToken
    refreshToken.value = res.refreshToken
    user.value = res.user
    sessionStorage.setItem('admin_token', res.accessToken)
    sessionStorage.setItem('admin_refresh_token', res.refreshToken)
    sessionStorage.setItem('admin_user', JSON.stringify(res.user))
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_refresh_token')
    localStorage.removeItem('admin_user')
  }

  async function fetchUser() {
    try {
      const res: any = await usersApi.getMe()
      user.value = res
      sessionStorage.setItem('admin_user', JSON.stringify(res))
    } catch {
      logout()
    }
  }

  function logout() {
    authApi.logout().catch(() => {})
    token.value = ''
    refreshToken.value = ''
    user.value = null
    sessionStorage.removeItem('admin_token')
    sessionStorage.removeItem('admin_refresh_token')
    sessionStorage.removeItem('admin_user')
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_refresh_token')
    localStorage.removeItem('admin_user')
    router.push('/login')
  }

  return { token, refreshToken, user, isLoggedIn, isAdmin, isSuperAdmin, login, fetchUser, logout }
})
