# TestAgent：基于 Skill 编排的软件测试智能体

本项目实现了一个基于 **Agent Harness 思想** 和 **Skill 编排机制** 的软件测试智能体，并以自研的 **FlexMD Markdown 演示文稿编译器** 作为被测程序，完成从测试任务理解、测试用例设计、用例执行、结果分析到测试报告生成的完整流程。

项目的核心目标不是简单调用大语言模型生成测试建议，而是让智能体通过工具真实地读写文件、运行被测程序、分析输出结果，并将测试全过程产物持久化保存。

---

## 1. 项目简介

本项目由两部分组成：

1. **TestAgent 测试智能体**

   - 使用 Go 实现的命令行对话式智能体。
   - 支持大语言模型对话、工具调用、Skill 动态读取、会话状态保存与恢复。
   - 能通过 `read_file`、`write_file`、`execute_command` 等工具与本地项目交互。
   - 采用“万物皆 Skill”的设计思想，将测试方法、工具接口、被测对象说明和任务流程统一组织为 Skill。
2. **FlexMD 被测程序**

   - 一个小型 Markdown 变体编译器。
   - 支持将 `.fmd` 文本文件编译为 HTML 演示文稿。
   - 支持分页、横向/纵向 flex 分栏、比例布局、嵌套布局、代码块隔离和 AST JSON 输出。
   - 作为本实验的软件测试对象，用于验证测试智能体的实际测试能力。

---

## 2. 主要特性

### 2.1 Agent Harness 思想

本项目中的 Agent Harness 负责管理智能体运行所需的基础设施，包括：

- 系统提示词加载；
- Skill 索引加载；
- 对话上下文维护；
- 工具注册与调用；
- 命令执行与结果反馈；
- 会话状态持久化；
- 终端日志与工具调用过程可视化。

该设计使测试智能体不是一次性脚本，而是一个可以持续对话、持续调用工具、持续推进测试任务的交互式测试系统。

### 2.2 “万物皆 Skill”的设计

项目将测试知识与任务语境封装为 Skill：

- 通用工具接口是 Skill；
- 黑盒、白盒、集成、系统测试方法是 Skill；
- FlexMD 被测对象说明是 Skill；
- FlexMD 完整测试任务编排也是 Skill。

智能体启动时不一次性读取所有 Skill 正文，而是先加载 Skill 索引。当用户提出测试任务后，智能体根据任务语义主动读取相关 Skill 文件，再按 Skill 中定义的流程执行测试。

### 2.3 一条龙测试流程

针对 FlexMD，智能体可以完成：

1. 读取 FlexMD 规格和接口说明；
2. 读取通用测试方法 Skill；
3. 制定测试计划；
4. 自主编写 `.fmd` 测试用例；
5. 调用 FlexMD 编译器执行测试；
6. 读取 AST JSON、HTML、返回码和错误信息；
7. 判断测试是否通过；
8. 构造失败复现或相邻验证用例；
9. 生成完整测试报告；
10. 在用户确认后执行修复和回归测试。

---

## 3. 项目结构

```text
TestAgent/
├── main.go                         # TestAgent 主程序
├── systemprompt.md                 # 大测试语境系统提示词
├── go.mod
├── go.sum
├── .env.example                    # 环境变量配置模板
├── .gitignore
│
├── skill_general/                  # 通用 Skill 层
│   ├── loader.go                   # Skill 索引与动态加载逻辑
│   ├── tool_interface.skill.yaml   # 工具接口规范 Skill
│   ├── code_testing.skill.yaml     # 通用代码测试 Skill
│   ├── black_box_testing.skill.yaml
│   ├── white_box_testing.skill.yaml
│   ├── integration_testing.skill.yaml
│   ├── system_testing.skill.yaml
│   └── tools/                      # 工具实现
│       ├── manager.go              # 工具注册与管理
│       ├── readfile.go             # 文件读取工具
│       ├── writefile.go            # 文件写入工具
│       └── executetool.go          # 命令执行工具
│
├── skill_custom/                   # 专用 Skill 层
│   └── flex_md/
│       ├── flexmd_program.skill.yaml
│       ├── flexmd_agentic_testing.skill.yaml
│       └── flexmd_spec.md          # FlexMD 规格说明
│
├── flexmd/                         # 被测程序：FlexMD 编译器
│   ├── __init__.py
│   ├── __main__.py                 # CLI 入口
│   └── compiler.py                 # FlexMD parser / AST / renderer
│
├── samples/
│   └── flexmd_agent_demo.fmd       # FlexMD 示例输入
│
├── prompts/
│   └── run_flexmd_agent_prompt.md  # 启动 FlexMD 测试任务的参考提示
│
├── reports/                        # 测试产物目录，运行后生成
│   └── flexmd_agent_session/
│       ├── 01_test_plan.md
│       ├── final_report.md
│       ├── cases/
│       ├── outputs/
│       └── repro/
│
└── legacy/
    └── student_score_manager/      # 原示例项目，保留为历史演示
```

---

## 4. 环境准备

### 4.1 基础环境

请确保本机已安装：

- Go 1.23 或以上版本；
- Python 3.10 或以上版本；
- 可访问 OpenAI 兼容接口的大语言模型服务。

Windows 用户建议使用：

- Windows Terminal；
- PowerShell 7；
- VS Code 终端。

传统 `cmd.exe` 对 ANSI 颜色支持可能不稳定。

### 4.2 配置环境变量

复制配置模板：

```bash
cp .env.example .env
```

Windows PowerShell：

```powershell
Copy-Item .env.example .env
```

编辑 `.env`：

```env
OPENAI_API_KEY=你的API_KEY
OPENAI_BASE_URL=你的模型服务地址
OPENAI_MODEL=你的模型名称

# 命令执行 shell，auto 会根据系统自动选择
EXECUTE_COMMAND_SHELL=auto

# 终端显示设置
AGENT_COLOR_OUTPUT=true
AGENT_RENDER_MARKDOWN=true
PLAIN_TEXT_OUTPUT=true

# Skill 上下文模式
SKILL_CONTEXT_MODE=dynamic
```

说明：

- `SKILL_CONTEXT_MODE=dynamic` 表示启动时只加载 Skill 索引，完整 Skill 正文由智能体按需读取；
- `AGENT_COLOR_OUTPUT=true` 会启用彩色终端输出；
- `PLAIN_TEXT_OUTPUT=true` 会尽量清理不适合终端显示的 Markdown 标记；
- `.env` 包含密钥信息，不应提交到仓库。

---

## 5. 运行方式

### 5.1 启动 TestAgent

在项目根目录执行：

```bash
go run .
```

如果只运行：

```bash
go run main.go
```

也可以启动主程序，但在部分环境下可能不会编译同目录下的辅助 Go 文件。推荐使用 `go run .`。

启动后，程序会加载：

- `systemprompt.md`；
- `skill_general/` 下的通用 Skill 索引；
- `skill_custom/` 下的专用 Skill 索引；
- `read_file`、`write_file`、`execute_command` 工具。

如果检测到 `conversation_state.json`，程序会询问是否恢复上次会话。

### 5.2 查看可用 Skill

启动后可以输入：

```text
你有什么 skills？
```

智能体会根据已加载的 Skill 索引说明当前可用能力。

### 5.3 启动 FlexMD 完整测试任务

可以直接输入：

```text
请执行 flexmd_agentic_testing，完成 FlexMD 的完整测试流程。
```

也可以复制 `prompts/run_flexmd_agent_prompt.md` 中的提示词。

智能体将按以下流程执行：

1. 读取 `skill_custom/flex_md/flexmd_agentic_testing.skill.yaml`；
2. 读取 `skill_custom/flex_md/flexmd_program.skill.yaml`；
3. 读取 `skill_custom/flex_md/flexmd_spec.md`；
4. 读取黑盒、白盒、集成、系统测试 Skill；
5. 读取 FlexMD 源码和 CLI 入口；
6. 创建 `reports/flexmd_agent_session/` 目录；
7. 编写测试计划和 `.fmd` 测试用例；
8. 调用 FlexMD 编译器执行测试；
9. 读取输出并分析；
10. 生成最终测试报告。

---

## 6. FlexMD 被测程序说明

FlexMD 是一个 Markdown 演示文稿 DSL，核心命令为：

```bash
python -m flexmd compile input.fmd -o output.html --ast output.ast.json
```

### 6.1 支持的核心语法

#### 幻灯片分页

根层级单独一行 `---` 表示分页：

```fmd
# Slide 1
---
# Slide 2
```

连续的 `---` 合法，会生成空 slide。

#### 横向 flex 布局

```fmd
\h
左侧内容
---
右侧内容
\\h
```

`\h` 表示 horizontal layout，渲染为左右分栏。

#### 纵向 flex 布局

```fmd
\v
上方内容
---
下方内容
\\v
```

`\v` 表示 vertical layout，渲染为上下分栏。

#### 比例布局

```fmd
\h 1:2
较窄区域
---
较宽区域
\\h
```

比例会进入 AST 的 `ratios` 字段，并在 HTML 中体现为 `flex-grow` 样式。

#### 代码块隔离

在 fenced code block 中，`---`、`\h`、`\v` 等特殊标记都应视为普通文本，不触发分页或布局解析。

---

## 7. 测试流程设计

本项目针对 FlexMD 设计了四类测试。

### 7.1 黑盒测试

黑盒测试只基于 FlexMD 规格，不依赖源码内部结构。主要覆盖：

- 基础 Markdown 渲染；
- 根层级分页；
- 连续分页生成空 slide；
- 横向 flex；
- 纵向 flex；
- 空 pane；
- 比例语法；
- 嵌套 flex；
- 代码块隔离；
- 非法输入错误处理。

### 7.2 白盒测试

白盒测试基于 `flexmd/compiler.py` 的内部逻辑设计，主要覆盖：

- parser 状态机；
- `in_code_fence` 状态切换；
- flex 栈 push / pop；
- ratio 解析路径；
- Markdown 渲染函数；
- 错误处理分支；
- AST 到 HTML 的渲染路径。

### 7.3 集成测试

集成测试验证：

```text
FlexMD 源文本
    ↓
Parser
    ↓
AST JSON
    ↓
HTML Renderer
    ↓
HTML 幻灯片
```

重点检查 AST 与 HTML 之间的数据一致性，例如：

- AST 中 `direction=row` 时，HTML 中应有 `flex-direction: row`；
- AST 中 `ratios=[1,2]` 时，HTML 中应体现 `flex-grow: 1` 和 `flex-grow: 2`；
- 代码块中的特殊标记不应影响 AST 结构。

### 7.4 系统测试

系统测试从用户视角执行完整端到端流程：

```bash
python -m flexmd compile samples/flexmd_agent_demo.fmd -o reports/flexmd_agent_session/outputs/demo.html --ast reports/flexmd_agent_session/outputs/demo.ast.json
```

验证输出文件、页面结构、CSS 样式、幻灯片数量、flex 数量和可展示性。

---

## 8. 测试产物

一次完整测试后，典型产物如下：

```text
reports/flexmd_agent_session/
├── 01_test_plan.md          # 测试计划
├── final_report.md          # 最终测试报告
├── cases/                   # Agent 编写的 .fmd 测试用例
│   ├── bb_test_001.fmd
│   ├── wb_test_001.fmd
│   └── ...
├── outputs/                 # HTML 和 AST JSON 输出
│   ├── bb_test_001.html
│   ├── bb_test_001.ast.json
│   └── ...
└── repro/                   # 缺陷最小复现用例
```

测试报告会记录：

- 测试类型；
- 用例名称；
- 测试目的；
- 执行命令；
- 预期结果；
- 实际结果；
- 通过/失败结论；
- 失败分析；
- 修改建议；
- 回归测试结果。

---

## 9. 常用命令

### 9.1 运行 FlexMD 编译器

```bash
python -m flexmd compile samples/flexmd_agent_demo.fmd -o reports/demo.html --ast reports/demo.ast.json
```

### 9.2 启动智能体

```bash
go run .
```

### 9.3 保存当前对话状态

在 Agent 对话中输入：

```text
save
```

会保存当前对话状态到：

```text
conversation_state.json
```

### 9.4 查看当前状态

```text
state
```

### 9.5 查看工具列表

```text
tools
```

### 9.6 查看 Skill 列表

```text
skills
```

---

## 10. 设计亮点

### 10.1 动态 Skill 读取

系统启动时只加载 Skill 索引，而不是一次性将所有 Skill 正文放入上下文。这样可以减少初始上下文占用，也更接近真实的 Skill 调度机制。

当用户提出 FlexMD 测试任务时，智能体会主动读取：

- FlexMD 任务 Skill；
- FlexMD 程序 Skill；
- FlexMD 规格文档；
- 工具接口 Skill；
- 黑盒测试 Skill；
- 白盒测试 Skill；
- 集成测试 Skill；
- 系统测试 Skill。

### 10.2 工具驱动的真实测试

智能体不是只生成自然语言建议，而是通过工具真实完成：

- 读取规格和源码；
- 写入测试用例；
- 执行编译命令；
- 读取 AST 和 HTML；
- 判断测试结果；
- 保存测试报告。

### 10.3 人机协作的测试边界

用户通过自然语言指定任务目标和产物目录，智能体在目标范围内自主完成测试设计与执行。

如果发现缺陷，智能体默认先生成分析和修改建议，不直接修改源码。只有用户确认后，才进行代码修改和回归测试。

### 10.4 可解释的测试过程

终端会打印：

- LLM 输出；
- 工具调用名称；
- 工具参数；
- 工具返回摘要；
- 命令执行结果。

这使测试过程可以被复查、截图和写入实验报告。

---

## 11. 实验结果摘要

在一次完整测试中，TestAgent 对 FlexMD 进行了：

- 黑盒测试；
- 白盒测试；
- 集成测试；
- 系统测试。

测试覆盖了 FlexMD 的核心语法和错误处理路径，包括分页、空 slide、横向/纵向 flex、空 pane、比例语法、嵌套 flex、代码块隔离、非法比例、闭合标记不匹配等。

最终测试报告显示，FlexMD 核心功能通过测试，并发现一个轻微语义问题：

- 有序列表 `1. item` 被渲染为 `<ul>` 而不是 `<ol>`。

该问题不影响 FlexMD 核心分页和布局功能，但影响 HTML 语义正确性。后续通过补充修复版本和回归测试验证了修改方案的可行性。

---

## 12. 注意事项

1. 请勿提交 `.env` 文件。
2. 若在 Windows 下命令执行异常，请确认 `.env` 中 `EXECUTE_COMMAND_SHELL=auto`。
3. 若终端颜色不可见，建议使用 Windows Terminal 或 PowerShell 7。
4. 如果需要严格区分黑盒和白盒测试，建议分会话执行：
   - 黑盒阶段禁止读取 `flexmd/compiler.py`；
   - 白盒阶段允许读取源码；
   - 最后再进行综合报告汇总。
5. 测试过程中生成的 `reports/` 目录可以作为实验报告证据。

---

## 13. 小组分工建议

可按实际情况填写：

- 成员 A：TestAgent 框架设计、工具调用与 Skill 加载机制；
- 成员 B：FlexMD 被测程序设计与实现；
- 成员 C：测试流程设计、实验报告整理与结果分析。

---

## 14. 项目定位

本项目不是简单的 Markdown 工具，也不是简单的测试脚本，而是一个面向课程实验的 **Skill 驱动软件测试智能体原型**。

它展示了如何将：

- 大语言模型；
- Agent Harness；
- Skill 编排；
- 工具调用；
- 命令行被测程序；
- 自动化测试方法；
- 测试报告生成；

组合成一个可以实际运行、可以复现实验过程、可以产生真实测试产物的软件测试智能体系统。
