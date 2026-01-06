package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

// FFmpegManager 管理 FFmpeg 可执行文件路径
type FFmpegManager struct {
	ctx        context.Context
	ffmpegPath string
}

// NewFFmpegManager 创建 FFmpeg 管理器
func NewFFmpegManager(ctx context.Context) *FFmpegManager {
	return &FFmpegManager{
		ctx: ctx,
	}
}

// GetFFmpegPath 获取 FFmpeg 可执行文件路径
func (m *FFmpegManager) GetFFmpegPath() (string, error) {
	// 如果已经找到路径，直接返回
	if m.ffmpegPath != "" {
		return m.ffmpegPath, nil
	}

	// 按优先级查找 FFmpeg
	paths := []string{
		m.getDevPath(),     // 开发环境路径
		m.getProdPath(),    // 生产环境路径
		m.getAppDataPath(), // 应用数据目录（可选）
	}

	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			m.ffmpegPath = path
			fmt.Printf("找到 FFmpeg: %s\n", path)
			return path, nil
		}
	}

	return "", errors.New("未找到 FFmpeg 可执行文件")
}

// getDevPath 获取开发环境路径
func (m *FFmpegManager) getDevPath() string {
	// 项目根目录下的 ffmpeg/ffmpeg.exe - 使用绝对路径
	if absPath, err := filepath.Abs(filepath.Join(".", "ffmpeg", "ffmpeg.exe")); err == nil {
		return absPath
	}
	return filepath.Join(".", "ffmpeg", "ffmpeg.exe")
}

// getProdPath 获取生产环境路径
func (m *FFmpegManager) getProdPath() string {
	// 可执行文件同级目录的 ffmpeg.exe
	if m.ctx != nil {
		exePath, err := os.Executable()
		if err == nil {
			return filepath.Join(filepath.Dir(exePath), "ffmpeg.exe")
		}
	}
	return ""
}

// getAppDataPath 获取应用数据目录路径
func (m *FFmpegManager) getAppDataPath() string {
	// 使用环境变量获取 AppData 目录
	appDataDir := os.Getenv("LOCALAPPDATA")
	if appDataDir != "" {
		return filepath.Join(appDataDir, "SilkRec", "ffmpeg.exe")
	}
	return ""
}

// CheckFFmpegAvailable 检查 FFmpeg 是否可用
func (m *FFmpegManager) CheckFFmpegAvailable() bool {
	path, err := m.GetFFmpegPath()
	if err != nil {
		return false
	}

	// 检查文件是否存在且可执行
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Windows 下检查 .exe 扩展名
	if goruntime.GOOS == "windows" && filepath.Ext(path) != ".exe" {
		return false
	}

	// 检查是否为常规文件
	return !info.IsDir()
}

// GetFFmpegVersion 获取 FFmpeg 版本
func (m *FFmpegManager) GetFFmpegVersion() (string, error) {
	path, err := m.GetFFmpegPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 解析版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", errors.New("无法解析 FFmpeg 版本")
}

// CheckEncoderAvailable 检查指定的编码器是否可用
func (m *FFmpegManager) CheckEncoderAvailable(encoder string) bool {
	path, err := m.GetFFmpegPath()
	if err != nil {
		return false
	}

	cmd := exec.Command(path, "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), encoder)
}

// GetBestEncoder 获取最佳可用编码器
// 优先级: h264_nvenc > libx264
func (m *FFmpegManager) GetBestEncoder() (string, error) {
	// 1. 尝试检测 h264_nvenc (NVIDIA GPU 加速)
	if m.CheckEncoderAvailable("h264_nvenc") {
		return "h264_nvenc", nil
	}

	// 2. 回退到 libx264 (CPU 编码)
	if m.CheckEncoderAvailable("libx264") {
		return "libx264", nil
	}

	return "", errors.New("未找到可用的视频编码器 (需要 h264_nvenc 或 libx264)")
}

// GetBestPreset 获取编码器的最佳预设
func (m *FFmpegManager) GetBestPreset(encoder string) string {
	switch encoder {
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
