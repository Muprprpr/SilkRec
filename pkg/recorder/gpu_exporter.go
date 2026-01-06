package recorder

import (
	"SmoothScreen/pkg/ffmpeg"
	"SmoothScreen/pkg/hook"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GPUExporter GPU 加速的视频导出器
// 使用 FFmpeg 硬件加速和滤镜链处理视频，无需前端渲染
type GPUExporter struct {
	ffmpegManager *ffmpeg.FFmpegManager
	config        ExportConfig
	mouseEvents   []hook.MouseEvent
	cameraFrames  []CameraFrame
	cmd           *exec.Cmd
	isExporting   bool
}

// NewGPUExporter 创建 GPU 加速导出器
func NewGPUExporter(ffmpegManager *ffmpeg.FFmpegManager) *GPUExporter {
	return &GPUExporter{
		ffmpegManager: ffmpegManager,
		isExporting:   false,
	}
}

// PrepareExport 准备导出
func (e *GPUExporter) PrepareExport(config ExportConfig) error {
	e.config = config

	// 加载鼠标数据
	if err := e.loadMouseData(); err != nil {
		return fmt.Errorf("加载鼠标数据失败: %w", err)
	}

	// 生成相机路径
	if err := e.generateCameraPath(); err != nil {
		return fmt.Errorf("生成相机路径失败: %w", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(config.OutputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	return nil
}

// loadMouseData 加载鼠标数据
func (e *GPUExporter) loadMouseData() error {
	data, err := os.ReadFile(e.config.MouseDataPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &e.mouseEvents); err != nil {
		return err
	}

	fmt.Printf("✓ 加载了 %d 个鼠标事件\n", len(e.mouseEvents))
	return nil
}

// generateCameraPath 生成相机路径
func (e *GPUExporter) generateCameraPath() error {
	e.cameraFrames = GenerateCameraPath(
		e.mouseEvents,
		e.config.ScreenWidth,
		e.config.ScreenHeight,
		e.config.FPS,
	)

	fmt.Printf("✓ 生成了 %d 个相机帧\n", len(e.cameraFrames))
	return nil
}

// ExportWithGPU 使用 GPU 加速导出视频
// 此方法使用 FFmpeg 的硬件加速和滤镜链，无需前端渲染
func (e *GPUExporter) ExportWithGPU() error {
	if e.isExporting {
		return fmt.Errorf("导出已在进行中")
	}

	// 获取 FFmpeg 路径
	ffmpegPath, err := e.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	// 获取最佳编码器
	codec, err := e.ffmpegManager.GetBestEncoder()
	if err != nil {
		codec = "libx264" // 回退到软件编码
	}

	// 获取预设
	preset := e.ffmpegManager.GetBestPreset(codec)

	// 构建 FFmpeg 命令
	args := e.buildGPUExportCommand(ffmpegPath, codec, preset)

	fmt.Printf("执行 GPU 加速导出命令:\n%s %s\n", ffmpegPath, strings.Join(args, " "))

	// 创建命令
	e.cmd = exec.Command(ffmpegPath, args...)
	e.cmd.Stdout = os.Stdout
	e.cmd.Stderr = os.Stderr

	e.isExporting = true

	// 启动 FFmpeg
	startTime := time.Now()
	if err := e.cmd.Run(); err != nil {
		e.isExporting = false
		return fmt.Errorf("FFmpeg 执行失败: %w", err)
	}

	e.isExporting = false
	duration := time.Since(startTime)

	fmt.Printf("✓ GPU 加速导出完成: %s (耗时: %.2f 秒)\n", e.config.OutputPath, duration.Seconds())
	return nil
}

// buildGPUExportCommand 构建 GPU 加速的 FFmpeg 命令
func (e *GPUExporter) buildGPUExportCommand(ffmpegPath, codec, preset string) []string {
	args := []string{}

	// 硬件加速选项
	if strings.Contains(codec, "nvenc") {
		// NVIDIA GPU 加速
		args = append(args, "-hwaccel", "cuda")
		args = append(args, "-hwaccel_output_format", "cuda")
	} else if strings.Contains(codec, "qsv") {
		// Intel QSV 加速
		args = append(args, "-hwaccel", "qsv")
		args = append(args, "-hwaccel_output_format", "qsv")
	} else if strings.Contains(codec, "amf") {
		// AMD AMF 加速
		args = append(args, "-hwaccel", "d3d11va")
		args = append(args, "-hwaccel_output_format", "d3d11")
	}

	// 输入文件
	args = append(args, "-i", e.config.VideoPath)

	// 构建复杂滤镜链
	filterComplex := e.buildFilterComplex()
	if filterComplex != "" {
		args = append(args, "-filter_complex", filterComplex)
	}

	// 编码器设置
	args = append(args, "-c:v", codec)

	// 编码器特定选项
	if strings.Contains(codec, "nvenc") {
		// NVIDIA 编码器优化
		args = append(args, "-preset", preset)
		args = append(args, "-rc", "vbr")        // 可变比特率
		args = append(args, "-cq", "23")         // 质量控制
		args = append(args, "-b:v", "5M")        // 目标比特率
		args = append(args, "-maxrate", "8M")    // 最大比特率
		args = append(args, "-bufsize", "10M")   // 缓冲区大小
		args = append(args, "-spatial_aq", "1")  // 空间自适应量化
		args = append(args, "-temporal_aq", "1") // 时间自适应量化
		args = append(args, "-gpu", "0")         // 使用第一个 GPU
	} else if strings.Contains(codec, "qsv") {
		// Intel QSV 优化
		args = append(args, "-preset", preset)
		args = append(args, "-global_quality", "23")
		args = append(args, "-look_ahead", "1")
	} else if strings.Contains(codec, "amf") {
		// AMD AMF 优化
		args = append(args, "-quality", "balanced")
		args = append(args, "-rc", "vbr_latency")
		args = append(args, "-qp_i", "22")
		args = append(args, "-qp_p", "23")
	} else {
		// 软件编码 (libx264) 优化
		args = append(args, "-preset", preset)
		args = append(args, "-crf", "23")
	}

	// 帧率
	args = append(args, "-r", fmt.Sprintf("%d", e.config.FPS))

	// Pixel format
	args = append(args, "-pix_fmt", "yuv420p")

	// 覆盖输出文件
	args = append(args, "-y")

	// 输出文件
	args = append(args, e.config.OutputPath)

	return args
}

// buildFilterComplex 构建复杂滤镜链
// 使用 zoompan 滤镜实现相机运动
func (e *GPUExporter) buildFilterComplex() string {
	if len(e.cameraFrames) == 0 {
		return ""
	}

	// 生成 zoompan 表达式
	zoomExpr := e.generateZoomExpression()
	xExpr := e.generateXExpression()
	yExpr := e.generateYExpression()

	// 构建滤镜链
	filters := []string{}

	// 1. zoompan 滤镜 - 实现缩放和平移
	zoompanFilter := fmt.Sprintf(
		"zoompan=z='%s':x='%s':y='%s':d=1:s=%dx%d:fps=%d",
		zoomExpr,
		xExpr,
		yExpr,
		e.config.ScreenWidth,
		e.config.ScreenHeight,
		e.config.FPS,
	)
	filters = append(filters, zoompanFilter)

	// 2. 如果需要绘制光标（可选）
	if e.config.ShowCursor {
		// 使用 drawtext 滤镜绘制光标指示器
		// 注意：这里简化处理，实际可以用 overlay 叠加光标图像
		cursorFilter := e.generateCursorFilter()
		if cursorFilter != "" {
			filters = append(filters, cursorFilter)
		}
	}

	return strings.Join(filters, ",")
}

// generateZoomExpression 生成缩放表达式
// 使用 FFmpeg 表达式根据时间戳计算缩放级别
func (e *GPUExporter) generateZoomExpression() string {
	// 简化版本：使用线性插值
	// 更复杂的版本可以生成完整的 if-then-else 表达式

	if len(e.cameraFrames) == 0 {
		return "1.0"
	}

	// 对于简单情况，返回固定缩放
	// 复杂的动态缩放需要生成长表达式
	avgZoom := 0.0
	for _, frame := range e.cameraFrames {
		avgZoom += frame.Zoom
	}
	avgZoom /= float64(len(e.cameraFrames))

	return fmt.Sprintf("%.3f", avgZoom)
}

// generateXExpression 生成 X 坐标表达式
func (e *GPUExporter) generateXExpression() string {
	if len(e.cameraFrames) == 0 {
		return "iw/2-(iw/zoom/2)"
	}

	// 计算平均 X 位置
	avgX := 0.0
	for _, frame := range e.cameraFrames {
		avgX += frame.X
	}
	avgX /= float64(len(e.cameraFrames))

	return fmt.Sprintf("%.1f-(iw/zoom/2)", avgX)
}

// generateYExpression 生成 Y 坐标表达式
func (e *GPUExporter) generateYExpression() string {
	if len(e.cameraFrames) == 0 {
		return "ih/2-(ih/zoom/2)"
	}

	// 计算平均 Y 位置
	avgY := 0.0
	for _, frame := range e.cameraFrames {
		avgY += frame.Y
	}
	avgY /= float64(len(e.cameraFrames))

	return fmt.Sprintf("%.1f-(ih/zoom/2)", avgY)
}

// generateCursorFilter 生成光标绘制滤镜
func (e *GPUExporter) generateCursorFilter() string {
	// 这里简化处理，实际应该使用 overlay 滤镜叠加光标图像
	// 或者在前端预处理时叠加光标
	return ""
}

// ExportWithSegments 分段导出（更精确的相机控制）
// 将视频分成多个段，每段应用不同的滤镜参数
func (e *GPUExporter) ExportWithSegments() error {
	if e.isExporting {
		return fmt.Errorf("导出已在进行中")
	}

	// 获取 FFmpeg 路径
	ffmpegPath, err := e.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	// 获取编码器
	codec, err := e.ffmpegManager.GetBestEncoder()
	if err != nil {
		codec = "libx264"
	}
	preset := e.ffmpegManager.GetBestPreset(codec)

	fmt.Printf("开始分段导出 (共 %d 帧)...\n", len(e.cameraFrames))

	// 创建临时目录
	tempDir := filepath.Join(filepath.Dir(e.config.OutputPath), "temp_segments")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir) // 清理临时文件

	// 分段导出
	segmentSize := 300 // 每段 300 帧 (10秒@30fps)
	segments := []string{}

	for i := 0; i < len(e.cameraFrames); i += segmentSize {
		end := i + segmentSize
		if end > len(e.cameraFrames) {
			end = len(e.cameraFrames)
		}

		segmentPath := filepath.Join(tempDir, fmt.Sprintf("segment_%04d.mp4", i/segmentSize))
		segments = append(segments, segmentPath)

		// 导出这一段
		if err := e.exportSegment(ffmpegPath, codec, preset, i, end, segmentPath); err != nil {
			return fmt.Errorf("导出段 %d-%d 失败: %w", i, end, err)
		}

		progress := float64(end) / float64(len(e.cameraFrames)) * 100
		fmt.Printf("进度: %.1f%% (%d/%d 帧)\n", progress, end, len(e.cameraFrames))
	}

	// 合并所有段
	fmt.Println("合并视频段...")
	if err := e.concatenateSegments(ffmpegPath, segments); err != nil {
		return fmt.Errorf("合并段失败: %w", err)
	}

	fmt.Printf("✓ 分段导出完成: %s\n", e.config.OutputPath)
	return nil
}

// exportSegment 导出单个视频段
func (e *GPUExporter) exportSegment(ffmpegPath, codec, preset string, startFrame, endFrame int, outputPath string) error {
	// 计算时间范围
	startTime := float64(e.cameraFrames[startFrame].Timestamp) / 1000.0
	duration := float64(e.cameraFrames[endFrame-1].Timestamp-e.cameraFrames[startFrame].Timestamp) / 1000.0

	// 计算此段的平均缩放和位置
	avgZoom := 0.0
	avgX := 0.0
	avgY := 0.0
	for i := startFrame; i < endFrame; i++ {
		avgZoom += e.cameraFrames[i].Zoom
		avgX += e.cameraFrames[i].X
		avgY += e.cameraFrames[i].Y
	}
	count := float64(endFrame - startFrame)
	avgZoom /= count
	avgX /= count
	avgY /= count

	// 构建命令
	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-t", fmt.Sprintf("%.3f", duration),
		"-i", e.config.VideoPath,
	}

	// 应用滤镜
	cropW := int(float64(e.config.ScreenWidth) / avgZoom)
	cropH := int(float64(e.config.ScreenHeight) / avgZoom)
	cropX := int(avgX - float64(cropW)/2)
	cropY := int(avgY - float64(cropH)/2)

	// 限制裁剪区域
	if cropX < 0 {
		cropX = 0
	}
	if cropY < 0 {
		cropY = 0
	}
	if cropX+cropW > e.config.ScreenWidth {
		cropX = e.config.ScreenWidth - cropW
	}
	if cropY+cropH > e.config.ScreenHeight {
		cropY = e.config.ScreenHeight - cropH
	}

	filterComplex := fmt.Sprintf("crop=%d:%d:%d:%d,scale=%d:%d",
		cropW, cropH, cropX, cropY,
		e.config.ScreenWidth, e.config.ScreenHeight)

	args = append(args, "-filter_complex", filterComplex)

	// 编码器设置
	args = append(args, "-c:v", codec)
	args = append(args, "-preset", preset)
	if codec == "libx264" {
		args = append(args, "-crf", "23")
	}
	args = append(args, "-r", fmt.Sprintf("%d", e.config.FPS))
	args = append(args, "-pix_fmt", "yuv420p")
	args = append(args, "-y")
	args = append(args, outputPath)

	// 执行
	cmd := exec.Command(ffmpegPath, args...)
	return cmd.Run()
}

// concatenateSegments 合并视频段
func (e *GPUExporter) concatenateSegments(ffmpegPath string, segments []string) error {
	// 创建合并列表文件
	listPath := filepath.Join(filepath.Dir(segments[0]), "concat_list.txt")
	listContent := ""
	for _, seg := range segments {
		listContent += fmt.Sprintf("file '%s'\n", filepath.Base(seg))
	}

	if err := os.WriteFile(listPath, []byte(listContent), 0644); err != nil {
		return err
	}

	// 使用 concat demuxer 合并
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy", // 直接复制，不重新编码
		"-y",
		e.config.OutputPath,
	}

	cmd := exec.Command(ffmpegPath, args...)
	return cmd.Run()
}

// Stop 停止导出
func (e *GPUExporter) Stop() error {
	if !e.isExporting {
		return nil
	}

	if e.cmd != nil && e.cmd.Process != nil {
		if err := e.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	e.isExporting = false
	return nil
}

// GetProgress 获取导出进度（估算）
func (e *GPUExporter) GetProgress() float64 {
	// 这里需要解析 FFmpeg 输出来获取真实进度
	// 简化处理，返回 0
	return 0.0
}

// IsExporting 检查是否正在导出
func (e *GPUExporter) IsExporting() bool {
	return e.isExporting
}
