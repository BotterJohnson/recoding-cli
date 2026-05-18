# 阶段四：多 Agent 与会话

## 目标

完整的多命令 Agent 系统（dev/review/plan/test/chat）+ 会话管理 + 中断恢复

## 前置条件

- 阶段三完成（上下文管理 + 记忆系统可用）

## 交付物

- 4 个专业 Agent（review/plan/test/chat）
- 交互式会话模式
- 会话保存、恢复、归档
- 安全中断与恢复

---

## 步骤拆分

### Step 1: Agent 定义系统

**输出文件：**
- `internal/agent/definition.go`（重构）— 通用 Agent 定义加载
- `prompts/review.tmpl`
- `prompts/plan.tmpl`
- `prompts/test.tmpl`
- `prompts/chat.tmpl`

**Agent 定义结构：**
```go
type AgentDefinition struct {
    Name         string
    Description  string
    PromptFile   string     // 模板文件路径
    Tools        []string   // 可用工具名列表
    MaxIter      int        // 最大循环次数
    OutputFormat string     // 输出格式提示
}
```

**验收：** 能通过配置加载不同 Agent 定义，各 Agent 使用不同 prompt 和工具集

---

### Step 2: Review Agent

**输出文件：**
- `cmd/review.go` — review 子命令
- `prompts/review.tmpl` — 审查 prompt

**工具集：** file_read, directory_list, code_search, git_diff

**工作流程：**
1. 接收目标路径（文件/目录）
2. 读取代码
3. 分析：逻辑错误、安全漏洞、性能问题、代码风格
4. 输出结构化审查报告（问题列表 + 严重程度 + 修复建议）

**验收：** `recoding-cli review ./internal/` 输出审查报告

---

### Step 3: Plan Agent

**输出文件：**
- `cmd/plan.go` — plan 子命令
- `prompts/plan.tmpl` — 规划 prompt

**工具集：** file_read, directory_list, web_search, url_fetch

**工作流程：**
1. 接收需求描述
2. 分析需求，搜索相关技术方案
3. 输出：模块划分、技术选型、目录结构、开发步骤

**验收：** `recoding-cli plan "用户管理系统"` 输出结构化计划

---

### Step 4: Test Agent

**输出文件：**
- `cmd/test.go` — test 子命令
- `prompts/test.tmpl` — 测试生成 prompt

**工具集：** file_read, file_write, directory_list, code_search, shell_exec

**工作流程：**
1. 读取目标代码
2. 分析函数签名、边界条件
3. 生成测试用例（单元测试 + 边界测试）
4. 可选执行测试验证

**验收：** `recoding-cli test ./internal/config/` 生成测试文件

---

### Step 5: Chat 交互模式

**输出文件：**
- `cmd/chat.go` — chat 子命令（交互式 REPL）
- `internal/session/manager.go` — 会话管理器

**交互功能：**
- 多轮对话循环（readline 输入）
- 会话内命令：/help, /clear, /save, /undo, /mode, /model, /usage, /exit
- 所有工具可用
- 启动时显示项目信息和模型信息

**验收：** `recoding-cli chat` 进入交互模式，支持多轮对话和会话命令

---

### Step 6: 会话管理

**输出文件：**
- `internal/session/history.go` — 对话历史存储
- `cmd/session.go` — session 子命令（list/resume）

**存储结构：**
```
~/.recoding/sessions/
├── active/session_xxx.json
├── archive/session_xxx.json
└── summaries/session_xxx_summary.md
```

**功能：**
- 自动保存活跃会话
- `session list` 查看历史
- `session resume <id>` 恢复会话

**验收：** 会话能保存和恢复，历史对话完整

---

### Step 7: 中断恢复

**输出文件：**
- `internal/session/interrupt.go` — 中断处理

**逻辑：**
1. 捕获 Ctrl+C（os.Signal）
2. 保存当前会话状态（标记为 interrupted）
3. 下次启动时检测 interrupted 会话，提示恢复

```go
type SessionState struct {
    ID          string
    Status      string    // active / interrupted / completed
    Messages    []Message
    PendingTool *ToolCall // 中断时正在执行的工具
    Checkpoint  *Checkpoint
}
```

**验收：** Ctrl+C 中断后重新启动，能恢复到中断前状态

---

## 集成验收

```bash
# Review
./recoding-cli review ./internal/provider/
# 预期：输出代码审查报告

# Plan
./recoding-cli plan "实现一个博客系统"
# 预期：输出模块划分和开发步骤

# Test
./recoding-cli test ./internal/config/config.go
# 预期：生成 config_test.go

# Chat 交互
./recoding-cli chat
> 帮我重构 provider 层
> /save
> /exit

# 恢复会话
./recoding-cli session list
./recoding-cli session resume <id>
```
