package recorder

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

// HttpPipeServer HTTP 管道服务器
// 用于接收前端通过 HTTP POST 发送的 JPEG Blob 数据
// 并直接写入 FFmpeg stdin，避免 Wails IPC 开销
type HttpPipeServer struct {
	ffmpegPath  string
	ffmpegCmd   *exec.Cmd
	ffmpegStdin io.WriteCloser
	server      *http.Server
	port        int
	isRunning   bool
	mu          sync.Mutex
}

// NewHttpPipeServer 创建 HTTP 管道服务器
func NewHttpPipeServer() *HttpPipeServer {
	return &HttpPipeServer{
		ffmpegPath:  "",
		ffmpegCmd:   nil,
		ffmpegStdin: nil,
		server:      nil,
		port:        13000, // 默认端口 13000
		isRunning:   false,
		mu:          sync.Mutex{},
	}
}

// Start 启动 HTTP 管道服务器和 FFmpeg 进程
func (s *HttpPipeServer) Start(ffmpegPath string, outputPath string, frameRate int, width int, height int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("HTTP 管道服务器已在运行中")
	}

	s.ffmpegPath = ffmpegPath

	// 强制修正宽高为偶数 (防止 libx264 崩溃)
	if width%2 != 0 {
		width--
	}
	if height%2 != 0 {
		height--
	}

	// 构建 FFmpeg 命令（使用 JPEG 输入，HTTP 管道）
	args := []string{
		"-y",
		"-f", "image2pipe",
		"-vcodec", "mjpeg", // 必须明确告诉 ffmpeg 输入流是 mjpeg 格式
		"-r", fmt.Sprintf("%d", frameRate),
		"-i", "-", // 从 stdin 读取

		// 编码参数
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		"-s", fmt.Sprintf("%dx%d", width, height),

		outputPath,
	}

	// 创建 FFmpeg 命令
	s.ffmpegCmd = exec.Command(s.ffmpegPath, args...)

	// 获取 stdin 管道
	stdin, err := s.ffmpegCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建 stdin 管道失败: %w", err)
	}
	s.ffmpegStdin = stdin

	// 启动 FFmpeg 进程
	if err := s.ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 进程失败: %w", err)
	}

	// 启动 HTTP 服务器
	go s.startHTTPServer()

	s.isRunning = true

	fmt.Printf("HTTP 管道服务器已启动: 端口 %d\n", s.port)
	fmt.Printf("FFmpeg 命令: %s %s\n", s.ffmpegPath, args)
	return nil
}

// startHTTPServer 启动 HTTP 服务器
func (s *HttpPipeServer) startHTTPServer() {
	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleFrameUpload)

	// 查找可用端口
	port := s.port
	for {
		s.server = &http.Server{Addr: fmt.Sprintf(":%d", port)}
		if err := s.server.ListenAndServe(); err == nil {
			break
		}
		port++
		if port > 13050 { // 最多尝试 50 个端口
			fmt.Printf("无法找到可用端口\n")
			return
		}
	}
}

// handleFrameUpload 处理帧上传请求
func (s *HttpPipeServer) handleFrameUpload(w http.ResponseWriter, r *http.Request) {
	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 读取请求体
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/jpg" {
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	// 直接将请求体复制到 FFmpeg stdin（零拷贝优化）
	written, err := io.Copy(s.ffmpegStdin, r.Body)
	if err != nil {
		fmt.Printf("写入 FFmpeg stdin 失败: %v\n", err)
		http.Error(w, "Failed to write to FFmpeg", http.StatusInternalServerError)
		return
	}

	fmt.Printf("已接收并写入帧: %d 字节\n", written)
	w.WriteHeader(http.StatusOK)
}

// Stop 停止 HTTP 管道服务器和 FFmpeg 进程
func (s *HttpPipeServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.isRunning = false

	// 关闭 HTTP 服务器
	if s.server != nil {
		if err := s.server.Close(); err != nil {
			fmt.Printf("关闭 HTTP 服务器失败: %v\n", err)
		}
	}

	// 关闭 FFmpeg stdin
	if s.ffmpegStdin != nil {
		if err := s.ffmpegStdin.Close(); err != nil {
			fmt.Printf("关闭 FFmpeg stdin 失败: %v\n", err)
		}
	}

	// 等待 FFmpeg 进程完成
	if s.ffmpegCmd != nil && s.ffmpegCmd.Process != nil {
		// 先尝试优雅等待 5 秒
		done := make(chan error, 1)
		go func() {
			done <- s.ffmpegCmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("FFmpeg 进程退出: %v\n", err)
			}
		case <-time.After(5 * time.Second):
			// 超时后强制终止
			if err := s.ffmpegCmd.Process.Kill(); err != nil {
				fmt.Printf("强制终止 FFmpeg 进程失败: %v\n", err)
			} else {
				fmt.Println("FFmpeg 进程已强制终止")
			}
		}
	}

	fmt.Println("HTTP 管道服务器已停止")
	return nil
}

// GetStatus 获取服务器状态
func (s *HttpPipeServer) GetStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return map[string]interface{}{
		"isRunning":  s.isRunning,
		"port":       s.port,
		"ffmpegPath": s.ffmpegPath,
	}
}

// IsRunning 检查是否正在运行
func (s *HttpPipeServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// GetPort 获取服务器端口
func (s *HttpPipeServer) GetPort() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}
