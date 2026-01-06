package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	stdpath "path/filepath"
	"runtime"
	"time"

	"SmoothScreen/pkg/ffmpeg"
	"SmoothScreen/pkg/hook"
	"SmoothScreen/pkg/io"
	"SmoothScreen/pkg/recorder"
	"SmoothScreen/pkg/server"
	"SmoothScreen/pkg/sys"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	ffmpegManager  *ffmpeg.FFmpegManager
	recorder       *recorder.Recorder
	mouseHook      *hook.MouseHook
	fileWriter     *io.FileWriter
	pipeWriter     *recorder.PipeWriter
	httpPipeServer *recorder.HttpPipeServer
	fileServer     *server.FileServer
	videoWriter    *recorder.VideoWriter
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		fileWriter:  io.NewFileWriter(),
		videoWriter: recorder.NewVideoWriter(),
	}
}

// startup is called when app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化 FFmpeg 管理器
	a.ffmpegManager = ffmpeg.NewFFmpegManager(ctx)

	// 检查 FFmpeg 是否可用
	if !a.ffmpegManager.CheckFFmpegAvailable() {
		wailsruntime.EventsEmit(ctx, "ffmpeg-error", "未找到 FFmpeg 可执行文件，请确保 ffmpeg.exe 在正确位置")
		fmt.Println("警告: 未找到 FFmpeg 可执行文件")
	} else {
		// 获取 FFmpeg 版本
		version, err := a.ffmpegManager.GetFFmpegVersion()
		if err == nil {
			fmt.Printf("FFmpeg 版本: %s\n", version)
		}

		// 检测最佳编码器
		codec, err := a.ffmpegManager.GetBestEncoder()
		if err == nil {
			fmt.Printf("最佳编码器: %s\n", codec)
		} else {
			fmt.Printf("编码器检测失败: %v\n", err)
		}
	}

	// 初始化增强的鼠标钩子
	a.mouseHook = hook.NewMouseHook(a.ctx)
	a.mouseHook.Start()

	// 初始化录制管理器
	a.recorder = recorder.NewRecorder(a.ffmpegManager, a.mouseHook, ctx)

	// 初始化文件服务器（用于提供视频文件访问）
	a.fileServer = server.NewFileServer("output", 8080)
	if err := a.fileServer.Start(); err != nil {
		fmt.Printf("启动文件服务器失败: %v\n", err)
	}

	fmt.Println("SilkRec 应用已启动")
}

// shutdown is called when app shuts down.
func (a *App) shutdown(ctx context.Context) {
	// 停止录制（如果正在录制）
	if a.recorder != nil && a.recorder.IsRecording() {
		a.recorder.StopRecording()
	}

	// 停止 HTTP 管道服务器
	if a.httpPipeServer != nil {
		a.httpPipeServer.Stop()
	}

	// 停止文件服务器
	if a.fileServer != nil {
		a.fileServer.Stop()
	}

	// 停止 H.264 写入器
	if a.videoWriter != nil && a.videoWriter.IsWriting() {
		a.videoWriter.StopWriter()
	}

	if a.mouseHook != nil {
		a.mouseHook.Stop()
	}

	if a.fileWriter != nil && a.fileWriter.IsOpen() {
		a.fileWriter.Close()
	}

	fmt.Println("SilkRec 应用已关闭")
}

// ========== 新的录制相关 API ==========

// CheckFFmpegAvailable 检查 FFmpeg 是否可用
func (a *App) CheckFFmpegAvailable() bool {
	if a.ffmpegManager == nil {
		return false
	}
	return a.ffmpegManager.CheckFFmpegAvailable()
}

// StartScreenRecording 开始屏幕录制
func (a *App) StartScreenRecording(videoPath string) error {
	if a.recorder == nil {
		return fmt.Errorf("录制器未初始化")
	}

	if err := a.recorder.StartRecording(videoPath); err != nil {
		return err
	}

	// 发送录制开始事件
	wailsruntime.EventsEmit(a.ctx, "recording-started", map[string]interface{}{
		"videoPath": videoPath,
		"timestamp": time.Now().Unix(),
	})

	return nil
}

// StopScreenRecording 停止屏幕录制
func (a *App) StopScreenRecording() (string, string, error) {
	if a.recorder == nil {
		return "", "", fmt.Errorf("录制器未初始化")
	}

	videoPath, mouseDataPath, err := a.recorder.StopRecording()
	if err != nil {
		return "", "", err
	}

	// 发送录制停止事件
	wailsruntime.EventsEmit(a.ctx, "recording-stopped", map[string]interface{}{
		"videoPath":     videoPath,
		"mouseDataPath": mouseDataPath,
		"timestamp":     time.Now().Unix(),
	})

	return videoPath, mouseDataPath, nil
}

// GetRecordingStatus 获取录制状态
func (a *App) GetRecordingStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if a.recorder != nil {
		recorderStatus := a.recorder.GetStatus()
		status["isRecording"] = recorderStatus.IsRecording
		status["outputPath"] = recorderStatus.OutputPath
		status["mouseDataPath"] = recorderStatus.MouseDataPath
		status["duration"] = recorderStatus.Duration
		status["mouseEventCount"] = recorderStatus.MouseEventCount
		status["ffmpegPID"] = recorderStatus.FFmpegPID
	}

	if a.ffmpegManager != nil {
		status["ffmpegAvailable"] = a.ffmpegManager.CheckFFmpegAvailable()
	}

	return status
}

// GetMouseData 获取录制的鼠标数据
func (a *App) GetMouseData() string {
	if a.recorder == nil {
		return "[]"
	}

	data, err := a.recorder.GetMouseDataJSON()
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}

	return data
}

// SaveMouseData 保存鼠标数据到JSON文件
func (a *App) SaveMouseData(filename string) error {
	if a.recorder == nil {
		return fmt.Errorf("录制器未初始化")
	}

	return a.recorder.SaveMouseData(filename)
}

// ========== 旧的录制相关 API（保留兼容性） ==========

// StartRecording 开始录制（旧版 API，保留兼容性）
func (a *App) StartRecording(filename string) error {
	if a.mouseHook == nil {
		return fmt.Errorf("鼠标钩子未初始化")
	}

	// 开始录制鼠标数据
	a.mouseHook.StartRecording()

	// 打开视频文件准备写入
	if err := a.fileWriter.Open(filename); err != nil {
		return fmt.Errorf("打开视频文件失败: %w", err)
	}

	fmt.Printf("开始录制: %s\n", filename)
	return nil
}

// StopRecording 停止录制（旧版 API，保留兼容性）
func (a *App) StopRecording() error {
	if a.mouseHook == nil {
		return fmt.Errorf("鼠标钩子未初始化")
	}

	// 停止录制鼠标数据
	a.mouseHook.StopRecording()

	// 关闭视频文件
	if a.fileWriter.IsOpen() {
		if err := a.fileWriter.Close(); err != nil {
			return fmt.Errorf("关闭视频文件失败: %w", err)
		}
	}

	fmt.Println("录制已停止")
	return nil
}

// WriteVideoChunk 写入视频数据块（旧版 API，保留兼容性）
func (a *App) WriteVideoChunk(data []byte) (int, error) {
	if a.fileWriter == nil || !a.fileWriter.IsOpen() {
		return 0, fmt.Errorf("文件未打开")
	}

	return a.fileWriter.Write(data)
}

// ========== 系统相关 API ==========

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetScreenInfo 获取屏幕信息
func (a *App) GetScreenInfo() (int, int, int) {
	info, err := sys.GetScreenInfo()
	if err != nil {
		fmt.Printf("获取屏幕信息失败: %v\n", err)
		return 1920, 1080, 96 // 返回默认值
	}

	return info.Width, info.Height, info.DPI
}

// SetWindowAlwaysOnTop 设置窗口置顶
func (a *App) SetWindowAlwaysOnTop(top bool) error {
	// 这里需要获取当前窗口的句柄
	// 由于Wails窗口句柄获取比较复杂，这里先返回成功
	// 实际实现可能需要通过runtime获取窗口句柄
	fmt.Printf("设置窗口置顶: %v\n", top)
	return nil
}

// ========== 文件操作 API ==========

// ReadVideoFile 读取视频文件并返回base64编码
func (a *App) ReadVideoFile(filepath string) (string, error) {
	data, err := a.fileWriter.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("读取视频文件失败: %w", err)
	}

	// 将二进制数据转换为base64编码
	base64Data := base64.StdEncoding.EncodeToString(data)
	return base64Data, nil
}

// ========== 导出相关 API ==========

// StartExport 开始导出视频
func (a *App) StartExport(outputPath string, frameRate int) error {
	if a.pipeWriter == nil {
		a.pipeWriter = recorder.NewPipeWriter(a.ffmpegManager)
	}

	return a.pipeWriter.StartExport(outputPath, frameRate)
}

// WriteExportFrame 写入导出帧
func (a *App) WriteExportFrame(base64Data string) error {
	if a.pipeWriter == nil {
		return fmt.Errorf("导出未启动")
	}

	return a.pipeWriter.WriteFrame(base64Data)
}

// FinishExport 完成导出
func (a *App) FinishExport() error {
	if a.pipeWriter == nil {
		return fmt.Errorf("导出未启动")
	}

	return a.pipeWriter.FinishExport()
}

// StopExport 停止导出
func (a *App) StopExport() error {
	if a.pipeWriter == nil {
		return fmt.Errorf("导出未启动")
	}

	return a.pipeWriter.StopExport()
}

// GetExportStatus 获取导出状态
func (a *App) GetExportStatus() map[string]interface{} {
	if a.pipeWriter == nil {
		return map[string]interface{}{
			"isWriting":   false,
			"totalFrames": 0,
			"outputPath":  "",
		}
	}

	return a.pipeWriter.GetStatus()
}

// ========== HTTP 管道服务器 API ==========

// StartHttpPipeServer 启动 HTTP 管道服务器
func (a *App) StartHttpPipeServer(outputPath string, width int, height int, frameRate int) map[string]interface{} {
	result := make(map[string]interface{})

	if a.httpPipeServer == nil {
		a.httpPipeServer = recorder.NewHttpPipeServer()
	}

	// 获取 FFmpeg 路径
	ffmpegPath, err := a.ffmpegManager.GetFFmpegPath()
	if err != nil {
		result["success"] = false
		result["error"] = fmt.Sprintf("获取 FFmpeg 路径失败: %v", err)
		return result
	}

	err = a.httpPipeServer.Start(ffmpegPath, outputPath, frameRate, width, height)
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result
	}

	result["success"] = true
	result["port"] = a.httpPipeServer.GetPort()
	return result
}

// StopHttpPipeServer 停止 HTTP 管道服务器
func (a *App) StopHttpPipeServer() error {
	if a.httpPipeServer == nil {
		return fmt.Errorf("HTTP 管道服务器未启动")
	}

	return a.httpPipeServer.Stop()
}

// GetHttpPipeServerStatus 获取 HTTP 管道服务器状态
func (a *App) GetHttpPipeServerStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if a.httpPipeServer == nil {
		status["running"] = false
		status["port"] = 0
		status["ffmpegPath"] = ""
		return status
	}

	serverStatus := a.httpPipeServer.GetStatus()
	status["running"] = serverStatus["isRunning"]
	status["port"] = serverStatus["port"]
	status["ffmpegPath"] = serverStatus["ffmpegPath"]

	return status
}

// GetFileURL 获取文件的 URL（通过文件服务器）
func (a *App) GetFileURL(filePath string) string {
	if a.fileServer == nil {
		return ""
	}
	// 提取文件名（移除路径前缀）
	filename := stdpath.Base(filePath)
	return a.fileServer.GetURL(filename)
}

// ========== WebCodecs 导出 API ==========

// StartH264Writer 启动 H.264 视频写入器（用于接收 WebCodecs 编码的 H.264 数据）
func (a *App) StartH264Writer(outputPath string) error {
	if a.videoWriter == nil {
		a.videoWriter = recorder.NewVideoWriter()
	}

	return a.videoWriter.StartWriter(outputPath)
}

// WriteH264Chunk 写入 H.264 数据块（从 WebCodecs 接收）
func (a *App) WriteH264Chunk(data []byte) error {
	if a.videoWriter == nil {
		return fmt.Errorf("H.264 写入器未初始化")
	}

	return a.videoWriter.WriteChunk(data)
}

// StopH264Writer 停止 H.264 写入器
func (a *App) StopH264Writer() error {
	if a.videoWriter == nil {
		return fmt.Errorf("H.264 写入器未初始化")
	}

	return a.videoWriter.StopWriter()
}

// GetH264WriterStatus 获取 H.264 写入器状态
func (a *App) GetH264WriterStatus() map[string]interface{} {
	if a.videoWriter == nil {
		return map[string]interface{}{
			"isWriting":  false,
			"totalBytes": 0,
			"outputPath": "",
		}
	}

	return a.videoWriter.GetStatus()
}

// FinalizeH264Export 最终化 H.264 导出（使用 FFmpeg 封装为 MP4）
func (a *App) FinalizeH264Export(h264Path string, audioPath string, outputPath string) error {
	if a.ffmpegManager == nil {
		return fmt.Errorf("FFmpeg 管理器未初始化")
	}

	ffmpegPath, err := a.ffmpegManager.GetFFmpegPath()
	if err != nil {
		return fmt.Errorf("获取 FFmpeg 路径失败: %w", err)
	}

	// 构建 FFmpeg 命令
	var args []string

	// 输入 H.264 文件
	args = append(args, "-i", h264Path)

	// 如果有音频，添加音频输入
	if audioPath != "" {
		args = append(args, "-i", audioPath)
	}

	// 使用 copy 模式快速封装（无需重新编码）
	args = append(args, "-c", "copy")

	// 输出文件
	args = append(args, outputPath)

	// 执行 FFmpeg 命令
	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("执行 FFmpeg 封装命令: %s %s\n", ffmpegPath, args)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 封装失败: %w", err)
	}

	fmt.Printf("视频封装完成: %s\n", outputPath)
	return nil
}

// MuxH264ToMp4 将 H.264 裸流封装为 MP4
// 参数:
//
//    inputPath: 临时生成的 .h264 文件路径
//    outputPath: 最终的 .mp4 输出路径
func (a *App) MuxH264ToMp4(inputPath string, outputPath string) error {
	// 1. 确定 FFmpeg 可执行文件路径
	ffmpegPath := "ffmpeg"

	// 如果在 Windows 下，且项目中有 ffmpeg/ffmpeg.exe，可以使用相对路径查找
	if runtime.GOOS == "windows" {
		if _, err := os.Stat("ffmpeg/ffmpeg.exe"); err == nil {
			absPath, _ := stdpath.Abs("ffmpeg/ffmpeg.exe")
			ffmpegPath = absPath
		}
	}

	// 2. 构建命令
	// -y: 覆盖输出文件
	// -i inputPath: 输入文件
	// -c copy: 核心参数！直接复制流，不重新编码（速度极快，无画质损失）
	cmd := exec.Command(ffmpegPath, "-y", "-i", inputPath, "-c", "copy", outputPath)

	// 3. 执行命令
	fmt.Printf("执行 FFmpeg 封装: %s -> %s\n", inputPath, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("FFmpeg 执行失败: %s\nOutput: %s\n", err, string(output))
		return fmt.Errorf("FFmpeg 封装失败: %v, 详细日志: %s", err, string(output))
	}

	fmt.Println("封装完成！")
	return nil
}

// DeleteFile 删除临时文件
func (a *App) DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}
