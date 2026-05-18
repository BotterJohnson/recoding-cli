# 阶段三：上下文与记忆系统

## 目标

Agent 理解项目上下文，具备跨会话记忆，上下文过长时自动压缩

## 前置条件

- 阶段二完成（工具系统 + ReAct 循环可用）

## 交付物

- 项目自动感知（语言/框架/结构）
- Token 窗口管理 + 摘要压缩
- 辅助模型（DeepSeek-Flash）集成
- 跨会话记忆读写

---

## 步骤拆分

### Step 1: 项目感知

**输出文件：**
- `internal/context/project.go` — 项目信息自动检测

**检测逻辑：**
1. 扫描项目根目录标志文件（go.mod, package.json, Cargo.toml, pom.xml 等）
2. 识别语言、框架、包管理器
3. 生成目录结构摘要（限制深度，排除 node_modules/.git 等）
4. 输出 ProjectInfo 结构体

```go
type ProjectInfo struct {
    Name       string
    Language   string
    Framework  string
    RootPath   string
    Structure  string   // 目录树文本
    KeyFiles   []string // 关键文件列表
}
```

**验收：** 在 Go 项目中运行，能正确识别语言和框架

---

### Step 2: 上下文管理器

**输出文件：**
- `internal/context/manager.go` — 上下文生命周期管理
- `internal/context/window.go` — Token 窗口计算

**分层上下文构建：**
```
总上下文 = System层(~2K) + 项目层(~4K) + 记忆层(~2K) + 会话层(动态) + 预留输出(~10K)
```

**Token 计算：**
- 使用 tiktoken-go 或简单估算（字符数 / 4）
- 跟踪每层 token 占用
- 提供 `RemainingTokens()` 查询

**验收：** 能正确构建分层上下文，token 计数准确

---

### Step 3: 辅助模型集成

**输出文件：**
- `internal/provider/selector.go` — 模型选择器（主/辅模型路由）

**辅助模型职责：**
- 摘要生成
- 记忆提取
- 意图分类（后续阶段用）

```go
type ModelSelector struct {
    Primary   Provider
    Assistant Provider
}

func (s *ModelSelector) Select(task TaskType) Provider
```

**验收：** 辅助模型能独立调用并返回结果

---

### Step 4: 摘要压缩

**输出文件：**
- `internal/context/summary.go` — 摘要生成与会话切换
- `prompts/summary.tmpl` — 摘要 prompt 模板

**触发条件：** Token 使用率 > 70%（可配置 `context.summary_threshold`）

**压缩流程：**
1. 检测 token 用量超阈值
2. 调用辅助模型生成结构化摘要（做了什么、关键决策、当前状态、待办）
3. 用摘要替换旧对话历史
4. 旧会话归档
5. 通知用户"上下文已压缩"

**验收：** 模拟长对话触发压缩，压缩后 Agent 仍能继续工作

---

### Step 5: 记忆存储

**输出文件：**
- `internal/memory/store.go` — 记忆存储接口
- `internal/memory/project.go` — 项目级记忆（.recoding/memory/）
- `internal/memory/global.go` — 全局记忆（~/.recoding/memory/global/）

**记忆文件结构：**
```
~/.recoding/memory/global/
├── preferences.md
├── tech_stack.md
└── patterns.md

.recoding/memory/
├── architecture.md
├── conventions.md
└── decisions.md
```

**验收：** 能读写记忆文件，支持追加和覆盖

---

### Step 6: 记忆提取

**输出文件：**
- `internal/memory/extractor.go` — 从对话中提取关键信息
- `prompts/memory_extract.tmpl` — 提取 prompt 模板

**提取时机：**
- 会话结束时自动提取
- 用户显式说"记住xxx"

**提取内容分类：**
- 架构决策 → architecture.md
- 编码约定 → conventions.md
- 用户偏好 → preferences.md

**验收：** 对话结束后，相关记忆被正确写入文件

---

### Step 7: 记忆注入

**修改文件：**
- `internal/context/manager.go`（扩展）— 会话开始时加载记忆

**注入逻辑：**
1. 会话开始时读取项目记忆 + 全局记忆
2. 拼接到上下文的记忆层（限制 max_inject_tokens: 2048）
3. 超出限制时截断最旧的记忆

**验收：** 新会话能感知到之前记录的项目约定和用户偏好

---

## 集成验收

```bash
# 1. 首次使用，Agent 自动感知项目
./recoding-cli dev "查看项目结构并告诉我这是什么项目"
# 预期：输出项目语言、框架、目录结构

# 2. 记忆写入
./recoding-cli dev "记住：这个项目使用 4 空格缩进，错误处理用 fmt.Errorf 包装"
# 预期：写入 .recoding/memory/conventions.md

# 3. 记忆读取（新会话）
./recoding-cli dev "帮我写一个工具函数"
# 预期：生成的代码遵循 4 空格缩进和 fmt.Errorf 约定

# 4. 长对话压缩（模拟）
# 连续多轮对话直到触发压缩
# 预期：提示"上下文已压缩"，后续对话正常
```
