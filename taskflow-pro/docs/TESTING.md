# TaskFlow Pro 测试说明

项目补充了 4 类适合 Go 后端实习面试讲解的测试。

## 1. 注册登录测试

文件：`backend/internal/service/flow_test.go`

测试点：

- `Register` 会创建用户、哈希密码、签发 JWT。
- `Login` 会校验邮箱和密码。
- 登录返回的 token 可以被 `ParseToken` 正确解析出用户 ID。
- 错误密码会返回 `ErrInvalidCredentials`。

讲解重点：

> 这类测试覆盖认证核心链路，不依赖真实 MySQL。通过 fake repository 模拟用户表，可以专注验证 service 层业务逻辑。

## 2. JWT 鉴权测试

文件：`backend/internal/middleware/middleware_test.go`

测试点：

- 带合法 `Authorization: Bearer <token>` 请求可以通过鉴权中间件。
- 中间件会把 token 中的 `user_id` 写入 Gin Context。
- 缺少 token 的请求会返回 `401 Unauthorized`。

讲解重点：

> JWT 签发在 service 层，鉴权拦截在 middleware 层。测试用 `httptest` 模拟 HTTP 请求，不需要启动真实服务器。

## 3. 任务状态流转测试

文件：`backend/internal/service/flow_test.go`

测试点：

- 新建任务默认进入 `todo` 状态。
- 任务可以从 `todo` 更新为 `doing`。
- 任务可以从 `doing` 更新为 `done`。
- 非法状态，例如 `blocked`，会返回 `ErrInvalidStatus`。

讲解重点：

> 任务状态是业务规则，不应该只依赖前端下拉框限制。后端 service 层必须再次校验，防止非法 API 调用写入脏数据。

## 4. 项目成员权限测试

文件：`backend/internal/service/flow_test.go`

测试点：

- 非项目成员不能创建任务。
- 项目 owner 可以添加成员。
- 非 owner 添加成员会返回 `ErrForbidden`。

讲解重点：

> 权限判断放在 service 层，而不是只放在 handler 层。这样无论接口从哪里调用，都会经过统一业务权限校验。

## 为了测试做的小重构

新增文件：`backend/internal/service/deps.go`

这里定义了 service 层依赖的接口：

- `UserStore`
- `ProjectStore`
- `TaskStore`
- `CommentStore`
- `Cache`

真实运行时使用 repository 和 Redis 适配器；测试时注入 fake store 和内存 cache。

这个设计体现了依赖倒置：

> service 层不关心数据来自真实 MySQL、Redis，还是测试 fake，只关心接口能提供什么能力。因此业务逻辑更容易单元测试。

## 运行测试

```powershell
cd C:\ashiyanqu\go-demo\items\taskflow-pro\backend
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
go test ./...
```

