# FlexMD 测试智能体说明

本项目采用“万物皆 skill”的组织方式：

- `systemprompt.md`：规定软件测试智能体的大测试语境和安全边界。
- `skill_general/`：通用技能，包括工具接口规范和通用测试方法。
- `skill_custom/flex_md/`：FlexMD 专用技能，包括被测对象说明、语言规格和完整测试任务编排。
- `flexmd/`：被测程序，提供 `.fmd -> HTML + AST JSON` 的命令行编译器。
- `reports/flexmd_agent_session/`：Agent 运行后生成的测试计划、测试用例、输出、复现和报告。

## 推荐运行方式

1. 复制 `.env.example` 为 `.env`，填写模型配置。
2. 在项目根目录运行：

```bash
go run main.go
```

3. 如果存在 `conversation_state.json`，程序会询问是否读取上次会话状态。
4. 启动后可以直接输入自然语言任务，例如：

```text
请使用 FlexMD 相关 skill，对当前仓库中的 FlexMD 项目完成一次完整的软件测试任务。
```

也可以粘贴：

```text
prompts/run_flexmd_agent_prompt.md
```

## Skill 分层

### skill_general

- `tool_interface.skill.yaml`：说明 read_file、write_file、execute_command 的交互规范。
- `black_box_testing.skill.yaml`：黑盒测试方法。
- `white_box_testing.skill.yaml`：白盒测试方法。
- `integration_testing.skill.yaml`：集成测试方法。
- `system_testing.skill.yaml`：系统测试方法。
- `code_testing.skill.yaml`：通用代码测试流程。

### skill_custom/flex_md

- `flexmd_program.skill.yaml`：说明 FlexMD 的规格文件、源码位置、CLI、AST/HTML 输出和判定方式。
- `flexmd_spec.md`：FlexMD 语言规格。
- `flexmd_agentic_testing.skill.yaml`：编排 FlexMD 的黑盒、白盒、集成、系统测试流程。

## 终端显示设置

`.env` 中可配置：

```env
AGENT_COLOR_OUTPUT=true
AGENT_RENDER_MARKDOWN=true
PLAIN_TEXT_OUTPUT=true
EXECUTE_COMMAND_SHELL=auto
```

在 Windows 上建议使用 Windows Terminal、PowerShell 7 或 VS Code Terminal。程序会尝试启用 Windows ANSI 颜色支持。
