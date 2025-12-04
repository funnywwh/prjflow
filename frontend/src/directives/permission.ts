/**
 * 权限指令
 * 用于控制按钮、链接等元素的显示/隐藏
 * 
 * 使用方式：
 * <a-button v-permission="'project:create'">创建项目</a-button>
 * <a-button v-permission="['project:create', 'project:update']">编辑项目</a-button>
 */

import type { Directive, DirectiveBinding } from 'vue'
import { usePermissionStore } from '@/stores/permission'

/**
 * 权限指令实现
 */
const permissionDirective: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
    const permissionStore = usePermissionStore()
    
    // 获取权限代码（支持字符串或数组）
    const permissionCodes = Array.isArray(binding.value) 
      ? binding.value 
      : [binding.value]
    
    // 检查是否有权限
    const hasPermission = permissionCodes.some(code => 
      permissionStore.hasPermission(code)
    )
    
    // 如果没有权限，隐藏元素
    if (!hasPermission) {
      el.style.display = 'none'
      // 保存原始display值，以便后续恢复
      ;(el as any).__originalDisplay = el.style.display || ''
    }
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
    const permissionStore = usePermissionStore()
    
    // 获取权限代码（支持字符串或数组）
    const permissionCodes = Array.isArray(binding.value) 
      ? binding.value 
      : [binding.value]
    
    // 检查是否有权限
    const hasPermission = permissionCodes.some(code => 
      permissionStore.hasPermission(code)
    )
    
    // 根据权限显示/隐藏元素
    if (!hasPermission) {
      el.style.display = 'none'
    } else {
      // 恢复原始display值
      const originalDisplay = (el as any).__originalDisplay
      el.style.display = originalDisplay || ''
    }
  }
}

export default permissionDirective

