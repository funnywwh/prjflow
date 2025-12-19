<template>
  <div class="bug-detail-content">
    <a-spin :spinning="loading">
      <!-- 基本信息 -->
      <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="编号">{{ bug?.id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="Bug标题">{{ bug?.title }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-space>
              <a-tag :color="getStatusColor(bug?.status || '')">
                {{ getStatusText(bug?.status || '') }}
              </a-tag>
              <a-tag v-if="bug?.confirmed" color="green">已确认</a-tag>
              <a-tag v-else-if="bug?.status === 'active'" color="orange">未确认</a-tag>
            </a-space>
          </a-descriptions-item>
          <a-descriptions-item label="优先级">
            <a-tag :color="getPriorityColor(bug?.priority || '')">
              {{ getPriorityText(bug?.priority || '') }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="严重程度">
            <a-tag :color="getSeverityColor(bug?.severity || '')">
              {{ getSeverityText(bug?.severity || '') }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="项目">
            {{ bug?.project?.name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="关联需求">
            <a v-if="bug?.requirement" @click="handleRequirementClick" style="cursor: pointer">
              {{ bug.requirement.title }}
            </a>
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="所属版本">
            <a-space v-if="bug?.versions && bug.versions.length > 0">
              <a-tag
                v-for="version in bug.versions"
                :key="version.id"
                color="blue"
                style="cursor: pointer"
                @click="handleVersionClick(version.id)"
              >
                {{ version.version_number }}
              </a-tag>
            </a-space>
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="指派给">
            <a-space>
              <a-tag
                v-for="assignee in bug?.assignees || []"
                :key="assignee.id"
              >
                {{ assignee.username }}{{ assignee.nickname ? `(${assignee.nickname})` : '' }}
              </a-tag>
              <span v-if="!bug?.assignees || bug.assignees.length === 0">-</span>
            </a-space>
          </a-descriptions-item>
          <a-descriptions-item label="创建人">
            {{ bug?.creator ? `${bug.creator.username}${bug.creator.nickname ? `(${bug.creator.nickname})` : ''}` : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="创建时间">
            {{ formatDateTime(bug?.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="更新时间">
            {{ formatDateTime(bug?.updated_at) }}
          </a-descriptions-item>
        </a-descriptions>
      </a-card>

      <!-- Bug描述 -->
      <a-card title="Bug描述" :bordered="false" style="margin-bottom: 16px">
        <div v-if="bug?.description" class="markdown-content">
          <MarkdownEditor
            :model-value="bug.description"
            :readonly="true"
          />
        </div>
        <a-empty v-else description="暂无描述" />
      </a-card>

      <!-- 附件 -->
      <AttachmentList :attachments="bug?.attachments || []" />

      <!-- 历史记录 -->
      <a-card :bordered="false" style="margin-bottom: 16px">
        <template #title>
          <span>历史记录</span>
          <a-button 
            type="link" 
            size="small"
            @click.stop="handleAddNote" 
            :disabled="historyLoading"
            style="margin-left: 8px; padding: 0"
          >
            添加备注
          </a-button>
        </template>
        <a-spin :spinning="historyLoading" :style="{ minHeight: '100px' }">
          <a-timeline v-if="historyList.length > 0">
            <a-timeline-item
              v-for="(action, index) in historyList"
              :key="action.id"
            >
              <template #dot>
                <span style="font-weight: bold; color: #1890ff">{{ historyList.length - index }}</span>
              </template>
              <div>
                <div style="margin-bottom: 8px">
                  <span style="color: #666; margin-right: 8px">{{ formatDateTime(action.date) }}</span>
                  <span>{{ getActionDescription(action) }}</span>
                  <a-button
                    v-if="hasHistoryDetails(action)"
                    type="link"
                    size="small"
                    @click="toggleHistoryDetail(action.id)"
                    style="padding: 0; height: auto; margin-left: 8px"
                  >
                    {{ expandedHistoryIds.has(action.id) ? '收起' : '展开' }}
                  </a-button>
                </div>
                <!-- 字段变更详情和备注内容（可折叠） -->
                <div
                  v-show="expandedHistoryIds.has(action.id)"
                  style="margin-left: 24px; margin-top: 8px"
                >
                  <!-- 字段变更详情 -->
                  <div v-if="action.histories && action.histories.length > 0">
                    <div
                      v-for="history in action.histories"
                      :key="history.id"
                      style="margin-bottom: 8px; color: #666"
                    >
                      <div>修改了{{ getFieldDisplayName(history.field) }}</div>
                      <div style="margin-left: 16px; margin-top: 4px;">
                        <div>旧值："{{ history.old_value || history.old || '-' }}"</div>
                        <div>新值："{{ history.new_value || history.new || '-' }}"</div>
                      </div>
                    </div>
                  </div>
                  <!-- 备注内容 -->
                  <div v-if="action.comment" style="margin-top: 8px; color: #666">
                    {{ action.comment }}
                  </div>
                </div>
              </div>
            </a-timeline-item>
          </a-timeline>
          <a-empty v-else description="暂无历史记录" />
        </a-spin>
      </a-card>
    </a-spin>

    <!-- 添加备注模态框 -->
    <a-modal
      v-model:open="noteModalVisible"
      title="添加备注"
      :mask-closable="true"
      :z-index="2100"
      @ok="handleNoteSubmit"
      @cancel="handleNoteCancel"
    >
      <a-form
        ref="noteFormRef"
        :model="noteFormData"
        :rules="noteFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="备注" name="comment">
          <a-textarea
            v-model:value="noteFormData.comment"
            placeholder="请输入备注"
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import { formatDateTime } from '@/utils/date'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentList from '@/components/AttachmentList.vue'
import {
  getBugHistory,
  addBugHistoryNote,
  type Bug,
  type Action
} from '@/api/bug'

interface Props {
  bug: Bug | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  'refresh': []
  'requirement-click': [requirementId: number]
  'version-click': [versionId: number]
}>()

// 历史记录相关
const historyLoading = ref(false)
const historyList = ref<Action[]>([])
const expandedHistoryIds = ref<Set<number>>(new Set())
const noteModalVisible = ref(false)
const noteFormRef = ref()
const noteFormData = reactive({
  comment: ''
})
const noteFormRules = {
  comment: [{ required: true, message: '请输入备注', trigger: 'blur' }]
}

// 加载历史记录
const loadBugHistory = async (bugId?: number) => {
  const id = bugId || props.bug?.id
  if (!id) return

  historyLoading.value = true
  try {
    const response = await getBugHistory(id)
    historyList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    historyLoading.value = false
  }
}

// 监听bug变化，重新加载历史记录
watch(() => props.bug?.id, (newId) => {
  if (newId) {
    loadBugHistory(newId)
    expandedHistoryIds.value = new Set()
  }
}, { immediate: true })

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    active: 'orange',
    resolved: 'green',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string | undefined) => {
  if (!status) return '-'
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return texts[status.toLowerCase()] || status
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

// 获取严重程度颜色
const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    critical: 'red'
  }
  return colors[severity] || 'default'
}

// 获取严重程度文本
const getSeverityText = (severity: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return texts[severity] || severity
}

// 获取操作描述
const getActionDescription = (action: Action): string => {
  const actorName = action.actor
    ? `${action.actor.username}${action.actor.nickname ? `(${action.actor.nickname})` : ''}`
    : '系统'

  switch (action.action) {
    case 'created':
      return `由 ${actorName} 创建。`
    case 'assigned':
      const assignHistory = action.histories?.find(h => h.field === 'assignee_ids')
      if (assignHistory) {
        return `由 ${actorName} 指派给 ${assignHistory.new_value || assignHistory.new || '-'}。`
      }
      return `由 ${actorName} 指派。`
    case 'resolved':
      let solution = ''
      if (action.extra) {
        try {
          const extra = JSON.parse(action.extra)
          if (extra.solution) {
            solution = extra.solution
          }
        } catch (e) {
          // 解析失败，忽略
        }
      }
      return `由 ${actorName} 解决${solution ? `, 方案为 ${solution}。` : '。'}`
    case 'closed':
      return `由 ${actorName} 关闭。`
    case 'confirmed':
      return `由 ${actorName} 确认。`
    case 'edited':
      return `由 ${actorName} 编辑。`
    case 'commented':
      return `由 ${actorName} 添加了备注：${action.comment || ''}`
    default:
      return `由 ${actorName} 执行了 ${action.action} 操作。`
  }
}

// 获取字段显示名称
const getFieldDisplayName = (fieldName: string): string => {
  const fieldNames: Record<string, string> = {
    title: 'Bug标题',
    description: 'Bug描述',
    status: 'Bug状态',
    priority: '优先级',
    severity: '严重程度',
    confirmed: '是否确认',
    project_id: '项目',
    requirement_id: '关联需求',
    module_id: '功能模块',
    assignee_ids: '指派给',
    estimated_hours: '预估工时',
    actual_hours: '实际工时',
    solution: '解决方案',
    solution_note: '解决方案备注',
    resolved_version_id: '解决版本'
  }
  return fieldNames[fieldName] || fieldName
}

// 判断历史记录是否有详情
const hasHistoryDetails = (action: Action): boolean => {
  return !!(action.histories && action.histories.length > 0) || !!action.comment
}

// 切换历史记录详情展开/收起
const toggleHistoryDetail = (actionId: number) => {
  const newSet = new Set(expandedHistoryIds.value)
  if (newSet.has(actionId)) {
    newSet.delete(actionId)
  } else {
    newSet.add(actionId)
  }
  expandedHistoryIds.value = newSet
}

// 添加备注
const handleAddNote = () => {
  if (!props.bug) {
    message.warning('Bug信息未加载完成，请稍候再试')
    return
  }
  noteFormData.comment = ''
  noteModalVisible.value = true
}

// 提交备注
const handleNoteSubmit = async () => {
  if (!props.bug) return
  try {
    await noteFormRef.value.validate()
    await addBugHistoryNote(props.bug.id, { comment: noteFormData.comment })
    message.success('添加备注成功')
    noteModalVisible.value = false
    await loadBugHistory(props.bug.id)
    emit('refresh')
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '添加备注失败')
  }
}

// 取消添加备注
const handleNoteCancel = () => {
  noteFormRef.value?.resetFields()
}

// 处理需求点击
const handleRequirementClick = () => {
  if (props.bug?.requirement?.id) {
    emit('requirement-click', props.bug.requirement.id)
  }
}

// 处理版本点击
const handleVersionClick = (versionId: number) => {
  emit('version-click', versionId)
}


// 暴露方法给父组件
defineExpose({
  loadBugHistory,
  refresh: () => {
    if (props.bug?.id) {
      loadBugHistory(props.bug.id)
    }
  }
})
</script>

<style scoped>
.bug-detail-content {
  width: 100%;
}

.markdown-content {
  min-height: 200px;
}

</style>

