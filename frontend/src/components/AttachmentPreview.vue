<template>
  <a-modal
    v-model:open="visible"
    :title="attachment?.file_name || '预览'"
    :width="modalWidth"
    :footer="null"
    :centered="true"
    :z-index="2200"
    @cancel="handleClose"
  >
    <template #title>
      <div class="preview-header">
        <span>{{ attachment?.file_name || '预览' }}</span>
        <span class="file-size">{{ formatFileSize(attachment?.file_size || 0) }}</span>
      </div>
    </template>

    <div v-if="loading" class="preview-loading">
      <a-spin size="large" />
    </div>

    <div v-else-if="error" class="preview-error">
      <a-result status="error" :title="error" />
    </div>

    <div v-else class="preview-content">
      <!-- 图片预览 -->
      <div v-if="isImage" class="preview-image" @wheel.prevent="handleImageWheel" ref="imageContainerRef">
        <img
          ref="previewImageRef"
          :src="previewUrl"
          :alt="attachment?.file_name"
          :style="{
            maxWidth: 'none',
            maxHeight: 'none',
            width: imageScale + '%',
            height: 'auto',
            cursor: imageScale !== 100 ? 'move' : 'default',
            transition: isDragging ? 'none' : 'width 0.1s ease-out, transform 0.1s ease-out',
            display: 'block',
            margin: '0 auto',
            userSelect: 'none',
            transform: `translate(${imagePosition.x}px, ${imagePosition.y}px)`
          }"
          @mousedown="handleImageMouseDown"
          @dragstart.prevent
        />
        <div class="zoom-controls">
          <a-button-group size="small">
            <a-button @click="zoomIn" :disabled="imageScale >= 500">
              <template #icon><PlusOutlined /></template>
            </a-button>
            <a-button @click="resetZoom">{{ Math.round(imageScale) }}%</a-button>
            <a-button @click="zoomOut" :disabled="imageScale <= 50">
              <template #icon><MinusOutlined /></template>
            </a-button>
          </a-button-group>
        </div>
      </div>

      <!-- 视频预览 -->
      <div v-else-if="isVideo" class="preview-video">
        <video
          :src="previewUrl"
          controls
          style="max-width: 100%; max-height: 70vh"
        >
          您的浏览器不支持视频播放
        </video>
      </div>

      <!-- PDF预览 -->
      <div v-else-if="isPdf" class="preview-pdf">
        <iframe
          :src="previewUrl"
          style="width: 100%; height: 70vh; border: none"
        />
      </div>

      <!-- 文本预览 -->
      <div v-else-if="isText" class="preview-text">
        <pre class="text-content">{{ textContent }}</pre>
      </div>

      <!-- 不支持预览的类型 -->
      <div v-else class="preview-unsupported">
        <a-result
          status="info"
          title="不支持预览此文件类型"
          sub-title="请下载后查看"
        />
      </div>
    </div>

    <template #footer>
      <div class="preview-footer">
        <a-button @click="handleDownload">下载</a-button>
        <a-button type="primary" @click="handleClose">关闭</a-button>
      </div>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, MinusOutlined } from '@ant-design/icons-vue'
import { getPreviewBlobUrl, downloadFile, type Attachment } from '@/api/attachment'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

interface Props {
  attachment: Attachment | null
  open: boolean
}

const props = withDefaults(defineProps<Props>(), {
  attachment: null,
  open: false
})

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const visible = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})

const loading = ref(false)
const error = ref('')
const previewUrl = ref('')
const textContent = ref('')
const blobUrl = ref<string | null>(null)
const imageScale = ref(100)
const previewImageRef = ref<HTMLImageElement | null>(null)
const imageContainerRef = ref<HTMLDivElement | null>(null)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const imagePosition = ref({ x: 0, y: 0 })

// 判断文件类型
const mimeType = computed(() => props.attachment?.mime_type || '')
const isImage = computed(() => mimeType.value.startsWith('image/'))
const isVideo = computed(() => mimeType.value.startsWith('video/'))
const isPdf = computed(() => mimeType.value === 'application/pdf')
const isText = computed(() => mimeType.value.startsWith('text/'))

// 弹窗宽度
const modalWidth = computed(() => {
  if (isImage.value || isVideo.value) {
    return '90%'
  }
  if (isPdf.value) {
    return '90%'
  }
  return 800
})

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// 加载预览内容
const loadPreview = async () => {
  if (!props.attachment) return

  loading.value = true
  error.value = ''
  textContent.value = ''

  try {
    const authStore = useAuthStore()
    const token = authStore.token

    if (isText.value) {
      // 文本文件：直接获取文本内容
      const response = await axios.get(`/api/attachments/${props.attachment.id}/preview`, {
        headers: {
          'Authorization': token ? `Bearer ${token}` : ''
        },
        responseType: 'text',
        transformResponse: [(data) => data] // 不转换响应数据
      })
      textContent.value = response.data
      previewUrl.value = ''
    } else if (isPdf.value) {
      // PDF：使用iframe直接加载（需要后端支持CORS或使用blob URL）
      // 由于iframe无法自动携带Authorization header，使用blob URL
      const url = await getPreviewBlobUrl(props.attachment.id)
      previewUrl.value = url
      blobUrl.value = url
    } else {
      // 图片和视频：使用blob URL
      const url = await getPreviewBlobUrl(props.attachment.id)
      previewUrl.value = url
      blobUrl.value = url
    }
  } catch (err: any) {
    error.value = err.message || '加载预览失败'
    message.error('加载预览失败')
  } finally {
    loading.value = false
  }
}

// 下载文件
const handleDownload = async () => {
  if (!props.attachment) return
  try {
    await downloadFile(props.attachment.id, props.attachment.file_name)
    message.success('下载成功')
  } catch (err: any) {
    message.error(err.message || '下载失败')
  }
}

// 图片缩放控制
const zoomIn = () => {
  imageScale.value = Math.min(500, imageScale.value + 25)
}

const zoomOut = () => {
  imageScale.value = Math.max(50, imageScale.value - 25)
}

const resetZoom = () => {
  imageScale.value = 100
  imagePosition.value = { x: 0, y: 0 }
}

// 处理图片滚轮缩放
const handleImageWheel = (e: WheelEvent) => {
  if (!isImage.value) return
  e.preventDefault()
  const delta = e.deltaY > 0 ? -10 : 10
  const newScale = Math.max(50, Math.min(500, imageScale.value + delta))
  imageScale.value = newScale
}

// 处理图片拖拽
const handleImageMouseDown = (e: MouseEvent) => {
  // 只有在缩放不等于100%时才允许拖拽（放大或缩小都可以拖拽）
  if (imageScale.value === 100) return
  e.preventDefault()
  e.stopPropagation()
  isDragging.value = true
  dragStart.value = { x: e.clientX - imagePosition.value.x, y: e.clientY - imagePosition.value.y }
  
  const handleMouseMove = (moveEvent: MouseEvent) => {
    if (!isDragging.value) return
    imagePosition.value = {
      x: moveEvent.clientX - dragStart.value.x,
      y: moveEvent.clientY - dragStart.value.y
    }
    if (previewImageRef.value) {
      previewImageRef.value.style.transform = `translate(${imagePosition.value.x}px, ${imagePosition.value.y}px)`
    }
  }
  
  const handleMouseUp = () => {
    isDragging.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

// 关闭弹窗
const handleClose = () => {
  // 清理blob URL
  if (blobUrl.value) {
    window.URL.revokeObjectURL(blobUrl.value)
    blobUrl.value = null
  }
  previewUrl.value = ''
  textContent.value = ''
  error.value = ''
  imageScale.value = 100
  imagePosition.value = { x: 0, y: 0 }
  visible.value = false
}

// 监听打开状态
watch(() => props.open, (newVal) => {
  if (newVal && props.attachment) {
    loadPreview()
    // 重置缩放和位置
    imageScale.value = 100
    imagePosition.value = { x: 0, y: 0 }
  } else {
    // 关闭时清理资源
    if (blobUrl.value) {
      window.URL.revokeObjectURL(blobUrl.value)
      blobUrl.value = null
    }
    // 重置缩放和位置
    imageScale.value = 100
    imagePosition.value = { x: 0, y: 0 }
  }
})

// 组件卸载时清理资源
onUnmounted(() => {
  if (blobUrl.value) {
    window.URL.revokeObjectURL(blobUrl.value)
  }
})
</script>

<style scoped>
.preview-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-size {
  color: #999;
  font-size: 12px;
  font-weight: normal;
}

.preview-loading,
.preview-error {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.preview-content {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.preview-image {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  max-height: 70vh;
  overflow: auto;
  padding: 20px;
}

.preview-image img {
  object-fit: contain;
}

.zoom-controls {
  position: absolute;
  bottom: 20px;
  right: 20px;
  z-index: 10;
  background: rgba(255, 255, 255, 0.9);
  padding: 8px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.preview-video {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
}

.preview-pdf {
  width: 100%;
}

.preview-text {
  width: 100%;
  max-height: 70vh;
  overflow: auto;
}

.text-content {
  margin: 0;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 4px;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.preview-unsupported {
  width: 100%;
}

.preview-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>

