import request from '@/utils/request'

const normalizeAnnouncement = (raw: any) => ({
  id: raw?.id ?? raw?.ID,
  title: raw?.title ?? raw?.Title,
  content: raw?.content ?? raw?.Content,
  summary: raw?.summary ?? raw?.Summary,
  type: raw?.type ?? raw?.Type,
  status: raw?.status ?? raw?.Status,
  isPinned: raw?.isPinned ?? raw?.IsPinned ?? false,
  isPopup: raw?.isPopup ?? raw?.IsPopup ?? false,
  authorId: raw?.authorId ?? raw?.AuthorID,
  publishedAt: raw?.publishedAt ?? raw?.PublishedAt,
  expiresAt: raw?.expiresAt ?? raw?.ExpiresAt,
  createdAt: raw?.createdAt ?? raw?.CreatedAt,
  updatedAt: raw?.updatedAt ?? raw?.UpdatedAt,
  author: raw?.author ?? raw?.Author,
})

const normalizeAnnouncementList = (raw: any) => {
  const itemsRaw = Array.isArray(raw?.items)
    ? raw.items
    : Array.isArray(raw?.data)
      ? raw.data
      : Array.isArray(raw)
        ? raw
        : []

  return {
    ...raw,
    items: itemsRaw.map(normalizeAnnouncement),
    total: raw?.total ?? itemsRaw.length,
  }
}

// 认证相关
export const authApi = {
  login: (data: { studentId: string; password: string }) =>
    request.post('/auth/login', data),
  refreshToken: (refreshToken: string) =>
    request.post('/auth/refresh', { refreshToken }),
  logout: () => request.post('/auth/logout'),
  changePassword: (oldPassword: string, newPassword: string) =>
    request.post('/auth/change-password', { oldPassword, newPassword }),
}

// 管理后台
export const adminApi = {
  getDashboardStats: () => request.get('/admin/dashboard/stats'),
  getPendingItems: () => request.get('/admin/dashboard/pending-items'),
  getFeatureFlags: () => request.get('/admin/features'),
  updateFeatureFlag: (name: string, isEnabled: boolean) =>
    request.post(`/admin/features/${name}`, { isEnabled }),
  getConfig: () => request.get('/admin/config'),
  updateConfig: (configs: Record<string, string>) =>
    request.post('/admin/config', { configs }),
}

// 用户管理
export const usersApi = {
  getList: (params: Record<string, any>) => request.get('/users', { params }),
  getById: (id: string) => request.get(`/users/${id}`),
  getMe: () => request.get('/users/me'),
  adminUpdate: (id: string, data: Record<string, any>) =>
    request.put(`/users/${id}/admin`, data),
  delete: (id: string) => request.delete(`/users/${id}`),
}

// 公告管理
export const announcementsApi = {
  getList: async (params: Record<string, any>) => {
    const res = await request.get('/announcements', { params })
    return normalizeAnnouncementList(res)
  },
  getById: async (id: string) => {
    const res = await request.get(`/announcements/${id}`)
    return normalizeAnnouncement(res)
  },
  create: async (data: Record<string, any>) => {
    const res = await request.post('/admin/announcements', data)
    return normalizeAnnouncement(res)
  },
  update: async (id: string, data: Record<string, any>) => {
    const res = await request.put(`/admin/announcements/${id}`, data)
    return normalizeAnnouncement(res)
  },
  delete: (id: string) => request.delete(`/admin/announcements/${id}`),
  publish: async (id: string) => {
    const res = await request.post(`/admin/announcements/${id}/publish`)
    return normalizeAnnouncement(res)
  },
  togglePin: async (id: string) => {
    const res = await request.post(`/admin/announcements/${id}/pin`)
    return normalizeAnnouncement(res)
  },
}

// 系统日志
export const logsApi = {
  getList: (params: Record<string, any>) =>
    request.get('/admin/system-logs', { params }),
  getActionTypes: () => request.get('/admin/system-logs/action-types'),
  getStats: () => request.get('/admin/system-logs/stats'),
}

// 成绩订阅管理
export const gradeSubscriptionApi = {
  triggerCheck: () => request.post('/admin/grade-subscription/trigger'),
}

// 邮件群发
export const emailApi = {
  broadcast: (data: { subject: string; content: string; target: string; role: string }) =>
    request.post('/admin/email/broadcast', data),
}

// 文件上传
export const uploadApi = {
  uploadAttachment: (formData: FormData) =>
    request.post('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }),
}
