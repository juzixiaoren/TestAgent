package tools

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// ToolFunc 工具函数类型
type ToolFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ToolWrapper 工具包装器
type ToolWrapper struct {
	Name        string
	Description string
	Func        ToolFunc
}

// Execute 执行工具
func (tw *ToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return tw.Func(ctx, args)
}

// ToolManager 工具管理器
type ToolManager struct {
	tools map[string]*ToolWrapper
}

// NewToolManager 创建新的工具管理器
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]*ToolWrapper),
	}
}

// RegisterTool 注册工具
func (tm *ToolManager) RegisterTool(name string, description string, toolFunc ToolFunc) {
	tm.tools[name] = &ToolWrapper{
		Name:        name,
		Description: description,
		Func:        toolFunc,
	}
}

// GetTool 获取工具
func (tm *ToolManager) GetTool(name string) (*ToolWrapper, bool) {
	tool, exists := tm.tools[name]
	return tool, exists
}

// GetAllTools 获取所有工具
func (tm *ToolManager) GetAllTools() map[string]*ToolWrapper {
	return tm.tools
}

// GetToolInfos 获取所有工具信息
func (tm *ToolManager) GetToolInfos() []*schema.ToolInfo {
	infos := make([]*schema.ToolInfo, 0, len(tm.tools))
	for _, tool := range tm.tools {
		// 为不同工具创建不同的参数定义
		var paramsOneOf *schema.ParamsOneOf

		switch tool.Name {
		case "read_file":
			paramsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"file_path": {
					Type: "string",
					Desc: "要读取的文件路径（绝对或相对路径）",
				},
			})
		case "write_file":
			paramsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"file_path": {
					Type: "string",
					Desc: "要写入的文件路径（绝对或相对路径）",
				},
				"content": {
					Type: "string",
					Desc: "要写入的文件内容",
				},
			})
		case "execute_command":
			paramsOneOf = schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"command": {
					Type: "string",
					Desc: "要执行的bash命令",
				},
				"working_dir": {
					Type: "string",
					Desc: "命令执行的工作目录（可选，默认当前目录）",
				},
				"timeout_seconds": {
					Type: "number",
					Desc: "命令执行超时时间，单位秒（可选，默认30秒）",
				},
			})
		}

		info := &schema.ToolInfo{
			Name:        tool.Name,
			Desc:        tool.Description,
			ParamsOneOf: paramsOneOf,
		}
		infos = append(infos, info)
	}
	return infos
}

// InitializeTools 初始化所有工具
func InitializeTools() *ToolManager {
	tm := NewToolManager()

	// 注册文件读取工具
	tm.RegisterTool("read_file", "读取文件内容", ReadFileTool)

	// 注册文件写入工具
	tm.RegisterTool("write_file", "写入内容到文件", WriteFileTool)

	// 注册执行命令工具
	tm.RegisterTool("execute_command", "执行系统命令", ExecuteCommandTool)

	return tm
}
