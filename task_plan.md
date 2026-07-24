# 任务计划：完成 Go 项目工程化分层

## 目标
在不改变现有业务行为的前提下，将 gin-basic 中的代码迁移为 cmd/server 与 internal 分层结构，并通过全量编译测试。

## 当前阶段
阶段 7

## 各阶段

### 阶段 1：盘点依赖
- [x] 确认用户要求一次完成全部迁移
- [x] 盘点现有文件、类型、函数和跨文件依赖
- [x] 将关键发现记录到 findings.md
- **状态：** complete

### 阶段 2：建立共享包
- [x] 迁移 config
- [x] 迁移 model
- [x] 规划 DTO、错误和响应包
- **状态：** complete

### 阶段 3：迁移 Repository
- [x] 迁移数据库初始化与 Product/User/RefreshToken 数据访问
- [x] 更新模型引用并编译验证
- **状态：** complete

### 阶段 4：迁移 Service
- [x] 迁移认证、JWT、Refresh Token 与商品业务逻辑
- [x] 保持依赖方向为 Service → Repository
- [x] 编译验证
- **状态：** complete

### 阶段 5：迁移 HTTP 层
- [x] 迁移 Handler、Middleware、DTO 和 Response
- [x] 注册现有全部路由
- [x] 编译验证
- **状态：** complete

### 阶段 6：入口与全量验证
- [x] 创建 cmd/server/main.go
- [x] 删除 gin-basic 中的过渡别名与孤立文件
- [x] 执行 gofmt、go test ./...、go vet ./...
- [x] 记录最终结构
- **状态：** complete

### 阶段 7：统一业务错误体系
- [x] 创建 AppError 与 HTTP 错误映射
- [x] 将 ProductService 和 AuthService 的错误转换为 AppError
- [x] 精简 ProductHandler、AuthHandler 与认证 Middleware
- [x] 执行 gofmt、go test ./...、go vet ./...
- [x] 记录对外错误响应格式变化
- **状态：** complete

## 关键问题
1. 当前业务代码是否已经形成可直接分包的依赖方向？
2. 哪些旧教学代码仍参与编译但不参与实际路由？

## 已做决策
| 决策 | 理由 |
|------|------|
| 保持单一 Go Module | internal 包只允许模块内部访问，适合当前服务 |
| 不改变数据库结构与接口行为 | 本次目标是工程化迁移，不是功能重写 |
| 分阶段编译 | 尽早发现循环依赖和可见性问题 |
| Service 返回业务 AppError，不包含 HTTP 状态码 | 保持 Service 与 HTTP 协议解耦 |
| Response 将 AppError 映射为 HTTP 状态码 | 集中管理错误响应和内部错误日志 |

## 遇到的错误
| 错误 | 尝试次数 | 解决方案 |
|------|---------|---------|
| 计划更新补丁上下文匹配失败 | 1 | 读取最新文件后拆分为精确补丁 |
| 阶段 7 发现更新补丁上下文匹配失败 | 1 | 读取准确区段并拆分更新 |
| Response 重复定义 HandleError/statusFromCode | 1 | 保留 error_handler.go，response.go 只保留基础响应 |
| 认证跨层大补丁上下文匹配失败 | 1 | 拆为 AuthService、AuthHandler、Middleware 三步修改 |

## 备注
- 所有新增或修改函数必须包含功能、参数和返回值说明。
- 只删除因本次迁移产生的孤立代码，不清理无关教学代码。
