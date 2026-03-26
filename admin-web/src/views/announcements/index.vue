<template>
  <div class="announcements-page">
    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-form :inline="true" :model="filters" @submit.prevent="fetchList">
          <el-form-item>
            <el-input v-model="filters.keyword" placeholder="搜索标题" clearable style="width: 200px" @clear="fetchList" />
          </el-form-item>
          <el-form-item>
            <el-select v-model="filters.type" placeholder="类型" clearable style="width: 120px">
              <el-option label="普通" value="NORMAL" />
              <el-option label="重要" value="IMPORTANT" />
              <el-option label="紧急" value="URGENT" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px">
              <el-option label="草稿" value="DRAFT" />
              <el-option label="已发布" value="PUBLISHED" />
              <el-option label="已归档" value="ARCHIVED" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchList"><el-icon><Search /></el-icon>搜索</el-button>
          </el-form-item>
        </el-form>
        <el-button type="primary" @click="router.push('/announcements/create')">
          <el-icon><Plus /></el-icon>发布公告
        </el-button>
      </div>
    </el-card>

    <!-- 公告表格 -->
    <el-card shadow="never" class="table-card">
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="标题" min-width="260">
          <template #default="{ row }">
            <div class="title-cell">
              <el-tag v-if="row.isPinned" type="danger" size="small" effect="dark">置顶</el-tag>
              <span class="title-text">{{ row.title }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.type === 'URGENT' ? 'danger' : row.type === 'IMPORTANT' ? 'warning' : 'info'"
              size="small"
            >{{ typeMap[row.type] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'PUBLISHED' ? 'success' : row.status === 'DRAFT' ? 'warning' : 'info'"
              size="small"
              effect="light"
            >{{ statusMap[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="作者" width="120" align="center">
          <template #default="{ row }">
            {{ row.author?.realName || row.author?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="170" align="center">
          <template #default="{ row }">
            {{ row.publishedAt ? dayjs(row.publishedAt).format('YYYY-MM-DD HH:mm') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="router.push(`/announcements/${row.id}/edit`)">编辑</el-button>
            <el-button
              v-if="row.status === 'DRAFT'"
              size="small"
              type="success"
              @click="handlePublish(row)"
            >发布</el-button>
            <el-button
              size="small"
              :type="row.isPinned ? 'warning' : 'default'"
              @click="handleTogglePin(row)"
            >{{ row.isPinned ? '取消置顶' : '置顶' }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          background
          @change="fetchList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { announcementsApi } from '@/api'
import dayjs from 'dayjs'

const router = useRouter()

const typeMap: Record<string, string> = { NORMAL: '普通', IMPORTANT: '重要', URGENT: '紧急' }
const statusMap: Record<string, string> = { DRAFT: '草稿', PUBLISHED: '已发布', ARCHIVED: '已归档' }

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', type: '', status: '' })

const fetchList = async () => {
  loading.value = true
  try {
    const res: any = await announcementsApi.getList({
      ...filters,
      page: page.value,
      pageSize: pageSize.value,
    })
    list.value = res.items || res.data || res
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handlePublish = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定发布「${row.title}」吗？`, '发布确认', { type: 'info' })
    await announcementsApi.publish(row.id)
    ElMessage.success('公告已发布')
    fetchList()
  } catch {}
}

const handleTogglePin = async (row: any) => {
  try {
    await announcementsApi.togglePin(row.id)
    ElMessage.success(row.isPinned ? '已取消置顶' : '已置顶')
    fetchList()
  } catch {}
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除「${row.title}」吗？此操作不可恢复。`, '删除确认', { type: 'warning' })
    await announcementsApi.delete(row.id)
    ElMessage.success('已删除')
    fetchList()
  } catch {}
}

onMounted(fetchList)
</script>

<style lang="scss" scoped>
.announcements-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .el-form-item {
    margin-bottom: 0;
  }
}

.table-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.title-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-text {
  font-weight: 500;
  color: #1e293b;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
