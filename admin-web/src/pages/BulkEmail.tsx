import { useState } from 'react'
import { api } from '../api'
import { Mail, Send, Loader2 } from 'lucide-react'

export default function BulkEmail() {
  const [subject, setSubject] = useState('')
  const [content, setContent] = useState('')
  const [target, setTarget] = useState('active')
  const [role, setRole] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<{ total: number; success: number; failed: number; skipped: number } | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!subject || !content) return
    setLoading(true); setResult(null)
    try {
      const r = await api.sendBulkEmail({ subject, content, target, role: role || undefined })
      setResult(r)
    } catch (err: any) {
      alert(err.message || '发送失败')
    } finally { setLoading(false) }
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <h2 className="text-xl font-bold text-gray-900">群发邮件</h2>

      <form onSubmit={handleSubmit} className="card space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">邮件主题</label>
          <input className="input-field" value={subject} onChange={e => setSubject(e.target.value)} placeholder="请输入邮件主题" required maxLength={200} />
        </div>
        <div className="flex gap-3">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1">目标用户</label>
            <select className="input-field" value={target} onChange={e => setTarget(e.target.value)}>
              <option value="active">活跃用户</option>
              <option value="inactive">非活跃用户</option>
              <option value="all">全部用户</option>
            </select>
          </div>
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1">角色筛选 (可选)</label>
            <select className="input-field" value={role} onChange={e => setRole(e.target.value)}>
              <option value="">不限</option>
              <option value="USER">USER</option>
              <option value="ADMIN">ADMIN</option>
              <option value="SUPER_ADMIN">SUPER_ADMIN</option>
            </select>
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">邮件内容 (支持HTML)</label>
          <textarea className="input-field h-64 resize-none font-mono text-sm" value={content} onChange={e => setContent(e.target.value)} placeholder="<p>邮件正文内容...</p>" required />
        </div>
        <button type="submit" disabled={loading || !subject || !content} className="btn-primary">
          {loading ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />}
          {loading ? '发送中...' : '发送邮件'}
        </button>
      </form>

      {result && (
        <div className="card space-y-2">
          <h3 className="font-semibold text-gray-900 flex items-center gap-2"><Mail size={18} />发送结果</h3>
          <div className="grid grid-cols-4 gap-3 text-center">
            <div className="bg-gray-50 rounded-lg p-3"><p className="text-2xl font-bold">{result.total}</p><p className="text-xs text-gray-500 mt-1">总数</p></div>
            <div className="bg-green-50 rounded-lg p-3"><p className="text-2xl font-bold text-green-600">{result.success}</p><p className="text-xs text-gray-500 mt-1">成功</p></div>
            <div className="bg-red-50 rounded-lg p-3"><p className="text-2xl font-bold text-red-600">{result.failed}</p><p className="text-xs text-gray-500 mt-1">失败</p></div>
            <div className="bg-yellow-50 rounded-lg p-3"><p className="text-2xl font-bold text-yellow-600">{result.skipped}</p><p className="text-xs text-gray-500 mt-1">跳过</p></div>
          </div>
        </div>
      )}
    </div>
  )
}
