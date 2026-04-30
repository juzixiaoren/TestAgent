package skill

import (
	"fmt"
	"os"
	"path/filepath"
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

// LoadSkills 从指定目录加载所有 .skill.yaml 文件
func LoadSkills(dir string) ([]*Skill, error) {
	// 确保目录存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("技能目录不存在: %s", dir)
	}

	var skills []*Skill

	// 遍历目录下的 .skill.yaml 文件
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取技能目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".skill.yaml") && !strings.HasSuffix(name, ".skill.yml") {
			continue
		}

		filePath := filepath.Join(dir, name)
		skill, err := loadSkillFile(filePath)
		if err != nil {
			fmt.Printf("警告：加载技能文件 %s 失败: %v\n", filePath, err)
			continue
		}

		skill.FilePath = filePath
		skills = append(skills, skill)
	}

	return skills, nil
}

// loadSkillFile 加载单个技能文件
func loadSkillFile(filePath string) (*Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var skill Skill
	if err := yaml.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %v", err)
	}

	if skill.Name == "" {
		return nil, fmt.Errorf("技能名称不能为空")
	}

	return &skill, nil
}

// FormatSkillsPrompt 将技能列表格式化为系统提示词内容
func FormatSkillsPrompt(skills []*Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## 可用技能\n")
	sb.WriteString("以下是你可以使用的技能，根据用户需求选择合适的技能和工具：\n")

	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("\n### %s (v%s)\n", s.Name, s.Version))
		sb.WriteString(fmt.Sprintf("- 描述：%s\n", s.Description))

		if len(s.Capabilities) > 0 {
			sb.WriteString("- 能力：\n")
			for _, cap := range s.Capabilities {
				sb.WriteString(fmt.Sprintf("  - %s\n", cap))
			}
		}

		if len(s.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("- 依赖工具：%s\n", strings.Join(s.Dependencies, ", ")))
		}

		if len(s.ToolInstructions) > 0 {
			sb.WriteString("- 工具使用说明：\n")
			for toolName, instr := range s.ToolInstructions {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", toolName, instr.Description))
				if len(instr.Parameters) > 0 {
					for paramName, paramDesc := range instr.Parameters {
						sb.WriteString(fmt.Sprintf("    - %s: %s\n", paramName, paramDesc))
					}
				}
			}
		}

		if len(s.Examples) > 0 {
			sb.WriteString("- 使用示例：\n")
			for _, ex := range s.Examples {
				sb.WriteString(fmt.Sprintf("  - %s：输入\"%s\" → %s\n", ex.Description, ex.Input, ex.Output))
			}
		}

		if s.Workflow != nil && len(s.Workflow.Steps) > 0 {
			sb.WriteString(fmt.Sprintf("- 工作流：%s\n", s.Workflow.Description))
			for _, step := range s.Workflow.Steps {
				sb.WriteString(fmt.Sprintf("  %d. %s： %s\n", step.Step, step.Name, step.Description))
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
