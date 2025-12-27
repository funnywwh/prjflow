<template>
  <a-space direction="vertical" style="width: 100%">
    <a-select
      :model-value="modelValue"
      :mode="multiple ? 'multiple' : undefined"
      :placeholder="placeholder"
      :allow-clear="!required"
      show-search
      :filter-option="filterVersionOption"
      :loading="loading"
      :disabled="disabled || createVersion"
      :getPopupContainer="getPopupContainer"
      :dropdownStyle="{ zIndex: 2100 }"
      @update:model-value="handleVersionChange"
      @focus="loadVersions"
    >
      <a-select-option
        v-for="version in versions"
        :key="version.id"
        :value="version.id"
      >
        {{ version.version_number }}
      </a-select-option>
    </a-select>
    <a-checkbox :checked="createVersion" @update:checked="handleCreateVersionChange">
      创建新版本
    </a-checkbox>
    <a-input
      v-if="createVersion"
      :model-value="versionNumber"
      :placeholder="versionNumberPlaceholder"
      @update:model-value="handleVersionNumberChange"
    />
  </a-space>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getVersions, type Version } from '@/api/version'

interface Props {
  modelValue?: number | number[] // 版本ID或版本ID数组
  projectId?: number // 项目ID
  multiple?: boolean // 是否多选
  placeholder?: string // 占位符
  required?: boolean // 是否必填
  disabled?: boolean // 是否禁用
  versionNumberPlaceholder?: string // 版本号输入框占位符
  createVersion?: boolean // 是否创建新版本
  versionNumber?: string // 版本号
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  placeholder: '选择版本',
  required: false,
  disabled: false,
  versionNumberPlaceholder: '请输入版本号（如：v1.0.0）',
  createVersion: false,
  versionNumber: ''
})

const emit = defineEmits<{
  'update:modelValue': [value?: number | number[]]
  'update:createVersion': [value: boolean]
  'update:versionNumber': [value: string]
}>()

const versions = ref<Version[]>([])
const loading = ref(false)

// 加载版本列表
const loadVersions = async () => {
  if (!props.projectId) {
    versions.value = []
    return
  }
  loading.value = true
  try {
    const response = await getVersions({ project_id: props.projectId, size: 1000 })
    versions.value = response.list || []
  } catch (error: any) {
    console.error('加载版本列表失败:', error)
    versions.value = []
  } finally {
    loading.value = false
  }
}

// 版本筛选
const filterVersionOption = (input: string, option: any) => {
  const version = versions.value.find(v => v.id === option.value)
  if (!version) return false
  const searchText = input.toLowerCase()
  return version.version_number.toLowerCase().includes(searchText)
}

// 获取下拉框容器（用于解决模态框中下拉框被遮挡的问题）
const getPopupContainer = (triggerNode: HTMLElement): HTMLElement => {
  return triggerNode.parentElement || document.body
}

// 处理版本变化
const handleVersionChange = (value: number | number[] | undefined) => {
  emit('update:modelValue', value)
}

// 处理创建新版本状态变化
const handleCreateVersionChange = (checked: boolean) => {
  emit('update:createVersion', checked)
  // 如果取消创建新版本，清空版本号
  if (!checked) {
    emit('update:versionNumber', '')
  }
}

// 处理版本号变化
const handleVersionNumberChange = (value: string) => {
  emit('update:versionNumber', value)
}

// 监听项目ID变化，重新加载版本列表
watch(() => props.projectId, (newProjectId) => {
  if (newProjectId) {
    loadVersions()
  } else {
    versions.value = []
  }
}, { immediate: true })

// 组件挂载时加载版本列表
onMounted(() => {
  if (props.projectId) {
    loadVersions()
  }
})
</script>

