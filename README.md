# eino 对话应用

一个基于 eino 框架的简单对话应用示例。

## 特性

- 使用 eino 框架的 OpenAI 模型组件
- 支持命令行交互式对话
- 简单的退出机制

## 前提条件

1. 安装 Go 1.26 或更高版本
2. 获取 OpenAI API 密钥（或其他兼容 OpenAI API 的 LLM 服务）

## 安装与运行

1. 克隆或下载此项目
2. 在项目目录中运行：
   ```bash
   go mod tidy
   go run main.go
   ```

3. 按照提示输入您的 OpenAI API 密钥
4. 开始对话！输入 'exit' 或 'quit' 退出

## 配置说明

### API 密钥

您可以在运行时输入 API 密钥，或者直接修改 `main.go` 文件中的默认配置：

```go
apiKey = "your-llm-token-here" // 请替换为您的实际API密钥
```

### 模型配置

在 `main.go` 中可以修改以下配置：

```go
Model:   "gpt-3.5-turbo", // 可改为 gpt-4, gpt-4o 等
BaseURL: "https://api.openai.com/v1", // 可改为其他兼容 OpenAI API 的服务地址
```

## 项目结构

- `main.go` - 主程序文件，包含对话逻辑
- `go.mod` - Go 模块定义文件
- `README.md` - 本说明文件

## 依赖

- [github.com/cloudwego/eino](https://github.com/cloudwego/eino) - eino 主框架
- [github.com/cloudwego/eino-ext](https://github.com/cloudwego/eino-ext) - eino 扩展组件

## 使用示例

```bash
$ go run main.go
=== eino 对话应用 ===
请输入您的OpenAI API密钥（或按回车使用默认配置）：
your-actual-api-key-here
模型初始化成功！
输入 'exit' 或 'quit' 退出对话
============================================
你: 你好
助手: 你好！有什么我可以帮助你的吗？
---
你: 解释一下什么是eino框架
助手: eino 是 CloudWeGo 生态中的一个 Go 语言 Agent 开发框架...
---
你: exit
再见！
```

## 注意事项

1. 请妥善保管您的 API 密钥，不要提交到公开仓库
2. 此示例仅用于演示目的，实际生产环境需要更多错误处理和配置管理
3. eino 框架仍在快速发展中，API 可能会有变化

## 下一步

基于此示例，您可以：

1. 添加对话历史记忆功能
2. 集成更多类型的模型（如本地模型）
3. 添加文件上传和处理能力
4. 实现多轮对话和上下文管理

## 许可证

MIT