<template>
  <div class="settings-page">
    <!-- 系统配置 -->
    <el-card v-if="authStore.isSuperAdmin" shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <div>
            <span class="card-title">系统配置</span>
            <p class="card-desc">修改后端 .env 环境变量配置，部分配置需要重启服务后生效</p>
          </div>
          <div class="header-actions">
            <el-button @click="fetchConfig" :loading="loadingConfig">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
            <el-button type="primary" @click="handleSaveConfig" :loading="savingConfig">
              <el-icon><Check /></el-icon>保存配置
            </el-button>
          </div>
        </div>
      </template>
      <div v-loading="loadingConfig">
        <div v-for="group in configGroups" :key="group.label" class="config-group">
          <h4 class="config-group-title">{{ group.label }}</h4>
          <div class="config-items">
            <div v-for="key in group.keys" :key="key" class="config-item">
              <label class="config-label">{{ key }}</label>
              <el-input
                v-model="configForm[key]"
                :placeholder="key"
                :type="isSensitive(key) ? 'password' : 'text'"
                :show-password="isSensitive(key)"
                clearable
                size="default"
              />
            </div>
          </div>
        </div>
        <el-empty v-if="configGroups.length === 0 && !loadingConfig" description="无法获取配置信息" />
      </div>
    </el-card>

    <!-- 功能开关 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <div>
            <span class="card-title">功能开关</span>
            <p class="card-desc">控制应用内各功能模块的启用状态</p>
          </div>
          <el-button @click="fetchFeatures" :loading="loadingFeatures">
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
        </div>
      </template>
      <div class="feature-list" v-loading="loadingFeatures">
        <div class="feature-item" v-for="flag in features" :key="flag.id">
          <div class="feature-info">
            <span class="feature-name">{{ flag.name }}</span>
            <span class="feature-desc">{{ flag.description || '暂无描述' }}</span>
          </div>
          <el-switch
            :model-value="flag.isEnabled"
            @change="(val: string | number | boolean) => handleToggleFeature(flag.name, !!val)"
            :loading="togglingFeature === flag.name"
          />
        </div>
        <el-empty v-if="features.length === 0 && !loadingFeatures" description="暂无功能开关" />
      </div>
    </el-card>

    <!-- 管理员信息 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <div>
            <span class="card-title">账号信息</span>
            <p class="card-desc">当前登录的管理员账号信息</p>
          </div>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户名">{{ authStore.user?.username }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ authStore.user?.realName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ authStore.user?.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="角色">
          <el-tag :type="authStore.isSuperAdmin ? 'danger' : 'primary'" size="small">
            {{ authStore.isSuperAdmin ? '超级管理员' : '管理员' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <div class="password-section">
        <h4>修改密码</h4>
        <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="100px" style="max-width: 500px">
          <el-form-item label="当前密码" prop="oldPassword">
            <el-input v-model="pwdForm.oldPassword" type="password" show-password />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input v-model="pwdForm.newPassword" type="password" show-password />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input v-model="pwdForm.confirmPassword" type="password" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleChangePassword" :loading="changingPwd">修改密码</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Check, Refresh } from '@element-plus/icons-vue'
import { adminApi, authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// ========== 系统配置 ==========
const configForm = reactive<Record<string, string>>({})
const configGroups = ref<{ label: string; keys: string[] }[]>([])
const loadingConfig = ref(false)
const savingConfig = ref(false)

const sensitiveKeys = ['MAIL_PASSWORD', 'JWXT_SERVICE_API_KEY']

const isSensitive = (key: string) => sensitiveKeys.includes(key)

const fetchConfig = async () => {
  loadingConfig.value = true
  try {
    const res: any = await adminApi.getConfig()
    configGroups.value = res.groups || []
    // 填充表单
    for (const [key, value] of Object.entries(res.configs || {})) {
      configForm[key] = value as string
    }
  } catch {
    // 非超级管理员会 403
  }
  loadingConfig.value = false
}

const handleSaveConfig = async () => {
  try {
    await ElMessageBox.confirm(
      '修改配置后，部分设置需要重启后端服务才能生效。确认保存？',
      '保存配置',
      { type: 'warning', confirmButtonText: '确认保存', cancelButtonText: '取消' },
    )
  } catch {
    return
  }

  savingConfig.value = true
  try {
    const res: any = await adminApi.updateConfig(configForm)
    ElMessage.success(res.message || '配置已保存')
    fetchConfig()
  } catch {
    ElMessage.error('保存失败')
  }
  savingConfig.value = false
}

// ========== 功能开关 ==========
const features = ref<any[]>([])
const loadingFeatures = ref(false)
const togglingFeature = ref('')

const fetchFeatures = async () => {
  loadingFeatures.value = true
  try {
    features.value = (await adminApi.getFeatureFlags()) as any
  } catch {}
  loadingFeatures.value = false
}

const handleToggleFeature = async (name: string, isEnabled: boolean) => {
  togglingFeature.value = name
  try {
    await adminApi.updateFeatureFlag(name, isEnabled)
    ElMessage.success(`功能「${name}」已${isEnabled ? '启用' : '禁用'}`)
    fetchFeatures()
  } catch {
    ElMessage.error('操作失败')
  }
  togglingFeature.value = ''
}

// ========== 修改密码 ==========
const pwdFormRef = ref<FormInstance>()
const changingPwd = ref(false)
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const pwdRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== pwdForm.newPassword) {
          callback(new Error('两次密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

const handleChangePassword = async () => {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return

  changingPwd.value = true
  try {
    await authApi.changePassword(pwdForm.oldPassword, pwdForm.newPassword)
    ElMessage.success('密码已修改，请重新登录')
    authStore.logout()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '修改失败')
  }
  changingPwd.value = false
}

onMounted(() => {
  fetchFeatures()
  if (authStore.isSuperAdmin) {
    fetchConfig()
  }
})
</script>

<style lang="scss" scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 900px;
}

.section-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.card-desc {
  font-size: 13px;
  color: #94a3b8;
  margin-top: 4px;
}

.config-group {
  &:not(:first-child) {
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid #f1f5f9;
  }
}

.config-group-title {
  font-size: 14px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 16px;
  padding-left: 10px;
  border-left: 3px solid #3b82f6;
}

.config-items {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.config-label {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  letter-spacing: 0.02em;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.feature-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-radius: 10px;
  transition: background 0.2s;

  &:hover {
    background: #f8fafc;
  }
}

.feature-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.feature-name {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
}

.feature-desc {
  font-size: 12px;
  color: #94a3b8;
}

.password-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e2e8f0;

  h4 {
    font-size: 15px;
    font-weight: 600;
    color: #1e293b;
    margin-bottom: 16px;
  }
}
</style>
