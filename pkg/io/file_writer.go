package io

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileWriter 管理视频文件的写入
type FileWriter struct {
	file       *os.File
	filePath   string
	mu         sync.Mutex
	isOpen     bool
	totalBytes int64
}

// NewFileWriter 创建新的文件写入器
func NewFileWriter() *FileWriter {
	return &FileWriter{
		isOpen: false,
	}
}

// Open 打开文件准备写入
func (fw *FileWriter) Open(filePath string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isOpen {
		return fmt.Errorf("文件已打开: %s", fw.filePath)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建或截断文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	fw.file = file
	fw.filePath = filePath
	fw.isOpen = true
	fw.totalBytes = 0

	fmt.Printf("文件写入器已打开: %s\n", filePath)
	return nil
}

// Write 写入数据块
func (fw *FileWriter) Write(data []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isOpen || fw.file == nil {
		return 0, fmt.Errorf("文件未打开")
	}

	n, err := fw.file.Write(data)
	if err != nil {
		return n, fmt.Errorf("写入文件失败: %w", err)
	}

	fw.totalBytes += int64(n)
	return n, nil
}

// WriteString 写入字符串
func (fw *FileWriter) WriteString(data string) (int, error) {
	return fw.Write([]byte(data))
}

// Append 追加写入数据
func (fw *FileWriter) Append(data []byte) error {
	_, err := fw.Write(data)
	return err
}

// Close 关闭文件
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isOpen || fw.file == nil {
		return nil
	}

	err := fw.file.Close()
	if err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}

	fmt.Printf("文件已关闭: %s, 总共写入 %d 字节\n", fw.filePath, fw.totalBytes)
	fw.isOpen = false
	fw.file = nil

	return nil
}

// Sync 同步文件到磁盘
func (fw *FileWriter) Sync() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isOpen || fw.file == nil {
		return fmt.Errorf("文件未打开")
	}

	return fw.file.Sync()
}

// GetTotalBytes 获取已写入的总字节数
func (fw *FileWriter) GetTotalBytes() int64 {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.totalBytes
}

// GetFilePath 获取当前文件路径
func (fw *FileWriter) GetFilePath() string {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.filePath
}

// IsOpen 检查文件是否打开
func (fw *FileWriter) IsOpen() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.isOpen
}

// DeleteFile 删除当前文件
func (fw *FileWriter) DeleteFile() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.filePath == "" {
		return fmt.Errorf("没有文件路径")
	}

	// 先关闭文件
	if fw.file != nil {
		fw.file.Close()
	}

	err := os.Remove(fw.filePath)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	fmt.Printf("文件已删除: %s\n", fw.filePath)
	fw.filePath = ""
	fw.isOpen = false
	fw.file = nil

	return nil
}

// WriteJSON 写入JSON数据
func (fw *FileWriter) WriteJSON(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	_, err = fw.Write(jsonData)
	return err
}

// ReadFile 读取文件内容
func (fw *FileWriter) ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return data, nil
}
