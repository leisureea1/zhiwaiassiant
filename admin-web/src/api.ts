const BASE = import.meta.env.VITE_API_BASE || 'http://localhost:3300/api/v1'

let token: string | null = localStorage.getItem('admin_token')

export const setToken = (t: string | null) => {
  token = t
  if (t) localStorage.setItem('admin_token', t)
  else localStorage.removeItem('admin_token')
}

export const getToken = () => token

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, { ...options, headers: { ...headers, ...options?.headers } })

  if (res.status === 401) {
    setToken(null)
    window.location.href = '/login'
    throw new Error('未登录')
  }

  const json = await res.json()
  if (!res.ok) throw new Error(json.message || `请求失败 (${res.status})`)
  return json.data ?? json
}

export const api = {
  // Auth
  login: (identifier: string, password: string) =>
    request<{ accessToken: string; refreshToken: string; user: UserInfo }>('/auth/login', {
      method: 'POST', body: JSON.stringify({ identifier, password }),
    }),

  // Dashboard
  getStats: () => request<DashboardStats>('/admin/dashboard/stats'),
  getPendingItems: () => request<PendingItems>('/admin/dashboard/pending-items'),

  // Users
  getUsers: (params?: Record<string, string>) =>
    request<Paginated<UserItem>>(`/users?${new URLSearchParams(params)}`),
  adminUpdateUser: (id: string, data: { role?: string; status?: string }) =>
    request(`/admin/users/${id}/admin`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteUser: (id: string) =>
    request(`/admin/users/${id}`, { method: 'DELETE' }),

  // Announcements
  getAnnouncements: (params?: Record<string, string>) =>
    request<Paginated<AnnouncementItem>>(`/announcements?${new URLSearchParams(params)}`),
  createAnnouncement: (data: { title: string; content: string; type?: string }) =>
    request('/admin/announcements', { method: 'POST', body: JSON.stringify(data) }),
  updateAnnouncement: (id: string, data: { title?: string; content?: string; type?: string }) =>
    request(`/admin/announcements/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteAnnouncement: (id: string) =>
    request(`/admin/announcements/${id}`, { method: 'DELETE' }),
  publishAnnouncement: (id: string) =>
    request(`/admin/announcements/${id}/publish`, { method: 'POST' }),
  pinAnnouncement: (id: string) =>
    request(`/admin/announcements/${id}/pin`, { method: 'POST' }),

  // System Logs
  getLogs: (params?: Record<string, string>) =>
    request<Paginated<LogItem>>(`/admin/system-logs?${new URLSearchParams(params)}`),
  getLogActionTypes: () => request<string[]>('/admin/system-logs/action-types'),
  getLogStats: () => request<LogStats>('/admin/system-logs/stats'),

  // Feature Flags
  getFeatureFlags: () => request<FeatureFlagItem[]>('/admin/features'),
  updateFeatureFlag: (name: string, isEnabled: boolean) =>
    request(`/admin/features/${name}`, { method: 'POST', body: JSON.stringify({ isEnabled }) }),

  // Grade Subscription
  triggerGradeCheck: () =>
    request('/grade-subscription/trigger', { method: 'POST' }),

  // Bulk Email
  sendBulkEmail: (data: { subject: string; content: string; target?: string; role?: string }) =>
    request<{ total: number; success: number; failed: number; skipped: number }>(
      '/admin/email/broadcast', { method: 'POST', body: JSON.stringify(data) }
    ),

  // Config (Super Admin)
  getConfig: () => request<ConfigResp>('/admin/config'),
  updateConfig: (configs: Record<string, string>) =>
    request<{ updated: string[]; message: string }>('/admin/config', { method: 'POST', body: JSON.stringify({ configs }) }),

  // Upload
  uploadFile: async (file: File): Promise<string> => {
    const form = new FormData()
    form.append('file', file)
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`
    const res = await fetch(`${BASE}/upload`, { method: 'POST', headers, body: form })
    const json = await res.json()
    if (!res.ok) throw new Error(json.message || '上传失败')
    return json.data?.fileURL ?? json.data?.url ?? json.url
  },
}

// Types matching backend JSON responses
export interface UserInfo {
  id: string; username: string; email: string | null; phone: string | null;
  studentId: string | null; realName: string | null; role: string;
}

export interface UserItem {
  id: string; username: string; email: string | null; realName: string | null;
  role: string; status: string; studentId: string | null;
  phone: string | null; avatar: string | null;
  college: string | null; major: string | null; className: string | null;
  jwxtBound: boolean;
  lastLoginAt: string | null; lastLoginIp: string | null;
  createdAt: string; updatedAt: string;
}

export interface AnnouncementItem {
  ID: string; Title: string; Content: string; Summary?: string;
  Type: string; Status: string; IsPinned: boolean;
  PublishedAt: string | null; CreatedAt: string; UpdatedAt: string;
}

export interface LogItem {
  ID: string; Level: string; Action: string; Module: string;
  Message: string; UserID?: string; CreatedAt: string;
}

export interface FeatureFlagItem {
  ID: string; Name: string; Description: string | null; IsEnabled: boolean;
}

export interface DashboardStats {
  totalUsers: number; activeUsers: number; totalAnnouncements: number; todayLogins: number;
}

export interface PendingItems {
  draftAnnouncements: number; inactiveUsers: number;
}

export interface ConfigResp {
  configs: Record<string, string>; groups: any[]; editableKeys: string[];
}

export interface LogStats {
  total: number; byLevel: Record<string, number>; recentActions: { action: string; count: number }[];
}

export interface Paginated<T> {
  data: T[]; total: number; page: number; size: number;
}
