import { useState } from 'react'
import { api } from '../api'

export default function Login({ onLogin }: { onLogin: (token: string) => void }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.login(username, password)
      onLogin(data.accessToken)
    } catch (err: any) {
      setError(err.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 via-white to-accent/10">
      <div className="w-full max-w-md mx-4">
        <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8">
          <div className="text-center mb-8">
            <div className="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-primary to-accent flex items-center justify-center mb-4">
              <span className="text-white text-2xl font-bold">知</span>
            </div>
            <h1 className="text-2xl font-bold text-gray-900">知外助手</h1>
            <p className="text-gray-500 text-sm mt-1">管理后台登录</p>
          </div>
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">用户名/邮箱/手机号</label>
              <input className="input-field" value={username} onChange={e => setUsername(e.target.value)} placeholder="支持用户名、邮箱、手机号、学号" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">密码</label>
              <input className="input-field" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="请输入密码" required />
            </div>
            {error && <div className="bg-red-50 text-red-600 text-sm px-4 py-2.5 rounded-lg">{error}</div>}
            <button type="submit" disabled={loading} className="btn-primary w-full justify-center py-3 text-base">
              {loading ? '登录中...' : '登录'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
