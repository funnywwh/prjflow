# 项目管理系统

基于Go + Gin + GORM和Vue3 + TypeScript + Ant Design Vue的全栈项目管理软件。

## 技术栈

- **后端**: Go + Gin + GORM + SQLite（支持MySQL）
- **前端**: Vue 3 + TypeScript + Ant Design Vue + Vite
- **认证**: 微信开放平台扫码登录
- **测试**: TDD开发，单元测试覆盖率目标100%

## 项目结构

```
project/
├── backend/          # Go后端
│   ├── cmd/server/  # 应用入口
│   ├── internal/    # 内部包
│   │   ├── api/     # API路由层
│   │   ├── service/ # 业务逻辑层
│   │   ├── model/   # 数据模型
│   │   ├── repository/ # 数据访问层
│   │   ├── middleware/ # 中间件
│   │   └── utils/   # 工具函数
│   ├── pkg/         # 可复用包
│   └── migrations/  # 数据库迁移
└── frontend/        # Vue3前端
    └── src/
        ├── api/     # API接口
        ├── views/   # 页面组件
        ├── components/ # 公共组件
        ├── stores/  # 状态管理
        └── router/  # 路由配置
```

## 已完成功能

### 后端
- ✅ 项目初始化和基础架构
- ✅ 数据库模型设计（所有表结构）
- ✅ 微信登录集成
- ✅ JWT认证机制
- ✅ RBAC权限系统
- ✅ 用户管理API
  - ✅ 用户CRUD操作
  - ✅ 用户昵称功能（username用于登录，nickname用于显示）
  - ✅ 扫码添加用户功能
  - ✅ 软删除用户恢复机制
  - ✅ 并发创建用户冲突处理
- ✅ 部门管理API
- ✅ 角色和权限管理API
- ✅ SQLite数据库迁移优化（支持NOT NULL字段添加）

### 前端
- ✅ 项目初始化和基础配置
- ✅ 路由配置和守卫
- ✅ 状态管理（Pinia）
- ✅ 登录页面框架
- ✅ 工作台页面框架
- ✅ API请求封装
- ✅ 用户管理页面
  - ✅ 用户列表展示（显示格式：username(nickname)）
  - ✅ 用户编辑功能（支持昵称编辑）
  - ✅ 扫码添加用户功能
  - ✅ 添加用户后昵称设置对话框

## 待开发功能

根据计划，以下功能模块需要继续开发：

1. **产品和项目管理** - 产品线、产品、项目集、项目的CRUD和关联关系
2. **需求管理和Bug追踪** - 需求关联产品/项目、Bug状态流转、分配功能
3. **任务管理和看板系统** - 任务CRUD、看板拖拽、项目进度跟踪、甘特图
4. **计划管理** - 产品计划、项目计划、计划执行分解
5. **版本和构建管理** - 构建创建、版本生成
6. **测试管理** - 测试单、测试报告、测试单关联Bug和测试报告
7. **人员资源管理** - 资源分配、统计功能（按小时管理）
8. **工作报告系统** - 日报、周报（支持Markdown）
9. **插件管理系统** - 插件安装、配置、前端界面支持、插件市场
10. **关系图生成** - 数据库ER图、业务关系图
11. **个人工作台** - 聚合显示用户相关事务

## 开发指南

### 后端开发

1. 启动开发服务器：
```bash
cd backend
go run cmd/server/main.go
```

2. 运行测试：
```bash
go test ./...
```

3. 数据库迁移：
数据库会在启动时自动迁移（通过GORM AutoMigrate）

### 前端开发

1. 安装依赖：
```bash
cd frontend
npm install
```

2. 启动开发服务器：
```bash
npm run dev
```

3. 构建生产版本：
```bash
npm run build
```

## 配置说明

### 后端配置

编辑 `backend/config.yaml`：
- 数据库配置（SQLite或MySQL）
- JWT密钥
- 微信开放平台AppID和AppSecret
- 服务器端口等

### 前端配置

编辑 `frontend/.env`：
- API基础URL

## API文档

API采用RESTful风格，主要端点：

- `/auth/*` - 认证相关
- `/permissions/*` - 权限管理
- `/users/*` - 用户管理
- `/departments/*` - 部门管理

## 开发规范

1. **TDD开发**：先写测试，再实现功能
2. **代码规范**：遵循Go和TypeScript最佳实践
3. **数据库兼容**：使用GORM抽象层，支持SQLite和MySQL
4. **Markdown支持**：需求、Bug、任务等模块支持Markdown格式

## 注意事项

- 数据库模型已全部定义，但部分功能模块的API和前端页面待实现
- 微信登录需要配置真实的AppID和AppSecret
- 生产环境请修改JWT密钥和数据库配置

## 下一步

按照计划继续实现剩余功能模块，每个模块遵循TDD开发流程：
1. 编写详细设计文档
2. 编写单元测试
3. 实现功能
4. 确保测试覆盖率100%

