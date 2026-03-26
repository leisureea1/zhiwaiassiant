<template>
  <div class="dashboard">
    <!-- 欢迎栏 -->
    <div class="welcome-card">
      <div class="welcome-info">
        <h2>{{ greeting }}，{{ authStore.user?.realName || authStore.user?.username }} 👋</h2>
        <p>欢迎回到知外助手管理后台，以下是今日数据概览。</p>
      </div>
      <div class="welcome-date">{{ currentDate }}</div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6" v-for="stat in statCards" :key="stat.title">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-value">{{ stat.value }}</div>
              <div class="stat-label">{{ stat.title }}</div>
            </div>
            <div class="stat-icon" :style="{ backgroundColor: stat.bgColor }">
              <el-icon :size="24" :color="stat.color"><component :is="stat.icon" /></el-icon>
            </div>
          </div>
          <div class="stat-footer" v-if="stat.extra">
            <span class="stat-extra">{{ stat.extra }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 待处理项 + 快捷操作 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">待处理事项</span>
              <el-tag type="warning" size="small" v-if="pendingTotal > 0">{{ pendingTotal }} 项</el-tag>
            </div>
          </template>
          <div class="pending-list">
            <div
              class="pending-item"
              v-for="item in pendingItems"
              :key="item.label"
              @click="router.push(item.route)"
            >
              <div class="pending-left">
                <el-icon :color="item.color"><component :is="item.icon" /></el-icon>
                <span>{{ item.label }}</span>
              </div>
              <el-badge :value="item.count" :type="item.count > 0 ? 'danger' : 'info'" />
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">快捷操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <div
              class="action-item"
              v-for="action in quickActions"
              :key="action.label"
              @click="router.push(action.route)"
            >
              <div class="action-icon" :style="{ backgroundColor: action.bgColor }">
                <el-icon :size="22" :color="action.color"><component :is="action.icon" /></el-icon>
              </div>
              <span class="action-label">{{ action.label }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { adminApi } from '@/api'
import dayjs from 'dayjs'

const router = useRouter()
const authStore = useAuthStore()

const stats = ref<any>({
  users: { total: 0, active: 0, newToday: 0 },
  announcements: { total: 0, published: 0 },
  logs: { todayCount: 0 },
})
const pending = ref<any>({ draftAnnouncements: 0, inactiveUsers: 0 })

const normalizeDashboardStats = (raw: any) => ({
  users: {
    total: raw?.users?.total ?? raw?.totalUsers ?? 0,
    active: raw?.users?.active ?? raw?.activeUsers ?? 0,
    newToday: raw?.users?.newToday ?? raw?.todayLogins ?? 0,
  },
  announcements: {
    total: raw?.announcements?.total ?? raw?.totalAnnouncements ?? 0,
    published: raw?.announcements?.published ?? 0,
  },
  logs: {
    todayCount: raw?.logs?.todayCount ?? raw?.todayLogs ?? 0,
  },
})

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '凌晨好'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const currentDate = computed(() => dayjs().format('YYYY年M月D日 dddd'))

const statCards = computed(() => [
  {
    title: '用户总数',
    value: stats.value.users.total,
    icon: 'User',
    color: '#3b82f6',
    bgColor: 'rgba(59,130,246,0.1)',
    extra: `今日新增 ${stats.value.users.newToday}`,
  },
  {
    title: '活跃用户',
    value: stats.value.users.active,
    icon: 'UserFilled',
    color: '#10b981',
    bgColor: 'rgba(16,185,129,0.1)',
    extra: null,
  },
  {
    title: '公告数量',
    value: stats.value.announcements.total,
    icon: 'Bell',
    color: '#f59e0b',
    bgColor: 'rgba(245,158,11,0.1)',
    extra: `已发布 ${stats.value.announcements.published}`,
  },
  {
    title: '今日日志',
    value: stats.value.logs.todayCount,
    icon: 'Document',
    color: '#6366f1',
    bgColor: 'rgba(99,102,241,0.1)',
    extra: null,
  },
])

const pendingTotal = computed(
  () => pending.value.draftAnnouncements + pending.value.inactiveUsers
)

const pendingItems = computed(() => [
  {
    label: '待发布公告',
    count: pending.value.draftAnnouncements,
    icon: 'EditPen',
    color: '#f59e0b',
    route: '/announcements',
  },
  {
    label: '未激活用户',
    count: pending.value.inactiveUsers,
    icon: 'Warning',
    color: '#ef4444',
    route: '/users',
  },
])

const quickActions = [
  { label: '发布公告', icon: 'Plus', color: '#3b82f6', bgColor: 'rgba(59,130,246,0.1)', route: '/announcements/create' },
  { label: '用户管理', icon: 'User', color: '#10b981', bgColor: 'rgba(16,185,129,0.1)', route: '/users' },
  { label: '系统日志', icon: 'Document', color: '#6366f1', bgColor: 'rgba(99,102,241,0.1)', route: '/logs' },
  { label: '系统设置', icon: 'Setting', color: '#64748b', bgColor: 'rgba(100,116,139,0.1)', route: '/settings' },
]

const fetchData = async () => {
  try {
    const [statsRes, pendingRes] = await Promise.all([
      adminApi.getDashboardStats(),
      adminApi.getPendingItems(),
    ])
    stats.value = normalizeDashboardStats(statsRes)
    pending.value = pendingRes as any
  } catch (e) {
    console.error('Failed to fetch dashboard data:', e)
  }
}

onMounted(fetchData)
</script>

<style lang="scss" scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.welcome-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  border-radius: 16px;
  padding: 28px 32px;
  color: #fff;

  h2 {
    font-size: 22px;
    font-weight: 700;
    margin-bottom: 6px;
  }

  p {
    font-size: 14px;
    opacity: 0.85;
  }
}

.welcome-date {
  font-size: 14px;
  opacity: 0.7;
  white-space: nowrap;
}

.stat-row {
  .el-col {
    margin-bottom: 0;
  }
}

.stat-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;

  .stat-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    color: #1e293b;
  }

  .stat-label {
    font-size: 13px;
    color: #64748b;
    margin-top: 4px;
  }

  .stat-icon {
    width: 52px;
    height: 52px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stat-footer {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f1f5f9;
  }

  .stat-extra {
    font-size: 12px;
    color: #94a3b8;
  }
}

.section-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;

  .card-title {
    font-size: 16px;
    font-weight: 600;
    color: #1e293b;
  }
}

.pending-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pending-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #f8fafc;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.2s;

  &:hover {
    background: #f1f5f9;
  }
}

.pending-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #475569;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: #f1f5f9;
    transform: translateY(-2px);
  }
}

.action-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-label {
  font-size: 14px;
  font-weight: 500;
  color: #475569;
}
</style>
