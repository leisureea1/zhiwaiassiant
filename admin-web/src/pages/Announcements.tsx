import { useState, useEffect, useRef } from 'react'
import { api, AnnouncementItem } from '../api'
import { Loader2, Plus, Pencil, Trash2, Send, Pin, PinOff, Image, Loader, Eye, Code } from 'lucide-react'

function mdToHtml(md: string): string {
  // 简单 Markdown 转 HTML：图片 ![alt](url) → <img>
  return md
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" style="max-width:100%;margin:12px 0;border-radius:8px;" />')
    .replace(/\n/g, '<br/>')
}

export default function Announcements() {
  const [items, setItems] = useState<AnnouncementItem[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<Partial<AnnouncementItem> | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [preview, setPreview] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const fetch = async () => {
    setLoading(true)
    try {
      const res = await api.getAnnouncements({ page: '1', pageSize: '100' })
      setItems(res.data ?? [])
    } catch { } finally { setLoading(false) }
  }

  useEffect(() => { fetch() }, [])

  const handleSave = async () => {
    if (!editing?.Title || !editing?.Content) return
    if (editing.ID) await api.updateAnnouncement(editing.ID, { title: editing.Title, content: editing.Content })
    else await api.createAnnouncement({ title: editing.Title!, content: editing.Content! })
    setShowForm(false); setEditing(null); setPreview(false); fetch()
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除？')) return
    await api.deleteAnnouncement(id); fetch()
  }

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const url = await api.uploadFile(file)
      const imgTag = `![${file.name}](${url})`
      setEditing(prev => ({ ...prev, Content: (prev?.Content ?? '') + '\n' + imgTag + '\n' }))
    } catch (err: any) {
      alert(err.message || '图片上传失败')
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const statusBadge = (s: string) => s === 'PUBLISHED' ? 'badge-green' : 'badge-yellow'

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">公告管理</h2>
        <button onClick={() => { setEditing({}); setShowForm(true) }} className="btn-primary"><Plus size={16} />新建公告</button>
      </div>

      <div className="card p-0 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 bg-gray-50">
              {['标题', '状态', '置顶', '创建时间', '操作'].map(h => (
                <th key={h} className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {loading ? (
              <tr><td colSpan={5} className="text-center py-12"><Loader2 className="animate-spin mx-auto text-primary" size={24} /></td></tr>
            ) : items.length === 0 ? (
              <tr><td colSpan={5} className="text-center py-12 text-gray-400">暂无公告</td></tr>
            ) : items.map(a => (
              <tr key={a.ID} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm font-medium">{a.Title}</td>
                <td className="px-4 py-3"><span className={`badge ${statusBadge(a.Status)}`}>{a.Status === 'PUBLISHED' ? '已发布' : '草稿'}</span></td>
                <td className="px-4 py-3">{a.IsPinned ? <Pin size={16} className="text-primary" /> : <PinOff size={16} className="text-gray-300" />}</td>
                <td className="px-4 py-3 text-sm text-gray-500">{new Date(a.CreatedAt).toLocaleDateString('zh-CN')}</td>
                <td className="px-4 py-3 flex gap-1">
                  <button onClick={() => { setEditing(a); setShowForm(true) }} className="btn-secondary text-xs py-1 px-2"><Pencil size={14} /></button>
                  {a.Status !== 'PUBLISHED' && <button onClick={() => api.publishAnnouncement(a.ID).then(fetch)} className="btn-primary text-xs py-1 px-2"><Send size={14} /></button>}
                  <button onClick={() => api.pinAnnouncement(a.ID).then(fetch)} className="btn-secondary text-xs py-1 px-2">{a.IsPinned ? <PinOff size={14} /> : <Pin size={14} />}</button>
                  <button onClick={() => handleDelete(a.ID)} className="btn-danger text-xs py-1 px-2"><Trash2 size={14} /></button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => { setShowForm(false); setEditing(null); setPreview(false) }}>
          <div className="bg-white rounded-xl p-6 w-full max-w-4xl mx-4 max-h-[90vh] flex flex-col" onClick={e => e.stopPropagation()}>
            <h3 className="font-bold text-lg mb-4">{editing?.ID ? '编辑公告' : '新建公告'}</h3>
            <input className="input-field mb-4" placeholder="标题" value={editing?.Title ?? ''} onChange={e => setEditing({ ...editing, Title: e.target.value })} />

            <div className="flex items-center gap-2 mb-2">
              <span className="text-sm text-gray-500 flex-1">内容 (支持 Markdown 图片语法)</span>
              <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={handleImageUpload} />
              <button type="button" onClick={() => fileInputRef.current?.click()} disabled={uploading} className="btn-secondary text-sm py-1 px-3">
                {uploading ? <Loader size={14} className="animate-spin" /> : <Image size={14} />}
                <span>{uploading ? '上传中...' : '插入图片'}</span>
              </button>
              <button type="button" onClick={() => setPreview(p => !p)} className={`text-sm py-1 px-3 rounded-lg border flex items-center gap-1 ${preview ? 'bg-primary text-white border-primary' : 'border-gray-300 text-gray-600 hover:bg-gray-50'}`}>
                {preview ? <Code size={14} /> : <Eye size={14} />}
                <span>{preview ? '编辑' : '预览'}</span>
              </button>
            </div>

            {preview ? (
              <div
                className="flex-1 border border-gray-300 rounded-lg p-4 bg-gray-50 overflow-auto prose prose-sm max-w-none"
                style={{ minHeight: '300px' }}
                dangerouslySetInnerHTML={{ __html: mdToHtml(editing?.Content ?? '') }}
              />
            ) : (
              <textarea
                className="input-field flex-1 resize-none font-mono text-sm overflow-auto"
                style={{ minHeight: '300px' }}
                value={editing?.Content ?? ''}
                onChange={e => setEditing({ ...editing, Content: e.target.value })}
                placeholder="正文内容，支持 Markdown 图片: ![描述](图片地址)"
              />
            )}

            <div className="flex gap-2 justify-end mt-4">
              <button onClick={() => { setShowForm(false); setEditing(null); setPreview(false) }} className="btn-secondary">取消</button>
              <button onClick={handleSave} className="btn-primary">保存</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
