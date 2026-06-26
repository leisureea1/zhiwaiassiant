import { useState, useEffect } from 'react'
import { api, LogItem } from '../api'
import { Loader2, Filter, Clock } from 'lucide-react'

const levelMap: Record<string, string> = { INFO: 'badge-blue', WARN: 'badge-yellow', ERROR: 'badge-red', DEBUG: 'badge-green' }

export default function SystemLogs() {
  const [logs, setLogs] = useState<LogItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [level, setLevel] = useState('')
  const [action, setAction] = useState('')
  const [actions, setActions] = useState<string[]>([])
  const [loading, setLoading] = useState(true)

  const fetch = async (p = page) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p), pageSize: '30' }
      if (level) params.level = level
      if (action) params.action = action
      const res = await api.getLogs(params)
      setLogs(res.data ?? [])
      setTotal(res.total)
    } catch { } finally { setLoading(false) }
  }

  useEffect(() => { fetch(); api.getLogActionTypes().then(setActions).catch(() => {}) }, [page, level, action])

  const pageSize = 30
  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold text-gray-900">系统日志</h2>

      <div className="card">
        <div className="flex flex-wrap items-center gap-3">
          <Filter size={16} className="text-gray-400" />
          <select className="input-field w-32" value={level} onChange={e => { setLevel(e.target.value); setPage(1) }}>
            <option value="">全部级别</option>
            <option value="INFO">INFO</option>
            <option value="WARN">WARN</option>
            <option value="ERROR">ERROR</option>
          </select>
          <select className="input-field w-48" value={action} onChange={e => { setAction(e.target.value); setPage(1) }}>
            <option value="">全部操作</option>
            {actions.map(a => <option key={a} value={a}>{a}</option>)}
          </select>
        </div>
      </div>

      <div className="space-y-2">
        {loading ? (
          <div className="flex justify-center py-12"><Loader2 className="animate-spin text-primary" size={24} /></div>
        ) : logs.length === 0 ? (
          <div className="card text-center py-12 text-gray-400">暂无日志</div>
        ) : logs.map(log => (
          <div key={log.ID} className="card py-3 px-4 flex items-center gap-4 hover:bg-gray-50 transition-colors">
            <span className={`badge ${levelMap[log.Level] || 'badge-blue'} shrink-0`}>{log.Level}</span>
            <span className="badge badge-yellow shrink-0">{log.Action}</span>
            <span className="text-sm text-gray-600 flex-1 truncate">{log.Message}</span>
            <span className="text-xs text-gray-400 flex items-center gap-1 shrink-0"><Clock size={12} />{new Date(log.CreatedAt).toLocaleString('zh-CN')}</span>
          </div>
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="btn-secondary text-sm">上一页</button>
          <span className="text-sm text-gray-500">{page}/{totalPages} (共{total}条)</span>
          <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} className="btn-secondary text-sm">下一页</button>
        </div>
      )}
    </div>
  )
}
