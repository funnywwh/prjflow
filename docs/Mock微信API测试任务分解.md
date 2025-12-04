# Mock微信API测试任务分解

## 一、概述

为了完成需要mock微信API的完整测试，需要创建mock对象来模拟微信API调用和WebSocket通信。本文档将任务分解为具体的、可执行的步骤。

## 二、需要Mock的组件

### 1. 微信API客户端（WeChatClient）

**位置**: `backend/pkg/wechat/client.go`

**需要Mock的方法**:
- `GetAccessToken(code string) (*AccessTokenResponse, error)` - 通过code获取access_token
- `GetUserInfo(accessToken, openID string) (*UserInfoResponse, error)` - 获取用户信息
- `GetQRCode(redirectURI string, customState ...string) (*QRCodeResponse, error)` - 获取二维码（可选，主要用于生成URL）

**Mock策略**:
- 方案1：创建接口，使用依赖注入（推荐）
- 方案2：使用httptest.Server模拟HTTP响应
- 方案3：使用gomock或其他mock框架

### 2. WebSocket Hub

**位置**: `backend/internal/websocket/hub.go`

**需要Mock的方法**:
- `GetHub().SendMessage(ticket, messageType, data, message)` - 发送WebSocket消息

**Mock策略**:
- 方案1：创建接口，使用依赖注入
- 方案2：使用测试专用的Hub实现
- 方案3：直接mock websocket包

## 三、任务分解

### 阶段1：创建Mock基础设施

#### 任务1.1：创建WeChatClient接口
- **文件**: `backend/pkg/wechat/client_interface.go`（新建）
- **内容**:
  ```go
  type WeChatClientInterface interface {
      GetAccessToken(code string) (*AccessTokenResponse, error)
      GetUserInfo(accessToken, openID string) (*UserInfoResponse, error)
      GetQRCode(redirectURI string, customState ...string) (*QRCodeResponse, error)
  }
  ```
- **工作量**: 小（30分钟）
- **依赖**: 无

#### 任务1.2：创建MockWeChatClient实现
- **文件**: `backend/tests/unit/mocks/wechat_client_mock.go`（新建）
- **内容**: 实现WeChatClientInterface，提供可配置的返回值
- **功能**:
  - 支持设置GetAccessToken的返回值
  - 支持设置GetUserInfo的返回值
  - 支持设置GetQRCode的返回值
  - 支持模拟错误场景
- **工作量**: 中（1-2小时）
- **依赖**: 任务1.1

#### 任务1.3：创建WebSocket Hub接口
- **文件**: `backend/internal/websocket/hub_interface.go`（新建）
- **内容**:
  ```go
  type HubInterface interface {
      SendMessage(ticket, messageType string, data interface{}, message string)
      // 其他必要方法
  }
  ```
- **工作量**: 小（30分钟）
- **依赖**: 无

#### 任务1.4：创建MockWebSocketHub实现
- **文件**: `backend/tests/unit/mocks/websocket_hub_mock.go`（新建）
- **内容**: 实现HubInterface，记录发送的消息
- **功能**:
  - 记录所有SendMessage调用
  - 提供方法查询是否发送了特定消息
  - 支持验证消息内容
- **工作量**: 中（1小时）
- **依赖**: 任务1.3

### 阶段2：重构现有代码以支持依赖注入

#### 任务2.1：重构WeChatCallbackContext使用接口
- **文件**: `backend/internal/api/wechat_callback.go`
- **修改**:
  - 将`WeChatClient *wechat.WeChatClient`改为`WeChatClient wechat.WeChatClientInterface`
  - 更新所有使用WeChatClient的地方
- **工作量**: 中（1小时）
- **依赖**: 任务1.1
- **风险**: 需要确保所有调用点都更新

#### 任务2.2：重构ProcessWeChatCallback使用接口
- **文件**: `backend/internal/api/wechat_callback.go`
- **修改**:
  - 更新函数签名，接受接口类型
  - 确保所有调用都使用接口
- **工作量**: 小（30分钟）
- **依赖**: 任务2.1

#### 任务2.3：重构WebSocket调用使用接口
- **文件**: 
  - `backend/internal/api/init_callback_handler.go`
  - `backend/internal/api/user_callback_handler.go`
  - `backend/internal/api/wechat_callback.go`
- **修改**:
  - 将websocket.GetHub()改为通过参数传入HubInterface
  - 或者创建全局接口变量，测试时替换
- **工作量**: 中（1-2小时）
- **依赖**: 任务1.3
- **风险**: 需要修改多个文件

### 阶段3：编写完整测试

#### 任务3.1：测试ProcessWeChatCallback成功场景
- **文件**: `backend/tests/unit/wechat_callback_test.go`（新建）
- **测试用例**:
  - 成功获取access_token和用户信息
  - 验证handler.Validate被调用
  - 验证handler.Process被调用
  - 验证返回结果正确
- **工作量**: 中（2小时）
- **依赖**: 任务1.2, 任务1.4, 任务2.2

#### 任务3.2：测试ProcessWeChatCallback错误场景
- **文件**: `backend/tests/unit/wechat_callback_test.go`
- **测试用例**:
  - code为空
  - 微信配置未设置
  - GetAccessToken失败
  - GetUserInfo失败
  - handler.Validate失败
  - handler.Process失败
- **工作量**: 中（2小时）
- **依赖**: 任务3.1

#### 任务3.3：测试InitCallbackHandlerImpl.Process成功场景
- **文件**: `backend/tests/unit/init_callback_handler_test.go`（补充）
- **测试用例**:
  - 成功创建管理员角色
  - 成功创建管理员用户
  - 成功分配管理员角色
  - 成功标记系统已初始化
  - 成功生成Token
  - 成功发送WebSocket消息
- **工作量**: 中（2小时）
- **依赖**: 任务1.2, 任务1.4, 任务2.3

#### 任务3.4：测试InitCallbackHandlerImpl.Process错误场景
- **文件**: `backend/tests/unit/init_callback_handler_test.go`（补充）
- **测试用例**:
  - 创建管理员角色失败
  - 创建管理员用户失败（唯一约束冲突）
  - 分配管理员角色失败
  - 标记系统初始化失败
  - 生成Token失败
  - 事务回滚验证
- **工作量**: 中（2小时）
- **依赖**: 任务3.3

#### 任务3.5：测试AddUserCallbackHandler.Process成功场景
- **文件**: `backend/tests/unit/user_callback_handler_test.go`（补充）
- **测试用例**:
  - 成功创建新用户
  - 成功恢复软删除的用户
  - 成功发送WebSocket消息
  - 用户名唯一性处理（重试机制）
- **工作量**: 中（2小时）
- **依赖**: 任务1.2, 任务1.4, 任务2.3

#### 任务3.6：测试AddUserCallbackHandler.Process错误场景
- **文件**: `backend/tests/unit/user_callback_handler_test.go`（补充）
- **测试用例**:
  - 用户已存在且未删除
  - 创建用户失败（非唯一约束错误）
  - 达到最大重试次数
- **工作量**: 中（1小时）
- **依赖**: 任务3.5

#### 任务3.7：测试InitCallbackHandler.HandleCallback完整流程
- **文件**: `backend/tests/unit/init_callback_test.go`（补充）
- **测试用例**:
  - 完整成功流程（从code到返回HTML）
  - 各种错误场景的HTML返回
- **工作量**: 中（1小时）
- **依赖**: 任务3.1, 任务3.3

### 阶段4：测试覆盖率和文档

#### 任务4.1：运行覆盖率检查
- **操作**: 运行`go test -cover`检查覆盖率
- **目标**: 相关模块覆盖率>80%
- **工作量**: 小（30分钟）
- **依赖**: 阶段3所有任务

#### 任务4.2：补充测试文档
- **文件**: `backend/tests/unit/README.md`（新建）
- **内容**:
  - Mock使用说明
  - 测试编写指南
  - 常见问题
- **工作量**: 小（1小时）
- **依赖**: 阶段3所有任务

## 四、实施建议

### 推荐方案：接口+依赖注入

**优点**:
- 代码清晰，易于测试
- 不依赖外部mock框架
- 符合Go的最佳实践

**缺点**:
- 需要重构现有代码
- 需要创建接口文件

### 备选方案：httptest.Server

**优点**:
- 不需要重构现有代码
- 可以模拟真实的HTTP请求

**缺点**:
- 测试代码较复杂
- 需要维护mock服务器

### 备选方案：gomock

**优点**:
- 自动生成mock代码
- 功能强大

**缺点**:
- 需要额外的工具和依赖
- 生成的代码可能较复杂

## 五、优先级和时间估算

### 高优先级（核心功能）
1. **任务1.1-1.4**: Mock基础设施（4小时）
2. **任务2.1-2.3**: 代码重构（3-4小时）
3. **任务3.1-3.2**: ProcessWeChatCallback测试（4小时）

### 中优先级（重要功能）
4. **任务3.3-3.4**: InitCallbackHandlerImpl.Process测试（4小时）
5. **任务3.5-3.6**: AddUserCallbackHandler.Process测试（3小时）

### 低优先级（完善功能）
6. **任务3.7**: InitCallbackHandler完整流程测试（1小时）
7. **任务4.1-4.2**: 覆盖率和文档（1.5小时）

**总估算时间**: 20-22小时

## 六、实施步骤

### 第一步：创建Mock基础设施（推荐先做）
1. 创建WeChatClient接口
2. 创建MockWeChatClient
3. 创建WebSocket Hub接口
4. 创建MockWebSocketHub

### 第二步：重构代码
1. 重构WeChatCallbackContext
2. 重构ProcessWeChatCallback
3. 重构WebSocket调用

### 第三步：编写测试
1. 先写ProcessWeChatCallback的测试
2. 再写各个Handler的Process方法测试
3. 最后写完整流程测试

### 第四步：验证和文档
1. 运行覆盖率检查
2. 补充测试文档

## 七、注意事项

1. **向后兼容**: 重构时要确保现有功能不受影响
2. **测试隔离**: 每个测试用例应该独立，不相互影响
3. **错误处理**: 要充分测试各种错误场景
4. **WebSocket**: 需要验证消息发送的正确性
5. **事务处理**: 需要验证数据库事务的正确性

## 八、参考资源

- Go接口最佳实践: https://go.dev/doc/effective_go#interfaces
- Go测试文档: https://go.dev/doc/code#Testing
- httptest包文档: https://pkg.go.dev/net/http/httptest

---

**文档版本**: v1.0  
**创建日期**: 2025年12月4日  
**维护者**: 开发团队

