# AGENTS.md - recoding-cli 多 Agent 协作规范

## 项目概述

recoding-cli 是一个基于 LLM 的终端 AI 编程助手（Go 语言），采用多 Agent 协作模式进行 vibecoding 开发。

## Agent 角色

### Planner（规划者）

**职责：** 分析需求，输出实现方案

**输入：**
- Step 需求描述（来自 `doc/project-process-breakdown/` 阶段文档）
- 项目当前代码
- `doc/project-plan/recoding-cli-plan.md` 中的相关设计

**输出格式：**
1. 需要创建/修改的文件列表（含路径）
2. 核心接口/结构体设计（Go 代码片段）
3. 文件间依赖关系
4. 实现注意事项
5. 测试要点

**约束：**
- 不写实现代码，只输出方案
- 方案必须与 DESIGN.md 中的架构一致
- 文件路径必须符合项目目录结构规范

---

### Implementer（实现者）

**职责：** 根据方案编写代码

**输入：** Planner 输出的实现方案

**行为：**
1. 按文件依赖顺序逐个实现
2. 每个文件写入前读取相关已有代码
3. 编写对应 `_test.go` 测试文件
4. 运行 `go build ./...` 确认编译通过
5. 运行 `go test ./...` 确认测试通过

**约束：**
- 遵循 DESIGN.md 中的编码规范
- 不偏离 Planner 方案（如需调整须说明原因）
- 每个公开函数/方法必须有注释
- 错误处理使用 `fmt.Errorf("xxx: %w", err)` 包装

---

### Reviewer（审查者）

**职责：** 审查代码质量

**输入：** Planner 方案 + Implementer 代码

**检查项：**
1. **完整性** — 方案中所有文件和接口是否都已实现
2. **正确性** — 逻辑是否正确，边界条件是否处理
3. **规范性** — 是否符合 DESIGN.md 编码约定
4. **测试覆盖** — 关键路径是否有测试
5. **安全性** — 是否有安全隐患

**输出：**
- `PASS` — 进入下一个 Step
- `FAIL` + 具体修复建议列表 → 返回 Implementer

**约束：**
- 最多重试 2 次，仍不通过则标记需人工介入
- 不自行修改代码

---

## 执行流程

```
Planner → Implementer → Reviewer
                ↑            │
                └── FAIL ────┘
```

### 质量门禁

每个 Step 完成必须满足：

| 检查 | 命令 |
|------|------|
| 编译 | `go build ./...` |
| 测试 | `go test ./...` |
| Lint | `golangci-lint run` |
| Review | Reviewer 输出 PASS |

### 并行规则

- 同阶段内无依赖的 Step 可并行执行
- 有依赖的 Step 必须串行（前一个 PASS 后才开始下一个）

---

## 开发阶段索引

按 `doc/project-process-breakdown/` 中的阶段文档顺序执行：

1. `phase-1-skeleton.md` — 基础骨架
2. `phase-2-tools.md` — 工具系统
3. `phase-3-context-memory.md` — 上下文与记忆
4. `phase-4-multi-agent.md` — 多 Agent 与会话
5. `phase-5-enhancement.md` — 增强与打磨

每个阶段的每个 Step 均使用上述三 Agent 流程执行。
