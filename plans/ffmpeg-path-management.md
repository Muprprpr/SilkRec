# FFmpeg 路径管理设计

## 目录结构

### 开发环境
```
SilkRec/
├── ffmpeg/
│   └── ffmpeg.exe          # 开发环境 FFmpeg
├── main.go
├── app.go
├── pkg/
│   └── ffmpeg/
│       └── ffmpeg.go       # FFmpeg 路径管理器
└── ...
```

### 生产环境（用户安装后）
```
SilkRec/
├── SilkRec.exe              # 主程序
├── ffmpeg.exe               # FFmpeg（与主程序同级）
└── output/                  # 录制输出目录（自动创建）
```

## 路径查找优先级

1. **开发环境**: `./ffmpeg/ffmpeg.exe`（相对于项目根目录）
2. **生产环境**: 可执行文件同级目录的 `ffmpeg.exe`
3. **备用**: 应用数据目录（可选）

## API 设计

```go
package ffmpeg

import (
    "errors"
    "os"
    "path/filepath"
    "runtime"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

// FFmpegManager 管理 FFmpeg 可执行文件路径
type FFmpegManager struct {
    ctx       context.Context
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
        m.getDevPath(),      // 开发环境路径
        m.getProdPath(),     // 生产环境路径
        m.getAppDataPath(),  // 应用数据目录（可选）
    }

    for _, path := range paths {
        if path == "" {
            continue
        }
        if _, err := os.Stat(path); err == nil {
            m.ffmpegPath = path
            return path, nil
        }
    }

    return "", errors.New("未找到 FFmpeg 可执行文件")
}

// getDevPath 获取开发环境路径
func (m *FFmpegManager) getDevPath() string {
    // 项目根目录下的 ffmpeg/ffmpeg.exe
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
    // 使用 Wails runtime 获取应用数据目录
    if m.ctx != nil {
        appDataDir := runtime.ApplicationDataDir(m.ctx)
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
    if runtime.GOOS == "windows" && filepath.Ext(path) != ".exe" {
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
```

## 使用示例

```go
// 在 app.go 中使用
type App struct {
    ctx           context.Context
    ffmpegManager *ffmpeg.FFmpegManager
    recorder      *recorder.Recorder
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // 初始化 FFmpeg 管理器
    a.ffmpegManager = ffmpeg.NewFFmpegManager(ctx)
    
    // 检查 FFmpeg 是否可用
    if !a.ffmpegManager.CheckFFmpegAvailable() {
        runtime.EventsEmit(ctx, "ffmpeg-error", "未找到 FFmpeg 可执行文件")
        return
    }
    
    // 获取 FFmpeg 路径
    ffmpegPath, _ := a.ffmpegManager.GetFFmpegPath()
    fmt.Printf("FFmpeg 路径: %s\n", ffmpegPath)
    
    // 初始化录制器
    a.recorder = recorder.NewRecorder(a.ffmpegManager)
}
```

## 打包说明

### 使用 Wails 打包

1. 将 `ffmpeg.exe` 复制到 `build/windows/` 目录
2. 修改 `build/windows/info.json` 添加文件复制配置
3. 或者在打包后手动将 `ffmpeg.exe` 复制到可执行文件同级目录

### NSIS 安装脚本（可选）

修改 `build/windows/installer/project.nsi`，添加 FFmpeg 文件复制：

```nsis
; 复制 FFmpeg
File "ffmpeg.exe"
```

## 注意事项

1. **开发环境**: 需要手动将 `ffmpeg.exe` 放在项目根目录的 `ffmpeg/` 文件夹中
2. **生产环境**: 需要将 `ffmpeg.exe` 与 `SilkRec.exe` 放在同一目录
3. **版本要求**: FFmpeg 需要 Windows 版本，支持以下功能：
   - `lavfi` 滤镜（用于 ddagrab）
   - `h264_nvenc` 或 `libx264` 编码器
   - `image2pipe` 格式（用于导出）
