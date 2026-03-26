<template>
  <div class="announcement-edit">
    <el-card shadow="never" class="edit-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">{{ isEdit ? '编辑公告' : '发布公告' }}</span>
          <el-button @click="router.back()">返回列表</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" size="large">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入公告标题" maxlength="100" show-word-limit />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width: 100%">
                <el-option label="普通" value="NORMAL" />
                <el-option label="重要" value="IMPORTANT" />
                <el-option label="紧急" value="URGENT" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="置顶">
              <el-switch v-model="form.isPinned" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="弹窗通知">
              <el-switch v-model="form.isPopup" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="摘要">
          <el-input v-model="form.summary" type="textarea" :rows="2" placeholder="可选，公告摘要" maxlength="200" show-word-limit />
        </el-form-item>

        <el-form-item label="正文" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="12" placeholder="请输入公告正文" />
        </el-form-item>

        <el-form-item label="正文图片">
          <div class="image-tools">
            <el-upload
              :show-file-list="false"
              :http-request="handleImageUpload"
              :before-upload="beforeImageUpload"
              accept="image/*"
            >
              <el-button :loading="uploadingImage">上传图片</el-button>
            </el-upload>
            <span class="image-tip">上传后会自动把图片链接插入正文末尾</span>
          </div>
          <div v-if="uploadedImages.length" class="image-preview-list">
            <el-image
              v-for="url in uploadedImages"
              :key="url"
              :src="url"
              fit="cover"
              class="image-preview"
              :preview-src-list="uploadedImages"
              preview-teleported
            />
          </div>
        </el-form-item>

        <el-form-item label="过期时间">
          <el-date-picker
            v-model="form.expiresAt"
            type="datetime"
            placeholder="可选，公告过期时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss.sssZ"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item>
          <div class="form-actions">
            <el-button @click="handleSaveDraft" :loading="saving">保存草稿</el-button>
            <el-button type="primary" @click="handlePublish" :loading="saving">
              {{ isEdit ? '保存并发布' : '立即发布' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules, type UploadRequestOptions } from 'element-plus'
import { announcementsApi, uploadApi } from '@/api'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)
const formRef = ref<FormInstance>()
const saving = ref(false)
const uploadingImage = ref(false)
const uploadedImages = ref<string[]>([])

const form = reactive({
  title: '',
  content: '',
  summary: '',
  type: 'NORMAL',
  isPinned: false,
  isPopup: false,
  expiresAt: '',
})

const rules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文内容', trigger: 'blur' }],
}

const imageRegex = /!\[[^\]]*\]\(([^)]+)\)|<img[^>]*src=["']([^"']+)["'][^>]*>/gi

const extractImageUrls = (text: string): string[] => {
  if (!text) return []
  const urls: string[] = []
  let matched: RegExpExecArray | null

  while ((matched = imageRegex.exec(text)) !== null) {
    const url = matched[1] || matched[2]
    if (url) {
      urls.push(url)
    }
  }

  return Array.from(new Set(urls))
}

const appendImageToContent = (url: string) => {
  const markdown = `![公告图片](${url})`
  form.content = form.content ? `${form.content}\n\n${markdown}` : markdown
  if (!uploadedImages.value.includes(url)) {
    uploadedImages.value.push(url)
  }
}

const beforeImageUpload = (file: File) => {
  const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowed.includes(file.type)) {
    ElMessage.error('仅支持 JPG/PNG/GIF/WebP 图片')
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

const handleImageUpload = async (options: UploadRequestOptions) => {
  uploadingImage.value = true
  try {
    const formData = new FormData()
    formData.append('file', options.file)
    const res: any = await uploadApi.uploadAttachment(formData)
    const url = res?.url

    if (!url) {
      throw new Error('上传成功但未返回图片地址')
    }

    appendImageToContent(url)
    options.onSuccess?.(res)
    ElMessage.success('图片上传成功，已插入正文')
  } catch (e: any) {
    options.onError?.(e)
    ElMessage.error(e?.message || '图片上传失败')
  } finally {
    uploadingImage.value = false
  }
}

const loadAnnouncement = async () => {
  if (!route.params.id) return
  try {
    const res: any = await announcementsApi.getById(route.params.id as string)
    Object.assign(form, {
      title: res.title,
      content: res.content,
      summary: res.summary || '',
      type: res.type,
      isPinned: res.isPinned,
      isPopup: res.isPopup,
      expiresAt: res.expiresAt || '',
    })
    uploadedImages.value = extractImageUrls(res.content || '')
  } catch {
    ElMessage.error('公告不存在')
    router.back()
  }
}

const submitForm = async (publish: boolean) => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const data = {
      ...form,
      expiresAt: form.expiresAt || undefined,
    }

    let announcementId = ''

    if (isEdit.value) {
      await announcementsApi.update(route.params.id as string, data)
      announcementId = route.params.id as string
    } else {
      const created: any = await announcementsApi.create(data)
      announcementId = created?.id || created?.ID || ''
    }

    if (publish) {
      if (!announcementId) {
        throw new Error('公告 ID 缺失，无法发布')
      }
      await announcementsApi.publish(announcementId)
      ElMessage.success('公告已发布')
    } else {
      ElMessage.success(isEdit.value ? '公告已保存' : '草稿已保存')
    }

    router.push('/announcements')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

const handleSaveDraft = () => submitForm(false)
const handlePublish = () => submitForm(true)

onMounted(loadAnnouncement)
</script>

<style lang="scss" scoped>
.announcement-edit {
  max-width: 900px;
}

.edit-card {
  border-radius: 14px;
  border: 1px solid #e2e8f0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.image-tools {
  display: flex;
  align-items: center;
  gap: 12px;
}

.image-tip {
  color: #64748b;
  font-size: 12px;
}

.image-preview-list {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.image-preview {
  width: 96px;
  height: 96px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}
</style>
