import { useState, useEffect } from 'react'
import { api, FeatureFlagItem, ConfigResp } from '../api'
import { Loader2, ToggleLeft, ToggleRight, Save, Play } from 'lucide-react'

export default function Settings() {
  const [flags, setFlags] = useState<FeatureFlagItem[]>([])
  const [configResp, setConfigResp] = useState<ConfigResp | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [triggering, setTriggering] = useState(false)

  const fetch = async () => {
    setLoading(true)
    try {
      const [f, c] = await Promise.all([api.getFeatureFlags().catch(() => [] as FeatureFlagItem[]), api.getConfig().catch(() => null as ConfigResp | null)])
      setFlags(f); setConfigResp(c)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetch() }, [])

  const toggleFlag = async (name: string, current: boolean) => {
    await api.updateFeatureFlag(name, !current)
    setFlags(prev => prev.map(f => f.Name === name ? { ...f, IsEnabled: !current } : f))
  }

  const handleTrigger = async () => {
    setTriggering(true)
    try { await api.triggerGradeCheck() } catch { } finally { setTriggering(false) }
  }

  const handleSaveConfig = async () => {
    if (!configResp) return
    setSaving(true)
    try {
      const res = await api.updateConfig(configResp.configs)
      alert(res.message)
    } catch { } finally { setSaving(false) }
  }

  if (loading) return <div className="flex justify-center py-12"><Loader2 className="animate-spin text-primary" size={24} /></div>

  return (
    <div className="space-y-6 max-w-3xl">
      <h2 className="text-xl font-bold text-gray-900">系统设置</h2>

      {/* Feature Flags */}
      <div className="card">
        <h3 className="font-semibold text-gray-900 mb-4">功能开关</h3>
        <div className="space-y-3">
          {flags.map(f => (
            <div key={f.Name} className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
              <div>
                <p className="text-sm font-medium text-gray-900">{f.Name}</p>
                {f.Description && <p className="text-xs text-gray-400 mt-0.5">{f.Description}</p>}
              </div>
              <button onClick={() => toggleFlag(f.Name, f.IsEnabled)} className="p-1 rounded-lg hover:bg-gray-100 transition-colors">
                {f.IsEnabled ? <ToggleRight size={32} className="text-primary" /> : <ToggleLeft size={32} className="text-gray-300" />}
              </button>
            </div>
          ))}
          {flags.length === 0 && <p className="text-sm text-gray-400">暂无功能开关</p>}
        </div>
      </div>

      {/* Config */}
      {configResp && (
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold text-gray-900">系统配置</h3>
            <button onClick={handleSaveConfig} disabled={saving} className="btn-primary text-sm py-1.5"><Save size={14} />{saving ? '保存中...' : '保存'}</button>
          </div>
          <div className="space-y-3">
            {Object.entries(configResp.configs).map(([key, value]) => (
              <div key={key}>
                <label className="block text-xs font-medium text-gray-500 mb-1">{key}</label>
                <input className="input-field" value={value} onChange={e => setConfigResp(prev => prev ? { ...prev, configs: { ...prev.configs, [key]: e.target.value } } : prev)} />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Grade Check Trigger */}
      <div className="card">
        <h3 className="font-semibold text-gray-900 mb-4">成绩检查触发</h3>
        <p className="text-sm text-gray-500 mb-3">手动触发一次成绩订阅检查周期</p>
        <button onClick={handleTrigger} disabled={triggering} className="btn-primary"><Play size={16} />{triggering ? '触发中...' : '立即触发检查'}</button>
      </div>
    </div>
  )
}
