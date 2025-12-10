<template>
  <div class="version-detail-content">
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
            <a-button v-if="version?.project" type="link" @click="$router.push(`/project/${version.project.id}`)">
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
      <a-card title="关联Bug" :bordered="false" style="margin-bottom: 16px">
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

      <!-- 附件 -->
      <AttachmentList :attachments="version?.attachments || []" />
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/date'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentList from '@/components/AttachmentList.vue'
import type { Version } from '@/api/version'

interface Props {
  version?: Version | null
  loading?: boolean
}

withDefaults(defineProps<Props>(), {
  version: null,
  loading: false
})

// 状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    wait: 'orange',
    normal: 'green',
    fail: 'red',
    terminate: 'default'
  }
  return colors[status] || 'default'
}

// 状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    wait: '未开始',
    normal: '已发布',
    fail: '发布失败',
    terminate: '停止维护'
  }
  return texts[status] || status
}

// 需求状态颜色
const getRequirementStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'default',
    reviewing: 'blue',
    active: 'green',
    changing: 'orange',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 需求状态文本
const getRequirementStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    reviewing: '评审中',
    active: '激活',
    changing: '变更中',
    closed: '已关闭'
  }
  return texts[status] || status
}

// Bug状态颜色
const getBugStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    active: 'orange',
    resolved: 'green',
    closed: 'default',
    reopened: 'red'
  }
  return colors[status] || 'default'
}

// Bug状态文本
const getBugStatusText = (status: string) => {
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭',
    reopened: '重新打开'
  }
  return texts[status] || status
}

// 优先级颜色
const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    normal: 'blue',
    high: 'orange',
    urgent: 'red'
  }
  return colors[priority] || 'default'
}

// 优先级文本
const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    normal: '普通',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

// 严重程度颜色
const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    critical: 'red'
  }
  return colors[severity] || 'default'
}

// 严重程度文本
const getSeverityText = (severity: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return texts[severity] || severity
}
</script>

<style scoped>
.version-detail-content {
  width: 100%;
}

</style>

