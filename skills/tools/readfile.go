package tools

import (
	"context"
	"fmt"
	"io"
	"os"
)

// ReadFileTool 文件读取工具
func ReadFileTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 获取文件路径参数
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// 返回结果
	result := map[string]interface{}{
		"file_path": filePath,
		"content":   string(content),
		"size":      len(content),
		"success":   true,
	}

	return result, nil
}
