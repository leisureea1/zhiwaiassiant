import { useState, useEffect } from 'react'
import { api, UserItem } from '../api'
import { Loader2, Search, UserCog, X, Eye } from 'lucide-react'

export default function Users() {
  const [users, setUsers] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<UserItem | null>(null)
  const [detailUser, setDetailUser] = useState<UserItem | null>(null)

  const fetchUsers = async (p = page, s = search) => {
    setLoading(true)
    try {
      const params: Record<string, string> = { page: String(p), pageSize: '20' }
      if (s) params.search = s
      const res = await api.getUsers(params)
      setUsers(res.data ?? [])
      setTotal(res.total)
    } catch { } finally { setLoading(false) }
  }

  useEffect(() => { fetchUsers() }, [page])

  const handleSearch = (e: React.FormEvent) => { e.preventDefault(); setPage(1); fetchUsers(1, search) }

  const roleBadge = (role: string) => {
    const map: Record<string, string> = { SUPER_ADMIN: 'badge-red', ADMIN: 'badge-blue', USER: 'badge-green' }
    return map[role] || 'badge-yellow'
  }

  const statusBadge = (status: string) => status === 'ACTIVE' ? 'badge-green' : 'badge-red'

  const pageSize = 20
  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">用户管理</h2>
        <form onSubmit={handleSearch} className="flex items-center gap-2">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input className="input-field pl-9 w-64" placeholder="搜索用户名/姓名/学号..." value={search} onChange={e => setSearch(e.target.value)} />
          </div>
          <button type="submit" className="btn-secondary">搜索</button>
        </form>
      </div>

      <div className="card p-0 overflow-auto">
        <table className="w-full min-w-[900px]">
          <thead>
            <tr className="border-b border-gray-200 bg-gray-50">
              {['用户名', '学号', '姓名', '学院', '角色', '状态', '注册时间', '操作'].map(h => (
                <th key={h} className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {loading ? (
              <tr><td colSpan={8} className="text-center py-12"><Loader2 className="animate-spin mx-auto text-primary" size={24} /></td></tr>
            ) : users.length === 0 ? (
              <tr><td colSpan={8} className="text-center py-12 text-gray-400">暂无数据</td></tr>
            ) : users.map(u => (
              <tr key={u.id} className="hover:bg-gray-50 transition-colors">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    {u.avatar && <img src={u.avatar} className="w-7 h-7 rounded-full object-cover" alt="" />}
                    <span className="text-sm font-medium text-gray-900">{u.username}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-sm text-gray-500">{u.studentId || '-'}</td>
                <td className="px-4 py-3 text-sm text-gray-700">{u.realName || '-'}</td>
                <td className="px-4 py-3 text-sm text-gray-500 max-w-[120px] truncate">{u.college || '-'}</td>
                <td className="px-4 py-3"><span className={`badge ${roleBadge(u.role)}`}>{u.role}</span></td>
                <td className="px-4 py-3"><span className={`badge ${statusBadge(u.status)}`}>{u.status === 'ACTIVE' ? '正常' : '禁用'}</span></td>
                <td className="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">{new Date(u.createdAt).toLocaleDateString('zh-CN')}</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1">
                    <button onClick={() => setDetailUser(u)} className="btn-secondary text-xs py-1 px-2" title="查看详情"><Eye size={14} /></button>
                    <button onClick={() => setEditing(u)} className="btn-primary text-xs py-1 px-2" title="编辑"><UserCog size={14} /></button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="btn-secondary text-sm">上一页</button>
          <span className="text-sm text-gray-500">第 {page}/{totalPages} 页 (共 {total} 条)</span>
          <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} className="btn-secondary text-sm">下一页</button>
        </div>
      )}

      {/* 编辑弹窗 */}
      {editing && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setEditing(null)}>
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4 space-y-4" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-lg">编辑用户: {editing.username}</h3>
              <button onClick={() => setEditing(null)} className="p-1 hover:bg-gray-100 rounded"><X size={18} /></button>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">角色</label>
              <select className="input-field" defaultValue={editing.role} id="edit-role">
                <option value="USER">USER</option>
                <option value="ADMIN">ADMIN</option>
                <option value="SUPER_ADMIN">SUPER_ADMIN</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select className="input-field" defaultValue={editing.status} id="edit-status">
                <option value="ACTIVE">正常</option>
                <option value="INACTIVE">禁用</option>
              </select>
            </div>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setEditing(null)} className="btn-secondary">取消</button>
              <button onClick={async () => {
                const role = (document.getElementById('edit-role') as HTMLSelectElement).value
                const status = (document.getElementById('edit-status') as HTMLSelectElement).value
                await api.adminUpdateUser(editing.id, { role, status })
                setEditing(null); fetchUsers()
              }} className="btn-primary">保存</button>
            </div>
          </div>
        </div>
      )}

      {/* 详情弹窗 */}
      {detailUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setDetailUser(null)}>
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4 max-h-[80vh] overflow-auto" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-5">
              <h3 className="font-bold text-lg">用户详情</h3>
              <button onClick={() => setDetailUser(null)} className="p-1 hover:bg-gray-100 rounded"><X size={18} /></button>
            </div>

            <div className="flex items-center gap-4 mb-5">
              {detailUser.avatar && <img src={detailUser.avatar} className="w-16 h-16 rounded-xl object-cover" alt="" />}
              <div>
                <p className="font-bold text-lg">{detailUser.realName || detailUser.username}</p>
                <p className="text-sm text-gray-500">@{detailUser.username}</p>
              </div>
            </div>

            <div className="space-y-3">
              {[
                ['学号', detailUser.studentId],
                ['邮箱', detailUser.email],
                ['手机号', detailUser.phone],
                ['学院', detailUser.college],
                ['专业', detailUser.major],
                ['班级', detailUser.className],
                ['角色', detailUser.role],
                ['状态', detailUser.status],
                ['教务绑定', detailUser.jwxtBound ? '已绑定' : '未绑定'],
                ['ID', detailUser.id],
                ['最后活跃', detailUser.lastLoginAt ? new Date(detailUser.lastLoginAt).toLocaleString('zh-CN') : '-'],
                ['活跃IP', detailUser.lastLoginIp],
                ['注册时间', new Date(detailUser.createdAt).toLocaleString('zh-CN')],
              ].map(([label, value]) => (
                <div key={label} className="flex justify-between py-1.5 border-b border-gray-50 last:border-0">
                  <span className="text-sm text-gray-500">{label}</span>
                  <span className="text-sm text-gray-900 text-right max-w-[60%] truncate">{value ?? '-'}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
