# 快速开始指南

## 🚀 5 分钟快速上手

### 前提条件

1. ✅ 已安装 Go 1.18+
2. ✅ 已安装 Node.js 16+
3. ✅ 已安装 Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
4. ✅ Windows 操作系统

### 步骤 1: 克隆项目

```bash
git clone https://github.com/Muprprpr/SilkRec.git
cd SilkRec
```

### 步骤 2: 安装 FFmpeg

**重要**: FFmpeg 未包含在仓库中，需要手动下载。

1. 下载 FFmpeg Windows 版本：
   - [ffmpeg.org](https://ffmpeg.org/download.html) 或
   - [gyan.dev](https://www.gyan.dev/ffmpeg/builds/)

2. 解压后找到 `ffmpeg.exe`

3. 放置到项目目录：
   ```bash
   mkdir ffmpeg
   copy path\to\ffmpeg.exe ffmpeg\ffmpeg.exe
   ```

### 步骤 3: 安装依赖

```bash
# 安装前端依赖
cd frontend
npm install
cd ..

# Go 依赖会自动下载
```

### 步骤 4: 启动开发服务器

```bash
wails dev
```

应用会自动打开！

### 步骤 5: 测试功能

#### 在浏览器控制台 (F12) 测试

```javascript
// 测试 Wails 绑定
const greeting = await window.go.main.App.Greet('World');
console.log(greeting); // "Hello World, It's show time!"

// 检查 FFmpeg
const ffmpegOk = await window.go.main.App.CheckFFmpegAvailable();
console.log('FFmpeg available:', ffmpegOk); // 应该是 true

// 获取屏幕信息
const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
console.log(`Screen: ${width}x${height}, DPI: ${dpi}`);
```

#### 运行内置示例

```javascript
// 导入示例
import examples from './src/examples/export-example.js';

// 快速测试所有功能
await examples.quickTest();

// 测试相机路径生成
await examples.debugCameraPathExample();
```

---

## 📚 下一步

### 学习使用

1. **阅读文档**:
   - [WAILS_INTEGRATION.md](WAILS_INTEGRATION.md) - Wails 架构集成
   - [CAMERA_MOVEMENT.md](CAMERA_MOVEMENT.md) - 相机运动原理
   - [frontend/FRONTEND_GUIDE.md](frontend/FRONTEND_GUIDE.md) - 前端开发指南

2. **查看示例**:
   - `frontend/src/examples/export-example.js` - 完整示例代码
   - `frontend/src/components/ExportPanel.vue` - UI 组件实现

3. **理解架构**:
   - `pkg/recorder/camera.go` - 相机运动算法
   - `pkg/recorder/exporter.go` - 导出管理
   - `app.go` - Wails API 绑定

### 录制和导出流程

#### 1. 录制视频

```javascript
// 开始录制
await window.go.main.App.StartScreenRecording('output/my_recording.mp4');

// ... 执行你的操作 ...

// 停止录制
const [videoPath, mouseDataPath, error] = await window.go.main.App.StopScreenRecording();
console.log('录制完成:', { videoPath, mouseDataPath });
```

#### 2. 导出带相机运动的视频

```javascript
import { ExportController } from './src/utils/exporter.js';

const controller = new ExportController();

// 获取屏幕尺寸
const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();

// 执行导出
await controller.export(
  {
    videoPath: 'output/my_recording.mp4',
    mouseDataPath: 'output/mouse_events.json',
    outputPath: 'output/final_export.mp4',
    screenWidth,
    screenHeight,
    fps: 30,
    showCursor: true
  },
  (progress, message) => {
    console.log(`[${progress.toFixed(1)}%] ${message}`);
  }
);

console.log('导出完成！');
```

---

## 🛠️ 开发工作流

### 修改代码

1. **后端 (Go)**:
   - 修改 `pkg/` 或 `app.go`
   - Wails 会自动重启 Go 进程
   - 刷新应用窗口

2. **前端 (Vue)**:
   - 修改 `frontend/src/`
   - Vite 会自动热重载
   - 无需刷新

### 调试

#### 前端调试
- 按 F12 打开 Chrome DevTools
- 使用 Console, Network, Sources 标签
- Vue DevTools 插件可用

#### 后端调试
- 使用 `fmt.Println()` 输出到终端
- 使用 `runtime.EventsEmit()` 发送日志到前端
- 或使用 Delve 调试器

### 添加新功能

#### 1. 添加后端 API

在 `app.go` 中添加方法：

```go
// MyNewFeature 新功能
func (a *App) MyNewFeature(param1 string, param2 int) (string, error) {
    // 实现逻辑
    result := fmt.Sprintf("Processed: %s, %d", param1, param2)
    return result, nil
}
```

#### 2. 从前端调用

```javascript
// 直接调用
const result = await window.go.main.App.MyNewFeature('test', 123);
console.log(result);

// 或使用生成的绑定 (推荐)
import { MyNewFeature } from '../wailsjs/go/main/App';
const result = await MyNewFeature('test', 123);
```

---

## 🏗️ 构建生产版本

### 构建应用

```bash
# 构建 Windows 版本
wails build

# 输出在 build/bin/ 目录
```

### 部署

需要包含：
1. ✅ `build/bin/SmoothScreen.exe`
2. ✅ `ffmpeg.exe` (放在与 exe 同目录)

不需要：
- ❌ `frontend/` 源码
- ❌ `pkg/` 源码
- ❌ `node_modules/`

运行时会自动创建：
- `output/` 目录

---

## 🔧 常见问题

### Q: wails: command not found

**A**: 安装 Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

确保 `$GOPATH/bin` 在 PATH 中。

### Q: FFmpeg 不可用

**A**: 
1. 确保 `ffmpeg.exe` 在 `./ffmpeg/` 目录
2. 运行测试:
   ```javascript
   const ok = await window.go.main.App.CheckFFmpegAvailable();
   console.log(ok); // 应该是 true
   ```

### Q: window.go 未定义

**A**: 
1. 确保 `wails dev` 正在运行
2. 刷新应用窗口 (Ctrl+R)
3. 检查终端是否有错误

### Q: 编译错误 (Linux/Mac)

**A**: 这是正常的，项目包含 Windows 特定代码：
- `pkg/sys/window.go` 使用 Windows API
- 只能在 Windows 上编译和运行

### Q: 导出很慢

**A**: 
- 这是正常的，逐帧渲染需要时间
- 降低 FPS (如 24 或 15) 可以加快速度
- 使用硬件加速的 FFmpeg 编码器

---

## 📖 核心概念

### Wails 架构

```
前端 (Vue)          后端 (Go)
    │                   │
    │   window.go.      │
    │   main.App.*      │
    │─────────────────► │  PrepareExport()
    │                   │  GetCameraFrames()
    │                   │  StartExport()
    │                   │  WriteExportFrame()
    │ ◄─────────────────│  FinishExport()
    │   Promise/Error   │
```

### 相机运动

相机通过**线性插值 (Lerp)** 实现平滑跟踪：

```
current = current + (target - current) * smoothFactor
```

- `smoothFactor = 0.15`: 平衡（默认）
- 越小越快，越大越平滑
- 点击时自动缩放 1.5 倍

### 导出流程

1. **PrepareExport** - 生成相机路径
2. **GetCameraFrames** - 获取相机帧数组
3. **循环**: 
   - 读取视频帧
   - 应用相机变换（缩放+平移）
   - 绘制光标
   - WriteExportFrame
4. **FinishExport** - 完成

---

## 🎯 推荐学习路径

### 第一天: 基础
1. ✅ 完成快速开始
2. ✅ 运行示例代码
3. ✅ 理解 Wails 绑定机制

### 第二天: 前端
1. 阅读 `FRONTEND_GUIDE.md`
2. 查看 `ExportPanel.vue` 实现
3. 修改 UI 组件

### 第三天: 后端
1. 阅读 `CAMERA_MOVEMENT.md`
2. 理解相机算法
3. 修改导出参数

### 第四天: 集成
1. 阅读 `WAILS_INTEGRATION.md`
2. 添加自定义功能
3. 完整测试流程

---

## 💡 最佳实践

### 1. 错误处理

```javascript
try {
  const result = await window.go.main.App.SomeMethod();
  // 成功处理
} catch (error) {
  console.error('Error:', error);
  alert('操作失败: ' + error);
}
```

### 2. 进度反馈

```javascript
await controller.export(config, (progress, message) => {
  // 更新 UI
  this.progress = progress;
  this.statusMessage = message;
});
```

### 3. 资源清理

```javascript
export default {
  beforeUnmount() {
    // 清理资源
    if (this.controller) {
      this.controller.cancel();
    }
  }
};
```

### 4. 异步处理

```javascript
// ✅ 正确: 使用 await
const result = await window.go.main.App.GetExportInfo();

// ❌ 错误: 忘记 await
const result = window.go.main.App.GetExportInfo(); // 这是 Promise!
```

---

## 🆘 获取帮助

### 文档资源

- **WAILS_INTEGRATION.md** - Wails 集成详解
- **WAILS_FEATURES_SUMMARY.md** - 功能总结
- **CAMERA_MOVEMENT.md** - 技术实现
- **EXPORT_GUIDE.md** - 导出详细指南
- **frontend/FRONTEND_GUIDE.md** - 前端开发指南

### 示例代码

- `frontend/src/examples/export-example.js` - 可运行的示例
- `frontend/src/components/ExportPanel.vue` - 完整 UI
- `frontend/src/utils/exporter.js` - 工具类

### 在线资源

- [Wails 官方文档](https://wails.io/docs/introduction)
- [Vue 3 文档](https://vuejs.org/)
- [FFmpeg 文档](https://ffmpeg.org/documentation.html)

---

## ✨ 开始创作！

现在你已经准备好开始使用 SilkRec 了！

```javascript
// 在控制台运行快速测试
import examples from './src/examples/export-example.js';
await examples.quickTest();
```

祝创作愉快！🎬✨

---

**提示**: 如果遇到问题，首先查看浏览器控制台和终端的错误信息。
