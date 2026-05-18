# recoding-cli 完整项目计划书

## 一、项目概述

### 1.1 项目定位

recoding-cli 是一个基于大语言模型的终端 AI 编程助手，通过命令行交互帮助开发者完成代码生成、代码审查、项目规划、测试生成等开发任务。

### 1.2 核心理念

- **命令即能力**：每个子命令对应一个专业 Agent，职责单一、能力明确
- **双模型架构**：主模型处理复杂任务，辅助模型处理简单任务，平衡质量与成本
- **项目感知**：自动理解当前项目上下文，减少用户重复描述
- **记忆持久化**：跨会话记忆，Agent 越用越懂你的项目和偏好
- **安全优先**：文件变更先 diff 确认，危险操作需用户授权

### 1.3 目标用户

- 独立开发者
- 学习编程的开发者
- 需要 AI 辅助提效的工程师

### 1.4 产品形态

- 终端 CLI 工具，单二进制分发
- 支持单次命令模式和交互式会话模式
- 中文界面，预留国际化能力

---

## 二、技术选型

### 2.1 核心技术栈

| 类别 | 选型 | 理由 |
|------|------|------|
| 语言 | Go 1.22+ | 编译单二进制、并发友好、CLI 生态成熟 |
| CLI 框架 | Cobra | Go CLI 标准选择，子命令路由天然支持 |
| 配置管理 | Viper | 支持 YAML、环境变量、多层配置覆盖 |
| LLM 主模型 | 用户自配（OpenAI/Anthropic/DeepSeek等） | 灵活适配 |
| LLM 辅助模型 | DeepSeek-Flash（内置） | 免费/极低成本，处理简单任务 |
| 终端渲染 | lipgloss + glamour | Markdown 渲染、语法高亮、美化输出 |
| 模板引擎 | Go text/template | Prompt 模板化管理 |
| 日志 | zerolog | 结构化日志，性能好 |
| 测试 | Go 标准 testing + testify | 断言库 + mock 支持 |

### 2.2 LLM API 依赖

| Provider | 库 | 说明 |
|----------|-----|------|
| OpenAI 兼容 | sashabaranov/go-openai | 兼容 DeepSeek、Moonshot、通义等 |
| Anthropic | anthropics/anthropic-sdk-go | Claude 系列 |

### 2.3 不使用的

- 不用 LangChainGo（不成熟，自研 Agent 循环）
- 不支持本地模型（不做 Ollama 适配）
- 不做 Web UI（纯 CLI）

---

## 三、系统架构

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                      用户终端                             │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   CLI 层 (Cobra)                          │
│  ┌─────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ │
│  │ dev │ │review│ │ plan │ │ test │ │ chat │ │config│ │
│  └──┬──┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ │
└─────┼────────┼────────┼────────┼────────┼────────┼──────┘
      │        │        │        │        │        │
┌─────▼────────▼────────▼────────▼────────▼────────┼──────┐
│                  Agent Runtime（核心引擎）          │      │
│  ┌────────────────────────────────────────────┐  │      │
│  │            Agent 执行循环                    │  │      │
│  │  Prompt构建 → LLM调用 → 解析响应 →          │  │      │
│  │  工具执行 → 结果拼装 → 继续/结束            │  │      │
│  └────────────────────────────────────────────┘  │      │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │      │
│  │上下文管理 │ │ 记忆系统 │ │ Prompt 模板引擎  │ │      │
│  └──────────┘ └──────────┘ └──────────────────┘ │      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   Provider 层                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │           统一 LLM Interface                      │   │
│  ├──────────┬──────────────┬────────────────────────┤   │
│  │ OpenAI   │  Anthropic   │  DeepSeek-Flash(辅助)  │   │
│  └──────────┴──────────────┴────────────────────────┘   │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    工具层 (Tools)                         │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │
│  │文件IO│ │命令行│ │代码搜│ │联网搜│ │数据库│  ...     │
│  │读写  │ │执行  │ │索    │ │索    │ │操作  │         │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘         │
└─────────────────────────────────────────────────────────┘
```

### 3.2 核心模块划分

```
recoding-cli/
├── cmd/                    # CLI 命令定义（Cobra）
│   ├── root.go            # 根命令、全局 flag
│   ├── dev.go             # dev 子命令
│   ├── review.go          # review 子命令
│   ├── plan.go            # plan 子命令
│   ├── test.go            # test 子命令
│   ├── chat.go            # chat 交互式会话
│   └── config.go          # 配置管理命令
├── internal/
│   ├── agent/             # Agent 核心引擎
│   │   ├── runtime.go     # Agent 执行循环
│   │   ├── definition.go  # Agent 定义（prompt、工具集、约束）
│   │   ├── router.go      # 意图路由（辅助模型分类）
│   │   └── pipeline.go    # Agent 间协作管道（预留）
│   ├── provider/          # LLM Provider 抽象
│   │   ├── interface.go   # 统一接口定义
│   │   ├── openai.go      # OpenAI 兼容实现
│   │   ├── anthropic.go   # Anthropic 实现
│   │   ├── model_meta.go  # 模型元信息（token限制等）
│   │   └── selector.go    # 主/辅模型选择器
│   ├── tools/             # 工具实现
│   │   ├── registry.go    # 工具注册中心
│   │   ├── interface.go   # 工具统一接口
│   │   ├── file.go        # 文件读写
│   │   ├── directory.go   # 目录操作
│   │   ├── shell.go       # 命令执行
│   │   ├── search.go      # 代码搜索
│   │   ├── web.go         # 联网搜索
│   │   ├── fetch.go       # URL 内容抓取
│   │   ├── database.go    # 数据库操作
│   │   └── git.go         # Git 操作
│   ├── context/           # 上下文管理
│   │   ├── manager.go     # 上下文生命周期管理
│   │   ├── project.go     # 项目感知（自动检测技术栈）
│   │   ├── summary.go     # 摘要生成（调辅助模型）
│   │   └── window.go      # Token 窗口计算与压缩
│   ├── memory/            # 记忆系统
│   │   ├── store.go       # 记忆存储接口
│   │   ├── project.go     # 项目级记忆
│   │   ├── global.go      # 全局记忆
│   │   ├── extractor.go   # 记忆提取（从对话中提取关键信息）
│   │   └── rag.go         # RAG 检索（后期）
│   ├── session/           # 会话管理
│   │   ├── manager.go     # 会话创建、恢复、归档
│   │   ├── history.go     # 对话历史存储
│   │   └── interrupt.go   # 中断与恢复
│   ├── render/            # 输出渲染
│   │   ├── terminal.go    # 终端 Markdown 渲染
│   │   ├── diff.go        # Diff 展示
│   │   ├── spinner.go     # 加载动画
│   │   └── stream.go      # 流式输出处理
│   ├── config/            # 配置管理
│   │   ├── config.go      # 配置结构定义
│   │   ├── loader.go      # 多层配置加载
│   │   └── migration.go   # 配置版本迁移
│   ├── security/          # 安全机制
│   │   ├── confirm.go     # 用户确认交互
│   │   ├── permission.go  # 工具权限等级
│   │   └── sandbox.go     # 命令执行沙箱
│   ├── logger/            # 日志系统
│   │   ├── logger.go      # 日志初始化
│   │   └── replay.go      # 执行回放
│   └── i18n/              # 国际化预留
│       ├── zh.go          # 中文文本
│       └── en.go          # 英文文本（预留）
├── prompts/               # Prompt 模板文件
│   ├── dev.tmpl           # 开发 Agent prompt
│   ├── review.tmpl        # 审查 Agent prompt
│   ├── plan.tmpl          # 规划 Agent prompt
│   ├── test.tmpl          # 测试 Agent prompt
│   ├── summary.tmpl       # 摘要生成 prompt
│   └── memory_extract.tmpl # 记忆提取 prompt
├── configs/               # 默认配置
│   └── default.yaml       # 默认配置模板
├── go.mod
├── go.sum
├── Makefile
└── README.md
```


---

## 四、核心模块详细设计

### 4.1 Agent Runtime（核心引擎）

Agent 执行循环是整个系统的心脏，采用 ReAct 模式：

```
┌─────────────────────────────────────────────┐
│              Agent Runtime Loop              │
│                                             │
│  1. 构建 Prompt                              │
│     ├── System Prompt（从模板加载）           │
│     ├── 项目上下文（自动感知）               │
│     ├── 记忆注入（相关记忆）                 │
│     ├── 工具定义（当前 Agent 可用工具）       │
│     └── 对话历史（含工具调用结果）           │
│                                             │
│  2. 调用 LLM                                 │
│     ├── 流式输出文本给用户                   │
│     └── 检测是否有工具调用请求               │
│                                             │
│  3. 解析响应                                 │
│     ├── 纯文本 → 展示给用户，结束本轮        │
│     └── 工具调用 → 进入步骤 4               │
│                                             │
│  4. 执行工具                                 │
│     ├── 权限检查（是否需要用户确认）         │
│     ├── 并行/串行执行工具                    │
│     ├── 收集执行结果                         │
│     └── 错误处理（失败信息反馈给 LLM）       │
│                                             │
│  5. 判断是否继续                             │
│     ├── 未达到目标 → 回到步骤 1             │
│     ├── 达到最大循环次数 → 强制停止并汇报    │
│     └── Token 接近上限 → 触发摘要压缩        │
│                                             │
└─────────────────────────────────────────────┘
```

**关键参数：**
- 最大循环次数：默认 20 轮（可配置）
- Token 阈值：模型上下文的 70% 触发摘要
- 工具调用超时：单个工具默认 30 秒（可配置）

### 4.2 Provider 层设计

#### 统一接口定义

```go
type Provider interface {
    // 流式对话，返回 channel
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)
    // 非流式对话
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    // 获取模型元信息
    ModelMeta() *ModelMeta
}

type ChatRequest struct {
    Model       string
    Messages    []Message
    Tools       []ToolDefinition
    Temperature float64
    MaxTokens   int
}

type StreamEvent struct {
    Type      EventType  // TextDelta / ToolCallStart / ToolCallDelta / Done / Error
    Text      string
    ToolCall  *ToolCall
    Error     error
}

type ModelMeta struct {
    MaxContextTokens  int
    MaxOutputTokens   int
    SupportsTools     bool
    SupportsStreaming  bool
    SupportsParallel  bool  // 是否支持并行工具调用
    CostPerInputToken float64
    CostPerOutputToken float64
}
```

#### 模型选择器

```go
type ModelSelector struct {
    Primary   Provider  // 主模型（用户配置）
    Assistant Provider  // 辅助模型（DeepSeek-Flash）
}

// 根据任务类型选择模型
func (s *ModelSelector) Select(task TaskType) Provider {
    switch task {
    case TaskSummary, TaskClassify, TaskExtractMemory, TaskFormatCheck:
        return s.Assistant  // 简单任务用便宜模型
    default:
        return s.Primary    // 复杂任务用主模型
    }
}
```

**辅助模型负责的任务：**
- 会话摘要生成
- 意图分类（路由到哪个 Agent）
- 记忆提取（从对话中提取关键信息）
- 格式校验（检查输出是否符合预期格式）
- Token 计数估算

### 4.3 工具系统设计

#### 工具接口

```go
type Tool interface {
    Name() string
    Description() string          // 给 LLM 看的描述
    Parameters() JSONSchema       // 参数 schema
    Permission() PermissionLevel  // safe / confirm / dangerous
    Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}

type PermissionLevel int
const (
    PermSafe      PermissionLevel = iota  // 只读操作，无需确认
    PermConfirm                           // 写操作，smart 模式下需确认
    PermDangerous                         // 危险操作，始终需确认
)

type ToolResult struct {
    Success bool
    Output  string    // 给 LLM 看的结果
    Error   string    // 错误信息
}
```

#### 工具注册中心

```go
type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(tool Tool)
func (r *ToolRegistry) Get(name string) Tool
func (r *ToolRegistry) GetForAgent(agentType string) []Tool  // 不同 Agent 可用不同工具子集
func (r *ToolRegistry) ToSchema() []ToolDefinition           // 转为 LLM 工具定义格式
```

#### 工具清单与权限

| 工具 | 权限 | 说明 |
|------|------|------|
| file_read | safe | 读取文件内容 |
| file_write | confirm | 写入/创建文件（先展示 diff） |
| file_delete | dangerous | 删除文件 |
| directory_list | safe | 列出目录结构 |
| directory_create | confirm | 创建目录 |
| shell_exec | confirm | 执行 shell 命令 |
| code_search | safe | 搜索代码符号/文本 |
| web_search | safe | 联网搜索信息 |
| url_fetch | safe | 抓取 URL 内容 |
| git_status | safe | 查看 git 状态 |
| git_diff | safe | 查看变更 |
| git_commit | confirm | 提交代码 |
| db_query | confirm | 执行数据库查询 |
| db_execute | dangerous | 执行数据库写操作 |

### 4.4 上下文管理

#### 分层上下文结构

```
总上下文 = System层 + 项目层 + 记忆层 + 会话层 + 工具结果层

Token 分配策略（以 128K 模型为例）：
├── System层：~2K（固定）
├── 项目层：~4K（项目信息、目录结构、技术栈）
├── 记忆层：~2K（相关记忆片段）
├── 会话层：~80K（对话历史，动态增长）
├── 工具结果层：~30K（工具返回内容，用完可压缩）
└── 预留输出：~10K（给模型输出空间）
```

#### 摘要压缩流程

```
1. 检测当前 token 用量 > 阈值（70%）
2. 调用辅助模型生成摘要：
   - 输入：当前完整对话历史
   - 输出：结构化摘要（做了什么、关键决策、当前状态、待办）
3. 创建新会话：
   - System Prompt + 项目上下文 + 记忆 + 摘要
4. 旧会话归档到 ~/.recoding/sessions/archive/
5. 通知用户："上下文已压缩，继续工作"
```

### 4.5 记忆系统

#### 记忆分类

```
记忆
├── 项目记忆（~/.recoding/memory/project/ 或 .recoding/memory/）
│   ├── architecture.md    # 架构决策
│   ├── conventions.md     # 编码约定
│   ├── decisions.md       # 历史决策记录
│   └── context.md         # 项目背景信息
├── 全局记忆（~/.recoding/memory/global/）
│   ├── preferences.md     # 用户偏好
│   ├── tech_stack.md      # 常用技术栈
│   └── patterns.md        # 常用模式
└── 会话摘要（~/.recoding/sessions/summaries/）
    ├── 2024-01-01_dev_用户模块.md
    └── ...
```

#### 记忆写入时机

1. **会话结束时**：辅助模型自动提取关键信息，写入对应记忆文件
2. **用户显式指令**：用户说"记住xxx" → 直接写入
3. **Agent 主动记录**：发现重要架构决策时，Agent 在工具调用中写入记忆

#### 记忆读取时机

1. **会话开始时**：加载项目记忆 + 全局记忆的摘要
2. **任务相关时**：根据当前任务关键词，检索相关记忆片段
3. **后期 RAG**：embedding 向量化记忆，语义检索最相关的片段

### 4.6 会话管理

#### 会话生命周期

```
创建 → 活跃 → [摘要压缩 → 新会话] → 结束 → 归档
                                         ↓
                                    提取记忆写入
```

#### 会话存储结构

```
~/.recoding/sessions/
├── active/
│   └── session_20240101_143000.json   # 当前活跃会话
├── archive/
│   └── session_20240101_120000.json   # 已归档会话
└── summaries/
    └── session_20240101_120000_summary.md  # 会话摘要
```

#### 中断恢复

```go
type SessionState struct {
    ID          string
    AgentType   string        // dev/review/plan...
    Status      string        // active/interrupted/completed
    Messages    []Message     // 对话历史
    PendingTool *ToolCall     // 中断时正在执行的工具
    Checkpoint  *Checkpoint   // 最后一个稳定状态
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

启动时检测是否有 interrupted 状态的会话，提示用户是否恢复。

### 4.7 安全机制

#### 用户确认模式

```yaml
# config.yaml
security:
  confirm_mode: smart  # auto / smart / manual
```

| 模式 | 行为 |
|------|------|
| auto | 所有操作自动执行，不询问 |
| smart（默认） | PermConfirm 和 PermDangerous 的操作需确认 |
| manual | 每个工具调用都需确认 |

#### Diff 确认流程（文件写入）

```
Agent 要写入文件 src/user.go:

┌─ 文件变更确认 ─────────────────────────────┐
│ 文件: src/user.go (新建)                    │
│                                            │
│ + package user                             │
│ +                                          │
│ + type User struct {                       │
│ +     ID    int64                          │
│ +     Name  string                         │
│ +     Email string                         │
│ + }                                        │
│                                            │
│ [y] 确认  [n] 拒绝  [e] 编辑  [a] 全部确认 │
└────────────────────────────────────────────┘
```

#### 命令执行沙箱

- 可配置允许/禁止的命令白名单/黑名单
- 默认禁止：`rm -rf /`、`format`、`shutdown` 等危险命令
- 命令执行超时限制（默认 30 秒）
- 工作目录限制（只能在项目目录内操作）


---

## 五、配置系统

### 5.1 配置文件位置与优先级

```
优先级从高到低：
1. 命令行 flag（--model, --provider 等）
2. 环境变量（RECODING_API_KEY 等）
3. 项目级配置（.recoding/config.yaml）
4. 全局配置（~/.recoding/config.yaml）
5. 内置默认值
```

### 5.2 完整配置结构

```yaml
# ~/.recoding/config.yaml
version: 1

# 主模型配置
provider:
  name: openai          # openai / anthropic / deepseek
  api_key: "sk-xxx"    # 或通过环境变量 RECODING_API_KEY
  base_url: ""         # 自定义 endpoint（兼容 OpenAI 格式的服务）
  model: "gpt-4o"      # 模型名称
  temperature: 0.7
  max_tokens: 4096     # 单次最大输出 token

# 辅助模型配置（内置默认值，用户可覆盖）
assistant_provider:
  name: deepseek
  api_key: "sk-xxx"    # 辅助模型的 key
  base_url: "https://api.deepseek.com"
  model: "deepseek-chat"  # DeepSeek-Flash
  temperature: 0.3

# Agent 配置
agent:
  max_iterations: 20       # 最大循环次数
  tool_timeout: 30s        # 工具执行超时
  parallel_tools: true     # 是否允许并行工具调用

# 上下文配置
context:
  summary_threshold: 0.7   # Token 使用率达到此值触发摘要
  max_tool_output: 4096    # 单个工具输出最大 token（超出截断）
  project_scan_depth: 3    # 项目结构扫描深度

# 安全配置
security:
  confirm_mode: smart      # auto / smart / manual
  blocked_commands:        # 禁止执行的命令
    - "rm -rf /"
    - "format"
    - "shutdown"
    - "reboot"
  allowed_paths: []        # 空表示不限制，否则只能操作列出的路径

# 会话配置
session:
  auto_save: true          # 自动保存会话
  max_history: 50          # 最多保留多少个历史会话

# 记忆配置
memory:
  enabled: true
  auto_extract: true       # 会话结束自动提取记忆
  max_inject_tokens: 2048  # 注入到上下文的记忆最大 token

# 输出配置
output:
  color: true              # 彩色输出
  markdown: true           # Markdown 渲染
  show_thinking: false     # 是否展示 Agent 思考过程
  show_token_usage: true   # 显示 token 用量
  language: zh             # 界面语言

# 命令别名
aliases:
  r: review
  d: dev
  p: plan
  t: test
```

### 5.3 项目级配置

```yaml
# .recoding/config.yaml（项目根目录）
version: 1

# 项目信息（自动检测 + 手动补充）
project:
  name: "my-project"
  language: "go"
  framework: "gin"
  description: "一个电商后端服务"
  conventions:
    - "使用 4 空格缩进"
    - "错误处理使用 pkg/errors 包装"
    - "API 返回统一用 Response 结构体"

# 项目级覆盖（可覆盖全局配置的任何字段）
agent:
  max_iterations: 30  # 这个项目允许更多循环
```

---

## 六、日志与可观测性

### 6.1 日志系统

```
~/.recoding/logs/
├── recoding.log           # 主日志（INFO 级别）
├── debug/
│   └── 2024-01-01.log    # Debug 日志（完整 prompt、响应）
└── usage/
    └── 2024-01.json      # 月度 token 用量统计
```

#### 日志级别

| 级别 | 内容 |
|------|------|
| ERROR | 工具执行失败、API 调用失败、不可恢复错误 |
| WARN | 重试、token 接近上限、配置缺失使用默认值 |
| INFO | 命令执行、Agent 循环开始/结束、工具调用摘要 |
| DEBUG | 完整 prompt、LLM 原始响应、工具调用详情 |

### 6.2 Token 用量统计

```go
type UsageRecord struct {
    Timestamp    time.Time
    SessionID    string
    Model        string
    InputTokens  int
    OutputTokens int
    EstCost      float64  // 估算费用
    AgentType    string   // dev/review/plan...
}
```

用户可通过 `recoding-cli usage` 查看：
- 今日/本周/本月用量
- 按模型分类统计
- 费用估算

### 6.3 执行回放

Debug 模式下记录完整的 Agent 执行过程，支持回放：

```bash
recoding-cli replay --session <session_id>
```

逐步展示 Agent 的思考过程、工具调用、结果，方便调试和面试演示。

---

## 七、错误处理与重试机制

### 7.1 错误分类与策略

```go
type ErrorType int
const (
    ErrNetwork      ErrorType = iota  // 网络错误
    ErrRateLimit                      // API 限流
    ErrAuthFailed                     // 认证失败
    ErrModelRefused                   // 模型拒绝回答
    ErrParseFailure                   // 响应解析失败
    ErrToolFailed                     // 工具执行失败
    ErrContextOverflow                // 上下文溢出
    ErrTimeout                        // 超时
)
```

| 错误类型 | 重试策略 | 最大重试 | 退避方式 |
|---------|---------|---------|---------|
| 网络错误 | 自动重试 | 3 次 | 指数退避（1s, 2s, 4s） |
| API 限流 | 等待后重试 | 5 次 | 按 Retry-After 头或指数退避 |
| 认证失败 | 不重试，提示用户检查配置 | 0 | - |
| 模型拒绝 | 改写 prompt 重试 | 1 次 | - |
| 解析失败 | 追加格式提示重试 | 2 次 | - |
| 工具失败 | 不重试同样操作，反馈错误给 LLM 重新推理 | - | - |
| 上下文溢出 | 触发摘要压缩后重试 | 1 次 | - |
| 超时 | 自动重试 | 2 次 | 固定间隔 |

### 7.2 工具失败的自愈流程

```
工具执行失败
    ↓
将错误信息作为工具结果返回给 LLM：
"工具 shell_exec 执行失败：command not found: npm"
    ↓
LLM 重新推理：
"npm 未安装，尝试使用 yarn 代替"
    ↓
调用新的工具/新的参数
    ↓
成功 → 继续
失败 → 再次反馈，直到达到最大循环次数
```

### 7.3 全局熔断

- 连续 5 次 API 调用失败 → 熔断，停止调用，提示用户检查网络/配置
- 单次会话 token 费用超过阈值（可配置）→ 警告用户是否继续

---

## 八、各 Agent 定义

### 8.1 Dev Agent（代码生成）

```yaml
name: dev
description: "根据用户需求生成代码"
system_prompt: prompts/dev.tmpl
tools:
  - file_read
  - file_write
  - directory_list
  - directory_create
  - shell_exec
  - code_search
  - web_search
  - url_fetch
  - git_status
  - git_diff
output_format: "代码 + 解释"
```

**工作流程：**
1. 理解用户需求
2. 查看项目结构和相关代码（自动）
3. 制定实现方案
4. 逐步生成代码，每个文件展示 diff 确认
5. 生成完毕后可选执行测试验证

### 8.2 Review Agent（代码审查）

```yaml
name: review
description: "审查代码质量、发现问题、给出改进建议"
system_prompt: prompts/review.tmpl
tools:
  - file_read
  - directory_list
  - code_search
  - git_diff
output_format: "问题列表 + 严重程度 + 修复建议"
```

**工作流程：**
1. 读取目标文件/目录
2. 分析代码质量（逻辑错误、安全漏洞、性能问题、代码风格）
3. 输出结构化审查报告
4. 可选：自动生成修复代码

### 8.3 Plan Agent（项目规划）

```yaml
name: plan
description: "根据需求拆解项目模块、生成技术方案"
system_prompt: prompts/plan.tmpl
tools:
  - file_read
  - directory_list
  - web_search
  - url_fetch
output_format: "结构化计划文档"
```

**工作流程：**
1. 理解用户需求
2. 搜索相关技术方案
3. 拆解模块和任务
4. 输出：模块划分、技术选型、目录结构、开发步骤、时间估算

### 8.4 Test Agent（测试生成）

```yaml
name: test
description: "为现有代码生成测试用例"
system_prompt: prompts/test.tmpl
tools:
  - file_read
  - file_write
  - directory_list
  - code_search
  - shell_exec
output_format: "测试代码 + 覆盖说明"
```

**工作流程：**
1. 读取目标代码
2. 分析函数签名、边界条件
3. 生成测试用例（单元测试 + 边界测试）
4. 执行测试验证是否通过

### 8.5 Chat Agent（自由对话）

```yaml
name: chat
description: "交互式对话，自由使用所有工具"
system_prompt: prompts/chat.tmpl
tools: all  # 所有工具
output_format: "自由格式"
```

进入交互式会话模式，用户可以自由对话，Agent 根据需要调用任何工具。


---

## 九、CLI 交互设计

### 9.1 命令结构

```bash
recoding-cli [command] [args] [flags]

# 核心命令
recoding-cli dev "实现用户注册接口"          # 代码生成
recoding-cli review ./src/                   # 代码审查
recoding-cli plan "电商系统"                 # 项目规划
recoding-cli test ./src/user/               # 生成测试
recoding-cli chat                           # 交互式会话

# 管理命令
recoding-cli config init                    # 初始化配置
recoding-cli config set provider.model gpt-4o
recoding-cli config show                    # 查看当前配置
recoding-cli session list                   # 查看历史会话
recoding-cli session resume <id>            # 恢复会话
recoding-cli usage                          # 查看用量统计
recoding-cli memory show                    # 查看记忆
recoding-cli memory clear                   # 清除记忆
recoding-cli replay <session_id>            # 回放执行过程

# 全局 flag
--model, -m       # 临时指定模型
--verbose, -v     # 详细输出（显示思考过程）
--debug           # Debug 模式（记录完整日志）
--no-confirm      # 跳过所有确认（等同 auto 模式）
--dry-run         # 只展示计划，不执行
```

### 9.2 交互式会话模式

```bash
$ recoding-cli chat

╭─ recoding-cli v0.1.0 ─────────────────────╮
│ 模型: gpt-4o | 项目: my-project (Go/Gin)  │
│ 输入 /help 查看命令，Ctrl+C 退出          │
╰───────────────────────────────────────────╯

> 帮我实现一个用户注册接口

🤔 正在分析项目结构...
📂 读取 internal/handler/
📂 读取 internal/model/

我来为你实现用户注册接口。基于项目现有结构，我会创建以下文件：

1. `internal/model/user.go` - 用户模型
2. `internal/handler/user.go` - 注册处理器
3. `internal/service/user.go` - 业务逻辑

┌─ 文件变更: internal/model/user.go (新建) ─┐
│ + package model                            │
│ +                                          │
│ + type User struct {                       │
│ +     ID        int64  `json:"id"`         │
│ +     Username  string `json:"username"`   │
│ +     Email     string `json:"email"`      │
│ +     Password  string `json:"-"`          │
│ + }                                        │
└────────────────────────────────────────────┘
[y] 确认  [n] 拒绝  [e] 编辑  [a] 全部确认

> /help

可用命令：
  /help          显示帮助
  /clear         清除当前会话
  /save          保存会话
  /undo          撤销上一次文件变更
  /mode [auto|smart|manual]  切换确认模式
  /model [name]  切换模型
  /usage         查看本次会话用量
  /exit          退出

> /usage

本次会话用量：
  输入 Token: 12,450
  输出 Token: 3,200
  工具调用: 8 次
  估算费用: ¥0.15
```

### 9.3 输出渲染规范

- **代码块**：语法高亮（基于语言自动检测）
- **Diff**：绿色新增、红色删除
- **状态提示**：emoji + 文字（🤔思考中、📂读取文件、✅完成、❌失败）
- **进度**：流式输出文字，工具调用时显示 spinner
- **错误**：红色高亮 + 建议操作

---

## 十、开发阶段规划

### 阶段一：基础骨架（第 1-2 周）

**目标：** 跑通最小闭环 —— 输入需求 → 调用 LLM → 输出代码

| 任务 | 详情 |
|------|------|
| 项目初始化 | Go module、目录结构、Makefile |
| CLI 框架搭建 | Cobra 根命令 + dev 子命令 |
| 配置系统 | Viper 加载 YAML 配置 |
| Provider 层 | OpenAI 兼容接口实现（含流式） |
| 基础 Agent 循环 | 单轮对话：prompt → LLM → 输出 |
| 终端渲染 | 流式文本输出 + Markdown 渲染 |

**交付物：** 能执行 `recoding-cli dev "hello world"` 并得到流式代码输出

### 阶段二：工具系统（第 3-4 周）

**目标：** Agent 能调用工具，实现真正的代码生成

| 任务 | 详情 |
|------|------|
| 工具接口定义 | Tool interface + Registry |
| 基础工具实现 | file_read、file_write、directory_list、shell_exec、code_search |
| 工具调用解析 | 解析 LLM 返回的 function_call |
| Agent 多轮循环 | ReAct 循环完整实现 |
| Diff 确认机制 | 文件写入前展示 diff，等待用户确认 |
| 权限系统 | 工具权限等级 + 确认模式 |

**交付物：** Agent 能读取项目文件、生成代码、写入文件（带确认）

### 阶段三：上下文与记忆（第 5-6 周）

**目标：** Agent 理解项目上下文，具备跨会话记忆

| 任务 | 详情 |
|------|------|
| 项目感知 | 自动检测语言/框架/目录结构 |
| 上下文管理 | Token 计数、分层上下文构建 |
| 摘要压缩 | 辅助模型生成摘要、会话切换 |
| 辅助模型集成 | DeepSeek-Flash 接入 + 模型选择器 |
| 记忆存储 | 项目记忆 + 全局记忆文件读写 |
| 记忆提取 | 会话结束自动提取关键信息 |
| 记忆注入 | 会话开始时加载相关记忆 |

**交付物：** Agent 能记住项目信息和用户偏好，上下文过长时自动压缩

### 阶段四：多 Agent 与会话（第 7-8 周）

**目标：** 完整的多命令 Agent 系统 + 会话管理

| 任务 | 详情 |
|------|------|
| Agent 定义系统 | Prompt 模板 + 工具集配置 |
| Review Agent | 代码审查功能 |
| Plan Agent | 项目规划功能 |
| Test Agent | 测试生成功能 |
| Chat 交互模式 | 交互式会话 + 会话内命令 |
| 会话管理 | 保存、恢复、归档、列表 |
| 中断恢复 | Ctrl+C 安全中断 + 下次恢复 |

**交付物：** 完整的多 Agent CLI 工具，支持 dev/review/plan/test/chat

### 阶段五：增强与打磨（第 9-10 周）

**目标：** 生产级质量，可作为日常工具使用

| 任务 | 详情 |
|------|------|
| Anthropic Provider | Claude 模型支持 |
| 增强工具 | web_search、url_fetch、git 操作、数据库操作 |
| 错误处理完善 | 全部重试策略、熔断机制 |
| 日志系统 | 结构化日志 + 用量统计 |
| 执行回放 | replay 命令 |
| 命令别名 | 用户自定义别名 |
| 配置迁移 | 版本升级自动迁移 |
| 性能优化 | 并行工具调用、响应缓存 |

**交付物：** 生产可用的完整工具

### 阶段六：扩展能力（后期）

| 任务 | 详情 |
|------|------|
| 插件系统 | 用户自定义工具 |
| Agent Pipeline | Agent 间协作（dev → review 自动触发） |
| RAG 记忆 | 向量化记忆 + 语义检索 |
| 多用户支持 | 团队配置共享（预留） |
| Docker/部署工具 | 运维类工具集成 |

---

## 十一、测试策略

### 11.1 测试分层

```
┌─────────────────────────────────────┐
│         E2E 测试（端到端）           │  少量，验证完整流程
├─────────────────────────────────────┤
│         集成测试                     │  模块间协作
├─────────────────────────────────────┤
│         单元测试                     │  大量，每个模块独立测试
└─────────────────────────────────────┘
```

### 11.2 各层测试方案

**单元测试：**
- Provider 层：mock HTTP 响应，测试请求构建和响应解析
- 工具层：mock 文件系统（afero），测试工具逻辑
- 上下文管理：测试 token 计算、压缩触发逻辑
- 记忆系统：测试读写、提取逻辑
- 配置系统：测试多层覆盖、默认值

**集成测试：**
- Agent 循环 + mock Provider：验证多轮对话、工具调用流程
- 录制真实 LLM 响应作为 fixture，回放测试

**E2E 测试：**
- 启动 CLI 进程，输入命令，验证输出和文件变更
- 使用便宜模型（DeepSeek-Flash）跑真实 E2E

### 11.3 测试工具

- `testing` + `testify`：断言和 mock
- `afero`：内存文件系统 mock
- `httptest`：HTTP mock server
- `go-cmp`：结构体深度比较

---

## 十二、项目质量保障

### 12.1 代码规范

- `golangci-lint` 静态检查（配置 .golangci.yml）
- `gofmt` / `goimports` 格式化
- 提交前 pre-commit hook 自动检查

### 12.2 CI/CD

```yaml
# GitHub Actions
on: [push, pull_request]
jobs:
  lint:    golangci-lint
  test:    go test ./...
  build:   go build（多平台交叉编译）
  release: goreleaser（tag 触发发布）
```

### 12.3 发布

- 使用 GoReleaser 自动构建多平台二进制
- 支持 `brew install`（macOS）、`scoop install`（Windows）
- GitHub Releases 分发

---

## 十三、关键设计决策总结

| 决策 | 选择 | 理由 |
|------|------|------|
| Agent 框架 | 自研 | Go 生态无成熟框架，自研可控且面试加分 |
| 工具调用格式 | 跟随 Provider 原生格式 | OpenAI function_calling / Anthropic tool_use |
| 记忆存储 | 先文件后 RAG | 文件方案简单可靠，RAG 作为演进方向 |
| 上下文压缩 | 辅助模型摘要 | 低成本，不消耗主模型额度 |
| 文件变更 | Diff 确认 | 安全第一，用户可控 |
| 配置格式 | YAML | 可读性好，Viper 原生支持 |
| 日志 | 结构化 JSON | 方便后期分析和回放 |
| 国际化 | 预留但先做中文 | 文本抽离到 i18n 包，后期加语言零成本 |

---

## 十四、风险与应对

| 风险 | 影响 | 应对 |
|------|------|------|
| LLM 工具调用不稳定 | Agent 循环失败 | 重试 + prompt 优化 + 格式校验 |
| Token 费用过高 | 用户弃用 | 双模型架构 + 用量统计 + 费用预警 |
| 上下文丢失关键信息 | 摘要后 Agent 遗忘 | 摘要质量优化 + 关键信息强制保留 |
| 工具执行破坏用户环境 | 数据丢失 | Diff 确认 + 备份 + 沙箱 |
| 模型 API 变更 | Provider 层失效 | 接口抽象隔离变更影响 |
| Go 生态 LLM 库不成熟 | 开发受阻 | 核心逻辑自研，只依赖 HTTP 调用 |

---

## 十五、面试亮点提炼

这个项目在面试中可以展开讲的技术点：

1. **架构设计**：分层架构、接口抽象、依赖注入
2. **设计模式**：策略模式（Provider 切换）、观察者模式（事件流）、模板方法（Agent 定义）
3. **并发编程**：Go channel 流式处理、并行工具调用、超时控制
4. **系统设计**：上下文窗口管理、记忆系统、会话状态机
5. **工程实践**：配置管理、错误处理、日志体系、测试策略
6. **AI 工程**：Prompt 工程、ReAct 循环、双模型架构、RAG（后期）
7. **产品思维**：安全机制、用户体验、成本控制

---

## 十六、第一步行动

确认计划书无误后，开发顺序：

```
1. go mod init github.com/xxx/recoding-cli
2. 搭建目录结构
3. 实现 Cobra 根命令 + dev 子命令（空壳）
4. 实现配置加载（Viper + YAML）
5. 实现 OpenAI Provider（流式）
6. 实现最简 Agent 循环（单轮）
7. 终端流式输出
8. 跑通第一个 "recoding-cli dev 'hello world'"
```

---

*计划书版本: v1.0*
*创建时间: 2026-05-18*
*状态: 待确认*
