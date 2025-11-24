<template>
  <div class="requirement-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="requirement?.title || '需求详情'"
            @back="() => router.push('/requirement')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEdit">编辑</a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="pending">待处理</a-menu-item>
                      <a-menu-item key="in_progress">进行中</a-menu-item>
                      <a-menu-item key="completed">已完成</a-menu-item>
                      <a-menu-item key="cancelled">已取消</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
                <a-popconfirm
                  title="确定要删除这个需求吗？"
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
                <a-descriptions-item label="需求标题">{{ requirement?.title }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="getStatusColor(requirement?.status || '')">
                    {{ getStatusText(requirement?.status || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="优先级">
                  <a-tag :color="getPriorityColor(requirement?.priority || '')">
                    {{ getPriorityText(requirement?.priority || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="产品">
                  <!-- {{ requirement?.product?.name || '-' }} -->
                </a-descriptions-item>
                <a-descriptions-item label="项目">
                  {{ requirement?.project?.name || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="负责人">
                  {{ requirement?.assignee ? `${requirement.assignee.username}${requirement.assignee.nickname ? `(${requirement.assignee.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建人">
                  {{ requirement?.creator ? `${requirement.creator.username}${requirement.creator.nickname ? `(${requirement.creator.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建时间">
                  {{ formatDateTime(requirement?.created_at) }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ formatDateTime(requirement?.updated_at) }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 需求描述 -->
            <a-card title="需求描述" :bordered="false" style="margin-bottom: 16px">
              <div v-if="requirement?.description" class="markdown-content">
                <MarkdownEditor
                  :model-value="requirement.description"
                  :readonly="true"
                />
              </div>
              <a-empty v-else description="暂无描述" />
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
  getRequirement,
  updateRequirementStatus,
  deleteRequirement,
  type Requirement
} from '@/api/requirement'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const requirement = ref<Requirement | null>(null)

// 加载需求详情
const loadRequirement = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('需求ID无效')
    router.push('/requirement')
    return
  }

  loading.value = true
  try {
    requirement.value = await getRequirement(id)
  } catch (error: any) {
    message.error(error.message || '加载需求详情失败')
    router.push('/requirement')
  } finally {
    loading.value = false
  }
}

// 编辑
const handleEdit = () => {
  router.push(`/requirement?edit=${requirement.value?.id}`)
}

// 状态变更
const handleStatusChange = async (status: string) => {
  if (!requirement.value) return
  try {
    await updateRequirementStatus(requirement.value.id, { status: status as any })
    message.success('状态更新成功')
    loadRequirement()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 删除
const handleDelete = async () => {
  if (!requirement.value) return
  try {
    await deleteRequirement(requirement.value.id)
    message.success('删除成功')
    router.push('/requirement')
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'orange',
    in_progress: 'blue',
    completed: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待处理',
    in_progress: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || status
}

// 获取优先级颜色
const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    urgent: 'red'
  }
  return colors[priority] || 'default'
}

// 获取优先级文本
const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

onMounted(() => {
  loadRequirement()
})
</script>

<style scoped>
.requirement-detail {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
}

.markdown-content {
  min-height: 200px;
}
</style>

