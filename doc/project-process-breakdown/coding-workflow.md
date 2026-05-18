# 编码流程：多 Agent 协作执行

## 概述

每个阶段的每个 Step 在实际编码时，采用多 Agent 协作流程执行。通过 Plan → Implement → Review 三阶段 DAG 管道，确保代码质量和一致性。

## Agent 角色定义

| Agent | 角色 | 职责 |
|-------|------|------|
| Planner | 规划者 | 分析 Step 需求，输出详细实现方案（文件列表、接口设计、依赖关系） |
| Implementer | 实现者 | 根据方案编写代码，创建文件，运行测试 |
| Reviewer | 审查者 | 审查实现代码，检查是否符合方案和项目规范 |

## 执行流程（DAG）

```
┌──────────────────────────────────────────────────────────────┐
│                    单个 Step 的执行流程                        │
│                                                              │
│  ┌─────────┐     ┌──────────────┐     ┌──────────┐          │
│  │ Planner │────▶│ Implementer  │────▶│ Reviewer │          │
│  └─────────┘     └──────────────┘     └──────────┘          │
│       │                  │                   │               │
│       ▼                  ▼                   ▼               │
│  实现方案文档       代码文件+测试        审查报告+修复建议    │
│                                                              │
│  如果 Review 不通过：                                         │
│  Reviewer ──修复建议──▶ Implementer ──修复──▶ Reviewer       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## 详细流程

### Stage 1: Plan（规划）

**输入：**
- Step 描述（来自阶段流程文档）
- 项目当前代码状态
- 项目计划书中的相关设计

**Planner Agent Prompt 模板：**
```
你是一个 Go 项目的技术方案规划者。

## 任务
根据以下 Step 需求，输出详细的实现方案。

## Step 需求
{step_description}

## 项目上下文
{project_context}

## 相关设计参考
{design_reference}

## 输出格式
1. 需要创建/修改的文件列表
2. 每个文件的核心接口/结构体设计
3. 文件间依赖关系
4. 实现注意事项
5. 测试要点
```

**输出：** 结构化实现方案（Markdown）

---

### Stage 2: Implement（实现）

**输入：**
- Planner 输出的实现方案
- 项目现有代码

**Implementer Agent 行为：**
1. 读取实现方案
2. 按文件依赖顺序逐个实现
3. 每个文件：读取相关代码 → 编写 → 写入
4. 编写对应测试文件
5. 运行 `go build ./...` 验证编译
6. 运行 `go test ./...` 验证测试

**工具集：** file_read, file_write, directory_list, directory_create, shell_exec, code_search

---

### Stage 3: Review（审查）

**输入：**
- Planner 的实现方案（作为标准）
- Implementer 生成的代码

**Reviewer Agent 检查项：**
1. **完整性**：方案中的所有文件和接口是否都已实现
2. **正确性**：逻辑是否正确，边界条件是否处理
3. **规范性**：是否符合项目编码约定（从记忆/配置读取）
4. **测试覆盖**：关键路径是否有测试
5. **安全性**：是否有安全隐患

**输出：**
- PASS：审查通过，进入下一个 Step
- FAIL + 修复建议：返回给 Implementer 修复

---

## 使用示例

### 执行阶段一 Step 4（Provider 层）

```
任务：实现阶段一 Step 4 - Provider 层（OpenAI 兼容）

Stage 配置：
┌─────────────────────────────────────────────────────────┐
│ Stage: plan                                             │
│ Role: planner                                           │
│ Prompt: 分析 Provider 层需求，输出实现方案               │
│ 输入: phase-1-skeleton.md Step 4 描述 + 项目计划书设计   │
│ depends_on: []                                          │
├─────────────────────────────────────────────────────────┤
│ Stage: implement                                        │
│ Role: implementer                                       │
│ Prompt: 根据方案实现 Provider 层代码                     │
│ 输入: plan 阶段输出                                     │
│ depends_on: [plan]                                      │
├─────────────────────────────────────────────────────────┤
│ Stage: review                                           │
│ Role: reviewer                                          │
│ Prompt: 审查 Provider 层实现是否符合方案和规范           │
│ 输入: plan 输出 + implement 输出                        │
│ depends_on: [implement]                                 │
└─────────────────────────────────────────────────────────┘
```

### 实际调用方式（subagent DAG）

```yaml
task: "实现 Provider 层 - OpenAI 兼容接口（含流式）"
stages:
  - name: plan
    role: kiro_planner
    prompt_template: |
      分析以下需求并输出详细实现方案：
      
      ## 需求
      实现 internal/provider/ 包：
      - interface.go: Provider 接口定义（ChatStream, Chat, ModelMeta）
      - openai.go: OpenAI 兼容实现（含流式 SSE 解析）
      - model_meta.go: 模型元信息结构体
      
      ## 设计参考
      {design_from_plan}
      
      ## 输出
      文件列表、接口设计、实现要点、测试要点
    depends_on: []

  - name: implement
    role: kiro_default
    prompt_template: |
      根据以下方案实现代码：
      
      {plan_output}
      
      要求：
      1. 按文件依赖顺序实现
      2. 编写单元测试
      3. 确保 go build 通过
    depends_on: [plan]

  - name: review
    role: kiro_default
    prompt_template: |
      审查以下实现是否符合方案：
      
      ## 原始方案
      {plan_output}
      
      ## 实现结果
      {implement_output}
      
      检查：完整性、正确性、规范性、测试覆盖
      输出：PASS 或 FAIL + 具体修复建议
    depends_on: [implement]
```

---

## 流程控制规则

### 1. 何时触发多 Agent 流程

- 每个阶段的每个 Step 都使用此流程
- 简单的配置修改或文档更新可跳过 Review 阶段

### 2. Review 不通过时

```
Review FAIL
    ↓
提取修复建议
    ↓
Implementer 根据建议修复
    ↓
再次 Review
    ↓
最多重试 2 次，仍不通过则人工介入
```

### 3. 跨 Step 依赖

```
Step N 完成（Review PASS）
    ↓
Step N+1 的 Planner 可以读取 Step N 的代码
    ↓
保证增量开发的连贯性
```

### 4. 并行执行

同一阶段内无依赖的 Step 可以并行执行：
```
例如阶段二：
- Step 2a: file.go（无依赖）  ─┐
- Step 2b: directory.go       ├─ 可并行
- Step 2c: shell.go           ─┘
```

---

## 质量门禁

每个 Step 完成后必须满足：

| 检查项 | 标准 |
|--------|------|
| 编译 | `go build ./...` 无错误 |
| 测试 | `go test ./...` 全部通过 |
| Lint | `golangci-lint run` 无 error |
| Review | Reviewer Agent 输出 PASS |

全部通过后才进入下一个 Step。

---

## 执行记录

每个 Step 执行后生成记录：

```
doc/project-process-breakdown/execution-log/
├── phase-1/
│   ├── step-1-plan.md
│   ├── step-1-review.md
│   ├── step-2-plan.md
│   └── ...
└── phase-2/
    └── ...
```

记录内容：
- 方案摘要
- 创建/修改的文件列表
- 测试结果
- Review 结论
- 耗时和 token 用量
