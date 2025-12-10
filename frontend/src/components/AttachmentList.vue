<template>
  <a-card title="附件" :bordered="false" style="margin-bottom: 16px">
    <div v-if="attachments && attachments.length > 0" class="attachment-list">
      <div
        v-for="attachment in attachments"
        :key="attachment.id"
        class="attachment-item"
      >
        <div class="attachment-info">
          <PaperClipOutlined class="attachment-icon" />
          <span class="attachment-name" :title="attachment.file_name">{{ attachment.file_name }}</span>
          <span class="attachment-size">{{ formatFileSize(attachment.file_size) }}</span>
        </div>
        <div class="attachment-actions">
          <!-- 预览按钮（仅支持预览的文件类型显示） -->
          <a-button
            v-if="isPreviewable(attachment.mime_type)"
            type="link"
            size="small"
            @click="handlePreview(attachment)"
          >
            <template #icon><EyeOutlined /></template>
            预览
          </a-button>
          <!-- 下载按钮 -->
          <a-button
            type="link"
            size="small"
            @click="handleDownload(attachment)"
          >
            <template #icon><DownloadOutlined /></template>
            下载
          </a-button>
          <!-- 删除按钮（非只读模式且需要权限） -->
          <a-button
            v-if="!readonly"
            v-permission="'attachment:delete'"
            type="link"
            size="small"
            danger
            @click="handleDelete(attachment)"
          >
            <template #icon><DeleteOutlined /></template>
            删除
          </a-button>
        </div>
      </div>
    </div>
    <a-empty v-else description="暂无附件" />
  </a-card>

  <!-- 预览弹窗 -->
  <AttachmentPreview
    :attachment="previewAttachment"
    :open="previewVisible"
    @update:open="previewVisible = $event"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PaperClipOutlined,
  DownloadOutlined,
  DeleteOutlined,
  EyeOutlined
} from '@ant-design/icons-vue'
import { downloadFile, deleteAttachment, type Attachment } from '@/api/attachment'
import AttachmentPreview from './AttachmentPreview.vue'

interface Props {
  attachments?: Attachment[]
  readonly?: boolean
}

withDefaults(defineProps<Props>(), {
  attachments: () => [],
  readonly: false
})

const emit = defineEmits<{
  'attachment-deleted': [attachmentId: number]
}>()

const previewVisible = ref(false)
const previewAttachment = ref<Attachment | null>(null)

// 判断文件是否可预览
const isPreviewable = (mimeType: string): boolean => {
  if (!mimeType) return false
  return (
    mimeType.startsWith('image/') ||
    mimeType.startsWith('video/') ||
    mimeType === 'application/pdf' ||
    mimeType.startsWith('text/')
  )
}

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// 预览附件
const handlePreview = (attachment: Attachment) => {
  previewAttachment.value = attachment
  previewVisible.value = true
}

// 下载附件
const handleDownload = async (attachment: Attachment) => {
  try {
    await downloadFile(attachment.id, attachment.file_name)
    message.success('下载成功')
  } catch (error: any) {
    message.error(error.message || '下载失败')
  }
}

// 删除附件
const handleDelete = (attachment: Attachment) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除附件 "${attachment.file_name}" 吗？`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteAttachment(attachment.id)
        message.success('删除成功')
        emit('attachment-deleted', attachment.id)
      } catch (error: any) {
        message.error(error.message || '删除失败')
      }
    }
  })
}
</script>

<style scoped>
.attachment-list {
  width: 100%;
}

.attachment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 8px;
  background: #fafafa;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  transition: all 0.3s;
}

.attachment-item:hover {
  background: #f0f0f0;
  border-color: #1890ff;
}

.attachment-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.attachment-icon {
  margin-right: 8px;
  color: #1890ff;
  font-size: 16px;
}

.attachment-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.attachment-size {
  color: #999;
  font-size: 12px;
  white-space: nowrap;
}

.attachment-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
</style>

