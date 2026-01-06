package recorder

import (
	"SmoothScreen/pkg/ffmpeg"
	"SmoothScreen/pkg/hook"
	"SmoothScreen/pkg/io"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Recorder 录制管理器
type Recorder struct {
	ffmpegManager *ffmpeg.FFmpegManager
	capture       *FFmpegCapture
	mouseHook     *hook.MouseHook
	fileWriter    *io.FileWriter
	startTime     time.Time
	isRecording   bool
	outputPath    string
	mouseDataPath string
	ctx           context.Context
}

// RecorderStatus 录制状态
type RecorderStatus struct {
	IsRecording     bool   `json:"isRecording"`
	OutputPath      string `json:"outputPath"`
	MouseDataPath   string `json:"mouseDataPath"`
	Duration        int64  `json:"duration"` // 录制时长（毫秒）
	MouseEventCount int    `json:"mouseEventCount"`
	FFmpegPID       int    `json:"ffmpegPID"`
}

// NewRecorder 创建录制管理器
func NewRecorder(ffmpegManager *ffmpeg.FFmpegManager, mouseHook *hook.MouseHook, ctx context.Context) *Recorder {
	return &Recorder{
		ffmpegManager: ffmpegManager,
		mouseHook:     mouseHook,
		fileWriter:    io.NewFileWriter(),
		ctx:           ctx,
		isRecording:   false,
	}
}

// StartRecording 开始录制
func (r *Recorder) StartRecording(outputPath string) error {
	if r.isRecording {
		return fmt.Errorf("录制已在进行中")
	}

	// 检查 FFmpeg 是否可用
	if !r.ffmpegManager.CheckFFmpegAvailable() {
		return fmt.Errorf("FFmpeg 不可用，请确保 ffmpeg.exe 在正确位置")
	}

	// 获取 FFmpeg 路径
	ffmpegPath, err := r.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成鼠标数据文件路径
	mouseDataPath := filepath.Join(filepath.Dir(outputPath), "mouse_events.json")

	// 检测最佳编码器
	codec, err := r.ffmpegManager.GetBestEncoder()
	if err != nil {
		return fmt.Errorf("检测编码器失败: %w", err)
	}

	// 获取最佳预设
	preset := r.ffmpegManager.GetBestPreset(codec)

	// 构建捕获配置
	config := DefaultCaptureConfig(outputPath)
	config.Codec = codec
	config.Preset = preset

	// 尝试使用 ddagrap（推荐）
	var args []string
	args = BuildDDAGrabCommand(ffmpegPath, config)

	// 创建 FFmpeg 捕获器
	r.capture = NewFFmpegCapture()

	// 启动 FFmpeg 进程
	if err := r.capture.Start(ffmpegPath, args); err != nil {
		// 如果 ddagrap 失败，尝试使用 gdigrab
		fmt.Println("ddagrap 失败，尝试使用 gdigrab...")
		args = BuildGDIRABCommand(ffmpegPath, config)
		if err := r.capture.Start(ffmpegPath, args); err != nil {
			return fmt.Errorf("启动 FFmpeg 捕获失败: %w", err)
		}
	}

	// 开始录制鼠标数据
	r.mouseHook.StartRecording()

	// 记录开始时间
	r.startTime = time.Now()
	r.isRecording = true
	r.outputPath = outputPath
	r.mouseDataPath = mouseDataPath

	fmt.Printf("录制已开始: %s\n", outputPath)
	fmt.Printf("使用编码器: %s\n", codec)
	return nil
}

// StopRecording 停止录制
func (r *Recorder) StopRecording() (string, string, error) {
	if !r.isRecording {
		return "", "", fmt.Errorf("没有正在进行的录制")
	}

	// 停止 FFmpeg 捕获
	if r.capture != nil {
		fmt.Println("正在停止 FFmpeg 捕获...")
		if err := r.capture.Stop(); err != nil {
			// 打印错误但不返回，继续处理鼠标数据
			fmt.Printf("停止 FFmpeg 捕获失败: %v\n", err)
			// 检查是否有错误日志
			if ffmpegErr := r.capture.GetError(); ffmpegErr != "" {
				fmt.Printf("FFmpeg 错误日志: %s\n", ffmpegErr)
			}
		} else {
			fmt.Println("FFmpeg 捕获已停止")
		}
	}

	// 等待 FFmpeg 完成文件写入（最多等待 3 秒）
	fmt.Println("等待 FFmpeg 完成文件写入...")
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(r.outputPath); err == nil {
			// 文件存在，检查文件大小
			if info, err := os.Stat(r.outputPath); err == nil && info.Size() > 1024 {
				// 文件大小大于 1KB，认为已写入完成
				fmt.Printf("视频文件已写入: %s (大小: %d 字节)\n", r.outputPath, info.Size())
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 停止鼠标录制
	r.mouseHook.StopRecording()

	// 保存鼠标数据到文件
	mouseData := r.mouseHook.GetMouseData()
	mouseDataJSON, err := json.MarshalIndent(mouseData, "", "  ")
	if err != nil {
		return r.outputPath, "", fmt.Errorf("序列化鼠标数据失败: %w", err)
	}

	// 写入鼠标数据文件
	if err := r.fileWriter.Open(r.mouseDataPath); err != nil {
		return r.outputPath, "", fmt.Errorf("打开鼠标数据文件失败: %w", err)
	}
	if _, err := r.fileWriter.Write(mouseDataJSON); err != nil {
		return r.outputPath, "", fmt.Errorf("写入鼠标数据失败: %w", err)
	}
	r.fileWriter.Close()

	// 更新状态
	r.isRecording = false
	duration := time.Since(r.startTime).Milliseconds()

	fmt.Printf("录制已停止: %s\n", r.outputPath)
	fmt.Printf("鼠标数据已保存: %s\n", r.mouseDataPath)
	fmt.Printf("录制时长: %d ms\n", duration)
	fmt.Printf("鼠标事件数量: %d\n", len(mouseData))

	// 返回原始视频路径和鼠标数据路径
	return r.outputPath, r.mouseDataPath, nil
}

// GetStatus 获取录制状态
func (r *Recorder) GetStatus() RecorderStatus {
	status := RecorderStatus{
		IsRecording: r.isRecording,
		OutputPath:  r.outputPath,
	}

	if r.isRecording {
		status.Duration = time.Since(r.startTime).Milliseconds()
	}

	if r.mouseHook != nil {
		mouseData := r.mouseHook.GetMouseData()
		status.MouseEventCount = len(mouseData)
	}

	if r.capture != nil {
		status.FFmpegPID = r.capture.GetPID()
	}

	return status
}

// GetMouseData 获取鼠标数据
func (r *Recorder) GetMouseData() []hook.MouseEvent {
	if r.mouseHook == nil {
		return nil
	}
	return r.mouseHook.GetMouseData()
}

// GetMouseDataJSON 获取鼠标数据的 JSON 格式
func (r *Recorder) GetMouseDataJSON() (string, error) {
	if r.mouseHook == nil {
		return "", fmt.Errorf("鼠标钩子未初始化")
	}
	return r.mouseHook.GetMouseDataJSON()
}

// SaveMouseData 保存鼠标数据到指定文件
func (r *Recorder) SaveMouseData(filePath string) error {
	if r.mouseHook == nil {
		return fmt.Errorf("鼠标钩子未初始化")
	}

	jsonData, err := r.mouseHook.GetMouseDataJSON()
	if err != nil {
		return fmt.Errorf("获取鼠标数据失败: %w", err)
	}

	if jsonData == "" {
		return fmt.Errorf("没有鼠标数据可保存")
	}

	// 创建临时文件写入器
	tempWriter := io.NewFileWriter()
	defer tempWriter.Close()

	if err := tempWriter.Open(filePath); err != nil {
		return fmt.Errorf("打开鼠标数据文件失败: %w", err)
	}

	if _, err := tempWriter.WriteString(jsonData); err != nil {
		return fmt.Errorf("写入鼠标数据失败: %w", err)
	}

	fmt.Printf("鼠标数据已保存到: %s\n", filePath)
	return nil
}

// GetOutputPath 获取输出视频路径
func (r *Recorder) GetOutputPath() string {
	return r.outputPath
}

// GetMouseDataPath 获取鼠标数据路径
func (r *Recorder) GetMouseDataPath() string {
	return r.mouseDataPath
}

// IsRecording 检查是否正在录制
func (r *Recorder) IsRecording() bool {
	return r.isRecording
}
