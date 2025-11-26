<template>
  <div class="module-management">
    <a-card v-if="showCard" :bordered="false">
      <template #title>
        <span>{{ title }}</span>
      </template>
      <template #extra>
        <a-button type="primary" @click="handleCreate">
          <template #icon><PlusOutlined /></template>
          新增模块
        </a-button>
      </template>

      <a-table
        :columns="columns"
        :data-source="modules"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '正常' : '禁用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleEdit(record)">
                编辑
              </a-button>
              <a-popconfirm
                title="确定要删除这个模块吗？"
                @confirm="handleDelete(record.id)"
              >
                <a-button type="link" size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 不显示卡片时，直接显示表格 -->
    <template v-else>
      <div style="margin-bottom: 16px; text-align: right">
        <a-button type="primary" @click="handleCreate">
          <template #icon><PlusOutlined /></template>
          新增模块
        </a-button>
      </div>
      <a-table
        :columns="columns"
        :data-source="modules"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '正常' : '禁用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleEdit(record)">
                编辑
              </a-button>
              <a-popconfirm
                title="确定要删除这个模块吗？"
                @confirm="handleDelete(record.id)"
              >
                <a-button type="link" size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </template>

    <!-- 模块编辑对话框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleSubmit"
      @cancel="handleCancel"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="模块名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入模块名称" />
        </a-form-item>
        <a-form-item label="模块编码" name="code">
          <a-input v-model:value="formData.code" placeholder="请输入模块编码（可选）" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea
            v-model:value="formData.description"
            placeholder="请输入模块描述"
            :rows="3"
          />
        </a-form-item>
        <a-form-item label="排序" name="sort">
          <a-input-number
            v-model:value="formData.sort"
            :min="0"
            placeholder="排序值（数字越小越靠前）"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-radio-group v-model:value="formData.status">
            <a-radio :value="1">正常</a-radio>
            <a-radio :value="0">禁用</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import type { FormInstance } from 'ant-design-vue'
import { getModules, createModule, updateModule, deleteModule, type Module } from '@/api/module'

// Props
interface Props {
  showCard?: boolean // 是否显示卡片包装
  title?: string // 卡片标题
}

// Props 在模板中使用，不需要在 script 中访问
withDefaults(defineProps<Props>(), {
  showCard: true,
  title: '功能模块管理'
})

// Emits
const emit = defineEmits<{
  change: [] // 数据变更事件
  created: [module: Module] // 创建成功事件
  updated: [module: Module] // 更新成功事件
  deleted: [id: number] // 删除成功事件
}>()

const loading = ref(false)
const modules = ref<Module[]>([])
const modalVisible = ref(false)
const modalTitle = ref('新增模块')
const currentModuleId = ref<number>()
const formRef = ref<FormInstance>()
const formData = ref({
  name: '',
  code: '',
  description: '',
  status: 1,
  sort: 0
})

const rules = {
  name: [{ required: true, message: '请输入模块名称', trigger: 'blur' }]
}

const columns = [
  { title: '模块名称', dataIndex: 'name', key: 'name' },
  { title: '模块编码', dataIndex: 'code', key: 'code' },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80 },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作', key: 'action', width: 150 }
]

// 加载模块列表
const loadModules = async () => {
  loading.value = true
  try {
    modules.value = await getModules()
  } catch (error: any) {
    message.error(error.message || '加载模块列表失败')
  } finally {
    loading.value = false
  }
}

// 创建模块
const handleCreate = () => {
  modalTitle.value = '新增模块'
  currentModuleId.value = undefined
  formData.value = {
    name: '',
    code: '',
    description: '',
    status: 1,
    sort: 0
  }
  modalVisible.value = true
}

// 编辑模块
const handleEdit = (record: Module) => {
  modalTitle.value = '编辑模块'
  currentModuleId.value = record.id
  formData.value = {
    name: record.name,
    code: record.code || '',
    description: record.description || '',
    status: record.status,
    sort: record.sort
  }
  modalVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    
    const data = {
      ...formData.value
    }

    if (modalTitle.value === '新增模块') {
      const newModule = await createModule(data)
      message.success('创建成功')
      emit('created', newModule)
    } else {
      if (currentModuleId.value) {
        const updatedModule = await updateModule(currentModuleId.value, data)
        message.success('更新成功')
        emit('updated', updatedModule)
      }
    }

    modalVisible.value = false
    await loadModules()
    emit('change')
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 取消
const handleCancel = () => {
  modalVisible.value = false
  formRef.value?.resetFields()
}

// 删除模块
const handleDelete = async (id: number) => {
  try {
    await deleteModule(id)
    message.success('删除成功')
    await loadModules()
    emit('deleted', id)
    emit('change')
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 暴露方法供父组件调用
defineExpose({
  loadModules,
  refresh: loadModules
})

onMounted(() => {
  loadModules()
})
</script>

<style scoped>
.module-management {
  padding: 0;
}
</style>

