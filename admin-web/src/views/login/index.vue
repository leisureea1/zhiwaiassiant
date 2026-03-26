<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-left">
        <div class="brand">
          <el-icon :size="48" color="#fff"><School /></el-icon>
          <h1>知外助手</h1>
          <p>管理后台 · 智慧校园服务平台</p>
        </div>
        <div class="features">
          <div class="feature-item">
            <el-icon :size="20"><User /></el-icon>
            <span>用户管理</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20"><Bell /></el-icon>
            <span>公告推送</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20"><Odometer /></el-icon>
            <span>数据监控</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20"><Setting /></el-icon>
            <span>系统配置</span>
          </div>
        </div>
      </div>

      <div class="login-right">
        <div class="login-form-wrapper">
          <h2>管理员登录</h2>
          <p class="subtitle">请使用管理员账号登录后台</p>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            size="large"
            @keyup.enter="handleLogin"
          >
            <el-form-item prop="studentId">
              <el-input
                v-model="form.studentId"
                placeholder="请输入学号"
                :prefix-icon="User"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="请输入密码"
                show-password
                :prefix-icon="Lock"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                class="login-btn"
                @click="handleLogin"
              >
                {{ loading ? '登录中...' : '登 录' }}
              </el-button>
            </el-form-item>
          </el-form>

          <div class="login-footer">
            <span>© 2026 知外工作室 · 西安外国语大学</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  studentId: '',
  password: '',
})

const rules: FormRules = {
  studentId: [{ required: true, message: '请输入学号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const handleLogin = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.studentId, form.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}

.login-container {
  display: flex;
  width: 900px;
  min-height: 520px;
  background: #fff;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.login-left {
  flex: 1;
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #fff;
}

.brand {
  margin-bottom: 48px;

  h1 {
    font-size: 32px;
    font-weight: 700;
    margin: 16px 0 8px;
  }

  p {
    font-size: 14px;
    opacity: 0.8;
  }
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  font-size: 14px;
  backdrop-filter: blur(10px);
}

.login-right {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.login-form-wrapper {
  width: 100%;
  max-width: 340px;

  h2 {
    font-size: 24px;
    font-weight: 700;
    color: #1e293b;
    margin-bottom: 8px;
  }

  .subtitle {
    color: #64748b;
    font-size: 14px;
    margin-bottom: 32px;
  }
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 10px;
}

.login-footer {
  text-align: center;
  margin-top: 32px;
  color: #94a3b8;
  font-size: 12px;
}
</style>
