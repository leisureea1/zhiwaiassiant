<template>
  <div class="users-page">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="filter-card">
      <el-form :inline="true" :model="filters" @submit.prevent="fetchUsers">
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="用户名 / 学号 / 姓名"
            clearable
            style="width: 200px"
            @clear="fetchUsers"
          />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="filters.role" placeholder="全部" clearable style="width: 130px">
            <el-option label="普通用户" value="USER" />
            <el-option label="管理员" value="ADMIN" />
            <el-option label="超级管理员" value="SUPER_ADMIN" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px">
            <el-option label="正常" value="ACTIVE" />
            <el-option label="未激活" value="INACTIVE" />
            <el-option label="已封禁" value="BANNED" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchUsers">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 用户表格 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">用户列表</span>
          <span class="card-total">共 {{ total }} 位用户</span>
        </div>
      </template>

      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column label="用户" min-width="200">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="36" :src="row.avatar">
                {{ (row.realName || row.username)?.[0] }}
              </el-avatar>
              <div class="user-cell-info">
                <span class="user-cell-name">{{ row.realName || row.username }}</span>
                <span class="user-cell-id">{{ row.studentId || row.username }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
        <el-table-column prop="college" label="院系" min-width="140" show-overflow-tooltip />
        <el-table-column label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.role === 'SUPER_ADMIN' ? 'danger' : row.role === 'ADMIN' ? 'warning' : 'info'"
              size="small"
            >
              {{ roleMap[row.role] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'ACTIVE' ? 'success' : row.status === 'BANNED' ? 'danger' : 'warning'"
              size="small"
              effect="light"
            >
              {{ statusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" width="170" align="center">
          <template #default="{ row }">
            {{ row.lastLoginAt ? dayjs(row.lastLoginAt).format('MM-DD HH:mm') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="170" align="center">
          <template #default="{ row }">
            {{ dayjs(row.createdAt).format('YYYY-MM-DD HH:mm') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-if="row.status === 'ACTIVE'"
                    command="ban"
                    :icon="CircleClose"
                  >封禁</el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.status === 'BANNED'"
                    command="unban"
                    :icon="CircleCheck"
                  >解封</el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.role === 'USER' && authStore.isSuperAdmin"
                    command="promote"
                    :icon="Promotion"
                  >设为管理员</el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.role === 'ADMIN' && authStore.isSuperAdmin"
                    command="demote"
                    :icon="Bottom"
                  >取消管理员</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          background
          @change="fetchUsers"
        />
      </div>
    </el-card>

    <!-- 用户详情抽屉 -->
    <el-drawer v-model="detailVisible" title="用户详情" size="420px">
      <template v-if="selectedUser">
        <div class="detail-header">
          <el-avatar :size="64" :src="selectedUser.avatar">
            {{ (selectedUser.realName || selectedUser.username)?.[0] }}
          </el-avatar>
          <div class="detail-name">
            <h3>{{ selectedUser.realName || selectedUser.username }}</h3>
            <el-tag
              :type="selectedUser.role === 'SUPER_ADMIN' ? 'danger' : selectedUser.role === 'ADMIN' ? 'warning' : 'info'"
              size="small"
            >{{ roleMap[selectedUser.role] }}</el-tag>
          </div>
        </div>
        <el-descriptions :column="1" border class="detail-desc">
          <el-descriptions-item label="用户ID">{{ selectedUser.id }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ selectedUser.username }}</el-descriptions-item>
          <el-descriptions-item label="学号">{{ selectedUser.studentId || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ selectedUser.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机">{{ selectedUser.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="院系">{{ selectedUser.college || '-' }}</el-descriptions-item>
          <el-descriptions-item label="专业">{{ selectedUser.major || '-' }}</el-descriptions-item>
          <el-descriptions-item label="班级">{{ selectedUser.className || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag
              :type="selectedUser.status === 'ACTIVE' ? 'success' : selectedUser.status === 'BANNED' ? 'danger' : 'warning'"
              size="small"
            >{{ statusMap[selectedUser.status] }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">
            {{ dayjs(selectedUser.createdAt).format('YYYY-MM-DD HH:mm:ss') }}
          </el-descriptions-item>
          <el-descriptions-item label="最后登录">
            {{ selectedUser.lastLoginAt ? dayjs(selectedUser.lastLoginAt).format('YYYY-MM-DD HH:mm:ss') : '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleClose, CircleCheck, Promotion, Bottom } from '@element-plus/icons-vue'
import { usersApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import dayjs from 'dayjs'

const authStore = useAuthStore()

const roleMap: Record<string, string> = {
  USER: '普通用户',
  ADMIN: '管理员',
  SUPER_ADMIN: '超级管理员',
}

const statusMap: Record<string, string> = {
  ACTIVE: '正常',
  INACTIVE: '未激活',
  BANNED: '已封禁',
}

const loading = ref(false)
const users = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filters = reactive({
  keyword: '',
  role: '',
  status: '',
})

const normalizeUser = (raw: any) => ({
  id: raw?.id ?? raw?.ID,
  username: raw?.username ?? raw?.Username,
  email: raw?.email ?? raw?.Email,
  phone: raw?.phone ?? raw?.Phone,
  studentId: raw?.studentId ?? raw?.StudentID,
  realName: raw?.realName ?? raw?.RealName,
  nickname: raw?.nickname ?? raw?.Nickname,
  avatar: raw?.avatar ?? raw?.Avatar,
  college: raw?.college ?? raw?.College,
  major: raw?.major ?? raw?.Major,
  className: raw?.className ?? raw?.ClassName,
  role: raw?.role ?? raw?.Role,
  status: raw?.status ?? raw?.Status,
  lastLoginAt: raw?.lastLoginAt ?? raw?.LastLoginAt,
  createdAt: raw?.createdAt ?? raw?.CreatedAt,
  updatedAt: raw?.updatedAt ?? raw?.UpdatedAt,
})

const detailVisible = ref(false)
const selectedUser = ref<any>(null)

const fetchUsers = async () => {
  loading.value = true
  try {
    const res: any = await usersApi.getList({
      search: filters.keyword,
      role: filters.role,
      status: filters.status,
      page: page.value,
      pageSize: pageSize.value,
    })
    const rawUsers = res.items || res.data || res
    users.value = Array.isArray(rawUsers) ? rawUsers.map(normalizeUser) : []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.keyword = ''
  filters.role = ''
  filters.status = ''
  page.value = 1
  fetchUsers()
}

const openDetail = (user: any) => {
  selectedUser.value = user
  detailVisible.value = true
}

const handleAction = async (command: string, user: any) => {
  const actionMap: Record<string, { confirmMsg: string; data: Record<string, any>; successMsg: string }> = {
    ban: { confirmMsg: `确定要封禁用户「${user.realName || user.username}」吗？`, data: { status: 'BANNED' }, successMsg: '用户已封禁' },
    unban: { confirmMsg: `确定要解封用户「${user.realName || user.username}」吗？`, data: { status: 'ACTIVE' }, successMsg: '用户已解封' },
    promote: { confirmMsg: `确定将「${user.realName || user.username}」设为管理员吗？`, data: { role: 'ADMIN' }, successMsg: '已设为管理员' },
    demote: { confirmMsg: `确定取消「${user.realName || user.username}」的管理员权限吗？`, data: { role: 'USER' }, successMsg: '已取消管理员' },
  }

  const action = actionMap[command]
  if (!action) return

  try {
    await ElMessageBox.confirm(action.confirmMsg, '操作确认', { type: 'warning' })
    await usersApi.adminUpdate(user.id, action.data)
    ElMessage.success(action.successMsg)
    fetchUsers()
  } catch {
    // 取消操作
  }
}

onMounted(fetchUsers)
</script>

<style lang="scss" scoped>
.users-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;

  .el-form-item {
    margin-bottom: 0;
  }
}

.table-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.card-total {
  font-size: 13px;
  color: #94a3b8;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-cell-info {
  display: flex;
  flex-direction: column;
}

.user-cell-name {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
}

.user-cell-id {
  font-size: 12px;
  color: #94a3b8;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;

  h3 {
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
    margin-bottom: 4px;
  }
}

.detail-desc {
  margin-top: 16px;
}
</style>
