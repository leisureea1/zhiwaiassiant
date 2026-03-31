<template>
  <div class="email-broadcast-page">
    <el-card shadow="never" class="form-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">邮件群发</span>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="邮件主题" prop="subject">
          <el-input
            v-model="form.subject"
            placeholder="请输入邮件主题"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="发送对象" prop="target">
          <el-radio-group v-model="form.target">
            <el-radio value="active">仅正常用户</el-radio>
            <el-radio value="inactive">未激活/封禁用户</el-radio>
            <el-radio value="all">全部用户</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="角色筛选" prop="role">
          <el-select v-model="form.role" placeholder="全部角色（不筛选）" clearable style="width: 200px">
            <el-option label="全部角色" value="" />
            <el-option label="普通用户" value="USER" />
            <el-option label="管理员" value="ADMIN" />
            <el-option label="超级管理员" value="SUPER_ADMIN" />
          </el-select>
        </el-form-item>

        <el-form-item label="邮件内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="10"
            placeholder="请输入邮件正文内容，支持HTML富文本。如：&lt;p&gt;尊敬的会员，您好！&lt;/p&gt;&lt;p&gt;这是一封来自知外助手的通知邮件...&lt;/p&gt;"
            maxlength="10000"
            show-word-limit
          />
          <div class="form-tip">
            支持富文本HTML标签，如 &lt;p&gt;、&lt;strong&gt;、&lt;br&gt; 等。邮件将自动带上统一的头部和尾部样式。
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="sending" @click="handleSend">
            {{ sending ? `发送中 (${sentCount}/${totalCount})` : '开始发送' }}
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 发送结果 -->
      <div v-if="result !== null" class="result-box">
        <el-alert
          :type="result.failed === 0 ? 'success' : 'warning'"
          :title="`发送完成：成功 ${result.success} 封，失败 ${result.failed} 封，跳过 ${result.skipped} 封（无邮箱），共 ${result.total} 位用户`"
          :closable="false"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { emailApi } from '@/api'

const formRef = ref()
const sending = ref(false)
const sentCount = ref(0)
const totalCount = ref(0)
const result = ref<{ total: number; success: number; failed: number; skipped: number } | null>(null)

const form = reactive({
  subject: '',
  content: '',
  target: 'active',
  role: '',
})

const rules = {
  subject: [{ required: true, message: '请输入邮件主题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入邮件内容', trigger: 'blur' }],
  target: [{ required: true, message: '请选择发送对象', trigger: 'change' }],
}

const handleSend = async () => {
  await formRef.value?.validate()

  if (!form.content.trim()) {
    ElMessage.warning('请输入邮件内容')
    return
  }

  try {
    sending.value = true
    result.value = null
    const res: any = await emailApi.broadcast({
      subject: form.subject,
      content: form.content,
      target: form.target,
      role: form.role,
    })
    result.value = res
    ElMessage.success('邮件发送完成')
  } catch (e: any) {
    ElMessage.error(e.message || '发送失败')
  } finally {
    sending.value = false
  }
}

const handleReset = () => {
  form.subject = ''
  form.content = ''
  form.target = 'active'
  form.role = ''
  result.value = null
  formRef.value?.resetFields()
}
</script>

<style lang="scss" scoped>
.email-broadcast-page {
  max-width: 800px;
}

.form-card {
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

.form-tip {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 6px;
  line-height: 1.5;
}

.result-box {
  margin-top: 20px;
}
</style>
