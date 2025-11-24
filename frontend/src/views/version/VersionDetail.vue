<template>
  <div class="version-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="version?.version_number || '版本详情'"
            @back="() => $router.push('/version')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEdit">编辑</a-button>
                <a-button v-if="version?.status === 'draft'" type="primary" @click="handleRelease">
                  发布
                </a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="draft">草稿</a-menu-item>
                      <a-menu-item key="released">已发布</a-menu-item>
                      <a-menu-item key="archived">已归档</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
                <a-popconfirm
                  title="确定要删除这个版本吗？"
                  @confirm="handleDelete"
                >
                  <a-button danger>删除</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <!-- 基本信息 -->
            <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
              <a-descriptions :column="2" bordered>
                <a-descriptions-item label="版本号">{{ version?.version_number }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="getStatusColor(version?.status || '')">
                    {{ getStatusText(version?.status || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="项目">
                  <a-button v-if="version?.project" type="link" @click="router.push(`/project/${version.project.id}`)">
                    {{ version.project.name }}
                  </a-button>
                  <span v-else>-</span>
                </a-descriptions-item>
                <a-descriptions-item label="发布日期">
                  {{ formatDateTime(version?.release_date) }}
                </a-descriptions-item>
                <a-descriptions-item label="创建时间">
                  {{ formatDateTime(version?.created_at) }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ formatDateTime(version?.updated_at) }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 发布说明 -->
            <a-card title="发布说明" :bordered="false" style="margin-bottom: 16px">
              <div v-if="version?.release_notes" class="markdown-content">
                <MarkdownEditor
                  :model-value="version.release_notes"
                  :readonly="true"
                />
              </div>
              <a-empty v-else description="暂无发布说明" />
            </a-card>

            <!-- 关联需求 -->
            <a-card title="关联需求" :bordered="false" style="margin-bottom: 16px">
              <a-list
                v-if="version?.requirements && version.requirements.length > 0"
                :data-source="version.requirements"
                :pagination="false"
              >
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #title>
                        <a-button type="link" @click="$router.push(`/requirement/${item.id}`)">
                          {{ item.title }}
                        </a-button>
                      </template>
                      <template #description>
                        <a-space>
                          <a-tag :color="getRequirementStatusColor(item.status)">
                            {{ getRequirementStatusText(item.status) }}
                          </a-tag>
                          <a-tag :color="getPriorityColor(item.priority)">
                            {{ getPriorityText(item.priority) }}
                          </a-tag>
                        </a-space>
                      </template>
                    </a-list-item-meta>
                  </a-list-item>
                </template>
              </a-list>
              <a-empty v-else description="暂无关联需求" />
            </a-card>

            <!-- 关联Bug -->
            <a-card title="关联Bug" :bordered="false">
              <a-list
                v-if="version?.bugs && version.bugs.length > 0"
                :data-source="version.bugs"
                :pagination="false"
              >
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #title>
                        <a-button type="link" @click="$router.push(`/bug/${item.id}`)">
                          {{ item.title }}
                        </a-button>
                      </template>
                      <template #description>
                        <a-space>
                          <a-tag :color="getBugStatusColor(item.status)">
                            {{ getBugStatusText(item.status) }}
                          </a-tag>
                          <a-tag :color="getPriorityColor(item.priority)">
                            {{ getPriorityText(item.priority) }}
                          </a-tag>
                          <a-tag :color="getSeverityColor(item.severity)">
                            {{ getSeverityText(item.severity) }}
                          </a-tag>
                        </a-space>
                      </template>
                    </a-list-item-meta>
                  </a-list-item>
                </template>
              </a-list>
              <a-empty v-else description="暂无关联Bug" />
            </a-card>
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getVersion,
  updateVersionStatus,
  deleteVersion,
  releaseVersion,
  type Version
} from '@/api/version'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const version = ref<Version | null>(null)

// 加载版本详情
const loadVersion = async () => {
  const id = route.params.id as string
  if (!id) {
    message.error('版本ID不存在')
    router.push('/version')
    return
  }

  loading.value = true
  try {
    const res = await getVersion(Number(id))
    version.value = res
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载失败')
    router.push('/version')
  } finally {
    loading.value = false
  }
}

// 编辑
const handleEdit = () => {
  if (version.value) {
    router.push(`/version?edit=${version.value.id}`)
  }
}

// 删除
const handleDelete = async () => {
  if (!version.value) return
  try {
    await deleteVersion(version.value.id)
    message.success('删除成功')
    router.push('/version')
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 状态变更
const handleStatusChange = async (status: string) => {
  if (!version.value) return
  try {
    await updateVersionStatus(version.value.id, status)
    message.success('状态更新成功')
    loadVersion()
  } catch (error: any) {
    message.error(error.response?.data?.message || '状态更新失败')
  }
}

// 发布版本
const handleRelease = async () => {
  if (!version.value) return
  try {
    await releaseVersion(version.value.id)
    message.success('发布成功')
    loadVersion()
  } catch (error: any) {
    message.error(error.response?.data?.message || '发布失败')
  }
}

// 状态颜色和文本
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'default',
    released: 'success',
    archived: 'default'
  }
  return colors[status] || 'default'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    released: '已发布',
    archived: '已归档'
  }
  return texts[status] || status
}

const getRequirementStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'default',
    in_progress: 'processing',
    completed: 'success',
    cancelled: 'default'
  }
  return colors[status] || 'default'
}

const getRequirementStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待处理',
    in_progress: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || status
}

const getBugStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    open: 'default',
    assigned: 'processing',
    in_progress: 'processing',
    resolved: 'success',
    closed: 'default'
  }
  return colors[status] || 'default'
}

const getBugStatusText = (status: string) => {
  const texts: Record<string, string> = {
    open: '待处理',
    assigned: '已分配',
    in_progress: '进行中',
    resolved: '已解决',
    closed: '已关闭'
  }
  return texts[status] || status
}

const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    urgent: 'red'
  }
  return colors[priority] || 'default'
}

const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'orange',
    high: 'red',
    critical: 'red'
  }
  return colors[severity] || 'default'
}

const getSeverityText = (severity: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return texts[severity] || severity
}

onMounted(() => {
  loadVersion()
})
</script>

<style scoped>
.version-detail {
  min-height: 100vh;
}
.content {
  padding: 24px;
}
.content-inner {
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
}
.markdown-content {
  padding: 16px 0;
}
</style>

