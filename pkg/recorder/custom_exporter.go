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
)

// CustomExportParams 自定义导出参数
type CustomExportParams struct {
	// 动画参数
	Smoothness      float64 `json:"smoothness"`      // 平滑强度 (0.01-0.5)
	ZoomLevel       float64 `json:"zoomLevel"`       // 缩放倍数 (1.0-3.0)
	Speed           float64 `json:"speed"`           // 运动速度 (0.5-2.0)
	VideoScale      float64 `json:"videoScale"`      // 视频画面大小 (0.8-1.0)
	CursorSize      int     `json:"cursorSize"`      // 光标大小 (16-64)
	ShowClickEffect bool    `json:"showClickEffect"` // 显示点击效果
}

// BackgroundParams 背景参数
type BackgroundParams struct {
	Type           string `json:"backgroundType"`  // 'solid', 'gradient', 'image'
	Color          string `json:"backgroundColor"` // 纯色背景颜色
	GradientColor1 string `json:"gradientColor1"`  // 渐变颜色1
	GradientColor2 string `json:"gradientColor2"`  // 渐变颜色2
	ImagePath      string `json:"backgroundImage"` // 背景图片路径（base64 或文件路径）
}

// CustomExporter 自定义参数导出器
type CustomExporter struct {
	ffmpegManager *ffmpeg.FFmpegManager
	config        ExportConfig
	customParams  CustomExportParams
	bgParams      BackgroundParams
	cursorImage   string // 光标图片（base64 或文件路径）
	mouseEvents   []hook.MouseEvent
	cameraFrames  []CameraFrame
	cmd           *exec.Cmd
	isExporting   bool
}

// NewCustomExporter 创建自定义导出器
func NewCustomExporter(ffmpegManager *ffmpeg.FFmpegManager) *CustomExporter {
	return &CustomExporter{
		ffmpegManager: ffmpegManager,
		isExporting:   false,
		customParams: CustomExportParams{
			Smoothness:      0.15,
			ZoomLevel:       1.5,
			Speed:           1.0,
			VideoScale:      1.0,
			CursorSize:      32,
			ShowClickEffect: true,
		},
	}
}

// PrepareCustomExport 准备自定义导出
func (e *CustomExporter) PrepareCustomExport(
	config ExportConfig,
	customParamsJSON string,
	bgParamsJSON string,
	cursorImage string,
) error {
	e.config = config
	e.cursorImage = cursorImage

	// 解析自定义参数
	if customParamsJSON != "" {
		if err := json.Unmarshal([]byte(customParamsJSON), &e.customParams); err != nil {
			return fmt.Errorf("解析自定义参数失败: %w", err)
		}
	}

	// 解析背景参数
	if bgParamsJSON != "" {
		if err := json.Unmarshal([]byte(bgParamsJSON), &e.bgParams); err != nil {
			return fmt.Errorf("解析背景参数失败: %w", err)
		}
	}

	// 加载鼠标数据
	data, err := os.ReadFile(config.MouseDataPath)
	if err != nil {
		return fmt.Errorf("加载鼠标数据失败: %w", err)
	}

	if err := json.Unmarshal(data, &e.mouseEvents); err != nil {
		return fmt.Errorf("解析鼠标数据失败: %w", err)
	}

	// 使用自定义参数生成相机路径
	e.cameraFrames = e.generateCustomCameraPath()

	fmt.Printf("✓ 自定义导出准备完成\n")
	fmt.Printf("  平滑强度: %.2f\n", e.customParams.Smoothness)
	fmt.Printf("  缩放倍数: %.2fx\n", e.customParams.ZoomLevel)
	fmt.Printf("  视频缩放: %.0f%%\n", e.customParams.VideoScale*100)
	fmt.Printf("  生成相机帧: %d\n", len(e.cameraFrames))

	return nil
}

// generateCustomCameraPath 使用自定义参数生成相机路径
func (e *CustomExporter) generateCustomCameraPath() []CameraFrame {
	controller := NewCameraController(
		e.config.ScreenWidth,
		e.config.ScreenHeight,
		e.customParams.Smoothness, // 使用自定义平滑度
		e.customParams.ZoomLevel,  // 使用自定义缩放
	)

	frames := []CameraFrame{}
	frameDuration := 1000.0 / float64(e.config.FPS)
	totalDuration := 0.0

	if len(e.mouseEvents) > 0 {
		totalDuration = float64(e.mouseEvents[len(e.mouseEvents)-1].Timestamp)
	}

	lastEventIndex := 0

	for t := 0.0; t < totalDuration; t += frameDuration {
		currentTime := int64(t)

		// 应用当前时间点的鼠标事件
		for lastEventIndex < len(e.mouseEvents) && e.mouseEvents[lastEventIndex].Timestamp <= currentTime {
			controller.Update(e.mouseEvents[lastEventIndex])
			lastEventIndex++
		}

		// 生成相机帧
		frame := controller.GetCurrentFrame(currentTime)

		// 应用速度调整（通过调整平滑度实现）
		if e.customParams.Speed != 1.0 {
			// 速度越大，跟随越快
			speedFactor := e.customParams.Speed
			frame.X = frame.X * speedFactor
			frame.Y = frame.Y * speedFactor
		}

		frames = append(frames, frame)
	}

	return frames
}

// ExportWithCustomParams 使用自定义参数导出
func (e *CustomExporter) ExportWithCustomParams() error {
	if e.isExporting {
		return fmt.Errorf("导出已在进行中")
	}

	ffmpegPath, err := e.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	codec, err := e.ffmpegManager.GetBestEncoder()
	if err != nil {
		codec = "libx264"
	}
	preset := e.ffmpegManager.GetBestPreset(codec)

	// 构建 FFmpeg 命令
	args := e.buildCustomExportCommand(ffmpegPath, codec, preset)

	fmt.Printf("执行自定义导出命令:\n%s %s\n", ffmpegPath, strings.Join(args, " "))

	e.cmd = exec.Command(ffmpegPath, args...)
	e.cmd.Stdout = os.Stdout
	e.cmd.Stderr = os.Stderr

	e.isExporting = true

	if err := e.cmd.Run(); err != nil {
		e.isExporting = false
		return fmt.Errorf("FFmpeg 执行失败: %w", err)
	}

	e.isExporting = false
	fmt.Printf("✓ 自定义导出完成: %s\n", e.config.OutputPath)
	return nil
}

// buildCustomExportCommand 构建自定义导出命令
func (e *CustomExporter) buildCustomExportCommand(ffmpegPath, codec, preset string) []string {
	args := []string{}

	// 硬件加速
	if strings.Contains(codec, "nvenc") {
		args = append(args, "-hwaccel", "cuda")
		args = append(args, "-hwaccel_output_format", "cuda")
	}

	// 输入视频
	args = append(args, "-i", e.config.VideoPath)

	// 构建复杂滤镜链
	filterComplex := e.buildCustomFilterComplex()
	if filterComplex != "" {
		args = append(args, "-filter_complex", filterComplex)
	}

	// 编码器设置
	args = append(args, "-c:v", codec)
	args = append(args, "-preset", preset)

	if codec == "libx264" {
		args = append(args, "-crf", "23")
	} else if strings.Contains(codec, "nvenc") {
		args = append(args, "-cq", "23")
		args = append(args, "-b:v", "5M")
	}

	args = append(args, "-r", fmt.Sprintf("%d", e.config.FPS))
	args = append(args, "-pix_fmt", "yuv420p")
	args = append(args, "-y")
	args = append(args, e.config.OutputPath)

	return args
}

// buildCustomFilterComplex 构建自定义滤镜链
func (e *CustomExporter) buildCustomFilterComplex() string {
	filters := []string{}

	// 1. 生成背景
	bgFilter := e.generateBackgroundFilter()
	if bgFilter != "" {
		filters = append(filters, bgFilter)
	}

	// 2. 缩放视频内容
	videoScale := e.customParams.VideoScale
	if videoScale != 1.0 {
		scaledWidth := int(float64(e.config.ScreenWidth) * videoScale)
		scaledHeight := int(float64(e.config.ScreenHeight) * videoScale)
		filters = append(filters, fmt.Sprintf(
			"[0:v]scale=%d:%d[scaled]",
			scaledWidth, scaledHeight,
		))
	} else {
		filters = append(filters, "[0:v]copy[scaled]")
	}

	// 3. 应用相机变换（crop + scale）
	// 这里简化处理，实际需要根据相机帧动态生成
	avgX, avgY, avgZoom := e.calculateAverageCameraParams()

	cropW := int(float64(e.config.ScreenWidth) / avgZoom)
	cropH := int(float64(e.config.ScreenHeight) / avgZoom)
	cropX := int(avgX - float64(cropW)/2)
	cropY := int(avgY - float64(cropH)/2)

	// 限制边界
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

	filters = append(filters, fmt.Sprintf(
		"[scaled]crop=%d:%d:%d:%d[cropped]",
		cropW, cropH, cropX, cropY,
	))

	// 4. 缩放回目标分辨率
	filters = append(filters, fmt.Sprintf(
		"[cropped]scale=%d:%d[final]",
		e.config.ScreenWidth, e.config.ScreenHeight,
	))

	// 5. 叠加到背景（如果有）
	if bgFilter != "" {
		offsetX := int(float64(e.config.ScreenWidth) * (1.0 - videoScale) / 2)
		offsetY := int(float64(e.config.ScreenHeight) * (1.0 - videoScale) / 2)

		filters = append(filters, fmt.Sprintf(
			"[bg][final]overlay=%d:%d[output]",
			offsetX, offsetY,
		))
	}

	return strings.Join(filters, ";")
}

// generateBackgroundFilter 生成背景滤镜
func (e *CustomExporter) generateBackgroundFilter() string {
	switch e.bgParams.Type {
	case "solid":
		// 纯色背景
		return fmt.Sprintf(
			"color=c=%s:s=%dx%d[bg]",
			e.bgParams.Color,
			e.config.ScreenWidth,
			e.config.ScreenHeight,
		)

	case "gradient":
		// 渐变背景（使用 FFmpeg 的 color 和 blend）
		return fmt.Sprintf(
			"color=c=%s:s=%dx%d[c1];color=c=%s:s=%dx%d[c2];[c1][c2]blend=all_mode='addition'[bg]",
			e.bgParams.GradientColor1,
			e.config.ScreenWidth,
			e.config.ScreenHeight,
			e.bgParams.GradientColor2,
			e.config.ScreenWidth,
			e.config.ScreenHeight,
		)

	case "image":
		// 图片背景
		if e.bgParams.ImagePath != "" {
			return fmt.Sprintf(
				"movie=%s[bg]",
				e.bgParams.ImagePath,
			)
		}
	}

	return ""
}

// calculateAverageCameraParams 计算平均相机参数
func (e *CustomExporter) calculateAverageCameraParams() (float64, float64, float64) {
	if len(e.cameraFrames) == 0 {
		return float64(e.config.ScreenWidth) / 2,
			float64(e.config.ScreenHeight) / 2,
			1.0
	}

	var sumX, sumY, sumZoom float64
	for _, frame := range e.cameraFrames {
		sumX += frame.X
		sumY += frame.Y
		sumZoom += frame.Zoom
	}

	count := float64(len(e.cameraFrames))
	return sumX / count, sumY / count, sumZoom / count
}

// Stop 停止导出
func (e *CustomExporter) Stop() error {
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

// GetCameraFrames 获取相机帧
func (e *CustomExporter) GetCameraFrames() []CameraFrame {
	return e.cameraFrames
}

// SaveCameraPath 保存相机路径
func (e *CustomExporter) SaveCameraPath(outputPath string) error {
	data, err := json.MarshalIndent(e.cameraFrames, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}
