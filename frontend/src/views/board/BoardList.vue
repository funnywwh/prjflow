<template>
  <div class="board-list">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="`项目看板 - ${project?.name || ''}`"
            @back="() => router.push(`/project/${projectId}`)"
          >
            <template #extra>
              <a-button type="primary" @click="handleCreateBoard">
                <template #icon><PlusOutlined /></template>
                创建看板
              </a-button>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <a-row :gutter="16" v-if="boards.length > 0">
              <a-col
                v-for="board in boards"
                :key="board.id"
                :span="8"
                style="margin-bottom: 16px"
              >
                <a-card
                  :title="board.name"
                  :bordered="false"
                  hoverable
                  @click="handleViewBoard(board.id)"
                >
                  <template #extra>
                    <a-dropdown>
                      <a-button type="text" size="small" @click.stop>
                        <MoreOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu>
                          <a-menu-item @click="handleEditBoard(board)">编辑</a-menu-item>
                          <a-menu-item @click="handleDeleteBoard(board.id)" danger>删除</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                  </template>
                  <p v-if="board.description" class="board-description">{{ board.description }}</p>
                  <p v-else class="board-description" style="color: #999">暂无描述</p>
                  <div class="board-meta">
                    <span>列数: {{ board.columns?.length || 0 }}</span>
                  </div>
                </a-card>
              </a-col>
            </a-row>
            <a-empty v-else description="暂无看板，点击创建看板开始使用" />
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 创建/编辑看板模态框 -->
    <a-modal
      v-model:open="boardModalVisible"
      :title="boardModalTitle"
      :width="700"
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
        <a-form-item label="默认列">
          <a-space direction="vertical" style="width: 100%">
            <div
              v-for="(column, index) in boardFormData.columns"
              :key="index"
              style="display: flex; gap: 8px; align-items: center"
            >
              <a-input v-model:value="column.name" placeholder="列名称" style="width: 150px" />
              <a-select v-model:value="column.status" placeholder="关联状态" style="width: 150px">
                <a-select-option value="todo">待办</a-select-option>
                <a-select-option value="in_progress">进行中</a-select-option>
                <a-select-option value="done">已完成</a-select-option>
                <a-select-option value="cancelled">已取消</a-select-option>
              </a-select>
              <a-input v-model:value="column.color" placeholder="颜色" style="width: 100px" />
              <a-button type="link" danger @click="handleRemoveColumn(index)" v-if="boardFormData.columns.length > 1">
                删除
              </a-button>
            </div>
            <a-button type="dashed" @click="handleAddColumn" style="width: 200px">
              <template #icon><PlusOutlined /></template>
              添加列
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, MoreOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import {
  getProjectBoards,
  createBoard,
  updateBoard,
  deleteBoard,
  type Board
} from '@/api/board'
import { getProject, type Project } from '@/api/project'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const boards = ref<Board[]>([])
const project = ref<Project | null>(null)
const projectId = ref<number>(0)

const boardModalVisible = ref(false)
const boardModalTitle = ref('创建看板')
const boardFormRef = ref()
const boardFormData = reactive<{
  id?: number
  name: string
  description: string
  columns: Array<{
    name: string
    status: string
    color: string
    sort: number
  }>
}>({
  name: '',
  description: '',
  columns: [
    { name: '待办', status: 'todo', color: '#1890ff', sort: 0 },
    { name: '进行中', status: 'in_progress', color: '#52c41a', sort: 1 },
    { name: '已完成', status: 'done', color: '#faad14', sort: 2 }
  ]
})

const boardFormRules = {
  name: [{ required: true, message: '请输入看板名称', trigger: 'blur' }]
}

// 加载项目信息
const loadProject = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('项目ID无效')
    router.push('/project')
    return
  }
  projectId.value = id
  try {
    project.value = await getProject(id)
  } catch (error: any) {
    message.error(error.message || '加载项目信息失败')
  }
}

// 加载看板列表
const loadBoards = async () => {
  const id = Number(route.params.id)
  if (!id) return
  loading.value = true
  try {
    boards.value = await getProjectBoards(id)
    // 如果没有看板，自动创建默认看板
    if (boards.value.length === 0) {
      await createDefaultBoard()
    }
  } catch (error: any) {
    message.error(error.message || '加载看板列表失败')
  } finally {
    loading.value = false
  }
}

// 创建默认看板
const createDefaultBoard = async () => {
  try {
    const defaultBoard = await createBoard(projectId.value, {
      name: '默认看板',
      description: '项目默认看板',
      columns: [
        { name: '待办', status: 'todo', color: '#1890ff', sort: 0 },
        { name: '进行中', status: 'in_progress', color: '#52c41a', sort: 1 },
        { name: '已完成', status: 'done', color: '#faad14', sort: 2 }
      ]
    })
    boards.value = [defaultBoard]
    message.success('已创建默认看板')
  } catch (error: any) {
    console.error('创建默认看板失败:', error)
  }
}

// 查看看板
const handleViewBoard = (boardId: number) => {
  router.push(`/board/${boardId}`)
}

// 创建看板
const handleCreateBoard = () => {
  boardModalTitle.value = '创建看板'
  boardFormData.id = undefined
  boardFormData.name = ''
  boardFormData.description = ''
  boardFormData.columns = [
    { name: '待办', status: 'todo', color: '#1890ff', sort: 0 },
    { name: '进行中', status: 'in_progress', color: '#52c41a', sort: 1 },
    { name: '已完成', status: 'done', color: '#faad14', sort: 2 }
  ]
  boardModalVisible.value = true
}

// 编辑看板
const handleEditBoard = (board: Board) => {
  boardModalTitle.value = '编辑看板'
  boardFormData.id = board.id
  boardFormData.name = board.name
  boardFormData.description = board.description || ''
  boardFormData.columns = (board.columns || []).map(col => ({
    name: col.name,
    status: col.status,
    color: col.color || '#1890ff',
    sort: col.sort
  }))
  if (boardFormData.columns.length === 0) {
    boardFormData.columns = [
      { name: '待办', status: 'todo', color: '#1890ff', sort: 0 },
      { name: '进行中', status: 'in_progress', color: '#52c41a', sort: 1 },
      { name: '已完成', status: 'done', color: '#faad14', sort: 2 }
    ]
  }
  boardModalVisible.value = true
}

// 看板提交
const handleBoardSubmit = async () => {
  try {
    await boardFormRef.value.validate()
    if (boardFormData.id) {
      await updateBoard(boardFormData.id, {
        name: boardFormData.name,
        description: boardFormData.description
      })
      message.success('更新成功')
    } else {
      await createBoard(projectId.value, {
        name: boardFormData.name,
        description: boardFormData.description,
        columns: boardFormData.columns
      })
      message.success('创建成功')
    }
    boardModalVisible.value = false
    await loadBoards()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 看板取消
const handleBoardCancel = () => {
  boardFormRef.value?.resetFields()
}

// 删除看板
const handleDeleteBoard = async (id: number) => {
  try {
    await deleteBoard(id)
    message.success('删除成功')
    await loadBoards()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 添加列
const handleAddColumn = () => {
  boardFormData.columns.push({
    name: '',
    status: 'todo',
    color: '#1890ff',
    sort: boardFormData.columns.length
  })
}

// 删除列
const handleRemoveColumn = (index: number) => {
  boardFormData.columns.splice(index, 1)
  // 重新排序
  boardFormData.columns.forEach((col, i) => {
    col.sort = i
  })
}

onMounted(() => {
  loadProject()
  loadBoards()
})
</script>

<style scoped>
.board-list {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 1400px;
  margin: 0 auto;
}

.board-description {
  margin: 8px 0;
  color: #666;
  font-size: 14px;
}

.board-meta {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}
</style>

