package recorder

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// FFmpegCapture FFmpeg 屏幕捕获器
type FFmpegCapture struct {
	cmd       *exec.Cmd
	process   *os.Process
	isRunning bool
	mu        sync.Mutex
	output    strings.Builder
	error     strings.Builder
	stdin     io.WriteCloser
}

// NewFFmpegCapture 创建 FFmpeg 捕获器
func NewFFmpegCapture() *FFmpegCapture {
	return &FFmpegCapture{
		isRunning: false,
	}
}

// Start 启动 FFmpeg 进程
func (c *FFmpegCapture) Start(ffmpegPath string, args []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("FFmpeg 进程已在运行")
	}

	// 创建命令
	c.cmd = exec.Command(ffmpegPath, args...)

	// 获取 stdin 管道（用于发送停止命令）
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取 stdin 管道失败: %w", err)
	}
	c.stdin = stdin

	// 获取输出管道
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取 stdout 管道失败: %w", err)
	}

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取 stderr 管道失败: %w", err)
	}

	// 启动进程
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 进程失败: %w", err)
	}

	// 保存进程引用
	c.process = c.cmd.Process
	c.isRunning = true
	c.output.Reset()
	c.error.Reset()

	// 启动 goroutine 读取输出
	go c.readOutput(stdout)
	go c.readOutput(stderr)

	// 等待进程结束
	go c.waitForProcess()

	fmt.Printf("FFmpeg 进程已启动: %s %s\n", ffmpegPath, strings.Join(args, " "))
	return nil
}

// readOutput 读取进程输出
func (c *FFmpegCapture) readOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		c.output.WriteString(line)
		c.output.WriteString("\n")
	}
}

// waitForProcess 等待进程结束
func (c *FFmpegCapture) waitForProcess() {
	err := c.cmd.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isRunning = false
	if err != nil {
		c.error.WriteString(fmt.Sprintf("FFmpeg 进程错误: %v\n", err))
	}
	fmt.Println("FFmpeg 进程已结束")
}

// Stop 停止 FFmpeg 进程
func (c *FFmpegCapture) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning || c.process == nil {
		return nil
	}

	fmt.Println("正在停止 FFmpeg 进程...")

	// 方法 1: 发送 'q' 命令到 FFmpeg 的 stdin（最优雅的方式）
	if c.stdin != nil {
		c.stdin.Write([]byte("q\n"))
		c.stdin.Close()
		c.stdin = nil
		fmt.Println("已发送 'q' 命令到 FFmpeg")
	}

	// 等待进程正常结束（最多等待 5 秒）
	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("FFmpeg 进程退出，错误: %v\n", err)
		} else {
			fmt.Println("FFmpeg 进程已正常退出")
		}
		c.isRunning = false
		return nil
	case <-time.After(5 * time.Second):
		// 超时，尝试发送信号
		fmt.Println("FFmpeg 进程未响应，尝试发送信号...")

		// 方法 2: 发送 SIGTERM 信号
		if err := c.process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("发送 SIGTERM 失败: %v\n", err)
		} else {
			fmt.Println("已发送 SIGTERM 信号")
		}

		// 再等待 3 秒
		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("FFmpeg 进程退出，错误: %v\n", err)
			} else {
				fmt.Println("FFmpeg 进程已正常退出")
			}
			c.isRunning = false
			return nil
		case <-time.After(3 * time.Second):
			// 方法 3: 强制终止
			fmt.Println("FFmpeg 进程仍未响应，强制终止...")
			if err := c.process.Kill(); err != nil {
				return fmt.Errorf("强制终止 FFmpeg 进程失败: %w", err)
			}
			<-done
			c.isRunning = false
			fmt.Println("FFmpeg 进程已被强制终止")
			return nil
		}
	}
}

// IsRunning 检查是否正在运行
func (c *FFmpegCapture) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}

// GetOutput 获取输出日志
func (c *FFmpegCapture) GetOutput() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.output.String()
}

// GetError 获取错误信息
func (c *FFmpegCapture) GetError() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.error.String()
}

// GetPID 获取进程 ID
func (c *FFmpegCapture) GetPID() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.process != nil {
		return c.process.Pid
	}
	return 0
}
