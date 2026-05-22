package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill 技能定义
type Skill struct {
	Name             string                     `yaml:"name"`
	Description      string                     `yaml:"description"`
	Version          string                     `yaml:"version"`
	Author           SkillAuthor                `yaml:"author"`
	Tags             []string                   `yaml:"tags"`
	Dependencies     []string                   `yaml:"dependencies"`
	Capabilities     []string                   `yaml:"capabilities"`
	Examples         []SkillExample             `yaml:"examples"`
	ToolInstructions map[string]ToolInstruction `yaml:"tool_instructions"`
	PromptTemplates  map[string]string          `yaml:"prompt_templates"`
	SecurityNotes    []string                   `yaml:"security_notes"`
	Config           SkillConfig                `yaml:"config"`
	Workflow         *SkillWorkflow             `yaml:"workflow"`

	// 运行时字段（不从 YAML 加载）
	FilePath string `yaml:"-"`
}

// SkillAuthor 技能作者信息
type SkillAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// SkillExample 技能使用示例
type SkillExample struct {
	Description string `yaml:"description"`
	Input       string `yaml:"input"`
	Output      string `yaml:"output"`
}

// ToolInstruction 工具使用说明
type ToolInstruction struct {
	Description string            `yaml:"description"`
	Parameters  map[string]string `yaml:"parameters"`
}

// SkillConfig 技能配置
type SkillConfig struct {
	AllowedDirectories []string `yaml:"allowed_directories"`
	MaxFileSize        int      `yaml:"max_file_size"`
	AllowedFileTypes   []string `yaml:"allowed_file_types"`
}

// SkillWorkflowStep 技能工作流步骤
type SkillWorkflowStep struct {
	Step        int      `yaml:"step"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	UseSkill    string   `yaml:"use_skill"`
	Actions     []string `yaml:"actions"`
}

// SkillWorkflow 技能工作流
type SkillWorkflow struct {
	Description string              `yaml:"description"`
	Steps       []SkillWorkflowStep `yaml:"steps"`
}

// LoadSkills 从指定目录递归加载所有 .skill.yaml / .skill.yml 文件。
func LoadSkills(dir string) ([]*Skill, error) {
	return LoadSkillsFromDirs([]string{dir})
}

// LoadSkillsFromDirs 从多个目录递归加载技能。目录不存在时跳过，至少加载失败时才返回错误。
func LoadSkillsFromDirs(dirs []string) ([]*Skill, error) {
	var files []string
	var missing []string

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			missing = append(missing, dir)
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if strings.HasSuffix(name, ".skill.yaml") || strings.HasSuffix(name, ".skill.yml") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("遍历技能目录 %s 失败: %v", dir, err)
		}
	}

	if len(files) == 0 {
		if len(missing) > 0 {
			return nil, fmt.Errorf("没有找到技能文件，缺失目录: %s", strings.Join(missing, ", "))
		}
		return []*Skill{}, nil
	}

	sort.Strings(files)

	var skills []*Skill
	seen := map[string]string{}
	for _, filePath := range files {
		s, err := loadSkillFile(filePath)
		if err != nil {
			fmt.Printf("警告：加载技能文件 %s 失败: %v\n", filePath, err)
			continue
		}
		if oldPath, ok := seen[s.Name]; ok {
			fmt.Printf("警告：技能 %s 重复，已加载 %s，跳过 %s\n", s.Name, oldPath, filePath)
			continue
		}
		s.FilePath = filePath
		seen[s.Name] = filePath
		skills = append(skills, s)
	}

	return skills, nil
}

// loadSkillFile 加载单个技能文件
func loadSkillFile(filePath string) (*Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var s Skill
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %v", err)
	}

	if s.Name == "" {
		return nil, fmt.Errorf("技能名称不能为空")
	}

	return &s, nil
}

// FormatSkillsPrompt 将技能列表格式化为系统提示词内容。
// 默认采用动态技能模式：这里只注入技能索引，不注入完整技能正文。
// Agent 需要执行某个任务时，应使用 read_file 读取相关 skill 文件全文，再按该 skill 的 workflow / dependencies / use_skill 继续加载依赖 skill。
func FormatSkillsPrompt(skills []*Skill) string {
	return FormatSkillIndexPrompt(skills)
}

// FormatSkillIndexPrompt 将技能列表格式化为“技能索引”。
// 这个索引应尽量短，只帮助模型知道有哪些技能、何时使用、完整文件在哪里。
func FormatSkillIndexPrompt(skills []*Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## 可用技能索引（动态加载）\n")
	sb.WriteString("系统启动时只注入技能索引，不注入完整技能正文。你应根据用户任务选择相关技能，并使用 read_file 读取技能文件全文。\n")
	sb.WriteString("执行任务前的推荐流程：\n")
	sb.WriteString("1. 根据用户目标从下方索引中选择最相关的任务 skill 或对象 skill。\n")
	sb.WriteString("2. 使用 read_file 读取该 skill 的 来源 文件，理解完整 workflow、dependencies、tool_instructions 和成功标准。\n")
	sb.WriteString("3. 如果该 skill 声明 dependencies 或 workflow.steps[].use_skill，继续从索引中找到对应 skill 并读取全文。\n")
	sb.WriteString("4. 读取完必要 skill 后，再开始编写测试计划、测试用例、执行命令和生成报告。\n")
	sb.WriteString("5. 不要声称已遵循某个 skill，除非你已经读取过它的完整文件或该 skill 内容已在当前上下文中。\n")

	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("\n### %s\n", s.Name))
		if s.FilePath != "" {
			sb.WriteString(fmt.Sprintf("- 来源：%s\n", filepath.ToSlash(s.FilePath)))
		}
		if s.Version != "" {
			sb.WriteString(fmt.Sprintf("- 版本：%s\n", s.Version))
		}
		sb.WriteString(fmt.Sprintf("- 描述：%s\n", s.Description))
		if len(s.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("- 标签：%s\n", strings.Join(s.Tags, ", ")))
		}
		if len(s.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("- 依赖：%s\n", strings.Join(s.Dependencies, ", ")))
		}
		if s.Workflow != nil && s.Workflow.Description != "" {
			sb.WriteString(fmt.Sprintf("- 工作流概览：%s\n", s.Workflow.Description))
		}
		if len(s.Capabilities) > 0 {
			maxCaps := len(s.Capabilities)
			if maxCaps > 3 {
				maxCaps = 3
			}
			sb.WriteString("- 能力摘要：\n")
			for i := 0; i < maxCaps; i++ {
				sb.WriteString(fmt.Sprintf("  - %s\n", s.Capabilities[i]))
			}
			if len(s.Capabilities) > maxCaps {
				sb.WriteString(fmt.Sprintf("  - ... 其余 %d 项请读取完整 skill 文件\n", len(s.Capabilities)-maxCaps))
			}
		}
	}

	return sb.String()
}

// FormatSkillsFullPrompt 将技能列表完整格式化为系统提示词内容。
// 仅在调试或显式需要 eager loading 时使用。
func FormatSkillsFullPrompt(skills []*Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## 可用技能（完整加载模式）\n")
	sb.WriteString("以下技能构成本次 Agent 的任务语境和执行规范。通用技能提供测试方法和工具接口，专用技能提供被测对象说明和任务编排。请根据用户意图选择并组合相关技能。\n")

	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("\n### %s (v%s)\n", s.Name, s.Version))
		if s.FilePath != "" {
			sb.WriteString(fmt.Sprintf("- 来源：%s\n", filepath.ToSlash(s.FilePath)))
		}
		sb.WriteString(fmt.Sprintf("- 描述：%s\n", s.Description))

		if len(s.Capabilities) > 0 {
			sb.WriteString("- 能力：\n")
			for _, cap := range s.Capabilities {
				sb.WriteString(fmt.Sprintf("  - %s\n", cap))
			}
		}

		if len(s.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("- 依赖：%s\n", strings.Join(s.Dependencies, ", ")))
		}

		if len(s.ToolInstructions) > 0 {
			sb.WriteString("- 工具使用说明：\n")
			toolNames := make([]string, 0, len(s.ToolInstructions))
			for toolName := range s.ToolInstructions {
				toolNames = append(toolNames, toolName)
			}
			sort.Strings(toolNames)
			for _, toolName := range toolNames {
				instr := s.ToolInstructions[toolName]
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", toolName, instr.Description))
				if len(instr.Parameters) > 0 {
					paramNames := make([]string, 0, len(instr.Parameters))
					for paramName := range instr.Parameters {
						paramNames = append(paramNames, paramName)
					}
					sort.Strings(paramNames)
					for _, paramName := range paramNames {
						sb.WriteString(fmt.Sprintf("    - %s: %s\n", paramName, instr.Parameters[paramName]))
					}
				}
			}
		}

		if len(s.Examples) > 0 {
			sb.WriteString("- 使用示例：\n")
			for _, ex := range s.Examples {
				sb.WriteString(fmt.Sprintf("  - %s：输入「%s」→ %s\n", ex.Description, ex.Input, ex.Output))
			}
		}

		if s.Workflow != nil && len(s.Workflow.Steps) > 0 {
			sb.WriteString(fmt.Sprintf("- 工作流：%s\n", s.Workflow.Description))
			for _, step := range s.Workflow.Steps {
				sb.WriteString(fmt.Sprintf("  %d. %s：%s\n", step.Step, step.Name, step.Description))
				if step.UseSkill != "" {
					sb.WriteString(fmt.Sprintf("     → 使用技能：%s\n", step.UseSkill))
				}
				if len(step.Actions) > 0 {
					sb.WriteString("     → 执行动作：\n")
					for _, action := range step.Actions {
						sb.WriteString(fmt.Sprintf("       - %s\n", action))
					}
				}
			}
		}
	}

	return sb.String()
}

// GetSkillNames 获取所有技能名称
func GetSkillNames(skills []*Skill) []string {
	names := make([]string, 0, len(skills))
	for _, s := range skills {
		names = append(names, s.Name)
	}
	return names
}
