package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ExecuteCommandTool 执行命令工具。
func ExecuteCommandTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok || strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("command parameter is required")
	}
	command = strings.TrimSpace(command)

	workingDir := "."
	if dir, ok := args["working_dir"].(string); ok && dir != "" {
		workingDir = dir
	}

	timeoutSeconds := 30.0
	if raw, ok := args["timeout_seconds"]; ok {
		switch v := raw.(type) {
		case float64:
			if v > 0 {
				timeoutSeconds = v
			}
		case int:
			if v > 0 {
				timeoutSeconds = float64(v)
			}
		case string:
			var parsed float64
			if _, err := fmt.Sscanf(v, "%f", &parsed); err == nil && parsed > 0 {
				timeoutSeconds = parsed
			}
		}
	}

	shellChoice := os.Getenv("EXECUTE_COMMAND_SHELL")
	if shellChoice == "" {
		shellChoice = "auto"
	}
	if raw, ok := args["shell"].(string); ok && strings.TrimSpace(raw) != "" {
		shellChoice = strings.TrimSpace(raw)
	}

	if !isCommandSafe(command) {
		return nil, fmt.Errorf("命令可能不安全：%s", command)
	}

	shellName, shellArgs, err := buildShellCommand(shellChoice, command)
	if err != nil {
		return nil, err
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds*float64(time.Second)))
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, shellName, shellArgs...)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	result := map[string]interface{}{
		"command":         command,
		"working_dir":     workingDir,
		"timeout_seconds": timeoutSeconds,
		"shell":           shellChoice,
		"shell_command":   shellName + " " + strings.Join(shellArgs, " "),
		"output":          string(output),
	}

	if cmdCtx.Err() == context.DeadlineExceeded {
		result["exit_code"] = -1
		result["error"] = fmt.Sprintf("命令执行超时，超过 %.0f 秒", timeoutSeconds)
		result["success"] = false
		return result, nil
	}

	if err != nil {
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		result["exit_code"] = exitCode
		result["error"] = err.Error()
		result["success"] = false
		return result, nil
	}

	result["exit_code"] = 0
	result["success"] = true
	result["message"] = fmt.Sprintf("命令执行成功，输出长度：%d 字符", len(output))
	return result, nil
}

func buildShellCommand(choice string, command string) (string, []string, error) {
	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "" || choice == "auto" {
		if runtime.GOOS == "windows" {
			if comspec := os.Getenv("COMSPEC"); comspec != "" {
				return comspec, []string{"/C", command}, nil
			}
			return "cmd", []string{"/C", command}, nil
		}
		if bash, err := exec.LookPath("bash"); err == nil {
			return bash, []string{"-lc", command}, nil
		}
		if sh, err := exec.LookPath("sh"); err == nil {
			return sh, []string{"-c", command}, nil
		}
		return "", nil, fmt.Errorf("auto shell 未找到可用的 bash 或 sh")
	}

	switch choice {
	case "cmd":
		if runtime.GOOS == "windows" {
			if comspec := os.Getenv("COMSPEC"); comspec != "" {
				return comspec, []string{"/C", command}, nil
			}
		}
		return "cmd", []string{"/C", command}, nil
	case "powershell":
		return "powershell", []string{"-NoProfile", "-Command", command}, nil
	case "pwsh":
		return "pwsh", []string{"-NoProfile", "-Command", command}, nil
	case "bash":
		return "bash", []string{"-lc", command}, nil
	case "sh":
		return "sh", []string{"-c", command}, nil
	default:
		return "", nil, fmt.Errorf("不支持的 shell: %s", choice)
	}
}

// isCommandSafe 基础命令安全检查。
func isCommandSafe(command string) bool {
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){:|:&};:",
		"dd if=/dev/zero",
		"mkfs",
		"fdisk",
		"format c:",
		"del /s /q c:",
		"shutdown /s",
	}

	command = strings.ToLower(strings.TrimSpace(command))
	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return false
		}
	}
	return true
}
