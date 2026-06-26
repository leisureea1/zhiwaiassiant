import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { setToken } from '../api'
import {
  LayoutDashboard, Users, Megaphone, ScrollText, Settings, Mail, LogOut, Menu, X, ChevronRight,
} from 'lucide-react'

const menuItems = [
  { path: '/', label: '仪表盘', icon: LayoutDashboard },
  { path: '/users', label: '用户管理', icon: Users },
  { path: '/announcements', label: '公告管理', icon: Megaphone },
  { path: '/logs', label: '系统日志', icon: ScrollText },
  { path: '/email', label: '群发邮件', icon: Mail },
  { path: '/settings', label: '系统设置', icon: Settings },
]

export default function Layout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const handleLogout = () => { setToken(null); navigate('/login') }

  const sidebar = (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-3 px-4 h-16 border-b border-gray-200">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-accent flex items-center justify-center">
          <span className="text-white font-bold text-sm">知</span>
        </div>
        {!collapsed && <span className="font-bold text-lg text-gray-900">知外助手</span>}
      </div>
      <nav className="flex-1 py-4 space-y-1 px-3">
        {menuItems.map(item => {
          const isActive = location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path))
          return (
            <button
              key={item.path}
              onClick={() => { navigate(item.path); setMobileOpen(false) }}
              className={`w-full sidebar-link ${isActive ? 'sidebar-link-active' : 'sidebar-link-inactive'}`}
              title={collapsed ? item.label : undefined}
            >
              <item.icon size={20} />
              {!collapsed && <span>{item.label}</span>}
            </button>
          )
        })}
      </nav>
      <div className="p-3 border-t border-gray-200">
        <button onClick={handleLogout} className="w-full sidebar-link sidebar-link-inactive text-red-500 hover:bg-red-50 hover:text-red-600">
          <LogOut size={20} />
          {!collapsed && <span>退出登录</span>}
        </button>
      </div>
    </div>
  )

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Mobile overlay */}
      {mobileOpen && <div className="fixed inset-0 bg-black/50 z-40 lg:hidden" onClick={() => setMobileOpen(false)} />}
      {/* Sidebar */}
      <aside className={`fixed lg:static inset-y-0 left-0 z-50 bg-white border-r border-gray-200 transition-all duration-300 ${mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'} ${collapsed ? 'w-16' : 'w-64'}`}>
        {sidebar}
      </aside>
      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-4 lg:px-6">
          <button className="lg:hidden p-2 rounded-lg hover:bg-gray-100" onClick={() => setMobileOpen(true)}>
            <Menu size={20} />
          </button>
          <button className="hidden lg:flex p-2 rounded-lg hover:bg-gray-100" onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? <ChevronRight size={20} /> : <Menu size={20} />}
          </button>
          <h1 className="text-sm font-medium text-gray-500 ml-auto">管理后台 v1.0</h1>
        </header>
        <main className="flex-1 overflow-auto p-4 lg:p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
