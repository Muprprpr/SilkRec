package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileServer HTTP 文件服务器
// 用于提供本地文件访问，避免 base64 编码
type FileServer struct {
	server    *http.Server
	port      int
	rootDir   string
	isRunning bool
	mu        sync.Mutex
}

// NewFileServer 创建文件服务器
func NewFileServer(rootDir string, port int) *FileServer {
	return &FileServer{
		rootDir:   rootDir,
		port:      port,
		isRunning: false,
		mu:        sync.Mutex{},
	}
}

// Start 启动文件服务器
func (s *FileServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("文件服务器已在运行中")
	}

	// 确保根目录存在
	if err := os.MkdirAll(s.rootDir, 0755); err != nil {
		return fmt.Errorf("创建根目录失败: %w", err)
	}

	// 创建文件服务器
	mux := http.NewServeMux()

	// 处理静态文件请求
	mux.HandleFunc("/", s.handleFileRequest)

	// 创建服务器
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// 启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("文件服务器错误: %v\n", err)
		}
	}()

	s.isRunning = true
	fmt.Printf("文件服务器已启动: http://localhost:%d (根目录: %s)\n", s.port, s.rootDir)
	return nil
}

// handleFileRequest 处理文件请求
func (s *FileServer) handleFileRequest(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 只接受 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取请求的文件路径
	requestPath := r.URL.Path
	if requestPath == "/" {
		// 根路径，返回空
		w.WriteHeader(http.StatusOK)
		return
	}

	// 移除开头的 /
	filePath := strings.TrimPrefix(requestPath, "/")

	// 构建完整文件路径
	fullPath := filepath.Join(s.rootDir, filePath)

	// 安全检查：确保文件在根目录内
	if !strings.HasPrefix(fullPath, filepath.Clean(s.rootDir)+string(filepath.Separator)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 如果是目录，返回 404
	if info.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 根据文件扩展名设置 Content-Type
	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType := "application/octet-stream"
	switch ext {
	case ".mp4":
		contentType = "video/mp4"
	case ".webm":
		contentType = "video/webm"
	case ".json":
		contentType = "application/json"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)

	// 提供文件
	http.ServeFile(w, r, fullPath)
}

// Stop 停止文件服务器
func (s *FileServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	if s.server != nil {
		if err := s.server.Close(); err != nil {
			return fmt.Errorf("关闭文件服务器失败: %w", err)
		}
	}

	s.isRunning = false
	fmt.Println("文件服务器已停止")
	return nil
}

// GetURL 获取文件的完整 URL
func (s *FileServer) GetURL(filename string) string {
	return fmt.Sprintf("http://localhost:%d/%s", s.port, filename)
}

// GetPort 获取服务器端口
func (s *FileServer) GetPort() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// IsRunning 检查是否正在运行
func (s *FileServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}
