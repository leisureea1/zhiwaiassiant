import { useState, useEffect } from 'react'
import { api, DashboardStats, PendingItems } from '../api'
import { Users, Megaphone, TrendingUp, UserCheck, FileText, UserX, Loader2 } from 'lucide-react'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [pending, setPending] = useState<PendingItems | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([api.getStats(), api.getPendingItems()]).then(([s, p]) => {
      setStats(s); setPending(p)
    }).finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="flex items-center justify-center h-64"><Loader2 className="animate-spin text-primary" size={32} /></div>

  const statCards = [
    { label: '总用户', value: stats?.totalUsers ?? 0, icon: Users, color: 'text-blue-500', bg: 'bg-blue-50' },
    { label: '活跃用户', value: stats?.activeUsers ?? 0, icon: UserCheck, color: 'text-green-500', bg: 'bg-green-50' },
    { label: '公告总数', value: stats?.totalAnnouncements ?? 0, icon: Megaphone, color: 'text-purple-500', bg: 'bg-purple-50' },
    { label: '今日活跃', value: stats?.todayLogins ?? 0, icon: TrendingUp, color: 'text-orange-500', bg: 'bg-orange-50' },
  ]

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold text-gray-900">仪表盘</h2>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map(c => (
          <div key={c.label} className="card flex items-center gap-4">
            <div className={`w-12 h-12 rounded-xl ${c.bg} flex items-center justify-center`}>
              <c.icon size={24} className={c.color} />
            </div>
            <div>
              <p className="text-sm text-gray-500">{c.label}</p>
              <p className="text-2xl font-bold text-gray-900">{c.value}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="card">
        <h3 className="font-semibold text-gray-900 mb-4">待处理事项</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
            <div className="flex items-center gap-3">
              <FileText size={20} className="text-yellow-600" />
              <span className="text-sm font-medium">草稿公告</span>
            </div>
            <span className="text-2xl font-bold text-yellow-700">{pending?.draftAnnouncements ?? 0}</span>
          </div>
          <div className="flex items-center justify-between p-4 bg-red-50 rounded-lg">
            <div className="flex items-center gap-3">
              <UserX size={20} className="text-red-600" />
              <span className="text-sm font-medium">非活跃用户</span>
            </div>
            <span className="text-2xl font-bold text-red-700">{pending?.inactiveUsers ?? 0}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
