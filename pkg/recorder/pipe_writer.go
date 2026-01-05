package recorder

import (
	"SmoothScreen/pkg/ffmpeg"
	"encoding/base64"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// PipeWriter FFmpeg 管道写入器
// 接收 base64 图像数据并写入 FFmpeg stdin
type PipeWriter struct {
	ffmpegManager *ffmpeg.FFmpegManager
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	isWriting     bool
	mu            sync.Mutex
	totalFrames   int
	outputPath    string
}

// NewPipeWriter 创建 FFmpeg 管道写入器
func NewPipeWriter(ffmpegManager *ffmpeg.FFmpegManager) *PipeWriter {
	return &PipeWriter{
		ffmpegManager: ffmpegManager,
		isWriting:     false,
		totalFrames:   0,
	}
}

// StartExport 开始导出
func (p *PipeWriter) StartExport(outputPath string, frameRate int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isWriting {
		return fmt.Errorf("导出已在进行中")
	}

	// 获取 FFmpeg 路径
	ffmpegPath, err := p.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	// 构建导出命令
	args := BuildExportCommand(ffmpegPath, outputPath, frameRate)

	// 创建命令
	p.cmd = exec.Command(ffmpegPath, args...)

	// 获取 stdin 管道
	stdin, err := p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建 stdin 管道失败: %w", err)
	}

	p.stdin = stdin
	p.outputPath = outputPath

	// 启动 FFmpeg 进程
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 进程失败: %w", err)
	}

	p.isWriting = true
	p.totalFrames = 0

	fmt.Printf("FFmpeg 导出进程已启动: %s\n", outputPath)
	fmt.Printf("命令: %s %s\n", ffmpegPath, args)
	return nil
}

// WriteFrame 写入一帧
func (p *PipeWriter) WriteFrame(base64Data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isWriting || p.stdin == nil {
		return fmt.Errorf("导出未启动或 stdin 未初始化")
	}

	// 解码 base64 数据
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("解码 base64 数据失败: %w", err)
	}

	// toDataURL() 返回的已经是完整的 PNG 文件格式（包含 PNG 文件头）
	// 直接写入即可
	if _, err := p.stdin.Write(imageData); err != nil {
		// 写入失败时，标记为不再写入
		p.isWriting = false
		return fmt.Errorf("写入图像数据失败: %w", err)
	}

	p.totalFrames++

	return nil
}

// WriteFrameBinary 写入二进制帧数据
func (p *PipeWriter) WriteFrameBinary(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isWriting || p.stdin == nil {
		return fmt.Errorf("导出未启动或 stdin 未初始化")
	}

	// 直接写入完整的 PNG 数据（已包含 PNG 文件头）
	if _, err := p.stdin.Write(data); err != nil {
		return fmt.Errorf("写入图像数据失败: %w", err)
	}

	p.totalFrames++

	return nil
}

// FinishExport 完成导出
func (p *PipeWriter) FinishExport() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isWriting {
		return fmt.Errorf("没有正在进行的导出")
	}

	// 关闭 stdin
	if p.stdin != nil {
		if err := p.stdin.Close(); err != nil {
			return fmt.Errorf("关闭 stdin 失败: %w", err)
		}
		p.stdin = nil
	}

	// 等待 FFmpeg 进程完成
	if err := p.cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg 进程退出失败: %w", err)
	}

	p.isWriting = false
	fmt.Printf("导出完成: %s, 共 %d 帧\n", p.outputPath, p.totalFrames)
	return nil
}

// StopExport 停止导出
func (p *PipeWriter) StopExport() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isWriting {
		return nil
	}

	// 先标记为不再写入，防止继续写入
	p.isWriting = false

	// 关闭 stdin（忽略错误，可能已经关闭）
	if p.stdin != nil {
		_ = p.stdin.Close()
		p.stdin = nil
	}

	// 等待 FFmpeg 进程自然退出（最多等待 5 秒）
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("FFmpeg 进程退出: %v\n", err)
		}
	case <-time.After(5 * time.Second):
		// 超时后强制终止
		if p.cmd.Process != nil {
			if err := p.cmd.Process.Kill(); err != nil {
				fmt.Printf("终止 FFmpeg 进程失败: %v\n", err)
			} else {
				fmt.Println("FFmpeg 进程已强制终止")
			}
		}
	}

	fmt.Println("导出已停止")
	return nil
}

// GetStatus 获取导出状态
func (p *PipeWriter) GetStatus() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"isWriting":   p.isWriting,
		"totalFrames": p.totalFrames,
		"outputPath":  p.outputPath,
	}
}

// GetTotalFrames 获取总帧数
func (p *PipeWriter) GetTotalFrames() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.totalFrames
}

// IsWriting 检查是否正在写入
func (p *PipeWriter) IsWriting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isWriting
}
