# Wails 架构功能总结

## 📋 概述

本文档总结了为 SilkRec（基于 Wails v2）新增的相机运动导出功能及其在 Wails 架构下的实现。

## 🏗️ Wails 架构说明

### 核心原理

```
┌─────────────────────────────────────────────────────┐
│                   Wails 应用                          │
├─────────────────────────────────────────────────────┤
│                                                       │
│  前端 (Vue 3)                后端 (Go)                │
│  ┌─────────────┐            ┌──────────────┐        │
│  │             │   IPC 通信  │              │        │
│  │  Vue 组件   │◄──────────►│  App 结构体   │        │
│  │             │            │              │        │
│  │ window.go.  │            │ 导出的方法    │        │
│  │ main.App.*  │            │ (首字母大写)  │        │
│  └─────────────┘            └──────────────┘        │
│                                     │                │
│                              ┌──────▼──────┐        │
│                              │   业务逻辑   │        │
│                              │  pkg/...    │        │
│                              └─────────────┘        │
└─────────────────────────────────────────────────────┘
```

### 绑定机制

1. **后端定义** (`app.go`):
   ```go
   type App struct {
       ctx context.Context
       exporter *recorder.Exporter
   }
   
   // 导出的方法（首字母大写）
   func (a *App) PrepareExport(...) (map[string]interface{}, error) {
       // 实现
   }
   ```

2. **Wails 绑定** (`main.go`):
   ```go
   wails.Run(&options.App{
       Bind: []interface{}{
           app,  // 自动绑定 App 的所有导出方法
       },
   })
   ```

3. **前端调用**:
   ```javascript
   // Wails 自动在 window.go.main.App 中注入方法
   const result = await window.go.main.App.PrepareExport(...);
   ```

## 🆕 新增功能

### 后端 API (Go)

在 `app.go` 中新增 4 个方法：

#### 1. PrepareExport
准备导出，加载鼠标数据并生成相机路径。

```go
func (a *App) PrepareExport(
    videoPath string,
    mouseDataPath string,
    outputPath string,
    screenWidth int,
    screenHeight int,
    fps int
) (map[string]interface{}, error)
```

**返回**:
```json
{
  "mouseEventCount": 1543,
  "cameraFrameCount": 900,
  "fps": 30,
  "enableZoom": true,
  "zoomLevel": 1.5,
  "smoothFactor": 0.15,
  "duration": 30.0,
  "estimatedFrames": 900
}
```

#### 2. GetCameraFrames
获取生成的相机帧数据（JSON 字符串）。

```go
func (a *App) GetCameraFrames() (string, error)
```

**返回**: JSON 字符串，需在前端解析
```javascript
const frames = JSON.parse(await GetCameraFrames());
// frames[i] = {
//   Timestamp: 0,
//   X: 960.0,
//   Y: 540.0,
//   Zoom: 1.0,
//   MouseX: 960,
//   MouseY: 540,
//   EventType: "move"
// }
```

#### 3. SaveCameraPath
保存相机路径到文件（调试用）。

```go
func (a *App) SaveCameraPath(outputPath string) error
```

#### 4. GetExportInfo
获取当前导出器的状态信息。

```go
func (a *App) GetExportInfo() map[string]interface{}
```

### 后端核心模块

#### 1. `pkg/recorder/camera.go`
相机运动系统：

- **CameraController**: 管理相机状态和平滑运动
- **CameraState**: 相机状态（位置、缩放）
- **CameraFrame**: 相机帧（包含时间戳、位置、鼠标坐标）
- **GenerateCameraPath()**: 生成相机运动路径

**关键算法**:
```go
// 线性插值（Lerp）实现平滑运动
current = current + (target - current) * smoothFactor

// 视口计算（修正了底部右侧聚焦问题）
viewportX = cameraX - viewportWidth/2
viewportY = cameraY - viewportHeight/2
// + 边界限制
```

#### 2. `pkg/recorder/exporter.go`
导出管理器：

- **Exporter**: 导出流程管理
- **ExportConfig**: 导出配置
- **LoadMouseData()**: 加载鼠标数据
- **GenerateCameraPath()**: 生成相机路径
- **GetCameraFrames()**: 返回相机帧

### 前端工具类

#### 1. `frontend/src/utils/exporter.js`

**ExportManager**:
```javascript
const manager = new ExportManager();

// 准备导出
await manager.prepareExport(videoPath, mouseDataPath, outputPath, width, height, fps);

// 获取相机帧
const frames = await manager.getCameraFrames();

// 保存调试数据
await manager.saveCameraPath('output/debug.json');
```

**CameraRenderer**:
```javascript
const renderer = new CameraRenderer(canvas);

// 加载视频
await renderer.loadVideo('/output/video.mp4');

// 渲染单帧
const imageData = await renderer.renderFrame(cameraFrame, width, height, showCursor);
```

**ExportController**:
```javascript
const controller = new ExportController();

// 执行完整导出（带进度回调）
await controller.export(config, (progress, message) => {
    console.log(`${progress}%: ${message}`);
});
```

#### 2. `frontend/src/components/ExportPanel.vue`

完整的 UI 组件，包含：
- 导出配置表单
- 实时进度显示
- 错误处理
- 结果展示
- 测试连接功能

#### 3. `frontend/src/examples/export-example.js`

可在浏览器控制台直接运行的示例：
- `basicExportExample()` - 基础流程
- `fullExportExample()` - 完整导出
- `debugCameraPathExample()` - 调试相机路径
- `testAllBindings()` - 测试所有 API
- `renderSingleFrameExample()` - 渲染单帧
- `quickTest()` - 快速测试

## 📝 使用流程

### 完整的导出流程

```javascript
// 1. 获取屏幕信息
const [screenWidth, screenHeight, dpi] = await window.go.main.App.GetScreenInfo();

// 2. 准备导出（生成相机路径）
const exportInfo = await window.go.main.App.PrepareExport(
    'output/recording.mp4',      // 输入视频
    'output/mouse_events.json',  // 鼠标数据
    'output/final.mp4',          // 输出路径
    screenWidth,
    screenHeight,
    30                           // FPS
);

console.log('准备完成:', exportInfo);

// 3. 获取相机帧
const framesJSON = await window.go.main.App.GetCameraFrames();
const cameraFrames = JSON.parse(framesJSON);

console.log(`获取到 ${cameraFrames.length} 个相机帧`);

// 4. 创建渲染器
const canvas = document.createElement('canvas');
canvas.width = screenWidth;
canvas.height = screenHeight;
const renderer = new CameraRenderer(canvas);

// 5. 加载视频
await renderer.loadVideo('/output/recording.mp4');

// 6. 启动导出管道
await window.go.main.App.StartExport('output/final.mp4', 30);

// 7. 渲染并导出每一帧
for (let i = 0; i < cameraFrames.length; i++) {
    const frame = cameraFrames[i];
    
    // 渲染帧（应用相机变换）
    const imageData = await renderer.renderFrame(
        frame,
        screenWidth,
        screenHeight,
        true  // 显示光标
    );
    
    // 写入帧
    await window.go.main.App.WriteExportFrame(imageData);
    
    // 更新进度
    if (i % 30 === 0) {
        console.log(`进度: ${(i / cameraFrames.length * 100).toFixed(1)}%`);
    }
}

// 8. 完成导出
await window.go.main.App.FinishExport();

console.log('导出完成!');
```

### 简化的使用方式

使用 `ExportController` 封装的方法：

```javascript
import { ExportController } from '@/utils/exporter.js';

const controller = new ExportController();

await controller.export(
    {
        videoPath: 'output/recording.mp4',
        mouseDataPath: 'output/mouse_events.json',
        outputPath: 'output/export.mp4',
        screenWidth: 1920,
        screenHeight: 1080,
        fps: 30,
        showCursor: true
    },
    (progress, message) => {
        console.log(`[${progress.toFixed(1)}%] ${message}`);
    }
);
```

## 🔧 开发指南

### 启动开发环境

```bash
# 在项目根目录
wails dev
```

### 测试 Wails 绑定

打开浏览器控制台 (F12):

```javascript
// 查看所有可用方法
console.log(Object.keys(window.go.main.App));

// 测试基本方法
const greeting = await window.go.main.App.Greet('Wails');
console.log(greeting);

// 测试 FFmpeg
const ffmpegOk = await window.go.main.App.CheckFFmpegAvailable();
console.log('FFmpeg available:', ffmpegOk);

// 测试导出信息
const info = await window.go.main.App.GetExportInfo();
console.log('Export info:', info);
```

### 运行示例代码

```javascript
// 导入示例
import examples from './examples/export-example.js';

// 运行快速测试
await examples.quickTest();

// 测试所有 API
await examples.testAllBindings();

// 完整导出
await examples.fullExportExample();
```

## 📚 文档导航

### 用户文档
- **README.md** - 项目介绍，FFmpeg 安装说明
- **TASK_COMPLETION_SUMMARY.md** - 任务完成总结（中文）

### 技术文档
- **CAMERA_MOVEMENT.md** - 相机运动系统技术文档
- **EXPORT_GUIDE.md** - 导出流程详细指南
- **CHANGELOG_CAMERA_MOVEMENT.md** - 完整变更日志

### Wails 文档
- **WAILS_INTEGRATION.md** - Wails 集成指南 ⭐
- **WAILS_FEATURES_SUMMARY.md** - 本文档 ⭐
- **frontend/FRONTEND_GUIDE.md** - 前端开发指南 ⭐

### 代码示例
- **frontend/src/utils/exporter.js** - 导出工具类
- **frontend/src/components/ExportPanel.vue** - UI 组件
- **frontend/src/examples/export-example.js** - 使用示例

## ✅ Wails 架构检查清单

### 后端 (Go)

- [x] 方法定义在 App 结构体上
- [x] 方法名首字母大写（导出）
- [x] 参数类型 JSON 可序列化
- [x] 返回值类型 JSON 可序列化
- [x] 使用 error 接口处理错误
- [x] 使用 context.Context 访问运行时
- [x] 在 main.go 中绑定 App

### 前端 (JavaScript/Vue)

- [x] 通过 window.go.main.App.* 调用方法
- [x] 所有调用使用 async/await
- [x] 正确处理 Promise rejection（错误）
- [x] 提供进度反馈
- [x] 资源清理（dispose/cancel）
- [x] 错误处理和用户提示

### 文档

- [x] API 使用说明
- [x] 完整代码示例
- [x] 常见问题解答
- [x] 调试指南
- [x] 性能优化建议

## 🎯 关键特性

### 1. 零配置绑定

Wails 自动绑定 Go 方法到前端，无需手动配置。

### 2. 类型安全

通过生成的 TypeScript 定义获得类型提示（如果使用 TS）。

### 3. 双向通信

- **前端 → 后端**: 方法调用
- **后端 → 前端**: 事件发送（`runtime.EventsEmit`）

### 4. 文件访问

通过 HTTP 静态文件服务访问输出文件：
```javascript
<video src="/output/recording.mp4"></video>
```

### 5. 开发体验

- 热重载
- 浏览器 DevTools
- 控制台调试
- 后端日志

## 🚀 性能特点

- **相机帧生成**: O(n)，n = 鼠标事件数
- **内存占用**: 极小，相机帧是轻量级结构
- **IPC 通信**: Wails 使用高效的二进制协议
- **渲染**: 使用 Canvas 2D（可升级到 WebGL）

## 🔍 常见问题

### Q: 如何确认 Wails 绑定正常工作？

A: 在控制台运行：
```javascript
console.log(window.go); // 应显示对象
console.log(Object.keys(window.go.main.App)); // 应显示方法列表
```

### Q: 为什么方法调用失败？

A: 检查：
1. 方法名大小写是否正确
2. 参数类型是否匹配
3. 是否使用了 await
4. 浏览器控制台是否有错误

### Q: 如何调试 Go 代码？

A: 
1. 使用 `fmt.Println()` 输出到终端
2. 使用 `runtime.EventsEmit()` 发送调试信息到前端
3. 使用 Go 调试器（Delve）

### Q: 如何处理大量数据？

A:
1. 不要在单次调用中传输大量数据
2. 使用流式处理（事件系统）
3. 在后端处理，只返回结果

## 📦 部署

### 构建生产版本

```bash
wails build
```

### 部署清单

必须包含：
- ✅ `SmoothScreen.exe` (或对应平台的可执行文件)
- ✅ `ffmpeg.exe` (与 exe 同目录)

自动创建：
- `output/` 目录

不需要：
- ❌ `frontend/` 源码
- ❌ `pkg/` 源码
- ❌ `node_modules/`

## 🎓 学习资源

### Wails
- [官方文档](https://wails.io/docs/introduction)
- [GitHub](https://github.com/wailsapp/wails)
- [示例项目](https://github.com/wailsapp/wails/tree/master/v2/examples)

### Vue 3
- [官方文档](https://vuejs.org/)
- [Composition API](https://vuejs.org/guide/extras/composition-api-faq.html)

### 本项目
- 阅读 `WAILS_INTEGRATION.md` 深入了解
- 运行 `examples/export-example.js` 中的示例
- 查看 `components/ExportPanel.vue` 的完整实现

## 🤝 贡献指南

开发新功能时：

1. **后端**: 在 `app.go` 添加导出方法
2. **前端**: 通过 `window.go.main.App.*` 调用
3. **测试**: 在 `examples/` 添加示例
4. **文档**: 更新相关 Markdown 文档

## ✨ 总结

通过 Wails 架构，我们实现了：

1. ✅ 高性能的 Go 后端
2. ✅ 现代化的 Vue 前端
3. ✅ 无缝的双向通信
4. ✅ 类型安全的 API
5. ✅ 优秀的开发体验
6. ✅ 简单的部署流程

相机运动导出功能完全集成到 Wails 架构中，提供了流畅的用户体验和强大的功能！

---

**最后更新**: 2024-01-06  
**Wails 版本**: v2  
**Vue 版本**: 3.2+
