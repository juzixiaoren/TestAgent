package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecuteCommandTool 执行命令工具
func ExecuteCommandTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 获取命令参数
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command parameter is required")
	}

	// 获取工作目录参数（可选）
	workingDir := "."
	if dir, ok := args["working_dir"].(string); ok && dir != "" {
		workingDir = dir
	}

	// 检查命令是否安全（基础检查）
	if !isCommandSafe(command) {
		return nil, fmt.Errorf("命令可能不安全：%s", command)
	}

	// 执行命令
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = workingDir

	// 设置环境变量
	cmd.Env = os.Environ()

	// 执行并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"command":     command,
			"working_dir": workingDir,
			"exit_code":   cmd.ProcessState.ExitCode(),
			"output":      string(output),
			"error":       err.Error(),
			"success":     false,
		}, nil
	}

	// 返回成功结果
	result := map[string]interface{}{
		"command":     command,
		"working_dir": workingDir,
		"exit_code":   0,
		"output":      string(output),
		"success":     true,
		"message":     fmt.Sprintf("命令执行成功，输出长度：%d 字符", len(output)),
	}

	return result, nil
}

// isCommandSafe 基础命令安全检查
func isCommandSafe(command string) bool {
	// 禁止的危险命令
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){:|:&};:", // fork炸弹
		"dd if=/dev/zero",
		"mkfs",
		"fdisk",
	}

	command = strings.ToLower(strings.TrimSpace(command))
	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return false
		}
	}

	// 基本安全检查通过
	return true
}
