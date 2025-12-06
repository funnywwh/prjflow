<template>
  <div>
    <a-select
      :key="`select-${projectId}-${selectKey}-${members.length}`"
      v-model:value="internalDisplayValue"
      :mode="multiple ? 'multiple' : undefined"
      :placeholder="placeholder"
      :allow-clear="allowClear"
      :show-search="showSearch"
      :filter-option="filterOption"
      :loading="loading"
      :disabled="disabled || !projectId"
      :style="style"
      :get-popup-container="getPopupContainer"
      @change="handleChange"
    >
      <a-select-option
        v-for="member in members"
        :key="member.user_id"
        :value="member.user_id"
      >
        {{ member.user?.username || '' }}{{ member.user?.nickname ? `(${member.user.nickname})` : '' }}
        <span v-if="showRole && member.role" style="color: #999; margin-left: 4px">
          ({{ getRoleText(member.role) }})
        </span>
      </a-select-option>
    </a-select>
    <div v-if="!projectId && showHint" style="color: #999; margin-top: 4px; font-size: 12px">
      {{ hintText || '请先选择项目' }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed, nextTick } from 'vue'
import { getProjectMembers, type ProjectMember } from '@/api/project'

interface Props {
  modelValue?: number | number[] | undefined
  projectId?: number | undefined
  multiple?: boolean
  placeholder?: string
  allowClear?: boolean
  showSearch?: boolean
  disabled?: boolean
  style?: Record<string, any>
  showRole?: boolean
  showHint?: boolean
  hintText?: string
  getPopupContainer?: (triggerNode: HTMLElement) => HTMLElement
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  placeholder: '选择项目成员',
  allowClear: true,
  showSearch: true,
  disabled: false,
  style: () => ({}),
  showRole: true,
  showHint: true,
  hintText: '',
  getPopupContainer: undefined
})

const emit = defineEmits<{
  'update:modelValue': [value: number | number[] | undefined]
  'change': [value: number | number[] | undefined]
}>()

const members = ref<ProjectMember[]>([])
const loading = ref(false)
const selectKey = ref(0)
const internalDisplayValue = ref<number | number[] | undefined>(undefined)

// 计算显示值：当成员列表加载完成且值在列表中时，才显示值
const displayValue = computed(() => {
  if (props.modelValue === undefined) {
    return undefined
  }
  
  // 如果成员列表还没加载完成，返回 undefined（不显示选中值）
  if (members.value.length === 0) {
    return undefined
  }
  
  // 检查值是否在成员列表中
  const valueExists = Array.isArray(props.modelValue)
    ? props.modelValue.every(v => members.value.some(m => m.user_id === v))
    : members.value.some(m => m.user_id === props.modelValue)
  
  if (valueExists) {
    return props.modelValue
  }
  
  return undefined
})

// 加载项目成员列表
const loadMembers = async (projectId: number | undefined) => {
  if (!projectId) {
    members.value = []
    loading.value = false
    return
  }
  
  loading.value = true
  try {
    members.value = await getProjectMembers(projectId)
  } catch (error: any) {
    console.error('加载项目成员失败:', error)
    members.value = []
  } finally {
    loading.value = false
  }
}

// 筛选函数
const filterOption = (input: string, option: any) => {
  const member = members.value.find(m => m.user_id === option.value)
  if (!member || !member.user) return false
  const searchText = input.toLowerCase()
  return (
    member.user.username.toLowerCase().includes(searchText) ||
    (member.user.nickname && member.user.nickname.toLowerCase().includes(searchText))
  )
}

// 获取角色文本
const getRoleText = (role: string): string => {
  const roleMap: Record<string, string> = {
    owner: '负责人',
    member: '成员',
    viewer: '查看者'
  }
  return roleMap[role] || role
}

// 处理值变化
const handleChange = (value: number | number[] | undefined) => {
  internalDisplayValue.value = value
  emit('update:modelValue', value)
  emit('change', value)
}

// 监听 displayValue 变化，同步到内部状态
watch(displayValue, (newValue) => {
  if (newValue !== internalDisplayValue.value) {
    internalDisplayValue.value = newValue
  }
}, { immediate: true })

// 监听项目ID变化
watch(() => props.projectId, (newProjectId) => {
  loadMembers(newProjectId)
}, { immediate: true })

// 监听 modelValue 变化
watch(() => props.modelValue, (newValue) => {
  // 当 modelValue 变化且成员列表已加载完成时，强制更新 selectKey
  if (newValue !== undefined && members.value.length > 0 && !loading.value) {
    const valueExists = Array.isArray(newValue)
      ? newValue.every(v => members.value.some(m => m.user_id === v))
      : members.value.some(m => m.user_id === newValue)
    
    if (valueExists) {
      nextTick(() => {
        selectKey.value++
      })
    }
  }
}, { immediate: true })

// 监听成员列表变化
watch(() => members.value, (newMembers) => {
  // 当成员列表加载完成且 modelValue 已设置时，强制更新 selectKey 以触发重新渲染
  if (newMembers.length > 0 && props.modelValue !== undefined && !loading.value) {
    const valueExists = Array.isArray(props.modelValue)
      ? props.modelValue.every(v => newMembers.some(m => m.user_id === v))
      : newMembers.some(m => m.user_id === props.modelValue)
    
    if (valueExists) {
      nextTick(() => {
        selectKey.value++
      })
    }
  }
}, { deep: true })

// 组件挂载时加载
onMounted(() => {
  if (props.projectId) {
    loadMembers(props.projectId)
  }
})
</script>

<style scoped>
/* 组件样式 */
</style>
