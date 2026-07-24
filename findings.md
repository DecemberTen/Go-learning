# 发现与决策

## 需求
- 将当前 gin-basic 工程一次性迁移为最终分层结构，供用户学习。
- 保留现有 Gin、GORM、认证、Refresh Token、商品接口和优雅关闭行为。
- 新增统一业务错误体系，减少 Handler 中重复的错误判断、响应和日志代码。

## 研究发现
- 模块路径为 `example.com/go-learning`。
- `internal/config` 与 `internal/model` 已迁移完成。
- gin-basic 仍通过类型别名兼容 `Product`、`ProductImage`、`User`、`RefreshToken`。
- DTO、Repository、Service、Handler、Middleware 与入口仍在 `package main`。
- 当前 ProductHandler 直接持有 `*gorm.DB` 并调用包级 GORM 函数，尚未形成 Service 边界。
- `database.go` 中的 raw SQL CRUD 没有被当前路由调用，生产路径统一使用 GORM。
- Product Repository 需要自有的筛选与更新输入类型，不能依赖 Handler DTO。
- AuthService 当前直接持有 GORM 并调用用户与 Refresh Token 包级函数。
- Refresh Token 轮换必须在一个事务中完成加锁查询、用户查询、撤销旧令牌和创建新令牌。
- Middleware 当前直接调用 JWT 函数和 User Repository，迁移后应统一依赖 AuthService。
- `gin-basic/sqls` 中存在 users 与 refresh_tokens 建表脚本，已作为迁移资产保留。
- 新入口 `cmd/server` 及所有 internal 包已通过全量测试与 go vet。
- ProductService 当前直接透传全部 Repository 错误，ProductHandler 重复完成 404/409/500 判断和日志。
- AuthService 当前混用 Service 哨兵错误与技术错误，AuthHandler 和 Middleware 分别重复映射。
- JSON、Query 和 Path 参数错误属于 HTTP 输入错误，仍由 Handler 直接返回 INVALID_REQUEST。
- AuthService.FindUserByID 只用于当前认证用户查询；用户不存在应解释为 Access Token 已失效。
- ProductHandler 与 AuthHandler 已不再识别 Repository/Service 哨兵错误，也不再分别记录内部错误。
- 认证 Middleware 使用 AbortAppError 中断 Gin 调用链，普通 Handler 使用 HandleError 后自行 return。
- 商品接口错误响应已从 `{"error":"..."}` 统一为 `{"code":"...","message":"..."}`。
- `go test ./...` 与 `go vet ./...` 均通过。

## 技术决策
| 决策 | 理由 |
|------|------|
| Repository 直接使用 internal/model | 消除 package main 的实体别名依赖 |
| HTTP DTO 单独放 internal/handler 或 dto 包 | 避免 model 混入接口校验标签 |
| main 只负责依赖组装和生命周期 | 保持依赖方向清晰 |
| 最终结构不迁移未使用的 raw SQL CRUD | 避免最终项目同时维护两套数据访问实现；连接池初始化仍保留 |
| ProductHandler 改为依赖 ProductService | 建立 Handler → Service → Repository 的单向依赖 |
| Repository 使用统一 Store | 事务回调可获得绑定到 tx 的 Store，同时保持 Service 不导入 GORM |
| AuthService 对 Middleware 暴露令牌解析和用户查询 | Middleware 不跨层访问 Repository |
| SQL 脚本迁移到根目录 migrations | 数据库结构资产不应放在旧应用源码目录 |
| AppError 只包含业务 Code、公开 Message 和内部 Cause | Service 不直接依赖 HTTP 状态码，同时避免泄露内部错误 |
| 只有 Response 层决定 HTTP 状态码 | HTTP 协议知识停留在传输层 |
| Handler 使用 HandleError，Middleware 使用 AbortAppError | Handler 返回即可，Middleware 还必须中断后续链路 |
| 只在统一 Response 层记录 500 Cause | 避免 Handler 和 Response 重复记录同一个内部错误 |

## 遇到的问题
| 问题 | 解决方案 |
|------|---------|
| 计划更新时把 findings.md 的决策行误认为在 task_plan.md | 读取当前文件后分别更新 |
| 阶段 7 发现更新再次引用了错误的文件上下文 | 读取三个规划文件的准确区段后拆分更新 |

## 资源
- 项目根目录：`/Users/pipi/CODE/go-learning`
- 当前应用入口：`/Users/pipi/CODE/go-learning/cmd/server`

## 视觉/浏览器发现
- 本任务不涉及视觉或浏览器内容。
