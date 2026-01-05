package recorder

import (
	"fmt"
	"os"
	"sync"
)

// VideoWriter H.264 视频写入器
// 接收 WebCodecs 编码的 H.264 数据块并写入文件
type VideoWriter struct {
	file       *os.File
	isWriting  bool
	mu         sync.Mutex
	totalBytes int64
	outputPath string
}

// NewVideoWriter 创建视频写入器
func NewVideoWriter() *VideoWriter {
	return &VideoWriter{
		isWriting:  false,
		totalBytes: 0,
	}
}

// StartWriter 开始写入
func (v *VideoWriter) StartWriter(filePath string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.isWriting {
		return fmt.Errorf("写入已在进行中")
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	v.file = file
	v.outputPath = filePath
	v.isWriting = true
	v.totalBytes = 0

	fmt.Printf("VideoWriter 已启动: %s\n", filePath)
	return nil
}

// WriteChunk 写入数据块
func (v *VideoWriter) WriteChunk(data []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.isWriting || v.file == nil {
		return fmt.Errorf("写入未启动或文件未打开")
	}

	// 写入数据
	n, err := v.file.Write(data)
	if err != nil {
		v.isWriting = false
		return fmt.Errorf("写入数据失败: %w", err)
	}

	v.totalBytes += int64(n)

	return nil
}

// StopWriter 停止写入
func (v *VideoWriter) StopWriter() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.isWriting {
		return nil
	}

	// 关闭文件
	if v.file != nil {
		if err := v.file.Close(); err != nil {
			return fmt.Errorf("关闭文件失败: %w", err)
		}
		v.file = nil
	}

	v.isWriting = false
	fmt.Printf("VideoWriter 已停止: %s, 共写入 %d bytes\n", v.outputPath, v.totalBytes)
	return nil
}

// GetStatus 获取写入状态
func (v *VideoWriter) GetStatus() map[string]interface{} {
	v.mu.Lock()
	defer v.mu.Unlock()

	return map[string]interface{}{
		"isWriting":  v.isWriting,
		"totalBytes": v.totalBytes,
		"outputPath": v.outputPath,
	}
}

// GetTotalBytes 获取总字节数
func (v *VideoWriter) GetTotalBytes() int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.totalBytes
}

// IsWriting 检查是否正在写入
func (v *VideoWriter) IsWriting() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.isWriting
}
