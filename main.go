package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"eino-chat-demo/skills/skill"
	"eino-chat-demo/skills/tools"
)

// loadEnv 从.env文件加载环境变量
func loadEnv() map[string]string {
	env := make(map[string]string)

	data, err := os.ReadFile(".env")
	if err != nil {
		env["OPENAI_API_KEY"] = "your-llm-token-here"
		env["OPENAI_BASE_URL"] = "https://api.openai.com/v1"
		env["OPENAI_MODEL"] = "gpt-3.5-turbo"
		return env
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if env["OPENAI_API_KEY"] == "" {
		env["OPENAI_API_KEY"] = "your-llm-token-here"
	}
	if env["OPENAI_BASE_URL"] == "" {
		env["OPENAI_BASE_URL"] = "https://api.openai.com/v1"
	}
	if env["OPENAI_MODEL"] == "" {
		env["OPENAI_MODEL"] = "gpt-3.5-turbo"
	}

	return env
}

// loadSystemPrompt 从systemprompt.md文件加载系统提示词
func loadSystemPrompt() string {
	data, err := os.ReadFile("systemprompt.md")
	if err != nil {
		return "你是一个智能助手，可以使用工具来帮助用户解决问题。"
	}
	return string(data)
}

// ConversationState 对话状态，用于记忆
type ConversationState struct {
	Messages    []*schema.Message `json:"messages"`
	ToolsUsed   []string          `json:"tools_used"`
	LastUpdated time.Time         `json:"last_updated"`
}

// ReactAgent React模式的Agent
type ReactAgent struct {
	model       *openai.ChatModel
	toolManager *tools.ToolManager
	skills      []*skill.Skill
	state       *ConversationState
	toolInfos   []*schema.ToolInfo
	ctx         context.Context
}

// NewReactAgent 创建新的ReactAgent
func NewReactAgent(ctx context.Context, apiKey string) (*ReactAgent, error) {
	env := loadEnv()

	if apiKey == "" {
		apiKey = env["OPENAI_API_KEY"]
	}

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   env["OPENAI_MODEL"],
		BaseURL: env["OPENAI_BASE_URL"],
	})
	if err != nil {
		return nil, fmt.Errorf("初始化模型失败: %v", err)
	}

	toolManager := tools.InitializeTools()
	toolInfos := toolManager.GetToolInfos()

	// 使用 eino 原生 tool calling：绑定工具到模型
	if err := model.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %v", err)
	}

	// 加载技能
	skillsDir := "skills/skill"
	loadedSkills, err := skill.LoadSkills(skillsDir)
	if err != nil {
		fmt.Printf("警告：加载技能失败: %v\n", err)
		loadedSkills = []*skill.Skill{}
	}
	if len(loadedSkills) > 0 {
		fmt.Printf("已加载 %d 个技能: %s\n", len(loadedSkills), strings.Join(skill.GetSkillNames(loadedSkills), ", "))
	}

	// 构建系统提示词：基础提示词 + 技能信息
	systemPrompt := loadSystemPrompt() + skill.FormatSkillsPrompt(loadedSkills)

	state := &ConversationState{
		Messages: []*schema.Message{
			schema.SystemMessage(systemPrompt),
		},
		ToolsUsed:   []string{},
		LastUpdated: time.Now(),
	}

	return &ReactAgent{
		model:       model,
		toolManager: toolManager,
		skills:      loadedSkills,
		state:       state,
		toolInfos:   toolInfos,
		ctx:         ctx,
	}, nil
}

// GetSkills 获取已加载的技能列表
func (agent *ReactAgent) GetSkills() []*skill.Skill {
	return agent.skills
}

// ProcessMessage 处理用户消息，支持多轮工具调用的 React 循环
func (agent *ReactAgent) ProcessMessage(userInput string) (string, error) {
	agent.state.Messages = append(agent.state.Messages, schema.UserMessage(userInput))

	// React 循环：LLM 回复 → 检查工具调用 → 执行工具 → 再次请求 LLM → 直到无工具调用
	for {
		response, err := agent.model.Generate(agent.ctx, agent.state.Messages)
		if err != nil {
			return "", fmt.Errorf("生成回复失败: %v", err)
		}

		// 将助手回复加入对话历史（包含 ToolCalls 信息）
		agent.state.Messages = append(agent.state.Messages, response)
		agent.state.LastUpdated = time.Now()

		// 检查是否有工具调用
		if len(response.ToolCalls) == 0 {
			return response.Content, nil
		}

		// 执行所有工具调用，将结果加入对话历史
		for _, toolCall := range response.ToolCalls {
			toolName := toolCall.Function.Name
			toolArgs, err := parseToolCallArgs(toolCall.Function.Arguments)
			if err != nil {
				agent.state.Messages = append(agent.state.Messages,
					schema.ToolMessage(fmt.Sprintf("参数解析失败: %v", err), toolCall.ID,
						schema.WithToolName(toolName)))
				continue
			}

			result, err := agent.executeTool(toolName, toolArgs)
			if err != nil {
				agent.state.Messages = append(agent.state.Messages,
					schema.ToolMessage(fmt.Sprintf("工具执行失败: %v", err), toolCall.ID,
						schema.WithToolName(toolName)))
				continue
			}

			resultJSON, _ := json.Marshal(result)
			agent.state.Messages = append(agent.state.Messages,
				schema.ToolMessage(string(resultJSON), toolCall.ID,
					schema.WithToolName(toolName)))
		}
	}
}

// parseToolCallArgs 解析工具调用参数
func parseToolCallArgs(argsStr string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return args, nil
}

// executeTool 执行工具并记录
func (agent *ReactAgent) executeTool(toolName string, args map[string]interface{}) (interface{}, error) {
	tool, exists := agent.toolManager.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("工具不存在: %s", toolName)
	}

	result, err := tool.Execute(agent.ctx, args)
	if err != nil {
		return nil, err
	}

	agent.state.ToolsUsed = append(agent.state.ToolsUsed, toolName)
	agent.state.LastUpdated = time.Now()
	return result, nil
}

// GetState 获取当前对话状态
func (agent *ReactAgent) GetState() *ConversationState {
	return agent.state
}

// SaveState 保存对话状态到文件
func (agent *ReactAgent) SaveState(filePath string) error {
	data, err := json.MarshalIndent(agent.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// LoadState 从文件加载对话状态
func (agent *ReactAgent) LoadState(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var state ConversationState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	agent.state = &state
	return nil
}

func main() {
	fmt.Println("=== eino React 模式对话应用 ===")

	var apiKey string

	if len(os.Args) > 1 {
		apiKey = os.Args[1]
	} else {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("请输入您的OpenAI API密钥（或按回车使用.env文件中的配置）：")
			reader := bufio.NewReader(os.Stdin)
			apiKeyInput, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKeyInput)
		}
	}

	if apiKey == "" {
		env := loadEnv()
		apiKey = env["OPENAI_API_KEY"]
		fmt.Printf("使用.env文件中的配置：模型=%s，API地址=%s\n", env["OPENAI_MODEL"], env["OPENAI_BASE_URL"])
	}

	ctx := context.Background()
	agent, err := NewReactAgent(ctx, apiKey)
	if err != nil {
		fmt.Printf("初始化Agent失败: %v\n", err)
		return
	}

	fmt.Println("ReactAgent初始化成功！")
	fmt.Println("输入 'exit' 或 'quit' 退出对话")
	fmt.Println("输入 'state' 查看当前对话状态")
	fmt.Println("输入 'tools' 查看可用工具")
	fmt.Println("输入 'skills' 查看已加载技能")
	fmt.Println("输入 'save' 保存对话状态")
	fmt.Println("============================================")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("你: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		if userInput == "" {
			continue
		}

		switch strings.ToLower(userInput) {
		case "exit", "quit":
			fmt.Println("再见！")
			return
		case "state":
			state := agent.GetState()
			stateJSON, _ := json.MarshalIndent(state, "", "  ")
			fmt.Printf("当前对话状态:\n%s\n", stateJSON)
		case "tools":
			toolInfos := agent.toolManager.GetToolInfos()
			fmt.Println("可用工具:")
			for _, info := range toolInfos {
				fmt.Printf("  - %s: %s\n", info.Name, info.Desc)
			}
		case "skills":
			loadedSkills := agent.GetSkills()
			if len(loadedSkills) == 0 {
				fmt.Println("未加载任何技能")
			} else {
				fmt.Printf("已加载 %d 个技能:\n", len(loadedSkills))
				for _, s := range loadedSkills {
					fmt.Printf("  - %s (v%s): %s\n", s.Name, s.Version, s.Description)
					if len(s.Capabilities) > 0 {
						fmt.Printf("    能力: %s\n", strings.Join(s.Capabilities, ", "))
					}
					if len(s.Dependencies) > 0 {
						fmt.Printf("    依赖: %s\n", strings.Join(s.Dependencies, ", "))
					}
				}
			}
		case "save":
			if err := agent.SaveState("conversation_state.json"); err != nil {
				fmt.Printf("保存状态失败: %v\n", err)
			} else {
				fmt.Println("对话状态已保存到 conversation_state.json")
			}
		default:
			reply, err := agent.ProcessMessage(userInput)
			if err != nil {
				fmt.Printf("处理消息失败: %v\n", err)
				continue
			}
			fmt.Printf("助手: %s\n", reply)
			fmt.Println("---")
		}
	}
}
