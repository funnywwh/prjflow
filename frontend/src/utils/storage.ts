/**
 * localStorage 工具函数
 * 用于保存和恢复用户的选择（项目、标签等）
 */

/**
 * 保存最后选择的值
 * @param key 存储的键名
 * @param value 要保存的值（会自动序列化为 JSON）
 */
export function saveLastSelected(key: string, value: any): void {
  try {
    const serialized = JSON.stringify(value)
    localStorage.setItem(key, serialized)
  } catch (error) {
    console.error('保存选择失败:', error)
  }
}

/**
 * 获取最后选择的值
 * @param key 存储的键名
 * @param defaultValue 默认值（如果不存在或解析失败时返回）
 * @returns 保存的值或默认值
 */
export function getLastSelected<T>(key: string, defaultValue?: T): T | undefined {
  try {
    const item = localStorage.getItem(key)
    if (item === null) {
      return defaultValue
    }
    return JSON.parse(item) as T
  } catch (error) {
    console.error('读取选择失败:', error)
    return defaultValue
  }
}

/**
 * 删除保存的选择
 * @param key 存储的键名
 */
export function removeLastSelected(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch (error) {
    console.error('删除选择失败:', error)
  }
}

/**
 * 清除所有保存的选择（可选，用于调试）
 */
export function clearAllSelected(): void {
  try {
    const keys = Object.keys(localStorage)
    keys.forEach(key => {
      if (key.startsWith('last_selected_')) {
        localStorage.removeItem(key)
      }
    })
  } catch (error) {
    console.error('清除选择失败:', error)
  }
}











