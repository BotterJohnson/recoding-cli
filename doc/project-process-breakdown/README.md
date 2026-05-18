# recoding-cli 项目开发流程拆分

## 概述

本目录将 recoding-cli 项目计划拆分为可执行的开发流程文档，每个阶段包含明确的输入、输出、步骤和验收标准。

## 阶段总览

| 阶段 | 文档 | 周期 | 目标 |
|------|------|------|------|
| 一 | [phase-1-skeleton.md](./phase-1-skeleton.md) | 第 1-2 周 | 跑通最小闭环：输入需求 → 调用 LLM → 输出代码 |
| 二 | [phase-2-tools.md](./phase-2-tools.md) | 第 3-4 周 | Agent 能调用工具，实现真正的代码生成 |
| 三 | [phase-3-context-memory.md](./phase-3-context-memory.md) | 第 5-6 周 | Agent 理解项目上下文，具备跨会话记忆 |
| 四 | [phase-4-multi-agent.md](./phase-4-multi-agent.md) | 第 7-8 周 | 完整的多命令 Agent 系统 + 会话管理 |
| 五 | [phase-5-enhancement.md](./phase-5-enhancement.md) | 第 9-10 周 | 生产级质量，可作为日常工具使用 |

## 编码流程

| 文档 | 说明 |
|------|------|
| [coding-workflow.md](./coding-workflow.md) | 多 Agent 协作编码流程（规划→实现→审查） |

## 依赖关系

```
阶段一 → 阶段二 → 阶段三 → 阶段四 → 阶段五
                                        ↓
                              每个阶段内部使用 coding-workflow 执行
```

## 技术栈速查

- 语言：Go 1.22+
- CLI：Cobra + Viper
- LLM：OpenAI 兼容 / Anthropic
- 渲染：lipgloss + glamour
- 日志：zerolog
- 测试：testing + testify
