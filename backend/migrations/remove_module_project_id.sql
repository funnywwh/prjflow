-- 迁移脚本：移除 modules 表的 project_id 字段
-- 功能模块改为系统资源，不再属于项目

-- 1. 创建新表（不包含 project_id）
CREATE TABLE `modules_new` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `name` text NOT NULL,
    `code` text,
    `description` text,
    `status` integer DEFAULT 1,
    `sort` integer DEFAULT 0
);

-- 2. 复制数据（忽略 project_id）
INSERT INTO `modules_new` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `code`, `description`, `status`, `sort`)
SELECT `id`, `created_at`, `updated_at`, `deleted_at`, `name`, `code`, `description`, `status`, `sort`
FROM `modules`;

-- 3. 删除旧表
DROP TABLE `modules`;

-- 4. 重命名新表
ALTER TABLE `modules_new` RENAME TO `modules`;

-- 5. 创建索引
CREATE UNIQUE INDEX `idx_modules_name` ON `modules`(`name`);
CREATE UNIQUE INDEX `idx_modules_code` ON `modules`(`code`);
CREATE INDEX `idx_modules_deleted_at` ON `modules`(`deleted_at`);

