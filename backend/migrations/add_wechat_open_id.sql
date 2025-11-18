-- 为 users 表添加 wechat_open_id 字段（如果不存在）
-- SQLite 版本
-- 注意：SQLite 不支持直接添加 UNIQUE 约束的列，需要重建表

-- 如果使用 MySQL，可以使用以下语句：
-- ALTER TABLE users ADD COLUMN wechat_open_id VARCHAR(100) UNIQUE;

-- SQLite 需要手动重建表（如果字段不存在）
-- 1. 创建新表
-- CREATE TABLE users_new (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     created_at DATETIME,
--     updated_at DATETIME,
--     deleted_at DATETIME,
--     wechat_open_id VARCHAR(100) UNIQUE,
--     username VARCHAR(50) NOT NULL,
--     email VARCHAR(100),
--     avatar VARCHAR(255),
--     phone VARCHAR(20),
--     status INTEGER DEFAULT 1,
--     department_id INTEGER
-- );

-- 2. 复制数据
-- INSERT INTO users_new (id, created_at, updated_at, deleted_at, username, email, avatar, phone, status, department_id)
-- SELECT id, created_at, updated_at, deleted_at, username, email, avatar, phone, status, department_id FROM users;

-- 3. 删除旧表
-- DROP TABLE users;

-- 4. 重命名新表
-- ALTER TABLE users_new RENAME TO users;

-- 建议：直接重启后端服务，让 GORM 的 AutoMigrate 自动处理

