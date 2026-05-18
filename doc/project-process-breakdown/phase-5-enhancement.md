# 阶段五：增强与打磨

## 目标

生产级质量，可作为日常工具使用。完善 Provider 支持、增强工具集、健壮错误处理、日志与可观测性。

## 前置条件

- 阶段四完成（多 Agent 系统 + 会话管理可用）

## 交付物

- Anthropic Provider 支持
- 增强工具集（web_search, url_fetch, git 操作, 数据库操作）
- 完整错误处理与重试机制
- 结构化日志 + 用量统计
- 执行回放功能
- 性能优化（并行工具调用）

---

## 步骤拆分

### Step 1: Anthropic Provider

**输出文件：**
- `internal/provider/anthropic.go` — Anthropic Claude 实现

**差异点：**
- 使用 anthropic-sdk-go
- tool_use 格式与 OpenAI 不同（content block 模式）
- 流式事件类型不同（content_block_start/delta/stop）

**验收：** 配置 `provider.name: anthropic` 后，所有 Agent 功能正常

---

### Step 2: 增强工具集

**输出文件：**
- `internal/tools/web.go` — web_search（联网搜索）
- `internal/tools/fetch.go` — url_fetch（URL 内容抓取）
- `internal/tools/git.go` — git_status, git_diff, git_commit
- `internal/tools/database.go` — db_query, db_execute

**权限分配：**
| 工具 | 权限 |
|------|------|
| web_search | safe |
| url_fetch | safe |
| git_status / git_diff | safe |
| git_commit | confirm |
| db_query | confirm |
| db_execute | dangerous |

**验收：** 各工具单元测试通过，Agent 能在对话中调用

---

### Step 3: 错误处理完善

**输出文件：**
- `internal/provider/retry.go` — 重试策略
- `internal/provider/circuit.go` — 熔断机制

**重试策略：**
| 错误类型 | 重试次数 | 退避方式 |
|---------|---------|---------|
| 网络错误 | 3 | 指数退避（1s, 2s, 4s） |
| API 限流 | 5 | Retry-After 或指数退避 |
| 认证失败 | 0 | 提示用户检查配置 |
| 解析失败 | 2 | 追加格式提示 |
| 超时 | 2 | 固定间隔 |

**熔断：** 连续 5 次 API 失败 → 停止调用，提示用户

**验收：** mock 网络错误，验证重试和熔断行为

---

### Step 4: 日志系统

**输出文件：**
- `internal/logger/logger.go` — zerolog 初始化 + 日志级别
- `internal/logger/replay.go` — 执行记录（用于回放）

**日志存储：**
```
~/.recoding/logs/
├── recoding.log        # 主日志
├── debug/YYYY-MM-DD.log  # Debug 日志
└── usage/YYYY-MM.json    # 月度用量
```

**用量统计结构：**
```go
type UsageRecord struct {
    Timestamp    time.Time
    SessionID    string
    Model        string
    InputTokens  int
    OutputTokens int
    EstCost      float64
    AgentType    string
}
```

**验收：** `recoding-cli usage` 显示今日/本月用量和费用估算

---

### Step 5: 执行回放

**输出文件：**
- `cmd/replay.go` — replay 子命令

**功能：**
- Debug 模式下记录完整 Agent 执行过程
- `recoding-cli replay <session_id>` 逐步回放
- 展示：思考过程 → 工具调用 → 结果 → 下一步

**验收：** 能回放一个完整的 dev 会话

---

### Step 6: 命令别名与配置迁移

**输出文件：**
- `cmd/root.go`（扩展）— 别名注册
- `internal/config/migration.go` — 配置版本迁移

**别名：**
```yaml
aliases:
  r: review
  d: dev
  p: plan
  t: test
```

**配置迁移：** 检测 config version，自动升级旧格式

**验收：** `recoding-cli r ./src/` 等同于 `recoding-cli review ./src/`

---

### Step 7: 性能优化

**修改文件：**
- `internal/agent/runtime.go`（扩展）— 并行工具调用

**优化点：**
1. 并行工具调用：LLM 返回多个 tool_calls 时并发执行
2. 响应缓存：相同 file_read 请求短期内不重复读取
3. 流式输出优化：减少渲染刷新频率

**验收：** 多工具调用场景下响应时间明显缩短

---

## 集成验收

```bash
# Anthropic 模型
./recoding-cli --model claude-3-5-sonnet dev "实现一个 HTTP 中间件"

# 联网搜索
./recoding-cli dev "用最新的 Go 1.22 路由特性重写路由层"

# Git 操作
./recoding-cli chat
> 查看当前 git 状态，帮我提交所有变更

# 用量统计
./recoding-cli usage

# 回放
./recoding-cli replay session_20240101_143000

# 别名
./recoding-cli r ./internal/
```
