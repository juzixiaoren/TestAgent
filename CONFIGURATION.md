# 配置管理说明

此应用现在支持使用.env文件管理配置，使配置更加灵活和安全。

## .env文件格式

.env文件应该包含以下配置项：

```env
# OpenAI API配置
OPENAI_API_KEY=your-llm-token-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-3.5-turbo
```

## 配置项说明

- `OPENAI_API_KEY`: OpenAI API密钥或其他兼容服务的API密钥
- `OPENAI_BASE_URL`: API服务地址，默认为OpenAI官方地址，可改为其他兼容服务地址
- `OPENAI_MODEL`: 使用的模型名称，默认为gpt-3.5-turbo，可改为gpt-4、gpt-4o等

## 使用方式

### 1. 使用.env文件配置

1. 在项目根目录创建.env文件
2. 在.env文件中设置您的配置
3. 运行程序时按回车使用.env中的配置

### 2. 运行时输入配置

1. 运行程序
2. 当提示输入API密钥时，输入您的API密钥
3. BaseURL和Model将使用.env中的配置

### 3. 修改.env文件示例

```env
# 使用其他兼容服务
OPENAI_API_KEY=your-api-key-here
OPENAI_BASE_URL=https://api.deepseek.com/v1
OPENAI_MODEL=deepseek-chat

# 或者使用本地模型服务
# OPENAI_BASE_URL=http://localhost:8080/v1
# OPENAI_MODEL=local-model
```

## 配置优先级

1. 运行时输入的API密钥（最高优先级）
2. .env文件中的配置
3. 程序硬编码的默认值（最低优先级）

## 安全注意事项

- .env文件包含敏感信息，请不要将其提交到版本控制系统
- 建议将.env文件添加到.gitignore中
- 为不同环境（开发、测试、生产）创建不同的.env文件

## 默认配置回退

如果.env文件不存在或配置项为空，程序将使用以下默认值：

- API密钥: "your-llm-token-here"（需要替换为实际值）
- BaseURL: "https://api.openai.com/v1"
- Model: "gpt-3.5-turbo"