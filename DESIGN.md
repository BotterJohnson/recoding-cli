# DESIGN.md - recoding-cli 架构与编码规范

## 架构概览

```
用户终端
    │
CLI 层 (Cobra)          ← 命令路由
    │
Agent Runtime           ← 核心引擎（ReAct 循环）
    │
Provider 层             ← LLM 调用抽象
    │
工具层 (Tools)          ← 文件/命令/搜索等能力
```

## 目录结构

```
recoding-cli/
├── main.go                     # 入口
├── cmd/                        # CLI 命令定义（Cobra）
│   ├── root.go                 # 根命令、全局 flag
│   ├── dev.go                  # dev 子命令
│   ├── review.go               # review 子命令
│   ├── plan.go                 # plan 子命令
│   ├── test.go                 # test 子命令
│   ├── chat.go                 # chat 交互式会话
│   └── config.go               # 配置管理命令
├── internal/
│   ├── agent/                  # Agent 核心引擎
│   │   ├── runtime.go          # Agent 执行循环（ReAct）
│   │   ├── definition.go       # Agent 定义
│   │   └── router.go           # 意图路由
│   ├── provider/               # LLM Provider 抽象
│   │   ├── interface.go        # 统一接口
│   │   ├── openai.go           # OpenAI 兼容实现
│   │   ├── anthropic.go        # Anthropic 实现
│   │   ├── model_meta.go       # 模型元信息
│   │   └── selector.go         # 主/辅模型选择器
│   ├── tools/                  # 工具实现
│   │   ├── interface.go        # 工具统一接口
│   │   ├── registry.go         # 工具注册中心
│   │   ├── file.go             # 文件读写
│   │   ├── directory.go        # 目录操作
│   │   ├── shell.go            # 命令执行
│   │   ├── search.go           # 代码搜索
│   │   ├── web.go              # 联网搜索
│   │   ├── fetch.go            # URL 抓取
│   │   └── git.go              # Git 操作
│   ├── context/                # 上下文管理
│   │   ├── manager.go          # 上下文生命周期
│   │   ├── project.go          # 项目感知
│   │   ├── summary.go          # 摘要生成
│   │   └── window.go           # Token 窗口计算
│   ├── memory/                 # 记忆系统
│   │   ├── store.go            # 记忆存储接口
│   │   ├── project.go          # 项目级记忆
│   │   ├── global.go           # 全局记忆
│   │   └── extractor.go        # 记忆提取
│   ├── session/                # 会话管理
│   │   ├── manager.go          # 会话创建/恢复/归档
│   │   ├── history.go          # 对话历史
│   │   └── interrupt.go        # 中断与恢复
│   ├── render/                 # 输出渲染
│   │   ├── terminal.go         # Markdown 渲染
│   │   ├── diff.go             # Diff 展示
│   │   └── stream.go           # 流式输出
│   ├── config/                 # 配置管理
│   │   ├── config.go           # 配置结构体
│   │   └── loader.go           # 多层配置加载
│   ├── security/               # 安全机制
│   │   ├── confirm.go          # 用户确认
│   │   └── permission.go       # 工具权限
│   └── logger/                 # 日志
│       └── logger.go           # 日志初始化
├── prompts/                    # Prompt 模板文件
│   ├── dev.tmpl
│   ├── review.tmpl
│   ├── plan.tmpl
│   ├── test.tmpl
│   └── summary.tmpl
├── configs/
│   └── default.yaml            # 默认配置
├── go.mod
├── go.sum
└── Makefile
```

## 核心接口

### Provider

```go
type Provider interface {
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ModelMeta() *ModelMeta
}
```

### Tool

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() JSONSchema
    Permission() PermissionLevel
    Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}
```

### PermissionLevel

```go
const (
    PermSafe      PermissionLevel = iota  // 只读，无需确认
    PermConfirm                           // 写操作，smart 模式需确认
    PermDangerous                         // 危险操作，始终确认
)
```

## 编码规范

### 命名

- 包名：小写单词，不用下划线（`provider`, `tools`, `config`）
- 文件名：小写 + 下划线（`model_meta.go`）
- 接口：动词或名词（`Provider`, `Tool`）
- 结构体：名词（`ChatRequest`, `ToolResult`）
- 方法：驼峰，公开首字母大写

### 错误处理

```go
// 使用 %w 包装错误，保留错误链
if err != nil {
    return fmt.Errorf("provider chat: %w", err)
}

// 自定义错误类型用于需要类型判断的场景
type ProviderError struct {
    Type    ErrorType
    Message string
    Err     error
}
```

### 依赖注入

- 通过接口传递依赖，不使用全局变量
- 构造函数接收依赖：`func NewRuntime(provider Provider, registry *ToolRegistry) *Runtime`

### 并发

- 使用 `context.Context` 传递取消信号和超时
- channel 用于流式数据传递
- 避免裸 goroutine，确保有退出机制

### 测试

- 每个 `.go` 文件对应 `_test.go`
- 使用 `testify/assert` 断言
- mock 外部依赖（Provider 用 interface mock，文件系统用 afero）
- 表驱动测试优先

### 注释

```go
// NewOpenAIProvider 创建 OpenAI 兼容的 Provider 实例。
// baseURL 为空时使用默认 OpenAI endpoint。
func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
```

### 配置

- 优先级：flag > 环境变量 > 项目配置 > 全局配置 > 默认值
- 环境变量前缀：`RECODING_`
- 配置文件：YAML 格式

## 技术选型

| 类别 | 选型 |
|------|------|
| 语言 | Go 1.22+ |
| CLI | Cobra + Viper |
| LLM (OpenAI 兼容) | sashabaranov/go-openai |
| LLM (Anthropic) | anthropics/anthropic-sdk-go |
| 终端渲染 | lipgloss + glamour |
| 日志 | zerolog |
| 测试 | testing + testify |

## 设计原则

1. **接口隔离** — Provider、Tool 均通过接口抽象，实现可替换
2. **单一职责** — 每个包/文件职责明确，不混杂
3. **显式依赖** — 构造函数注入，不隐式依赖全局状态
4. **安全优先** — 写操作先 diff 确认，危险操作需授权
5. **渐进增强** — 先跑通最小闭环，再逐步增加能力
