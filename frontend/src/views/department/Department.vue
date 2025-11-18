<template>
  <div class="department-management">
    <a-layout>
      <a-layout-header class="header">
        <div class="logo">项目管理系统</div>
        <a-menu
          mode="horizontal"
          :selected-keys="selectedKeys"
          :style="{ lineHeight: '64px' }"
        >
          <a-menu-item key="dashboard" @click="$router.push('/dashboard')">
            工作台
          </a-menu-item>
          <a-menu-item key="user" @click="$router.push('/user')">
            用户管理
          </a-menu-item>
          <a-menu-item key="permission" @click="$router.push('/permission')">
            权限管理
          </a-menu-item>
          <a-menu-item key="department" @click="$router.push('/department')">
            部门管理
          </a-menu-item>
        </a-menu>
      </a-layout-header>
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="部门管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增部门
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false">
            <a-table
              :columns="columns"
              :data-source="departmentList"
              :loading="loading"
              :pagination="false"
              row-key="id"
              :default-expand-all-rows="false"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <a-tag :color="record.status === 1 ? 'green' : 'red'">
                    {{ record.status === 1 ? '正常' : '禁用' }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleAddChild(record)">
                      添加子部门
                    </a-button>
                    <a-button type="link" size="small" @click="handleManageMembers(record)">
                      成员管理
                    </a-button>
                    <a-button type="link" size="small" @click="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-popconfirm
                      title="确定要删除这个部门吗？"
                      @confirm="handleDelete(record.id)"
                    >
                      <a-button type="link" size="small" danger>删除</a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </template>
            </a-table>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 部门编辑对话框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleSubmit"
      @cancel="handleCancel"
      :confirm-loading="submitting"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="部门名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入部门名称" />
        </a-form-item>
        <a-form-item label="部门代码" name="code">
          <a-input v-model:value="formData.code" placeholder="请输入部门代码" />
        </a-form-item>
        <a-form-item label="父部门">
          <a-tree-select
            v-model:value="formData.parent_id"
            :tree-data="departmentTreeData"
            placeholder="选择父部门（不选则为顶级部门）"
            allow-clear
            :field-names="{ children: 'children', label: 'name', value: 'id' }"
            :disabled="!!formData.id"
          />
        </a-form-item>
        <a-form-item label="排序" name="sort">
          <a-input-number v-model:value="formData.sort" :min="0" placeholder="排序值，数字越小越靠前" style="width: 100%" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 部门成员管理对话框 -->
    <a-modal
      v-model:open="memberModalVisible"
      title="部门成员管理"
      @cancel="handleCloseMemberModal"
      @ok="handleCloseMemberModal"
      ok-text="关闭"
      width="800px"
    >
      <a-spin :spinning="memberLoading">
        <div style="margin-bottom: 16px">
          <a-space>
            <a-select
              v-model:value="selectedUserIds"
              mode="multiple"
              placeholder="选择用户"
              style="width: 300px"
            >
              <a-select-option
                v-for="user in availableUsers"
                :key="user.id"
                :value="user.id"
              >
                {{ user.username }}
              </a-select-option>
            </a-select>
            <a-button type="primary" @click="handleAddMembers">添加成员</a-button>
          </a-space>
        </div>
        <a-table
          :columns="memberColumns"
          :data-source="departmentMembers"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'username'">
              {{ record.username }}
            </template>
            <template v-else-if="column.key === 'email'">
              {{ record.email || '-' }}
            </template>
            <template v-else-if="column.key === 'roles'">
              <a-tag
                v-for="role in record.roles"
                :key="role.id"
                style="margin-right: 4px"
              >
                {{ role.name }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-popconfirm
                title="确定要移除这个成员吗？"
                @confirm="handleRemoveMember(record.id)"
              >
                <a-button type="link" size="small" danger>移除</a-button>
              </a-popconfirm>
            </template>
          </template>
        </a-table>
      </a-spin>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import {
  getDepartments,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  getDepartmentMembers,
  addDepartmentMembers,
  removeDepartmentMember,
  type Department
} from '@/api/department'
import { getUsers, type User } from '@/api/user'

const route = useRoute()
const router = useRouter()
const selectedKeys = ref([route.name as string])

const loading = ref(false)
const submitting = ref(false)
const memberLoading = ref(false)
const departments = ref<Department[]>([])
const departmentMembers = ref<User[]>([])
const availableUsers = ref<User[]>([])
const currentDepartmentId = ref<number>()
const memberModalVisible = ref(false)
const selectedUserIds = ref<number[]>([])

const columns = [
  {
    title: '部门名称',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: '部门代码',
    dataIndex: 'code',
    key: 'code'
  },
  {
    title: '层级',
    dataIndex: 'level',
    key: 'level',
    width: 80
  },
  {
    title: '排序',
    dataIndex: 'sort',
    key: 'sort',
    width: 80
  },
  {
    title: '状态',
    key: 'status',
    width: 80
  },
  {
    title: '操作',
    key: 'action',
    width: 300,
    fixed: 'right' as const
  }
]

const memberColumns = [
  { title: '用户名', key: 'username', width: 150 },
  { title: '邮箱', key: 'email', width: 200 },
  { title: '角色', key: 'roles', width: 200 },
  { title: '操作', key: 'action', width: 100 }
]

// 将树形结构转换为扁平列表（用于表格显示）
const departmentList = computed(() => {
  if (!departments.value || departments.value.length === 0) {
    return []
  }
  const flatten = (list: Department[], level = 0): Department[] => {
    const result: Department[] = []
    list.forEach(item => {
      result.push({ ...item, level })
      if (item.children && item.children.length > 0) {
        result.push(...flatten(item.children, level + 1))
      }
    })
    return result
  }
  return flatten(departments.value)
})

// 树形数据（用于选择器）
const departmentTreeData = computed(() => {
  if (!departments.value || departments.value.length === 0) {
    return []
  }
  const buildTree = (list: Department[], parentId?: number): any[] => {
    return list
      .filter(item => {
        if (parentId === undefined) {
          return !item.parent_id
        }
        return item.parent_id === parentId
      })
      .map(item => ({
        id: item.id,
        name: item.name,
        children: buildTree(list, item.id)
      }))
  }
  return buildTree(departments.value)
})

const modalVisible = ref(false)
const modalTitle = ref('新增部门')
const formRef = ref()
const formData = reactive<Partial<Department> & { id?: number }>({
  name: '',
  code: '',
  parent_id: undefined,
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入部门代码', trigger: 'blur' }]
}

// 加载部门列表
const loadDepartments = async () => {
  loading.value = true
  try {
    departments.value = await getDepartments()
  } catch (error: any) {
    message.error(error.message || '加载部门列表失败')
  } finally {
    loading.value = false
  }
}

// 加载用户列表（用于选择）
const loadAvailableUsers = async () => {
  try {
    const response = await getUsers()
    availableUsers.value = response.list || []
  } catch (error: any) {
    console.error('加载用户列表失败:', error)
  }
}

// 加载部门成员
const loadDepartmentMembers = async (departmentId: number) => {
  memberLoading.value = true
  try {
    departmentMembers.value = await getDepartmentMembers(departmentId)
  } catch (error: any) {
    message.error(error.message || '加载部门成员失败')
  } finally {
    memberLoading.value = false
  }
}

// 新增
const handleCreate = () => {
  modalTitle.value = '新增部门'
  Object.assign(formData, {
    name: '',
    code: '',
    parent_id: undefined,
    sort: 0,
    status: 1
  })
  delete formData.id
  modalVisible.value = true
}

// 添加子部门
const handleAddChild = (record: Department) => {
  modalTitle.value = '新增子部门'
  Object.assign(formData, {
    name: '',
    code: '',
    parent_id: record.id,
    sort: 0,
    status: 1
  })
  delete formData.id
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: Department) => {
  modalTitle.value = '编辑部门'
  Object.assign(formData, {
    id: record.id,
    name: record.name,
    code: record.code,
    parent_id: record.parent_id,
    sort: record.sort,
    status: record.status
  })
  modalVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    
    const data: any = {
      name: formData.name,
      code: formData.code,
      sort: formData.sort,
      status: formData.status
    }
    if (formData.parent_id) {
      data.parent_id = formData.parent_id
    }
    
    if (formData.id) {
      await updateDepartment(formData.id, data)
      message.success('更新成功')
    } else {
      await createDepartment(data)
      message.success('创建成功')
    }
    
    modalVisible.value = false
    loadDepartments()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// 取消
const handleCancel = () => {
  modalVisible.value = false
  formRef.value?.resetFields()
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteDepartment(id)
    message.success('删除成功')
    loadDepartments()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 管理成员
const handleManageMembers = async (record: Department) => {
  currentDepartmentId.value = record.id
  memberModalVisible.value = true
  selectedUserIds.value = []
  await loadDepartmentMembers(record.id)
}

// 添加成员
const handleAddMembers = async () => {
  if (!currentDepartmentId.value || selectedUserIds.value.length === 0) {
    message.warning('请选择用户')
    return
  }
  try {
    await addDepartmentMembers(currentDepartmentId.value, selectedUserIds.value)
    message.success('添加成功')
    selectedUserIds.value = []
    await loadDepartmentMembers(currentDepartmentId.value)
  } catch (error: any) {
    message.error(error.message || '添加失败')
  }
}

// 移除成员
const handleRemoveMember = async (userId: number) => {
  if (!currentDepartmentId.value) return
  try {
    await removeDepartmentMember(currentDepartmentId.value, userId)
    message.success('移除成功')
    await loadDepartmentMembers(currentDepartmentId.value)
  } catch (error: any) {
    message.error(error.message || '移除失败')
  }
}

// 关闭成员管理对话框
const handleCloseMemberModal = () => {
  memberModalVisible.value = false
  selectedUserIds.value = []
}

onMounted(() => {
  loadDepartments()
  loadAvailableUsers()
})
</script>

<style scoped>
.department-management {
  min-height: 100vh;
}

.header {
  background: #001529;
  color: white;
  display: flex;
  align-items: center;
  padding: 0 24px;
}

.logo {
  color: white;
  font-size: 20px;
  font-weight: bold;
  margin-right: 24px;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 64px);
}

.content-inner {
  background: white;
  padding: 24px;
  border-radius: 4px;
}
</style>

