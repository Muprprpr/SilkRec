package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// AudioRecorder 音频录制器
// 使用 FFmpeg 录制系统音频和麦克风
type AudioRecorder struct {
	systemAudioPath string
	micAudioPath    string
	outputPath      string

	systemCmd *exec.Cmd
	micCmd    *exec.Cmd

	isRecording bool
	isPaused    bool
	mu          sync.Mutex

	ffmpegPath string
}

// AudioConfig 音频配置
type AudioConfig struct {
	RecordSystemAudio bool   // 是否录制系统音频
	RecordMicrophone  bool   // 是否录制麦克风
	SystemDevice      string // 系统音频设备名称
	MicDevice         string // 麦克风设备名称
	SampleRate        int    // 采样率 (默认 48000)
	Channels          int    // 声道数 (默认 2)
	Bitrate           string // 比特率 (默认 192k)
}

// DefaultAudioConfig 默认音频配置
func DefaultAudioConfig() AudioConfig {
	return AudioConfig{
		RecordSystemAudio: true,
		RecordMicrophone:  true,
		SystemDevice:      "", // 空字符串表示默认设备
		MicDevice:         "", // 空字符串表示默认设备
		SampleRate:        48000,
		Channels:          2,
		Bitrate:           "192k",
	}
}

// NewAudioRecorder 创建音频录制器
func NewAudioRecorder(ffmpegPath string, config AudioConfig) *AudioRecorder {
	basePath := "output/audio_" + time.Now().Format("20060102_150405")

	return &AudioRecorder{
		systemAudioPath: basePath + "_system.wav",
		micAudioPath:    basePath + "_mic.wav",
		outputPath:      basePath + "_merged.wav",
		ffmpegPath:      ffmpegPath,
		isRecording:     false,
		isPaused:        false,
	}
}

// StartRecording 开始录制音频
func (a *AudioRecorder) StartRecording(config AudioConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRecording {
		return fmt.Errorf("音频录制已在进行中")
	}

	// 确保输出目录存在
	os.MkdirAll(filepath.Dir(a.systemAudioPath), 0755)

	var err error

	// 录制系统音频
	if config.RecordSystemAudio {
		a.systemCmd, err = a.startSystemAudioRecording(config)
		if err != nil {
			return fmt.Errorf("启动系统音频录制失败: %w", err)
		}
		fmt.Println("✓ 系统音频录制已启动")
	}

	// 录制麦克风
	if config.RecordMicrophone {
		a.micCmd, err = a.startMicrophoneRecording(config)
		if err != nil {
			// 如果麦克风失败，停止系统音频
			if a.systemCmd != nil {
				a.systemCmd.Process.Kill()
			}
			return fmt.Errorf("启动麦克风录制失败: %w", err)
		}
		fmt.Println("✓ 麦克风录制已启动")
	}

	a.isRecording = true
	return nil
}

// startSystemAudioRecording 启动系统音频录制
func (a *AudioRecorder) startSystemAudioRecording(config AudioConfig) (*exec.Cmd, error) {
	// Windows: 使用 dshow 捕获音频
	// 自动检测音频设备
	deviceName := config.SystemDevice
	if deviceName == "" {
		deviceName = "audio=Stereo Mix" // Windows 默认立体声混音
	}

	args := []string{
		"-f", "dshow",
		"-i", deviceName,
		"-acodec", "pcm_s16le",
		"-ar", fmt.Sprintf("%d", config.SampleRate),
		"-ac", fmt.Sprintf("%d", config.Channels),
		"-y",
		a.systemAudioPath,
	}

	cmd := exec.Command(a.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// startMicrophoneRecording 启动麦克风录制
func (a *AudioRecorder) startMicrophoneRecording(config AudioConfig) (*exec.Cmd, error) {
	// Windows: 使用 dshow 捕获麦克风
	deviceName := config.MicDevice
	if deviceName == "" {
		deviceName = "audio=Microphone" // Windows 默认麦克风
	}

	args := []string{
		"-f", "dshow",
		"-i", deviceName,
		"-acodec", "pcm_s16le",
		"-ar", fmt.Sprintf("%d", config.SampleRate),
		"-ac", fmt.Sprintf("%d", config.Channels),
		"-y",
		a.micAudioPath,
	}

	cmd := exec.Command(a.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// StopRecording 停止录制并合并音频
func (a *AudioRecorder) StopRecording() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRecording {
		return "", fmt.Errorf("音频录制未在进行中")
	}

	// 停止系统音频
	if a.systemCmd != nil && a.systemCmd.Process != nil {
		a.systemCmd.Process.Kill()
		a.systemCmd.Wait()
		fmt.Println("✓ 系统音频录制已停止")
	}

	// 停止麦克风
	if a.micCmd != nil && a.micCmd.Process != nil {
		a.micCmd.Process.Kill()
		a.micCmd.Wait()
		fmt.Println("✓ 麦克风录制已停止")
	}

	a.isRecording = false

	// 合并音频文件
	mergedPath, err := a.mergeAudioFiles()
	if err != nil {
		return "", fmt.Errorf("合并音频失败: %w", err)
	}

	return mergedPath, nil
}

// mergeAudioFiles 合并系统音频和麦克风音频
func (a *AudioRecorder) mergeAudioFiles() (string, error) {
	// 检查文件是否存在
	systemExists := fileExists(a.systemAudioPath)
	micExists := fileExists(a.micAudioPath)

	// 如果只有一个音频源，直接返回
	if systemExists && !micExists {
		return a.systemAudioPath, nil
	}
	if !systemExists && micExists {
		return a.micAudioPath, nil
	}
	if !systemExists && !micExists {
		return "", fmt.Errorf("没有录制到音频")
	}

	// 合并两个音频流
	// 使用 amix 滤镜混合音频
	args := []string{
		"-i", a.systemAudioPath,
		"-i", a.micAudioPath,
		"-filter_complex", "amix=inputs=2:duration=longest:dropout_transition=2",
		"-acodec", "pcm_s16le",
		"-y",
		a.outputPath,
	}

	cmd := exec.Command(a.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("合并音频失败: %w, 输出: %s", err, string(output))
	}

	fmt.Printf("✓ 音频已合并: %s\n", a.outputPath)
	return a.outputPath, nil
}

// PauseRecording 暂停录制（通过停止进程实现）
// 注意：FFmpeg 不支持真正的暂停，这里使用停止/重启模拟
func (a *AudioRecorder) PauseRecording() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRecording || a.isPaused {
		return fmt.Errorf("无法暂停")
	}

	// 这里可以实现更复杂的暂停逻辑
	// 简化处理：标记为暂停
	a.isPaused = true
	return nil
}

// ResumeRecording 恢复录制
func (a *AudioRecorder) ResumeRecording() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isPaused {
		return fmt.Errorf("未暂停")
	}

	a.isPaused = false
	return nil
}

// GetSystemAudioPath 获取系统音频文件路径
func (a *AudioRecorder) GetSystemAudioPath() string {
	return a.systemAudioPath
}

// GetMicAudioPath 获取麦克风音频文件路径
func (a *AudioRecorder) GetMicAudioPath() string {
	return a.micAudioPath
}

// GetMergedAudioPath 获取合并后的音频文件路径
func (a *AudioRecorder) GetMergedAudioPath() string {
	return a.outputPath
}

// IsRecording 检查是否正在录制
func (a *AudioRecorder) IsRecording() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isRecording
}

// IsPaused 检查是否暂停
func (a *AudioRecorder) IsPaused() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isPaused
}

// ListAudioDevices 列出可用的音频设备（Windows dshow）
func ListAudioDevices(ffmpegPath string) ([]string, error) {
	// 使用 FFmpeg 列出 dshow 设备
	args := []string{
		"-list_devices", "true",
		"-f", "dshow",
		"-i", "dummy",
	}

	cmd := exec.Command(ffmpegPath, args...)
	output, _ := cmd.CombinedOutput()

	// 解析输出（简化处理）
	devices := []string{}
	// 这里可以解析 FFmpeg 输出，提取设备列表
	// 实际实现需要解析 FFmpeg 的输出格式

	fmt.Println("可用音频设备:")
	fmt.Println(string(output))

	return devices, nil
}

// MergeAudioWithVideo 将音频和视频合并
func MergeAudioWithVideo(ffmpegPath, videoPath, audioPath, outputPath string) error {
	args := []string{
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy", // 视频流不重编码
		"-c:a", "aac", // 音频编码为 AAC
		"-b:a", "192k", // 音频比特率
		"-strict", "experimental",
		"-y",
		outputPath,
	}

	cmd := exec.Command(ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("合并音视频失败: %w, 输出: %s", err, string(output))
	}

	fmt.Printf("✓ 音视频已合并: %s\n", outputPath)
	return nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
