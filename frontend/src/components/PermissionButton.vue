<template>
  <component v-if="hasPermission" :is="tag" v-bind="$attrs">
    <slot />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { usePermissionStore } from '@/stores/permission'

interface Props {
  /**
   * 权限代码，可以是单个权限或权限数组
   * 如果是数组，只要有一个权限即可显示
   */
  permission: string | string[]
  /**
   * 渲染的标签，默认为 'div'
   * 可以是 'div', 'span', 'a-button' 等
   */
  tag?: string
}

const props = withDefaults(defineProps<Props>(), {
  tag: 'div'
})

const permissionStore = usePermissionStore()

// 检查是否有权限
const hasPermission = computed(() => {
  const permissionCodes = Array.isArray(props.permission)
    ? props.permission
    : [props.permission]
  
  return permissionCodes.some(code => permissionStore.hasPermission(code))
})
</script>

