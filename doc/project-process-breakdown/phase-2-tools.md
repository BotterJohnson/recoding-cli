# 阶段二：工具系统开发

## 目标

Agent 能调用工具，实现真正的代码生成（读取项目文件 → 生成代码 → 写入文件带确认）

## 前置条件

- 阶段一完成（CLI + Provider + 基础 Agent 循环可用）

## 交付物

- Agent 能读取项目文件、生成代码、写入文件（带 Diff 确认）
- 完整的 ReAct 多轮循环

---

## 步骤拆分

### Step 1: 工具接口与注册中心

**输出文件：**
- `internal/tools/interface.go` — Tool 接口 + ToolResult + PermissionLevel
- `internal/tools/registry.go` — ToolRegistry（注册、获取、按 Agent 过滤）

```go
// 核心接口
type Tool interface {
    Name() string
    Description() string
    Parameters() JSONSchema
    Permission() PermissionLevel
    Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}
```

**验收：** 能注册工具并按名称获取

---

### Step 2: 基础工具实现

**输出文件：**
- `internal/tools/file.go` — file_read（safe）、file_write（confirm）
- `internal/tools/directory.go` — directory_list（safe）、directory_create（confirm）
- `internal/tools/shell.go` — shell_exec（confirm）
- `internal/tools/search.go` — code_search（safe）

**每个工具需要：**
1. 实现 Tool 接口
2. 参数 JSON Schema 定义
3. 错误处理（返回错误信息给 LLM）

**验收：** 各工具单元测试通过

---

### Step 3: 工具调用解析

**输出文件：**
- `internal/provider/openai.go`（扩展）— 解析 function_call / tool_calls 响应

**逻辑：**
1. StreamEvent 新增 ToolCallStart / ToolCallDelta 类型
2. 累积 tool_call 参数 JSON
3. 解析完整后返回 ToolCall 结构体

**验收：** mock 一个含 tool_calls 的 SSE 响应，能正确解析出工具名和参数

---

### Step 4: Agent 多轮循环（ReAct）

**输出文件：**
- `internal/agent/runtime.go`（重构）— 完整 ReAct 循环

```
循环流程：
1. 构建 messages（含历史工具调用结果）
2. 调用 LLM
3. 如果响应含 tool_calls → 执行工具 → 结果追加到 messages → 回到 1
4. 如果响应是纯文本 → 输出给用户 → 结束
5. 超过 max_iterations → 强制停止
```

**关键参数：**
- max_iterations: 20
- tool_timeout: 30s

**验收：** Agent 能自主决定读取文件 → 生成代码 → 写入文件的多步流程

---

### Step 5: Diff 确认机制

**输出文件：**
- `internal/render/diff.go` — Diff 展示（绿色新增、红色删除）
- `internal/security/confirm.go` — 用户确认交互

```
确认流程：
1. file_write 工具执行前拦截
2. 计算 diff（新建文件则全部为新增）
3. 渲染 diff 到终端
4. 等待用户输入：[y] 确认 / [n] 拒绝 / [e] 编辑 / [a] 全部确认
5. 确认 → 执行写入；拒绝 → 返回"用户拒绝"给 LLM
```

**验收：** 文件写入前能看到 diff 并交互确认

---

### Step 6: 权限系统

**输出文件：**
- `internal/security/permission.go` — 权限检查逻辑

```
confirm_mode:
- auto: 所有操作自动执行
- smart（默认）: PermConfirm + PermDangerous 需确认
- manual: 每个工具调用都确认
```

**验收：** smart 模式下 file_read 自动执行，file_write 需确认

---

## 集成验收

```bash
./recoding-cli dev "读取当前项目的 go.mod，然后创建一个 utils/string.go 工具函数"

# 预期流程：
# 1. Agent 调用 file_read 读取 go.mod（自动）
# 2. Agent 调用 directory_list 查看项目结构（自动）
# 3. Agent 调用 file_write 创建文件（展示 diff，等待确认）
# 4. 用户确认后写入文件
# 5. Agent 输出总结
```
