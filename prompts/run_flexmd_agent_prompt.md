请使用 FlexMD 相关 skill，对当前仓库中的 FlexMD 项目完成一次完整的软件测试任务。

被测对象：
- FlexMD 编译器源码：flexmd/compiler.py
- 命令行入口：flexmd/__main__.py
- 被测程序说明与规格：skill_custom/flex_md/flexmd_program.skill.yaml 和 skill_custom/flex_md/flexmd_spec.md

测试要求：
1. 使用 flexmd_agentic_testing 作为任务编排 skill。
2. 依次完成黑盒测试、白盒测试、集成测试和系统测试。
3. 每类测试都需要由你自主设计测试用例，并实际写入 .fmd 文件。
4. 每个用例都要送入 FlexMD 编译器执行，不能只描述不运行。
5. 读取 AST JSON、HTML、返回码和错误信息来判断通过/失败。
6. 测试产物保存到 reports/flexmd_agent_session/。
7. 如果发现失败，请构造最小复现并写出缺陷分析。
8. 暂时不要修改源码；完成测试和报告后向我汇报。
