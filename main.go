package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	skill "eino-chat-demo/skill_general"
	"eino-chat-demo/skill_general/tools"
)

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiItalic  = "\033[3m"
	ansiDim     = "\033[2m"
	ansiGray    = "\033[90m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
)

// loadEnv 从 .env 文件和系统环境变量加载配置。
// 优先级：系统环境变量 > .env > 默认值。
func loadEnv() map[string]string {
	env := make(map[string]string)

	if data, err := os.ReadFile(".env"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				env[strings.TrimSpace(parts[0])] = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}

	keys := []string{
		"OPENAI_API_KEY", "OPENAI_BASE_URL", "OPENAI_MODEL",
		"EXECUTE_COMMAND_SHELL", "PLAIN_TEXT_OUTPUT", "AGENT_COLOR_OUTPUT", "AGENT_RENDER_MARKDOWN", "SKILL_CONTEXT_MODE",
	}
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			env[key] = val
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
	if env["EXECUTE_COMMAND_SHELL"] == "" {
		env["EXECUTE_COMMAND_SHELL"] = "auto"
	}
	if env["SKILL_CONTEXT_MODE"] == "" {
		env["SKILL_CONTEXT_MODE"] = "dynamic"
	}
	return env
}

func envBool(env map[string]string, key string, defaultValue bool) bool {
	val := strings.TrimSpace(strings.ToLower(env[key]))
	if val == "" {
		return defaultValue
	}
	switch val {
	case "1", "true", "yes", "y", "on", "enable", "enabled":
		return true
	case "0", "false", "no", "n", "off", "disable", "disabled":
		return false
	default:
		return defaultValue
	}
}

// loadSystemPrompt 从 systemprompt.md 文件加载系统提示词。
func loadSystemPrompt() string {
	data, err := os.ReadFile("systemprompt.md")
	if err != nil {
		return "你是一个软件测试智能体，可以使用工具读取文件、写入文件和执行命令。"
	}
	return string(data)
}

// ConversationState 对话状态，用于会话内上下文和可选持久化。
type ConversationState struct {
	Messages    []*schema.Message `json:"messages"`
	ToolsUsed   []string          `json:"tools_used"`
	LastUpdated time.Time         `json:"last_updated"`
}

// ReactAgent ReAct 模式 Agent。
type ReactAgent struct {
	model          *openai.ChatModel
	toolManager    *tools.ToolManager
	skills         []*skill.Skill
	state          *ConversationState
	toolInfos      []*schema.ToolInfo
	ctx            context.Context
	systemPrompt   string
	colorOutput    bool
	plainText      bool
	renderMarkdown bool
}

// NewReactAgent 创建新的 ReactAgent。
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

	if err := model.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %v", err)
	}

	loadedSkills, err := skill.LoadSkillsFromDirs([]string{"skill_general", "skill_custom"})
	if err != nil {
		fmt.Printf("警告：加载 skill 目录失败: %v\n", err)
		loadedSkills = []*skill.Skill{}
	}
	if len(loadedSkills) > 0 {
		fmt.Printf("已加载 %d 个技能: %s\n", len(loadedSkills), strings.Join(skill.GetSkillNames(loadedSkills), ", "))
	}

	skillContextMode := strings.TrimSpace(strings.ToLower(env["SKILL_CONTEXT_MODE"]))
	var skillsPrompt string
	switch skillContextMode {
	case "full", "eager", "all":
		skillsPrompt = skill.FormatSkillsFullPrompt(loadedSkills)
		fmt.Println("技能上下文模式：full（启动时注入完整 skill 正文）")
	default:
		skillContextMode = "dynamic"
		skillsPrompt = skill.FormatSkillIndexPrompt(loadedSkills)
		fmt.Println("技能上下文模式：dynamic（启动时只注入 skill 索引，按需 read_file 读取全文）")
	}

	systemPrompt := loadSystemPrompt() + skillsPrompt

	colorOutput := envBool(env, "AGENT_COLOR_OUTPUT", true)
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		colorOutput = false
	}

	state := &ConversationState{
		Messages: []*schema.Message{
			schema.SystemMessage(systemPrompt),
		},
		ToolsUsed:   []string{},
		LastUpdated: time.Now(),
	}

	return &ReactAgent{
		model:          model,
		toolManager:    toolManager,
		skills:         loadedSkills,
		state:          state,
		toolInfos:      toolInfos,
		ctx:            ctx,
		systemPrompt:   systemPrompt,
		colorOutput:    colorOutput,
		plainText:      envBool(env, "PLAIN_TEXT_OUTPUT", true),
		renderMarkdown: envBool(env, "AGENT_RENDER_MARKDOWN", true),
	}, nil
}

// GetSkills 获取已加载的技能列表。
func (agent *ReactAgent) GetSkills() []*skill.Skill {
	return agent.skills
}

func truncateText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("...<已截断，原长度 %d 字符>", len(s))
}

func prettyJSON(v interface{}, max int) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncateText(fmt.Sprintf("%v", v), max)
	}
	return truncateText(string(data), max)
}

func summarizeToolResult(result interface{}) string {
	m, ok := result.(map[string]interface{})
	if !ok {
		return prettyJSON(result, 3000)
	}
	view := map[string]interface{}{}
	for k, v := range m {
		switch k {
		case "content", "output":
			if s, ok := v.(string); ok {
				view[k] = truncateText(s, 1200)
			} else {
				view[k] = v
			}
		default:
			view[k] = v
		}
	}
	return prettyJSON(view, 4000)
}

func (agent *ReactAgent) style(text string, codes ...string) string {
	if !agent.colorOutput || len(codes) == 0 {
		return text
	}
	return strings.Join(codes, "") + text + ansiReset
}

func (agent *ReactAgent) stripMarkdownLite(text string) string {
	replacements := []struct {
		re   *regexp.Regexp
		repl string
	}{
		{regexp.MustCompile("(?m)^#{1,6}\\s+"), ""},
		{regexp.MustCompile("`([^`\\n]+)`"), "$1"},
		{regexp.MustCompile("\\*\\*\\*([^*]+)\\*\\*\\*"), "$1"},
		{regexp.MustCompile("\\*\\*([^*]+)\\*\\*"), "$1"},
		{regexp.MustCompile("__([^_]+)__"), "$1"},
		{regexp.MustCompile("\\*([^*\\n]+)\\*"), "$1"},
		{regexp.MustCompile("_([^_\\n]+)_"), "$1"},
		{regexp.MustCompile("(?m)^>\\s?"), ""},
	}
	for _, r := range replacements {
		text = r.re.ReplaceAllString(text, r.repl)
	}
	return text
}

func (agent *ReactAgent) renderMarkdownANSI(text string) string {
	if !agent.colorOutput || !agent.renderMarkdown {
		if agent.plainText {
			return agent.stripMarkdownLite(text)
		}
		return text
	}

	// 先渲染行级结构，再渲染行内结构。
	lines := strings.SplitAfter(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lines[i] = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`).ReplaceAllStringFunc(line, func(match string) string {
				parts := regexp.MustCompile(`^(#{1,6})\s+(.+)$`).FindStringSubmatch(strings.TrimRight(match, "\r\n"))
				if len(parts) != 3 {
					return match
				}
				newline := ""
				if strings.HasSuffix(match, "\n") {
					newline = "\n"
				}
				return agent.style(parts[2], ansiBold, ansiYellow) + newline
			})
		}
		if strings.HasPrefix(trimmed, ">") {
			lines[i] = regexp.MustCompile(`(?m)^>\s?`).ReplaceAllString(line, agent.style("│ ", ansiGray))
		}
	}
	text = strings.Join(lines, "")

	rules := []struct {
		re    *regexp.Regexp
		style []string
	}{
		{regexp.MustCompile("`([^`\\n]+)`"), []string{ansiCyan}},
		{regexp.MustCompile("\\*\\*\\*([^*]+)\\*\\*\\*"), []string{ansiBold, ansiItalic}},
		{regexp.MustCompile("\\*\\*([^*]+)\\*\\*"), []string{ansiBold}},
		{regexp.MustCompile("__([^_]+)__"), []string{ansiBold}},
		{regexp.MustCompile("\\*([^*\\n]+)\\*"), []string{ansiItalic}},
		{regexp.MustCompile("_([^_\\n]+)_"), []string{ansiItalic}},
	}
	for _, rule := range rules {
		re := rule.re
		codes := rule.style
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			sub := re.FindStringSubmatch(match)
			if len(sub) < 2 {
				return match
			}
			return agent.style(sub[1], codes...)
		})
	}
	return text
}

func (agent *ReactAgent) printLLMStart() {
	fmt.Println(agent.style("\n[LLM] 开始生成...", ansiDim, ansiGray))
}

func (agent *ReactAgent) printToolCall(toolName string, toolArgs map[string]interface{}) {
	line := "──────────────── 工具调用 ────────────────"
	fmt.Println(agent.style("\n"+line, ansiBlue, ansiBold))
	fmt.Println(agent.style("[Agent -> Tool] 调用 ", ansiBlue) + agent.style(toolName, ansiBold, ansiCyan))
	fmt.Println(agent.style("参数：", ansiBlue))
	fmt.Println(agent.style(prettyJSON(toolArgs, 2000), ansiGray))
}

func (agent *ReactAgent) printToolResult(toolName string, result interface{}) {
	status := "执行完成"
	color := ansiGreen
	if m, ok := result.(map[string]interface{}); ok {
		if success, exists := m["success"].(bool); exists && !success {
			status = "执行返回失败状态"
			color = ansiYellow
		}
	}
	fmt.Println(agent.style("[Tool -> Agent] ", color) + agent.style(toolName+" "+status, ansiBold, color))
	fmt.Println(agent.style("结果摘要：", color))
	fmt.Println(agent.style(summarizeToolResult(result), ansiGray))
	fmt.Println(agent.style("────────────────────────────────────────", ansiBlue, ansiBold))
}

func (agent *ReactAgent) printToolError(toolName string, err error) {
	fmt.Println(agent.style("[Tool -> Agent] ", ansiRed) + agent.style(toolName+" 执行失败", ansiBold, ansiRed) + "：" + err.Error())
	fmt.Println(agent.style("────────────────────────────────────────", ansiBlue, ansiBold))
}

// ProcessMessage 处理用户消息，支持多轮工具调用的 ReAct 循环（非流式）。
func (agent *ReactAgent) ProcessMessage(userInput string) (string, error) {
	agent.state.Messages = append(agent.state.Messages, schema.UserMessage(userInput))

	for {
		response, err := agent.model.Generate(agent.ctx, agent.state.Messages)
		if err != nil {
			return "", fmt.Errorf("生成回复失败: %v", err)
		}

		agent.state.Messages = append(agent.state.Messages, response)
		agent.state.LastUpdated = time.Now()

		if len(response.ToolCalls) == 0 {
			return response.Content, nil
		}

		if err := agent.runToolCalls(response.ToolCalls); err != nil {
			return "", err
		}
	}
}

// ProcessMessageStream 处理用户消息，并把模型正文按行流式输出到终端。
func (agent *ReactAgent) ProcessMessageStream(userInput string) (string, error) {
	agent.state.Messages = append(agent.state.Messages, schema.UserMessage(userInput))

	for {
		agent.printLLMStart()
		streamResult, err := agent.model.Stream(agent.ctx, agent.state.Messages)
		if err != nil {
			return "", fmt.Errorf("流式生成回复失败: %v", err)
		}

		chunks := []*schema.Message{}
		var lineBuf strings.Builder
		for {
			chunk, err := streamResult.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				streamResult.Close()
				return "", fmt.Errorf("读取流式回复失败: %v", err)
			}
			if chunk == nil {
				continue
			}
			chunks = append(chunks, chunk)
			if chunk.Content != "" {
				lineBuf.WriteString(chunk.Content)
				for {
					current := lineBuf.String()
					idx := strings.IndexByte(current, '\n')
					if idx < 0 {
						break
					}
					line := current[:idx+1]
					fmt.Print(agent.renderMarkdownANSI(line))
					lineBuf.Reset()
					lineBuf.WriteString(current[idx+1:])
				}
			}
		}
		streamResult.Close()
		if lineBuf.Len() > 0 {
			fmt.Print(agent.renderMarkdownANSI(lineBuf.String()))
		}
		fmt.Println()

		var response *schema.Message
		if len(chunks) == 0 {
			response = schema.AssistantMessage("", nil)
		} else {
			response, err = schema.ConcatMessages(chunks)
			if err != nil {
				return "", fmt.Errorf("拼接流式回复失败: %v", err)
			}
		}

		agent.state.Messages = append(agent.state.Messages, response)
		agent.state.LastUpdated = time.Now()

		if len(response.ToolCalls) == 0 {
			return response.Content, nil
		}

		if err := agent.runToolCalls(response.ToolCalls); err != nil {
			return "", err
		}
	}
}

func (agent *ReactAgent) runToolCalls(toolCalls []schema.ToolCall) error {
	for _, toolCall := range toolCalls {
		toolName := toolCall.Function.Name
		toolArgs, err := parseToolCallArgs(toolCall.Function.Arguments)
		if err != nil {
			agent.printToolError(toolName, fmt.Errorf("参数解析失败: %v", err))
			agent.state.Messages = append(agent.state.Messages,
				schema.ToolMessage(fmt.Sprintf("参数解析失败: %v", err), toolCall.ID,
					schema.WithToolName(toolName)))
			continue
		}

		agent.printToolCall(toolName, toolArgs)
		result, err := agent.executeTool(toolName, toolArgs)
		if err != nil {
			agent.printToolError(toolName, err)
			agent.state.Messages = append(agent.state.Messages,
				schema.ToolMessage(fmt.Sprintf("工具执行失败: %v", err), toolCall.ID,
					schema.WithToolName(toolName)))
			continue
		}

		agent.printToolResult(toolName, result)
		resultJSON, _ := json.Marshal(result)
		agent.state.Messages = append(agent.state.Messages,
			schema.ToolMessage(string(resultJSON), toolCall.ID,
				schema.WithToolName(toolName)))
	}
	return nil
}

// parseToolCallArgs 解析工具调用参数。
func parseToolCallArgs(argsStr string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return args, nil
}

// executeTool 执行工具并记录。
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

// GetState 获取当前对话状态。
func (agent *ReactAgent) GetState() *ConversationState {
	return agent.state
}

// SaveState 保存对话状态到文件。
func (agent *ReactAgent) SaveState(filePath string) error {
	data, err := json.MarshalIndent(agent.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// LoadState 从文件加载对话状态，并用当前系统提示词替换旧系统提示词。
func (agent *ReactAgent) LoadState(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var state ConversationState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	if len(state.Messages) == 0 {
		state.Messages = []*schema.Message{schema.SystemMessage(agent.systemPrompt)}
	} else {
		state.Messages[0] = schema.SystemMessage(agent.systemPrompt)
	}
	if state.ToolsUsed == nil {
		state.ToolsUsed = []string{}
	}
	state.LastUpdated = time.Now()
	agent.state = &state
	return nil
}

func maybeLoadConversationState(agent *ReactAgent, reader *bufio.Reader) {
	const statePath = "conversation_state.json"
	if _, err := os.Stat(statePath); err != nil {
		return
	}
	fmt.Printf("发现 %s，是否读取上次会话状态？(y/N): ", statePath)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		if err := agent.LoadState(statePath); err != nil {
			fmt.Printf("读取状态失败: %v\n", err)
			return
		}
		fmt.Printf("已读取 %s，并使用当前 system prompt 与 skill 重新初始化系统消息。\n", statePath)
	}
}

func main() {
	fmt.Println("=== TestAgent 软件测试智能体 ===")

	var apiKey string

	if len(os.Args) > 1 {
		apiKey = os.Args[1]
	} else {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("请输入您的 OpenAI API 密钥（或按回车使用 .env 文件中的配置）：")
			reader := bufio.NewReader(os.Stdin)
			apiKeyInput, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKeyInput)
		}
	}

	if apiKey == "" {
		env := loadEnv()
		apiKey = env["OPENAI_API_KEY"]
		fmt.Printf("使用 .env 文件中的配置：模型=%s，API地址=%s\n", env["OPENAI_MODEL"], env["OPENAI_BASE_URL"])
	}

	ctx := context.Background()
	agent, err := NewReactAgent(ctx, apiKey)
	if err != nil {
		fmt.Printf("初始化 Agent 失败: %v\n", err)
		return
	}

	fmt.Println("Agent 初始化成功。")
	fmt.Println("输入 'exit' 或 'quit' 退出对话")
	fmt.Println("输入 'state' 查看当前对话状态")
	fmt.Println("输入 'tools' 查看可用工具")
	fmt.Println("输入 'skills' 查看已加载技能")
	fmt.Println("输入 'save' 保存对话状态到 conversation_state.json")
	fmt.Println("输入 'load' 读取 conversation_state.json")
	fmt.Println("============================================")

	reader := bufio.NewReader(os.Stdin)
	maybeLoadConversationState(agent, reader)

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
					if s.FilePath != "" {
						fmt.Printf("    来源: %s\n", s.FilePath)
					}
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
		case "load":
			if err := agent.LoadState("conversation_state.json"); err != nil {
				fmt.Printf("读取状态失败: %v\n", err)
			} else {
				fmt.Println("已读取 conversation_state.json")
			}
		default:
			fmt.Println("============================================")
			_, err := agent.ProcessMessageStream(userInput)
			if err != nil {
				fmt.Printf("处理消息失败: %v\n", err)
				continue
			}
			fmt.Println("============================================")
		}
	}
}
