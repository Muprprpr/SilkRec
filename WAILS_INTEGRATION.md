# Wails 架构集成指南

## 概述

本文档说明如何在 Wails v2 架构下正确使用 SilkRec 的相机运动导出功能。

## Wails 架构说明

### 后端 (Go)

后端代码在 `app.go` 中定义，所有导出的方法都会自动绑定到前端。

#### 关键点：
1. **方法必须是 App 结构体的导出方法**（首字母大写）
2. **参数和返回值类型必须是 JSON 可序列化的**
3. **使用 context.Context 来访问 Wails 运行时**
4. **错误处理使用 Go 的 error 接口**

#### 导出的 API 方法：

```go
// 准备导出
func (a *App) PrepareExport(
    videoPath string, 
    mouseDataPath string, 
    outputPath string, 
    screenWidth int, 
    screenHeight int, 
    fps int
) (map[string]interface{}, error)

// 获取相机帧
func (a *App) GetCameraFrames() (string, error)

// 保存相机路径
func (a *App) SaveCameraPath(outputPath string) error

// 获取导出信息
func (a *App) GetExportInfo() map[string]interface{}

// 其他已有方法
func (a *App) CheckFFmpegAvailable() bool
func (a *App) GetScreenInfo() (int, int, int)
func (a *App) StartExport(outputPath string, frameRate int) error
func (a *App) WriteExportFrame(base64Data string) error
func (a *App) FinishExport() error
func (a *App) StopExport() error
```

### 前端 (Vue 3)

#### 方法 1: 直接调用（开发环境）

在开发环境下，Wails 在 `window.go` 对象中注入所有绑定：

```javascript
// 调用 Go 方法
const result = await window.go.main.App.PrepareExport(
    videoPath,
    mouseDataPath,
    outputPath,
    screenWidth,
    screenHeight,
    fps
);
```

#### 方法 2: 使用生成的绑定（生产环境推荐）

Wails 构建时会在 `frontend/wailsjs` 目录生成类型安全的绑定：

```javascript
// 从生成的绑定导入
import { PrepareExport, GetCameraFrames } from '../wailsjs/go/main/App';

// 调用
const result = await PrepareExport(
    videoPath,
    mouseDataPath,
    outputPath,
    screenWidth,
    screenHeight,
    fps
);
```

**注意**: `wailsjs` 目录在 `wails dev` 或 `wails build` 时自动生成。

## 完整的集成示例

### 1. 创建工具类

文件：`frontend/src/utils/exporter.js`

这个文件已经创建，包含：
- `ExportManager`: 管理导出流程
- `CameraRenderer`: 处理视频渲染
- `ExportController`: 完整流程控制

### 2. 创建 Vue 组件

文件：`frontend/src/components/ExportPanel.vue`

这个组件已经创建，提供完整的 UI：
- 导出配置表单
- 进度显示
- 错误处理
- 结果展示

### 3. 在 App.vue 中使用

```vue
<template>
  <div id="app">
    <ExportPanel />
  </div>
</template>

<script>
import ExportPanel from './components/ExportPanel.vue';

export default {
  name: 'App',
  components: {
    ExportPanel
  }
};
</script>
```

## Wails 绑定验证

### 检查绑定是否正确

在浏览器控制台中：

```javascript
// 检查 window.go 对象是否存在
console.log(window.go);

// 检查 App 绑定
console.log(window.go.main.App);

// 列出所有可用方法
console.log(Object.keys(window.go.main.App));
```

预期输出应包含：
```
[
  "CheckFFmpegAvailable",
  "DeleteFile",
  "FinishExport",
  "GetCameraFrames",
  "GetExportInfo",
  "GetScreenInfo",
  "Greet",
  "PrepareExport",
  "SaveCameraPath",
  "StartExport",
  "StopExport",
  "WriteExportFrame",
  ...
]
```

### 测试连接

```javascript
// 测试简单方法
const greeting = await window.go.main.App.Greet('Test');
console.log(greeting); // 应该返回 "Hello Test, It's show time!"

// 测试 FFmpeg
const ffmpegOk = await window.go.main.App.CheckFFmpegAvailable();
console.log('FFmpeg available:', ffmpegOk);

// 测试屏幕信息
const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
console.log(`Screen: ${width}x${height}, DPI: ${dpi}`);
```

## 完整的导出流程

### JavaScript 实现

```javascript
import { ExportController } from './utils/exporter.js';

async function exportVideo() {
  const controller = new ExportController();
  
  try {
    const result = await controller.export(
      {
        videoPath: 'output/recording.mp4',
        mouseDataPath: 'output/mouse_events.json',
        outputPath: 'output/final_export.mp4',
        screenWidth: 1920,
        screenHeight: 1080,
        fps: 30,
        showCursor: true
      },
      (progress, message) => {
        console.log(`[${progress.toFixed(1)}%] ${message}`);
        // 更新 UI
      }
    );
    
    console.log('Export complete:', result);
    
  } catch (error) {
    console.error('Export failed:', error);
  }
}
```

### Vue 组件实现

```vue
<script>
import { ExportController } from '../utils/exporter.js';

export default {
  data() {
    return {
      controller: new ExportController(),
      progress: 0,
      message: ''
    };
  },
  
  methods: {
    async startExport() {
      try {
        await this.controller.export(
          this.config,
          (progress, message) => {
            this.progress = progress;
            this.message = message;
          }
        );
        
        alert('导出完成！');
        
      } catch (error) {
        alert('导出失败: ' + error.message);
      }
    }
  }
};
</script>
```

## Wails 事件系统

### 后端发送事件

在 Go 代码中使用 `runtime.EventsEmit`:

```go
import wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

func (a *App) someMethod() {
    // 发送事件到前端
    wailsruntime.EventsEmit(a.ctx, "export-progress", map[string]interface{}{
        "progress": 50,
        "message": "Processing...",
    })
}
```

### 前端监听事件

```javascript
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// 开始监听
const unlisten = EventsOn('export-progress', (data) => {
  console.log('Progress:', data.progress, data.message);
});

// 停止监听
unlisten();
```

或者使用 window.runtime:

```javascript
window.runtime.EventsOn('export-progress', (data) => {
  console.log('Progress:', data);
});

window.runtime.EventsOff('export-progress');
```

## 文件路径处理

### 相对路径 vs 绝对路径

Wails 应用的工作目录是可执行文件所在目录：

```javascript
// ✅ 推荐：相对路径
videoPath: 'output/recording.mp4'

// ✅ 也可以：绝对路径
videoPath: 'C:/Users/xxx/output/recording.mp4'

// ❌ 不要：浏览器路径
videoPath: 'file:///C:/Users/xxx/output/recording.mp4'
```

### 访问输出文件

在 `main.go` 中已经配置了静态文件服务：

```go
mux.Handle("/output/", http.StripPrefix("/output/", http.FileServer(http.Dir("output"))))
```

所以在前端可以这样访问：

```javascript
// 在 video 标签中使用
<video src="/output/recording.mp4"></video>

// 或在 JavaScript 中
const videoUrl = window.location.origin + '/output/recording.mp4';
```

## 开发调试

### 1. 启动开发服务器

```bash
# 在项目根目录
wails dev
```

这会：
- 启动 Go 后端
- 启动前端开发服务器（Vite）
- 打开应用窗口
- 支持热重载

### 2. 使用浏览器调试工具

在开发模式下按 F12 打开 Chrome DevTools：
- Console: 查看日志和错误
- Network: 检查 API 调用
- Sources: 调试 JavaScript 代码

### 3. Go 后端日志

后端的 `fmt.Println()` 和 `log.Println()` 输出会显示在终端。

### 4. 常见问题排查

#### 问题：找不到 window.go

**原因**: Wails 绑定未正确加载

**解决**:
1. 确保 `wails dev` 正在运行
2. 检查浏览器控制台是否有错误
3. 刷新页面（Ctrl+R）

#### 问题：Go 方法调用失败

**原因**: 方法签名不正确或未导出

**检查**:
1. 方法名首字母大写
2. 参数类型 JSON 可序列化
3. 在 `main.go` 的 `Bind` 中包含了 `app`

#### 问题：FFmpeg 不可用

**原因**: `ffmpeg.exe` 未放在正确位置

**解决**:
1. 开发环境：放在 `./ffmpeg/ffmpeg.exe`
2. 生产环境：与可执行文件同目录
3. 调用 `CheckFFmpegAvailable()` 验证

## 构建生产版本

### 1. 构建命令

```bash
# Windows
wails build

# 构建特定平台
wails build -platform windows/amd64
```

### 2. 输出位置

构建产物在 `build/bin/` 目录：
```
build/bin/
├── SmoothScreen.exe
└── ... (其他文件)
```

### 3. 部署清单

确保包含：
- ✅ `SmoothScreen.exe`
- ✅ `ffmpeg.exe` (与 exe 同目录)
- ✅ `output/` 目录（会自动创建）

**不要包含**:
- ❌ `frontend/` 源码（已编译进 exe）
- ❌ `pkg/` 源码（已编译进 exe）
- ❌ `node_modules/`

## 类型安全（TypeScript）

如果使用 TypeScript，Wails 会生成类型定义：

```typescript
// frontend/wailsjs/go/main/App.d.ts
export function PrepareExport(
  videoPath: string,
  mouseDataPath: string,
  outputPath: string,
  screenWidth: number,
  screenHeight: number,
  fps: number
): Promise<{[key: string]: any}>;

export function GetCameraFrames(): Promise<string>;
```

使用：

```typescript
import { PrepareExport, GetCameraFrames } from '../wailsjs/go/main/App';

async function export() {
  const info = await PrepareExport(
    'video.mp4',
    'mouse.json',
    'out.mp4',
    1920,
    1080,
    30
  );
  
  const frames: string = await GetCameraFrames();
  const cameraData = JSON.parse(frames);
}
```

## 最佳实践

### 1. 错误处理

```javascript
try {
  const result = await window.go.main.App.PrepareExport(...);
  // 成功
} catch (error) {
  // Go 返回的 error 会作为 Promise rejection
  console.error('Error:', error);
  alert('操作失败: ' + error);
}
```

### 2. 异步操作

所有 Go 方法调用都是异步的，总是返回 Promise：

```javascript
// ✅ 使用 await
const info = await window.go.main.App.GetExportInfo();

// ✅ 或使用 .then()
window.go.main.App.GetExportInfo().then(info => {
  console.log(info);
});

// ❌ 不要同步调用
const info = window.go.main.App.GetExportInfo(); // 这是 Promise，不是结果
```

### 3. 资源清理

```javascript
// 组件卸载时清理
export default {
  beforeUnmount() {
    // 停止正在进行的导出
    if (this.controller) {
      this.controller.cancel();
    }
  }
};
```

### 4. 进度反馈

```javascript
// 使用回调函数提供实时反馈
await controller.export(config, (progress, message) => {
  // 更新 UI
  this.progress = progress;
  this.statusMessage = message;
  
  // 或发送到 Vuex/Pinia
  store.commit('setExportProgress', { progress, message });
});
```

## 性能优化

### 1. 避免频繁的 Go 调用

```javascript
// ❌ 不好：每帧都调用
for (const frame of frames) {
  await window.go.main.App.ProcessFrame(frame); // 太慢
}

// ✅ 好：在前端处理，批量发送
const processedFrames = frames.map(f => processFrame(f));
await window.go.main.App.ProcessBatch(processedFrames);
```

### 2. 使用流式处理

对于大量数据，使用 Wails 的事件系统：

```go
// Go: 流式发送
for _, frame := range frames {
    runtime.EventsEmit(ctx, "frame-data", frame)
    time.Sleep(10 * time.Millisecond)
}
```

```javascript
// JS: 流式接收
EventsOn('frame-data', (frame) => {
  processFrame(frame);
});
```

## 参考资源

- [Wails 官方文档](https://wails.io/)
- [Wails v2 API 参考](https://wails.io/docs/reference/runtime/intro)
- [Vue 3 文档](https://vuejs.org/)
- 本项目文档:
  - `CAMERA_MOVEMENT.md` - 技术实现
  - `EXPORT_GUIDE.md` - 导出流程
  - `README.md` - 项目说明

## 总结

Wails 架构下的关键点：

1. **后端**: Go 方法自动绑定，无需手动配置
2. **前端**: 通过 `window.go.main.App.*` 调用
3. **类型安全**: 使用生成的 TypeScript 定义
4. **事件系统**: 用于实时通信
5. **文件访问**: 使用相对路径，通过 HTTP 访问输出文件
6. **调试**: 使用浏览器 DevTools 和终端日志

遵循这些规范，确保应用在 Wails 架构下正常运行！
