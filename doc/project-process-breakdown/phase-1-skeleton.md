# 阶段一：基础骨架搭建

## 目标

跑通最小闭环 —— `recoding-cli dev "hello world"` 输入需求 → 调用 LLM → 流式输出代码

## 前置条件

- Go 1.22+ 已安装
- 有可用的 OpenAI 兼容 API Key

## 交付物

- 可编译的 CLI 二进制
- 执行 `recoding-cli dev "实现一个hello world"` 能得到流式代码输出

---

## 步骤拆分

### Step 1: 项目初始化

**输入：** 无
**输出：** Go module + 目录结构 + Makefile

```
操作：
1. go mod init github.com/xxx/recoding-cli
2. 创建目录结构：
   - cmd/
   - internal/agent/
   - internal/provider/
   - internal/config/
   - internal/render/
   - prompts/
   - configs/
3. 创建 Makefile（build、run、test、lint）
4. 创建 .gitignore
```

**验收：** `go build ./...` 无报错

---

### Step 2: CLI 框架搭建（Cobra）

**输入：** 项目骨架
**输出：** 根命令 + dev 子命令（空壳）

```
文件：
- cmd/root.go      # 根命令，全局 flag（--model, --verbose, --debug）
- cmd/dev.go       # dev 子命令，接收 positional arg 作为需求描述
- main.go          # 入口，调用 cmd.Execute()
```

**验收：** `recoding-cli dev "test"` 能运行，输出占位文本

---

### Step 3: 配置系统（Viper）

**输入：** CLI 框架
**输出：** 多层配置加载

```
文件：
- internal/config/config.go   # 配置结构体定义
- internal/config/loader.go   # 加载逻辑（全局 → 项目 → 环境变量 → flag）
- configs/default.yaml        # 默认配置模板

配置结构（最小集）：
- provider.name
- provider.api_key
- provider.base_url
- provider.model
- provider.temperature
- provider.max_tokens
```

**验收：** 能从 `~/.recoding/config.yaml` 或环境变量读取 API Key

---

### Step 4: Provider 层（OpenAI 兼容）

**输入：** 配置系统
**输出：** 能调用 OpenAI 兼容 API 的 Provider

```
文件：
- internal/provider/interface.go  # Provider 接口定义
- internal/provider/openai.go     # OpenAI 兼容实现（含流式）
- internal/provider/model_meta.go # 模型元信息

核心接口：
- ChatStream(ctx, req) → <-chan StreamEvent
- Chat(ctx, req) → *ChatResponse
- ModelMeta() → *ModelMeta
```

**验收：** 单元测试通过，能调用真实 API 获取响应

---

### Step 5: 基础 Agent 循环（单轮）

**输入：** Provider 层
**输出：** 最简 Agent Runtime（单轮对话，无工具）

```
文件：
- internal/agent/runtime.go     # Agent 执行循环
- internal/agent/definition.go  # Agent 定义（dev agent 的 system prompt）
- prompts/dev.tmpl              # Dev Agent 的 prompt 模板

流程：
1. 加载 system prompt 模板
2. 拼装 messages（system + user）
3. 调用 Provider.ChatStream
4. 流式输出结果
```

**验收：** 输入需求描述，能得到 LLM 的代码响应

---

### Step 6: 终端渲染

**输入：** Agent 循环输出
**输出：** 美化的终端输出

```
文件：
- internal/render/terminal.go  # Markdown 渲染（glamour）
- internal/render/stream.go    # 流式输出处理

功能：
- 流式逐字输出
- 代码块语法高亮
- 状态提示（emoji）
```

**验收：** 输出有语法高亮和格式化

---

## 集成验收

```bash
# 1. 配置 API Key
export RECODING_API_KEY="sk-xxx"

# 2. 构建
make build

# 3. 运行
./recoding-cli dev "用 Go 实现一个 hello world HTTP 服务器"

# 预期：流式输出带语法高亮的 Go 代码
```

## 关键依赖

```
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/sashabaranov/go-openai
go get github.com/charmbracelet/glamour
go get github.com/charmbracelet/lipgloss
```
