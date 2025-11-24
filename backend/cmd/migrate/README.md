# zentao到goproject数据迁移工具

## 功能说明

此工具用于将zentao数据库中的数据迁移到goproject数据库，支持以下数据的迁移：

- 部门 (zt_dept -> departments)
- 角色 (zt_group + zt_grouppriv -> roles + permissions)
- 用户 (zt_user -> users)
- 项目 (zt_project -> projects)
- 需求 (zt_story -> requirements)
- 任务 (zt_task -> tasks)
- Bug (zt_bug -> bugs)

## 使用方法

### 1. 配置文件

复制 `migrate-config.yaml.example` 为 `migrate-config.yaml`，并修改相应的数据库连接信息：

```yaml
zentao:
  host: localhost
  port: 3306
  user: root
  password: "your_password"
  dbname: zentao_db

goproject:
  type: sqlite
  dsn: data.db  # SQLite数据库文件路径（相对于goproject/backend目录）
```

### 2. 运行迁移

在 `goproject/backend` 目录下运行：

```bash
go run cmd/migrate/main.go -config cmd/migrate/migrate-config.yaml
```

或者编译后运行：

```bash
go build -o migrate cmd/migrate/main.go
./migrate -config cmd/migrate/migrate-config.yaml
```

### 3. 迁移顺序

迁移工具会按照以下顺序执行：

1. 部门（需要先迁移，因为用户依赖部门）
2. 角色和权限（需要先迁移，因为用户需要角色）
3. 用户（依赖部门和角色）
4. 项目（依赖用户）
5. 需求（依赖项目和用户）
6. 任务（依赖项目、需求和用户）
7. Bug（依赖项目、需求和用户）

## 数据映射规则

### 部门映射
- `name` -> `name`
- `parent` -> `parent_id` (通过ID映射)
- `grade` -> `level`
- `order` -> `sort`
- `code`: 自动生成（格式：dept_{id}）
- `status`: 默认1（正常）

### 角色映射
- 如果角色名称包含"admin"、"管理员"或"管理"，则映射到goproject的默认admin角色
- 其他角色创建新角色，并根据`zt_grouppriv`表映射权限

### 用户映射
- `account` -> `username`
- `realname` -> `nickname` (如果为空则使用account)
- `email` -> `email`
- `mobile` -> `phone`
- `avatar` -> `avatar`
- `password`: 默认设置为"123"（使用bcrypt加密）
- `dept` -> `department_id` (通过ID映射)
- `deleted` -> `status` (0->1正常, 1->0禁用)

### 项目映射
- `name` -> `name`
- `code` -> `code`
- `desc` -> `description`
- `begin` -> `start_date`
- `end` -> `end_date`
- `status`: 转换（doing->1正常, done/closed->0禁用, wait->1）
- 只迁移 `deleted='0'` 且 `type='sprint'` 或 `type='project'` 的项目

### 需求映射
- `title` -> `title`
- `zt_storyspec.spec` -> `description`
- `status`: 转换（active->in_progress, closed->completed, draft->pending）
- `pri`: 转换（1->urgent, 2->high, 3->medium, 4->low）
- `estimate` -> `estimated_hours` (天转小时，乘以8)

### 任务映射
- `name` -> `title`
- `desc` -> `description`
- `status`: 转换（wait->todo, doing->in_progress, done->done, pause/cancel->cancelled）
- `pri`: 转换（1->urgent, 2->high, 3->medium, 4->low）
- `estimate` -> `estimated_hours` (天转小时)
- `consumed` -> `actual_hours` (天转小时)

### Bug映射
- `title` -> `title`
- `steps` -> `description`
- `status`: 转换（active->open, resolved->resolved, closed->closed）
- `severity`: 转换（1->critical, 2->high, 3->medium, 4->low）
- `pri`: 转换（1->urgent, 2->high, 3->medium, 4->low）
- `resolution` -> `solution`
- `resolvedBuild` -> `solution_note`

## 注意事项

1. **密码重置**: 所有迁移的用户密码都设置为"123"，首次登录后请修改密码
2. **角色映射**: 管理员角色会映射到goproject的默认admin角色，其他角色会创建新角色
3. **ID映射**: 工具会维护zentao ID到goproject ID的映射关系，确保外键关联正确
4. **重复数据**: 如果记录已存在（基于唯一约束），会跳过并记录日志
5. **缺失引用**: 如果引用的记录不存在（如项目、用户等），会跳过该记录并记录日志

## 错误处理

- 迁移过程中会记录详细的日志
- 如果某个记录迁移失败，会记录错误但继续处理其他记录
- 建议在迁移前备份goproject数据库

## 依赖关系

迁移工具依赖以下Go包：
- gorm.io/gorm
- gorm.io/driver/mysql
- gorm.io/driver/sqlite
- gopkg.in/yaml.v3
- project-management/internal/model
- project-management/internal/utils

