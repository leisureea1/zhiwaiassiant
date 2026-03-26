import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
})

let isRefreshing = false
let pendingQueue: Array<(token: string | null) => void> = []

const flushPendingQueue = (token: string | null) => {
  pendingQueue.forEach((cb) => cb(token))
  pendingQueue = []
}

const clearAuthStorage = () => {
  sessionStorage.removeItem('admin_token')
  sessionStorage.removeItem('admin_refresh_token')
  sessionStorage.removeItem('admin_user')
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_refresh_token')
  localStorage.removeItem('admin_user')
}

const refreshAccessToken = async (): Promise<string | null> => {
  const refreshToken = sessionStorage.getItem('admin_refresh_token') || localStorage.getItem('admin_refresh_token')
  if (!refreshToken) return null

  try {
    const resp = await axios.post(
      `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
      { refreshToken },
      { timeout: 10000 }
    )
    const payload = resp.data?.code === 0 ? resp.data.data : resp.data
    if (!payload?.accessToken || !payload?.refreshToken) {
      return null
    }

    sessionStorage.setItem('admin_token', payload.accessToken)
    sessionStorage.setItem('admin_refresh_token', payload.refreshToken)
    return payload.accessToken
  } catch {
    return null
  }
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = sessionStorage.getItem('admin_token') || localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // 清理空字符串的 query 参数，避免后端校验报错
    if (config.params) {
      const cleaned: Record<string, any> = {}
      for (const [key, value] of Object.entries(config.params)) {
        if (value !== '' && value !== undefined && value !== null) {
          cleaned[key] = value
        }
      }
      config.params = cleaned
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 后端 TransformInterceptor 包装格式: { code, message, data, timestamp }
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code === 0) {
        return res.data // 自动解包，直接返回内层 data
      }
      // 业务错误
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    const { response } = error
    if (response) {
      switch (response.status) {
        case 401:
          {
            const originalRequest = error.config || {}
            if (originalRequest._retry) {
              clearAuthStorage()
              router.push('/login')
              ElMessage.error('登录已过期，请重新登录')
              break
            }

            if (isRefreshing) {
              return new Promise((resolve, reject) => {
                pendingQueue.push((token) => {
                  if (!token) {
                    reject(error)
                    return
                  }
                  originalRequest.headers = originalRequest.headers || {}
                  originalRequest.headers.Authorization = `Bearer ${token}`
                  resolve(request(originalRequest))
                })
              })
            }

            isRefreshing = true
            originalRequest._retry = true
            return refreshAccessToken()
              .then((newToken) => {
                isRefreshing = false
                flushPendingQueue(newToken)
                if (!newToken) {
                  clearAuthStorage()
                  router.push('/login')
                  ElMessage.error('登录已过期，请重新登录')
                  return Promise.reject(error)
                }
                originalRequest.headers = originalRequest.headers || {}
                originalRequest.headers.Authorization = `Bearer ${newToken}`
                return request(originalRequest)
              })
              .catch((e) => {
                isRefreshing = false
                flushPendingQueue(null)
                clearAuthStorage()
                router.push('/login')
                ElMessage.error('登录已过期，请重新登录')
                return Promise.reject(e)
              })
          }
        case 403:
          ElMessage.error('没有权限执行此操作')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break
        default:
          ElMessage.error(response.data?.message || '请求失败')
      }
    } else {
      ElMessage.error('网络连接失败')
    }
    return Promise.reject(error)
  }
)

export default request
