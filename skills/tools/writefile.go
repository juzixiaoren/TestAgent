package tools

import (
	"context"
	"fmt"
	"os"
)

// WriteFileTool 文件写入工具
func WriteFileTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 获取文件路径参数
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	// 获取文件内容参数
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required")
	}

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	// 返回结果
	result := map[string]interface{}{
		"file_path": filePath,
		"size":      len(content),
		"success":   true,
		"message":   fmt.Sprintf("文件 %s 写入成功，大小：%d 字节", filePath, len(content)),
	}

	return result, nil
}
