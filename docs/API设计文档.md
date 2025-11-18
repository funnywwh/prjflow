# API设计文档

## API规范

### 基础规范

- **协议**: HTTP/HTTPS
- **格式**: JSON
- **编码**: UTF-8
- **风格**: RESTful

### 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 状态码定义

- `200`: 成功
- `400`: 参数错误
- `401`: 未授权
- `403`: 没有权限
- `404`: 资源不存在
- `500`: 服务器错误

### 分页格式

请求参数：
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）

响应格式：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```

## 认证相关API

### 获取微信登录二维码

**GET** `/auth/wechat/qrcode`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "ticket": "ticket_string",
    "qr_code_url": "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=...",
    "expire_seconds": 300
  }
}
```

### 微信登录

**POST** `/auth/wechat/login`

**请求体**:
```json
{
  "code": "wechat_auth_code",
  "state": "optional_state"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "jwt_token_string",
    "user": {
      "id": 1,
      "username": "用户名",
      "email": "email@example.com",
      "avatar": "avatar_url",
      "roles": ["admin", "user"]
    }
  }
}
```

### 获取当前用户信息

**GET** `/auth/user/info`

**Headers**: `Authorization: Bearer {token}`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "用户名",
    "email": "email@example.com",
    "avatar": "avatar_url",
    "phone": "13800138000",
    "department": {
      "id": 1,
      "name": "技术部"
    },
    "roles": ["admin", "user"]
  }
}
```

### 登出

**POST** `/auth/logout`

**Headers**: `Authorization: Bearer {token}`

## 权限管理API

### 获取角色列表

**GET** `/permissions/roles`

### 创建角色

**POST** `/permissions/roles`

**请求体**:
```json
{
  "name": "角色名称",
  "code": "role_code",
  "description": "角色描述",
  "status": 1
}
```

### 更新角色

**PUT** `/permissions/roles/:id`

### 删除角色

**DELETE** `/permissions/roles/:id`

### 获取权限列表

**GET** `/permissions/permissions`

### 创建权限

**POST** `/permissions/permissions`

**请求体**:
```json
{
  "code": "permission_code",
  "name": "权限名称",
  "resource": "resource_type",
  "action": "action_type",
  "description": "权限描述"
}
```

### 分配角色权限

**POST** `/permissions/roles/:id/permissions`

**请求体**:
```json
{
  "permission_ids": [1, 2, 3]
}
```

### 分配用户角色

**POST** `/permissions/users/:id/roles`

**请求体**:
```json
{
  "role_ids": [1, 2]
}
```

## 用户管理API

### 获取用户列表

**GET** `/users?keyword=&department_id=&page=1&page_size=20`

### 获取用户详情

**GET** `/users/:id`

### 创建用户

**POST** `/users`

**请求体**:
```json
{
  "username": "用户名",
  "email": "email@example.com",
  "phone": "13800138000",
  "department_id": 1,
  "status": 1
}
```

### 更新用户

**PUT** `/users/:id`

### 删除用户

**DELETE** `/users/:id`

## 部门管理API

### 获取部门列表（树形结构）

**GET** `/departments`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "技术部",
      "code": "tech",
      "parent_id": null,
      "level": 1,
      "children": [
        {
          "id": 2,
          "name": "前端组",
          "parent_id": 1,
          "level": 2,
          "children": []
        }
      ]
    }
  ]
}
```

### 获取部门详情

**GET** `/departments/:id`

### 创建部门

**POST** `/departments`

**请求体**:
```json
{
  "name": "部门名称",
  "code": "dept_code",
  "parent_id": 1,
  "sort": 0,
  "status": 1
}
```

### 更新部门

**PUT** `/departments/:id`

### 删除部门

**DELETE** `/departments/:id`

## 产品管理API（待实现）

### 获取产品线列表

**GET** `/product-lines`

### 创建产品线

**POST** `/product-lines`

### 获取产品列表

**GET** `/products?product_line_id=&keyword=&page=1&page_size=20`

### 创建产品

**POST** `/products`

### 关联产品到项目

**POST** `/products/:id/projects`

**请求体**:
```json
{
  "project_ids": [1, 2, 3]
}
```

## 项目管理API（待实现）

### 获取项目集列表

**GET** `/project-groups`

### 获取项目列表

**GET** `/projects?project_group_id=&product_id=&keyword=&page=1&page_size=20`

### 创建项目

**POST** `/projects`

**请求体**:
```json
{
  "name": "项目名称",
  "code": "project_code",
  "description": "项目描述",
  "project_group_id": 1,
  "product_id": 1,
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "status": 1
}
```

### 添加项目成员

**POST** `/projects/:id/members`

**请求体**:
```json
{
  "user_ids": [1, 2, 3],
  "role": "member"
}
```

## 需求管理API ✅

### 获取需求列表

**GET** `/requirements?product_id=&project_id=&status=&priority=&assignee_id=&creator_id=&keyword=&page=1&page_size=20`

### 获取需求统计

**GET** `/requirements/statistics?product_id=&project_id=&status=&keyword=`

### 获取需求详情

**GET** `/requirements/:id`

### 创建需求

**POST** `/requirements`

**请求体**:
```json
{
  "title": "需求标题",
  "description": "需求描述（Markdown）",
  "product_id": 1,
  "project_id": 1,
  "status": "pending",
  "priority": "high",
  "assignee_id": 1
}
```

### 更新需求

**PUT** `/requirements/:id`

### 删除需求

**DELETE** `/requirements/:id`

### 更新需求状态

**PATCH** `/requirements/:id/status`

**请求体**:
```json
{
  "status": "in_progress"
}
```

## Bug管理API ✅

### 获取Bug列表

**GET** `/bugs?project_id=&status=&priority=&severity=&requirement_id=&creator_id=&keyword=&page=1&page_size=20`

### 获取Bug统计

**GET** `/bugs/statistics?project_id=&requirement_id=&keyword=`

### 获取Bug详情

**GET** `/bugs/:id`

### 创建Bug

**POST** `/bugs`

**请求体**:
```json
{
  "title": "Bug标题",
  "description": "Bug描述（Markdown）",
  "project_id": 1,
  "status": "open",
  "priority": "high",
  "severity": "critical",
  "requirement_id": 1,
  "assignee_ids": [1, 2]
}
```

### 更新Bug

**PUT** `/bugs/:id`

### 删除Bug

**DELETE** `/bugs/:id`

### 更新Bug状态

**PATCH** `/bugs/:id/status`

**请求体**:
```json
{
  "status": "in_progress"
}
```

### 分配Bug

**POST** `/bugs/:id/assign`

**请求体**:
```json
{
  "assignee_ids": [1, 2]
}
```

## 任务管理API（已实现）

### 获取任务列表

**GET** `/tasks?project_id=&status=&assignee_id=&keyword=&page=1&page_size=20`

### 创建任务

**POST** `/tasks`

**请求体**:
```json
{
  "title": "任务标题",
  "description": "任务描述（Markdown）",
  "project_id": 1,
  "status": "todo",
  "priority": "medium",
  "assignee_id": 1,
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "dependency_ids": [1, 2]
}
```

### 更新任务进度

**PATCH** `/tasks/:id/progress`

**请求体**:
```json
{
  "progress": 50
}
```

## 看板API（已实现）

### 获取项目看板

**GET** `/projects/:id/boards`

### 创建看板

**POST** `/projects/:id/boards`

### 拖拽任务到不同列

**PATCH** `/boards/:id/tasks/:task_id/move`

**请求体**:
```json
{
  "column_id": 2,
  "position": 0
}
```

## 甘特图API（已实现）

### 获取项目甘特图数据

**GET** `/projects/:id/gantt`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tasks": [
      {
        "id": 1,
        "title": "任务1",
        "start_date": "2024-01-01",
        "end_date": "2024-01-15",
        "progress": 50,
        "dependencies": [2]
      }
    ]
  }
}
```

## 项目进度跟踪API（已实现）

### 获取项目进度跟踪数据

**GET** `/projects/:id/progress`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "statistics": {
      "total_tasks": 10,
      "todo_tasks": 3,
      "in_progress_tasks": 4,
      "done_tasks": 3,
      "total_bugs": 5,
      "open_bugs": 2,
      "in_progress_bugs": 1,
      "resolved_bugs": 2,
      "total_requirements": 8,
      "in_progress_requirements": 3,
      "completed_requirements": 5,
      "total_members": 6
    },
    "task_progress_trend": [
      {
        "date": "2024-01-01",
        "average": 45.5,
        "count": 2
      }
    ],
    "task_status_distribution": [
      {
        "status": "todo",
        "count": 3
      },
      {
        "status": "in_progress",
        "count": 4
      },
      {
        "status": "done",
        "count": 3
      }
    ],
    "task_priority_distribution": [
      {
        "priority": "low",
        "count": 2
      },
      {
        "priority": "medium",
        "count": 5
      },
      {
        "priority": "high",
        "count": 3
      }
    ],
    "task_completion_trend": [
      {
        "week": "2024-01-01",
        "total": 10,
        "completed": 2,
        "completion_rate": 20.0
      }
    ],
    "member_workload": [
      {
        "user_id": 1,
        "username": "user1",
        "nickname": "用户1",
        "total": 5,
        "completed": 2,
        "in_progress": 2
      }
    ],
    "bug_trend": [
      {
        "date": "2024-01-01",
        "count": 2
      }
    ],
    "requirement_trend": [
      {
        "date": "2024-01-01",
        "count": 1
      }
    ]
  }
}
```

## 构建管理API ✅

### 获取构建列表

**GET** `/builds?project_id=&status=&branch=&creator_id=&keyword=&page=1&page_size=20`

### 获取构建详情

**GET** `/builds/:id`

### 创建构建

**POST** `/builds`

**请求体**:
```json
{
  "build_number": "build-2024-01-01-001",
  "status": "pending",
  "branch": "main",
  "commit": "abc123def456",
  "build_time": "2024-01-01 10:00:00",
  "project_id": 1
}
```

### 更新构建

**PUT** `/builds/:id`

### 删除构建

**DELETE** `/builds/:id`

### 更新构建状态

**PATCH** `/builds/:id/status`

**请求体**:
```json
{
  "status": "building"
}
```

**状态值**:
- `pending`: 待构建
- `building`: 构建中
- `success`: 成功
- `failed`: 失败

## 版本管理API ✅

### 获取版本列表

**GET** `/versions?build_id=&project_id=&status=&keyword=&page=1&page_size=20`

### 获取版本详情

**GET** `/versions/:id`

### 创建版本

**POST** `/versions`

**请求体**:
```json
{
  "version_number": "v1.0.0",
  "release_notes": "发布说明（Markdown）",
  "status": "draft",
  "build_id": 1,
  "release_date": "2024-01-01",
  "requirement_ids": [1, 2],
  "bug_ids": [3, 4]
}
```

### 更新版本

**PUT** `/versions/:id`

### 删除版本

**DELETE** `/versions/:id`

### 更新版本状态

**PATCH** `/versions/:id/status`

**请求体**:
```json
{
  "status": "released"
}
```

**状态值**:
- `draft`: 草稿
- `released`: 已发布
- `archived`: 已归档

### 发布版本

**POST** `/versions/:id/release`

**说明**: 只有构建状态为`success`的版本才能发布，发布后状态自动变为`released`，并自动设置发布日期。

## 个人工作台API（待实现）

### 获取工作台数据

**GET** `/dashboard`

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tasks": {
      "todo": 5,
      "in_progress": 3,
      "done": 10
    },
    "bugs": {
      "open": 2,
      "in_progress": 1,
      "resolved": 5
    },
    "projects": [
      {
        "id": 1,
        "name": "项目1",
        "role": "owner"
      }
    ],
    "reports": {
      "pending": 2,
      "submitted": 1
    },
    "statistics": {
      "total_tasks": 18,
      "total_bugs": 8,
      "total_projects": 3,
      "week_hours": 40
    }
  }
}
```

## 错误处理

所有API错误统一返回格式：

```json
{
  "code": 400,
  "message": "错误描述"
}
```

常见错误码：
- `400`: 参数错误
- `401`: 未授权，需要登录
- `403`: 没有权限
- `404`: 资源不存在
- `500`: 服务器内部错误

## 认证方式

所有需要认证的API都需要在请求头中携带Token：

```
Authorization: Bearer {jwt_token}
```

## 版本控制

未来如果需要API版本控制，可以在URL中添加版本号：

```
/api/v1/users
/api/v2/users
```

---

**文档版本**: v1.0  
**最后更新**: 2024年

