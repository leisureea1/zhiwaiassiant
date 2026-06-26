import { useState } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { setToken, getToken } from './api'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Announcements from './pages/Announcements'
import SystemLogs from './pages/SystemLogs'
import Settings from './pages/Settings'
import BulkEmail from './pages/BulkEmail'

export default function App() {
  const [loggedIn, setLoggedIn] = useState(!!getToken())

  const handleLogin = (t: string) => { setToken(t); setLoggedIn(true) }

  if (!loggedIn) return <Login onLogin={handleLogin} />

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/users" element={<Users />} />
        <Route path="/announcements" element={<Announcements />} />
        <Route path="/logs" element={<SystemLogs />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/email" element={<BulkEmail />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  )
}
