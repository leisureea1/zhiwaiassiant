<template>
  <div class="logs-page">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="never" class="mini-stat">
          <div class="mini-stat-value">{{ logStats.todayLogs || 0 }}</div>
          <div class="mini-stat-label">今日日志</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="mini-stat">
          <div class="mini-stat-value">{{ logStats.weekLogs || 0 }}</div>
          <div class="mini-stat-label">本周日志</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="mini-stat">
          <div class="mini-stat-value text-warning">{{ errorCount }}</div>
          <div class="mini-stat-label">今日错误</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="mini-stat">
          <div class="mini-stat-value text-success">{{ infoCount }}</div>
          <div class="mini-stat-label">今日信息</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选 + 表格 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="filter-bar">
          <el-form :inline="true" :model="filters" @submit.prevent="fetchLogs">
            <el-form-item>
              <el-select v-model="filters.level" placeholder="日志级别" clearable style="width: 120px">
                <el-option label="DEBUG" value="DEBUG" />
                <el-option label="INFO" value="INFO" />
                <el-option label="WARN" value="WARN" />
                <el-option label="ERROR" value="ERROR" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-select v-model="filters.action" placeholder="操作类型" clearable style="width: 130px">
                <el-option v-for="t in actionTypes" :key="t" :label="t" :value="t" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-input v-model="filters.module" placeholder="模块" clearable style="width: 140px" />
            </el-form-item>
            <el-form-item>
              <el-date-picker
                v-model="dateRange"
                type="daterange"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
                @change="onDateChange"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchLogs"><el-icon><Search /></el-icon>搜索</el-button>
            </el-form-item>
          </el-form>
        </div>
      </template>

      <el-table :data="logs" v-loading="loading" stripe size="small">
        <el-table-column label="级别" width="90" align="center">
          <template #default="{ row }">
            <el-tag
              :type="levelType(row.level)"
              size="small"
              effect="dark"
            >{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" prop="action" />
        <el-table-column label="模块" width="120" prop="module" />
        <el-table-column label="消息" min-width="280" prop="message" show-overflow-tooltip />
        <el-table-column label="用户" width="120" align="center">
          <template #default="{ row }">
            {{ row.user?.realName || row.user?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="IP" width="130" prop="ipAddress" />
        <el-table-column label="耗时" width="80" align="center">
          <template #default="{ row }">
            {{ row.duration ? row.duration + 'ms' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="时间" width="170" align="center">
          <template #default="{ row }">
            {{ dayjs(row.createdAt).format('MM-DD HH:mm:ss') }}
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
          @change="fetchLogs"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { logsApi } from '@/api'
import dayjs from 'dayjs'

const loading = ref(false)
const logs = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const actionTypes = ref<string[]>([])
const logStats = ref<any>({})
const dateRange = ref<string[]>([])

const filters = reactive({
  level: '',
  action: '',
  module: '',
  startAt: '',
  endAt: '',
})

const toNumber = (value: unknown, fallback = 0) => {
  const n = Number(value)
  return Number.isFinite(n) ? n : fallback
}

const pickFirst = (obj: any, keys: string[], fallback?: unknown) => {
  for (const key of keys) {
    if (obj?.[key] !== undefined && obj?.[key] !== null) {
      return obj[key]
    }
  }
  return fallback
}

const normalizeLevelStats = (raw: any) => {
  const result: Record<string, number> = {}
  if (!raw) return result

  if (Array.isArray(raw)) {
    raw.forEach((item: any) => {
      const level = String(item?.level ?? item?.name ?? item?.key ?? '').toUpperCase()
      if (!level) return
      result[level] = toNumber(item?.count ?? item?.total ?? item?.value, 0)
    })
    return result
  }

  Object.entries(raw).forEach(([key, value]) => {
    result[key.toUpperCase()] = toNumber(value, 0)
  })
  return result
}

const normalizeStats = (raw: any) => {
  const source = raw?.data && typeof raw.data === 'object' && !Array.isArray(raw.data)
    ? raw.data
    : raw
  const byLevelRaw = pickFirst(source, ['byLevel', 'by_level', 'levelStats', 'levels'], {})

  return {
    todayLogs: toNumber(pickFirst(source, ['todayLogs', 'todayCount', 'today', 'TodayLogs'], 0), 0),
    weekLogs: toNumber(
      pickFirst(source, ['weekLogs', 'weekCount', 'thisWeekLogs', 'sevenDaysLogs', 'totalLogs', 'TotalLogs'], 0),
      0
    ),
    byLevel: normalizeLevelStats(byLevelRaw),
  }
}

const normalizeLogItem = (item: any) => {
  const details = item?.details ?? item?.Details ?? {}
  const duration = item?.duration ?? item?.Duration ?? details?.duration
  const rawUser = item?.user ?? item?.User
  const normalizedUser = rawUser
    ? {
        id: rawUser?.id ?? rawUser?.ID ?? item?.userId ?? item?.UserID,
        username: rawUser?.username ?? rawUser?.Username,
        realName: rawUser?.realName ?? rawUser?.RealName,
      }
    : undefined

  return {
    ...item,
    id: item?.id ?? item?.ID,
    level: item?.level ?? item?.Level,
    action: item?.action ?? item?.Action,
    module: item?.module ?? item?.Module,
    message: item?.message ?? item?.Message,
    ipAddress: item?.ipAddress ?? item?.IPAddress,
    duration,
    createdAt: item?.createdAt ?? item?.CreatedAt,
    user: normalizedUser,
  }
}

const normalizeLogsResult = (raw: any) => {
  const source = raw?.data && typeof raw.data === 'object' && !Array.isArray(raw.data)
    ? raw.data
    : raw
  const items = pickFirst(source, ['items', 'data', 'list', 'records', 'rows'], [])

  return {
    items: Array.isArray(items) ? items.map(normalizeLogItem) : [],
    total: toNumber(pickFirst(source, ['total', 'count', 'Total', 'Count'], 0), 0),
  }
}

const getLevelCount = (level: string) => {
  const byLevel = logStats.value?.byLevel
  if (!byLevel) return 0
  return byLevel[String(level).toUpperCase()] || 0
}

const errorCount = computed(() => {
  return getLevelCount('ERROR')
})

const infoCount = computed(() => {
  return getLevelCount('INFO')
})

const levelType = (level: string) => {
  const map: Record<string, string> = { DEBUG: 'info', INFO: 'success', WARN: 'warning', ERROR: 'danger' }
  return (map[level] || 'info') as any
}

const onDateChange = (val: string[] | null) => {
  filters.startAt = val?.[0] ? `${val[0]}T00:00:00+08:00` : ''
  filters.endAt = val?.[1] ? `${val[1]}T23:59:59+08:00` : ''
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const res: any = await logsApi.getList({
      ...filters,
      page: page.value,
      pageSize: pageSize.value,
    })
    const normalized = normalizeLogsResult(res)
    logs.value = normalized.items
    total.value = normalized.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchMeta = async () => {
  try {
    const [types, stats] = await Promise.all([logsApi.getActionTypes(), logsApi.getStats()])
    actionTypes.value = Array.isArray(types)
      ? types
      : (types as any)?.data || (types as any)?.items || []
    logStats.value = normalizeStats(stats)
  } catch {}
}

onMounted(() => {
  fetchLogs()
  fetchMeta()
})
</script>

<style lang="scss" scoped>
.logs-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.stats-row {
  .mini-stat {
    border-radius: 14px;
    border: 1px solid #e2e8f0;
    text-align: center;

    .mini-stat-value {
      font-size: 28px;
      font-weight: 700;
      color: #1e293b;

      &.text-warning { color: #f59e0b; }
      &.text-success { color: #10b981; }
    }

    .mini-stat-label {
      font-size: 13px;
      color: #94a3b8;
      margin-top: 4px;
    }
  }
}

.table-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.filter-bar {
  .el-form-item {
    margin-bottom: 0;
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
