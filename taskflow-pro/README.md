# TaskFlow Pro

TaskFlow Pro 是一个用于后端面试展示的前后端分离团队任务协作系统。

## 技术栈

- 后端：Go、Gin、GORM、MySQL、Redis、JWT、bcrypt
- 前端：独立静态 HTML/CSS/JavaScript，通过 Fetch 调用后端 API
- 部署：Docker Compose，一键启动 MySQL、Redis、后端、前端

## 快速启动

```powershell
cd C:\ashiyanqu\go-demo\items\taskflow-pro
docker compose up -d --build
```

访问：

- 前端：http://localhost:3001
- 后端健康检查：http://localhost:8090/health
- MySQL：localhost:3308，用户 `taskflow`，密码 `taskflow123`
- Redis：localhost:6381

查看日志：

```powershell
docker compose logs -f backend
```

## 本地后端运行

需要先启动 MySQL 和 Redis，或者只启动依赖服务：

```powershell
docker compose up -d mysql redis
cd backend
go run .\cmd\server
```

## 面试讲解重点

1. 前后端分离：`frontend` 是独立静态页面，只通过 REST API 调用 `backend`。
2. Gin：路由分组、中间件、参数绑定、统一响应。
3. GORM/MySQL：用户、项目、成员、任务、评论多表关系，事务和分页查询。
4. Redis：接口限流、用户信息缓存、项目统计缓存。
5. JWT：登录签发 token，中间件解析用户身份。
6. 分层架构：handler -> service -> repository -> model。

## 默认功能

- 注册 / 登录
- 创建项目
- 添加项目成员
- 创建、筛选、更新、删除任务
- 任务看板
- 任务评论
- 项目任务统计

