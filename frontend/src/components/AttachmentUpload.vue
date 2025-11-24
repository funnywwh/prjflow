<template>
  <div class="attachment-upload">
    <!-- 上传按钮 -->
    <a-upload
      v-if="!readonly"
      :file-list="fileList"
      :before-upload="beforeUpload"
      :custom-request="handleUpload"
      :show-upload-list="false"
      multiple
    >
      <a-button>
        <template #icon><UploadOutlined /></template>
        选择文件
      </a-button>
    </a-upload>

    <!-- 文件列表 -->
    <div v-if="fileList.length > 0" class="file-list">
      <div
        v-for="(file, index) in fileList"
        :key="file.uid || index"
        class="file-item"
      >
        <div class="file-info">
          <PaperClipOutlined class="file-icon" />
          <span class="file-name" :title="file.name">{{ file.name }}</span>
          <span class="file-size">{{ formatFileSize(file.size || 0) }}</span>
        </div>
        <div class="file-actions">
          <!-- 上传进度 -->
          <a-progress
            v-if="file.status === 'uploading'"
            :percent="file.percent || 0"
            :size="small"
            style="width: 100px; margin-right: 8px"
          />
          <!-- 下载按钮（已上传） -->
          <a-button
            v-if="file.status === 'done' && file.id"
            type="link"
            size="small"
            @click="handleDownload(file)"
          >
            <template #icon><DownloadOutlined /></template>
          </a-button>
          <!-- 删除按钮 -->
          <a-button
            v-if="!readonly"
            type="link"
            size="small"
            danger
            @click="handleRemove(file, index)"
          >
            <template #icon><DeleteOutlined /></template>
          </a-button>
        </div>
      </div>
    </div>

    <!-- 已存在的附件列表（只读模式） -->
    <div v-if="readonly && existingAttachments.length > 0" class="file-list">
      <div
        v-for="(attachment, index) in existingAttachments"
        :key="attachment.id"
        class="file-item"
      >
        <div class="file-info">
          <PaperClipOutlined class="file-icon" />
          <span class="file-name" :title="attachment.file_name">{{ attachment.file_name }}</span>
          <span class="file-size">{{ formatFileSize(attachment.file_size) }}</span>
        </div>
        <div class="file-actions">
          <a-button
            type="link"
            size="small"
            @click="handleDownloadAttachment(attachment)"
          >
            <template #icon><DownloadOutlined /></template>
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  UploadOutlined,
  DeleteOutlined,
  DownloadOutlined,
  PaperClipOutlined
} from '@ant-design/icons-vue'
import type { UploadFile, UploadRequestOption } from 'ant-design-vue'
import { uploadFile as uploadFileAPI, deleteAttachment, downloadFile, type Attachment } from '@/api/attachment'

interface Props {
  projectId: number
  modelValue?: number[] // 已上传的附件ID列表
  readonly?: boolean
  existingAttachments?: Attachment[] // 已存在的附件列表（只读模式）
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => [],
  readonly: false,
  existingAttachments: () => []
})

const emit = defineEmits<{
  'update:modelValue': [value: number[]]
}>()

const fileList = ref<UploadFile[]>([])

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// 上传前验证
const beforeUpload = (file: File): boolean => {
  // 文件大小限制：100MB
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    message.error('文件大小不能超过 100MB')
    return false
  }
  return true
}

// 自定义上传
const handleUpload = async (options: UploadRequestOption) => {
  const { file, onProgress, onSuccess, onError } = options

  if (!(file instanceof File)) {
    onError?.(new Error('无效的文件'))
    return
  }

  // 添加到文件列表
  const uploadFileItem: UploadFile = {
    uid: `${Date.now()}-${Math.random()}`,
    name: file.name,
    size: file.size,
    status: 'uploading',
    percent: 0
  }
  fileList.value.push(uploadFileItem)

  try {
    // 上传文件
    const attachment = await uploadFileAPI(
      file,
      props.projectId,
      (progress) => {
        uploadFileItem.percent = progress
        onProgress?.({ percent: progress })
      }
    )

    // 更新文件状态
    uploadFileItem.status = 'done'
    uploadFileItem.percent = 100
    uploadFileItem.id = attachment.id
    uploadFileItem.response = attachment

    // 更新已上传的附件ID列表
    const currentIds = [...props.modelValue]
    currentIds.push(attachment.id)
    emit('update:modelValue', currentIds)

    onSuccess?.(attachment)
    message.success('上传成功')
  } catch (error: any) {
    uploadFileItem.status = 'error'
    onError?.(error)
    message.error(error.message || '上传失败')
  }
}

// 删除文件
const handleRemove = async (file: UploadFile, index: number) => {
  // 如果已上传，需要调用删除API
  if (file.status === 'done' && file.id) {
    try {
      await deleteAttachment(file.id)
      // 从已上传的附件ID列表中移除
      const currentIds = props.modelValue.filter(id => id !== file.id)
      emit('update:modelValue', currentIds)
      message.success('删除成功')
    } catch (error: any) {
      message.error(error.message || '删除失败')
      return
    }
  }

  // 从文件列表中移除
  fileList.value.splice(index, 1)
}

// 下载文件
const handleDownload = async (file: UploadFile) => {
  if (!file.id || !file.name) return

  try {
    await downloadFile(file.id, file.name)
  } catch (error: any) {
    message.error(error.message || '下载失败')
  }
}

// 下载已存在的附件
const handleDownloadAttachment = async (attachment: Attachment) => {
  try {
    await downloadFile(attachment.id, attachment.file_name)
  } catch (error: any) {
    message.error(error.message || '下载失败')
  }
}

// 监听 modelValue 变化，清理已删除的附件
watch(() => props.modelValue, (newIds) => {
  // 移除不在新列表中的已上传文件
  fileList.value = fileList.value.filter(file => {
    if (file.status === 'done' && file.id) {
      return newIds.includes(file.id)
    }
    return true
  })
}, { deep: true })
</script>

<style scoped>
.attachment-upload {
  width: 100%;
}

.file-list {
  margin-top: 16px;
}

.file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  margin-bottom: 8px;
  background: #fafafa;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}

.file-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.file-icon {
  margin-right: 8px;
  color: #1890ff;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.file-size {
  color: #999;
  font-size: 12px;
  white-space: nowrap;
}

.file-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

