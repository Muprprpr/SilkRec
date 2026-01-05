package recorder

import (
	"fmt"
	"os/exec"
	"strings"
)

// CaptureConfig 屏幕捕获配置
type CaptureConfig struct {
	OutputPath string // 输出文件路径
	FrameRate  int    // 帧率（默认 60）
	Codec      string // 编码器（h264_nvenc 或 libx264）
	Quality    int    // 质量参数
	Preset     string // 编码预设
}

// DefaultCaptureConfig 返回默认的捕获配置
func DefaultCaptureConfig(outputPath string) CaptureConfig {
	return CaptureConfig{
		OutputPath: outputPath,
		FrameRate:  60,
		Codec:      "", // 将自动检测
		Quality:    20, // qp 值，越小质量越高
		Preset:     "", // 将自动选择
	}
}

// BuildDDAGrabCommand 构建 ddagrap 命令（推荐，使用 lavfi 滤镜）
// ddagrab 是 Windows 10/11 的高性能屏幕捕获方法
func BuildDDAGrabCommand(ffmpegPath string, config CaptureConfig) []string {
	// 基础命令
	args := []string{
		"-y", // 覆盖输出文件
		"-f", "lavfi",
		"-i", fmt.Sprintf("ddagrab=framerate=%d:draw_mouse=0", config.FrameRate),
	}

	// 添加编码器参数
	args = append(args, buildEncoderArgs(config)...)

	// 输出文件
	args = append(args, config.OutputPath)

	return args
}

// BuildGDIRABCommand 构建 gdigrab 命令（回退方案）
// gdigrab 是 Windows 的传统屏幕捕获方法
func BuildGDIRABCommand(ffmpegPath string, config CaptureConfig) []string {
	// 基础命令
	args := []string{
		"-y", // 覆盖输出文件
		"-f", "gdigrab",
		"-framerate", fmt.Sprintf("%d", config.FrameRate),
		"-i", "desktop", // 捕获整个桌面
	}

	// 添加编码器参数
	args = append(args, buildEncoderArgs(config)...)

	// 输出文件
	args = append(args, config.OutputPath)

	return args
}

// buildEncoderArgs 构建编码器参数
func buildEncoderArgs(config CaptureConfig) []string {
	var args []string

	switch config.Codec {
	case "h264_nvenc":
		// NVIDIA NVENC 编码器参数
		args = []string{
			"-c:v", "h264_nvenc",
			"-qp", fmt.Sprintf("%d", config.Quality),
			"-preset", config.Preset,
		}
	case "libx264":
		// libx264 编码器参数
		args = []string{
			"-c:v", "libx264",
			"-preset", config.Preset,
			"-crf", fmt.Sprintf("%d", config.Quality),
		}
	default:
		// 默认使用 libx264
		args = []string{
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-crf", "23",
		}
	}

	return args
}

// DetectBestCodec 检测最佳编码器
// 优先级: h264_nvenc > libx264
func DetectBestCodec(ffmpegPath string) (string, error) {
	// 1. 尝试检测 h264_nvenc
	cmd := exec.Command(ffmpegPath, "-encoders")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "h264_nvenc") {
		return "h264_nvenc", nil
	}

	// 2. 回退到 libx264
	cmd = exec.Command(ffmpegPath, "-encoders")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "libx264") {
		return "libx264", nil
	}

	return "", fmt.Errorf("未找到可用的视频编码器")
}

// GetBestPreset 获取编码器的最佳预设
func GetBestPreset(codec string) string {
	switch codec {
	case "h264_nvenc":
		// NVIDIA 编码器预设: p1 (fastest) - p7 (slowest, best quality)
		return "p4" // 平衡速度和质量
	case "libx264":
		// libx264 预设: ultrafast - veryslow
		return "ultrafast" // 快速编码
	default:
		return ""
	}
}

// BuildExportCommand 构建导出命令（从 stdin 接收图像数据）
func BuildExportCommand(ffmpegPath string, outputPath string, frameRate int) []string {
	return []string{
		"-y", // 覆盖输出文件
		"-f", "image2pipe",
		"-framerate", fmt.Sprintf("%d", frameRate),
		"-i", "-", // 从 stdin 读取
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "20",
		outputPath,
	}
}
