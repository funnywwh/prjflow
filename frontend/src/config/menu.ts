// 菜单配置
export interface MenuItem {
  key: string
  title: string
  icon?: string
  path?: string
  permission?: string | string[] // 权限代码，可以是单个或多个（任一即可）
  children?: MenuItem[]
  order?: number // 排序
}

// 菜单配置列表
export const menuConfig: MenuItem[] = [
  {
    key: 'dashboard',
    title: '工作台',
    icon: 'DashboardOutlined',
    path: '/dashboard'
    // 工作台不需要权限，所有登录用户都可以访问
  },
  {
    key: 'project-management',
    title: '项目管理',
    icon: 'ProjectOutlined',
    permission: 'project:read',
    children: [
      {
        key: 'project',
        title: '项目列表',
        path: '/project',
        permission: 'project:read'
      },
      {
        key: 'requirement',
        title: '需求管理',
        path: '/requirement',
        permission: 'requirement:read'
      },
      {
        key: 'task',
        title: '任务管理',
        path: '/task',
        permission: 'task:read'
      },
      {
        key: 'bug',
        title: 'Bug管理',
        path: '/bug',
        permission: 'bug:read'
      },
      {
        key: 'version',
        title: '版本管理',
        path: '/version',
        permission: 'project:read' // 版本管理使用项目权限
      },
      {
        key: 'test-case',
        title: '测试管理',
        path: '/test-case',
        permission: 'project:read' // 测试管理使用项目权限
      },
      {
        key: 'product',
        title: '产品',
        path: '/product',
        permission: 'project:read' // 产品管理使用项目权限
      }
    ]
  },
  {
    key: 'resource-management',
    title: '资源管理',
    icon: 'TeamOutlined',
    permission: 'resource:read',
    children: [
      {
        key: 'resource-statistics',
        title: '资源统计',
        path: '/resource/statistics',
        permission: 'resource:read'
      }
    ]
  },
  {
    key: 'system-management',
    title: '系统管理',
    icon: 'SettingOutlined',
    permission: ['user:read', 'department:read', 'permission:manage'], // 任一权限即可显示
    children: [
      {
        key: 'user',
        title: '用户管理',
        path: '/user',
        permission: 'user:read'
      },
      {
        key: 'department',
        title: '部门管理',
        path: '/department',
        permission: 'department:read'
      },
      {
        key: 'permission',
        title: '权限管理',
        path: '/permission',
        permission: 'permission:manage'
      },
      {
        key: 'wechat-settings',
        title: '微信设置',
        path: '/system/wechat-settings',
        permission: 'permission:manage' // 只有管理员可以配置微信
      }
    ]
  }
]

// 根据权限过滤菜单
export function filterMenuByPermission(menuItems: MenuItem[], hasPermission: (code: string) => boolean): MenuItem[] {
  return menuItems
    .filter(item => {
      // 如果没有权限要求，则显示
      if (!item.permission) {
        return true
      }
      
      // 如果是数组，检查是否有任一权限
      if (Array.isArray(item.permission)) {
        return item.permission.some(code => hasPermission(code))
      }
      
      // 如果是字符串，检查是否有该权限
      return hasPermission(item.permission)
    })
    .map(item => {
      // 如果有子菜单，递归过滤
      if (item.children && item.children.length > 0) {
        const filteredChildren = filterMenuByPermission(item.children, hasPermission)
        // 如果过滤后还有子菜单，则保留父菜单
        if (filteredChildren.length > 0) {
          return {
            ...item,
            children: filteredChildren
          }
        }
        // 如果过滤后没有子菜单，但父菜单有路径，则保留父菜单（作为单级菜单）
        if (item.path) {
          return {
            ...item,
            children: undefined
          }
        }
        // 如果过滤后没有子菜单且父菜单没有路径，则不显示
        return null
      }
      return item
    })
    .filter((item): item is MenuItem => item !== null)
    .sort((a, b) => (a.order || 0) - (b.order || 0))
}

