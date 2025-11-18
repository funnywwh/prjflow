<template>
  <div class="board-view">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="board?.name || '看板'"
            :sub-title="board?.description"
            @back="() => router.push(`/project/${projectId}`)"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEditBoard">编辑看板</a-button>
                <a-button @click="handleManageColumns">管理列</a-button>
                <a-button type="primary" @click="handleCreateTask">新建任务</a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <div class="board-container" v-if="board && board.columns">
              <div
                v-for="column in sortedColumns"
                :key="column.id"
                class="board-column"
                @drop="handleDrop($event, column)"
                @dragover.prevent
                @dragenter.prevent
              >
                <div class="column-header" :style="{ borderTopColor: column.color || '#1890ff' }">
                  <span class="column-title">{{ column.name }}</span>
                  <span class="column-count">({{ getColumnTasks(column.id).length }})</span>
                </div>
                <div class="column-content">
                  <div
                    v-for="task in getColumnTasks(column.id)"
                    :key="task.id"
                    class="task-card"
                    draggable="true"
                    @dragstart="handleDragStart($event, task)"
                    @click="handleViewTask(task)"
                  >
                    <div class="task-header">
                      <span class="task-title">{{ task.title }}</span>
                      <a-tag :color="getPriorityColor(task.priority)" size="small">
                        {{ getPriorityText(task.priority) }}
                      </a-tag>
                    </div>
                    <div class="task-body">
                      <div v-if="task.assignee" class="task-assignee">
                        <a-avatar :size="20" style="margin-right: 4px">
                          {{ (task.assignee.username || '').charAt(0).toUpperCase() }}
                        </a-avatar>
                        <span>{{ task.assignee.username }}{{ task.assignee.nickname ? `(${task.assignee.nickname})` : '' }}</span>
                      </div>
                      <div v-if="task.progress !== undefined" class="task-progress">
                        <a-progress :percent="task.progress" :show-info="false" size="small" />
                      </div>
                      <div v-if="task.due_date" class="task-due-date" :style="{ color: isOverdue(task.due_date, task.status) ? 'red' : '' }">
                        <CalendarOutlined /> {{ task.due_date }}
                      </div>
                    </div>
                  </div>
                  <div v-if="getColumnTasks(column.id).length === 0" class="empty-column">
                    暂无任务
                  </div>
                </div>
              </div>
            </div>
            <a-empty v-else description="看板不存在或没有列" />
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 编辑看板模态框 -->
    <a-modal
      v-model:open="boardModalVisible"
      :title="boardModalTitle"
      @ok="handleBoardSubmit"
      @cancel="handleBoardCancel"
    >
      <a-form
        ref="boardFormRef"
        :model="boardFormData"
        :rules="boardFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="看板名称" name="name">
          <a-input v-model:value="boardFormData.name" placeholder="请输入看板名称" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="boardFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 管理列模态框 -->
    <a-modal
      v-model:open="columnsModalVisible"
      title="管理看板列"
      :width="800"
      @ok="handleColumnsSubmit"
      @cancel="handleColumnsCancel"
    >
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item
          v-for="(column, index) in columnsFormData"
          :key="index"
          :label="`列 ${index + 1}`"
        >
          <a-space style="width: 100%">
            <a-input v-model:value="column.name" placeholder="列名称" style="width: 150px" />
            <a-select v-model:value="column.status" placeholder="关联状态" style="width: 150px">
              <a-select-option value="todo">待办</a-select-option>
              <a-select-option value="in_progress">进行中</a-select-option>
              <a-select-option value="done">已完成</a-select-option>
              <a-select-option value="cancelled">已取消</a-select-option>
            </a-select>
            <a-input v-model:value="column.color" placeholder="颜色" style="width: 100px" />
            <a-input-number v-model:value="column.sort" placeholder="排序" :min="0" style="width: 100px" />
            <a-button type="link" danger @click="handleRemoveColumn(index)" v-if="columnsFormData.length > 1">
              删除
            </a-button>
          </a-space>
        </a-form-item>
        <a-form-item>
          <a-button type="dashed" @click="handleAddColumn" style="width: 100%">
            <template #icon><PlusOutlined /></template>
            添加列
          </a-button>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, CalendarOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import {
  getBoard,
  getBoardTasks,
  updateBoard,
  createBoardColumn,
  updateBoardColumn,
  deleteBoardColumn,
  moveTask,
  type Board,
  type BoardColumn
} from '@/api/board'
import { type Task } from '@/api/task'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const board = ref<Board | null>(null)
const tasksByColumn = ref<Record<number, Task[]>>({})
const projectId = ref<number>(0)
const draggedTask = ref<Task | null>(null)

const boardModalVisible = ref(false)
const boardModalTitle = ref('编辑看板')
const boardFormRef = ref()
const boardFormData = reactive({
  name: '',
  description: ''
})

const boardFormRules = {
  name: [{ required: true, message: '请输入看板名称', trigger: 'blur' }]
}

const columnsModalVisible = ref(false)
const columnsFormData = reactive<Array<{
  id?: number
  name: string
  status: string
  color: string
  sort: number
}>>([])

// 排序后的列
const sortedColumns = computed(() => {
  if (!board.value || !board.value.columns) return []
  return [...board.value.columns].sort((a, b) => a.sort - b.sort)
})

// 获取列的任务
const getColumnTasks = (columnId: number) => {
  return tasksByColumn.value[columnId] || []
}

// 加载看板数据
const loadBoard = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('看板ID无效')
    router.push('/project')
    return
  }

  loading.value = true
  try {
    board.value = await getBoard(id)
    projectId.value = board.value.project_id
    await loadBoardTasks()
  } catch (error: any) {
    message.error(error.message || '加载看板失败')
    router.push('/project')
  } finally {
    loading.value = false
  }
}

// 加载看板任务
const loadBoardTasks = async () => {
  if (!board.value) return
  try {
    const response = await getBoardTasks(board.value.id)
    tasksByColumn.value = response.tasks_by_column || {}
  } catch (error: any) {
    console.error('加载看板任务失败:', error)
  }
}

// 拖拽开始
const handleDragStart = (event: DragEvent, task: Task) => {
  draggedTask.value = task
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

// 拖拽放置
const handleDrop = async (event: DragEvent, column: BoardColumn) => {
  event.preventDefault()
  if (!draggedTask.value || !board.value) return

  // 如果任务已经在同一列，不处理
  const currentTasks = getColumnTasks(column.id)
  if (currentTasks.some(t => t.id === draggedTask.value!.id)) {
    draggedTask.value = null
    return
  }

  try {
    await moveTask(board.value.id, draggedTask.value.id, {
      column_id: String(column.id),
      position: 0
    })
    message.success('任务移动成功')
    await loadBoardTasks()
  } catch (error: any) {
    message.error(error.message || '移动任务失败')
  } finally {
    draggedTask.value = null
  }
}

// 查看任务
const handleViewTask = (task: Task) => {
  router.push(`/task/${task.id}`)
}

// 创建任务
const handleCreateTask = () => {
  router.push(`/task?project_id=${projectId.value}`)
}

// 编辑看板
const handleEditBoard = () => {
  if (!board.value) return
  boardModalTitle.value = '编辑看板'
  boardFormData.name = board.value.name
  boardFormData.description = board.value.description || ''
  boardModalVisible.value = true
}

// 看板提交
const handleBoardSubmit = async () => {
  if (!board.value) return
  try {
    await boardFormRef.value.validate()
    await updateBoard(board.value.id, {
      name: boardFormData.name,
      description: boardFormData.description
    })
    message.success('更新成功')
    boardModalVisible.value = false
    await loadBoard()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '更新失败')
  }
}

// 看板取消
const handleBoardCancel = () => {
  boardFormRef.value?.resetFields()
}

// 管理列
const handleManageColumns = () => {
  if (!board.value || !board.value.columns) return
  columnsFormData.splice(0, columnsFormData.length)
  board.value.columns.forEach(col => {
    columnsFormData.push({
      id: col.id,
      name: col.name,
      status: col.status,
      color: col.color || '',
      sort: col.sort
    })
  })
  columnsModalVisible.value = true
}

// 添加列
const handleAddColumn = () => {
  columnsFormData.push({
    name: '',
    status: 'todo',
    color: '#1890ff',
    sort: columnsFormData.length
  })
}

// 删除列
const handleRemoveColumn = (index: number) => {
  columnsFormData.splice(index, 1)
  // 重新排序
  columnsFormData.forEach((col, i) => {
    col.sort = i
  })
}

// 列提交
const handleColumnsSubmit = async () => {
  if (!board.value) return
  try {
    // 更新现有列
    for (const col of columnsFormData) {
      if (col.id) {
        await updateBoardColumn(board.value.id, col.id, {
          name: col.name,
          status: col.status,
          color: col.color,
          sort: col.sort
        })
      } else {
        await createBoardColumn(board.value.id, {
          name: col.name,
          status: col.status,
          color: col.color,
          sort: col.sort
        })
      }
    }
    // 删除已移除的列
    if (board.value.columns) {
      const existingIds = columnsFormData.filter(c => c.id).map(c => c.id!)
      const toDelete = board.value.columns.filter(c => !existingIds.includes(c.id))
      for (const col of toDelete) {
        await deleteBoardColumn(board.value.id, col.id)
      }
    }
    message.success('更新成功')
    columnsModalVisible.value = false
    await loadBoard()
  } catch (error: any) {
    message.error(error.message || '更新失败')
  }
}

// 列取消
const handleColumnsCancel = () => {
  columnsFormData.splice(0, columnsFormData.length)
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

// 判断是否逾期
const isOverdue = (dueDate: string, status?: string) => {
  if (status === 'done' || status === 'cancelled') {
    return false
  }
  const due = dayjs(dueDate)
  const now = dayjs()
  return due.isBefore(now, 'day')
}

onMounted(() => {
  loadBoard()
})
</script>

<style scoped>
.board-view {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
}

.board-container {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 16px;
  min-height: calc(100vh - 200px);
}

.board-column {
  flex: 0 0 300px;
  background: #f5f5f5;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 200px);
}

.column-header {
  padding: 12px 16px;
  background: white;
  border-top: 3px solid #1890ff;
  border-radius: 4px 4px 0 0;
  font-weight: 600;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.column-title {
  font-size: 14px;
}

.column-count {
  color: #999;
  font-size: 12px;
}

.column-content {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
  min-height: 200px;
}

.task-card {
  background: white;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 8px;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
}

.task-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.task-card[draggable="true"] {
  cursor: move;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.task-title {
  font-weight: 500;
  flex: 1;
  margin-right: 8px;
}

.task-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-assignee {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #666;
}

.task-progress {
  margin-top: 4px;
}

.task-due-date {
  font-size: 12px;
  color: #999;
  display: flex;
  align-items: center;
  gap: 4px;
}

.empty-column {
  text-align: center;
  color: #999;
  padding: 40px 0;
  font-size: 14px;
}
</style>

