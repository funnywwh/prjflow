# 前端项目

基于 Vue 3 + TypeScript + Vite + Ant Design Vue 的项目管理系统前端。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **UI库**: Ant Design Vue (v4.2.6)
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios
- **日期处理**: dayjs
- **图表**: ECharts + vue-echarts
- **Markdown**: marked + highlight.js

## 开发指南

### 安装依赖

```bash
npm install
# 或使用 yarn
yarn install
```

### 启动开发服务器

```bash
npm run dev
# 或使用 yarn
yarn dev
```

开发服务器会在 `http://localhost:5173` 启动，通过 Vite 代理转发 API 请求到后端。

### 构建生产版本

```bash
npm run build
# 或使用 yarn
yarn build
```

构建后的文件在 `dist` 目录，后端会自动服务这些静态文件。

### TypeScript 类型检查

```bash
npm run type-check
# 或使用 yarn
yarn type-check
```

## 项目结构

```
frontend/src/
├── api/                # API接口定义
├── views/              # 页面组件
│   ├── auth/          # 登录页面
│   ├── init/          # 系统初始化页面
│   ├── dashboard/     # 工作台
│   ├── project/       # 项目管理
│   ├── requirement/   # 需求管理
│   ├── bug/           # Bug管理
│   ├── task/          # 任务管理
│   ├── version/       # 版本管理
│   └── ...
├── components/         # 公共组件
│   ├── AppHeader.vue  # 顶部导航栏
│   ├── MarkdownEditor.vue  # Markdown编辑器
│   └── WeChatQRCode.vue    # 微信二维码组件
├── stores/            # Pinia状态管理
│   ├── auth.ts        # 认证状态
│   └── permission.ts   # 权限状态
├── router/            # 路由配置
└── utils/             # 工具函数
    ├── request.ts     # HTTP请求封装
    └── date.ts        # 日期处理
```

## 主要功能

- ✅ 系统初始化（微信扫码和密码登录两种方式）
- ✅ 用户登录（微信登录和密码登录，默认微信登录）
- ✅ 个人工作台（聚合显示用户相关事务、统计信息）
- ✅ 项目管理（项目列表、详情、统计、成员管理、标签管理）
- ✅ 需求管理（需求列表、创建、编辑、详情、统计）
- ✅ Bug管理（Bug列表、创建、编辑、详情、分配、统计）
- ✅ 任务管理（任务列表、创建、编辑、详情、进度管理、依赖关系、工时管理）
- ✅ 版本管理（版本列表、创建、编辑、关联需求和Bug）
- ✅ 测试管理（测试单、测试报告）
- ✅ 资源管理（资源统计、资源分配、日历视图）
- ✅ 工作报告（日报、周报，支持Markdown）
- ✅ 附件管理（文件上传、下载、关联实体）
- ✅ 功能模块管理（模块CRUD）
- ✅ 看板视图（任务看板、拖拽排序）
- ✅ 甘特图（任务时间线、依赖关系）
- ✅ 进度跟踪（统计图表、进度报表）

## 开发规范

1. **组件结构**: 使用 `<script setup>` 语法（Composition API）
2. **类型定义**: 所有 API 接口都有完整的 TypeScript 类型定义
3. **API调用**: 使用统一的 `request` 实例，自动处理错误和认证
4. **日期处理**: 统一使用 `dayjs` 进行日期格式化
5. **响应格式**: API 响应自动提取 `data` 字段

## 注意事项

- API 参数使用 `size` 而不是 `page_size`（已统一）
- 日期字段在表单中使用 `Dayjs` 类型，提交时自动转换为字符串
- 标签使用独立表和关联表，不是 JSON 数组
- 所有 TypeScript 类型错误已修复，构建通过
