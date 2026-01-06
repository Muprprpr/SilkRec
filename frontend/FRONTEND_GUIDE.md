# 前端开发指南

## 快速开始

本指南帮助前端开发者快速上手 SilkRec 的相机运动导出功能。

## 项目结构

```
frontend/
├── src/
│   ├── App.vue                      # 根组件
│   ├── components/
│   │   └── ExportPanel.vue          # 导出面板组件 ✨新增
│   ├── utils/
│   │   └── exporter.js              # 导出工具类 ✨新增
│   ├── examples/
│   │   └── export-example.js        # 使用示例 ✨新增
│   └── assets/                      # 静态资源
├── wailsjs/                         # Wails 自动生成的绑定
│   ├── go/
│   │   └── main/
│   │       └── App.js               # Go App 绑定
│   └── runtime/
│       └── runtime.js               # Wails 运行时
├── index.html
└── package.json
```

## 核心文件说明

### 1. `utils/exporter.js` - 导出工具类

提供三个核心类：

#### ExportManager
管理导出流程的类：
```javascript
import { ExportManager } from '@/utils/exporter.js';

const manager = new ExportManager();

// 准备导出
const info = await manager.prepareExport(
  videoPath, mouseDataPath, outputPath,
  screenWidth, screenHeight, fps
);

// 获取相机帧
const frames = await manager.getCameraFrames();

// 保存相机路径（调试用）
await manager.saveCameraPath('output/camera.json');
```

#### CameraRenderer
渲染视频帧的类：
```javascript
import { CameraRenderer } from '@/utils/exporter.js';

const canvas = document.getElementById('canvas');
const renderer = new CameraRenderer(canvas);

// 加载视频
await renderer.loadVideo('/output/recording.mp4');

// 渲染单帧
const imageData = await renderer.renderFrame(
  cameraFrame,
  screenWidth,
  screenHeight,
  showCursor
);
```

#### ExportController
完整流程控制器：
```javascript
import { ExportController } from '@/utils/exporter.js';

const controller = new ExportController();

// 执行完整导出
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
    console.log(`${progress}% - ${message}`);
  }
);
```

### 2. `components/ExportPanel.vue` - UI 组件

完整的导出 UI，包含：
- 配置表单
- 进度显示
- 错误处理
- 结果展示

使用方式：
```vue
<template>
  <div id="app">
    <ExportPanel />
  </div>
</template>

<script>
import ExportPanel from './components/ExportPanel.vue';

export default {
  components: { ExportPanel }
};
</script>
```

### 3. `examples/export-example.js` - 示例代码

包含多个使用示例，可在浏览器控制台直接运行：

```javascript
// 在控制台运行
import examples from './examples/export-example.js';

// 快速测试
examples.quickTest();

// 测试所有 API
examples.testAllBindings();

// 完整导出
examples.fullExportExample();

// 调试相机路径
examples.debugCameraPathExample();

// 渲染单帧
examples.renderSingleFrameExample();
```

## Wails API 调用

### 可用的后端方法

所有方法通过 `window.go.main.App.*` 调用：

```javascript
// 系统信息
const greeting = await window.go.main.App.Greet('Name');
const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
const ffmpegOk = await window.go.main.App.CheckFFmpegAvailable();

// 导出相关 ✨新增
const exportInfo = await window.go.main.App.PrepareExport(
  videoPath, mouseDataPath, outputPath,
  screenWidth, screenHeight, fps
);
const framesJSON = await window.go.main.App.GetCameraFrames();
await window.go.main.App.SaveCameraPath(outputPath);
const info = await window.go.main.App.GetExportInfo();

// FFmpeg 管道导出
await window.go.main.App.StartExport(outputPath, frameRate);
await window.go.main.App.WriteExportFrame(base64ImageData);
await window.go.main.App.FinishExport();
await window.go.main.App.StopExport();

// 录制相关
await window.go.main.App.StartScreenRecording(videoPath);
const [videoPath, mouseDataPath, error] = await window.go.main.App.StopScreenRecording();
const status = await window.go.main.App.GetRecordingStatus();
```

### 类型安全（推荐）

使用 Wails 生成的绑定：

```javascript
// 从生成的绑定导入
import * as App from '../wailsjs/go/main/App';

// 调用（带类型提示）
const info = await App.PrepareExport(
  videoPath,
  mouseDataPath,
  outputPath,
  screenWidth,
  screenHeight,
  fps
);
```

## 开发流程

### 1. 启动开发服务器

```bash
# 在项目根目录
wails dev
```

应用会自动打开，支持热重载。

### 2. 测试 API 连接

打开浏览器控制台 (F12)，运行：

```javascript
// 测试连接
const result = await window.go.main.App.Greet('Test');
console.log(result); // "Hello Test, It's show time!"

// 检查 FFmpeg
const available = await window.go.main.App.CheckFFmpegAvailable();
console.log('FFmpeg:', available);

// 获取屏幕信息
const [w, h, d] = await window.go.main.App.GetScreenInfo();
console.log(`Screen: ${w}x${h}, DPI: ${d}`);
```

### 3. 集成到你的组件

```vue
<template>
  <div>
    <button @click="startExport" :disabled="isExporting">
      {{ isExporting ? '导出中...' : '开始导出' }}
    </button>
    <div v-if="isExporting">
      进度: {{ progress }}%
    </div>
  </div>
</template>

<script>
import { ExportController } from '@/utils/exporter.js';

export default {
  data() {
    return {
      isExporting: false,
      progress: 0,
      controller: null
    };
  },
  
  mounted() {
    this.controller = new ExportController();
  },
  
  methods: {
    async startExport() {
      this.isExporting = true;
      
      try {
        const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
        
        await this.controller.export(
          {
            videoPath: 'output/recording.mp4',
            mouseDataPath: 'output/mouse_events.json',
            outputPath: 'output/final.mp4',
            screenWidth,
            screenHeight,
            fps: 30
          },
          (progress, message) => {
            this.progress = progress;
            console.log(message);
          }
        );
        
        alert('导出完成！');
        
      } catch (error) {
        alert('导出失败: ' + error.message);
      } finally {
        this.isExporting = false;
      }
    }
  }
};
</script>
```

## 调试技巧

### 1. 查看 Wails 绑定

```javascript
// 列出所有可用方法
console.log(Object.keys(window.go.main.App));

// 查看完整对象
console.log(window.go.main.App);
```

### 2. 保存相机路径用于调试

```javascript
// 生成并保存相机路径
const [w, h] = await window.go.main.App.GetScreenInfo();

await window.go.main.App.PrepareExport(
  'output/recording.mp4',
  'output/mouse_events.json',
  'output/test.mp4',
  w, h, 30
);

await window.go.main.App.SaveCameraPath('output/camera_debug.json');
console.log('相机路径已保存');

// 然后可以查看 output/camera_debug.json 文件
```

### 3. 单帧渲染测试

```javascript
// 渲染单个帧进行测试
import examples from './examples/export-example.js';
await examples.renderSingleFrameExample();

// Canvas 会添加到页面底部，可以查看渲染效果
```

### 4. 使用浏览器 DevTools

- **Console**: 查看日志和错误
- **Network**: 检查 API 调用（实际上是 IPC，不会显示）
- **Sources**: 设置断点调试
- **Vue DevTools**: 查看组件状态（需要安装插件）

## 常见问题

### Q: window.go 未定义

**A**: 确保 `wails dev` 正在运行，刷新页面。

### Q: 方法调用失败

**A**: 检查：
1. 方法名是否正确（大小写敏感）
2. 参数类型是否正确
3. 浏览器控制台是否有错误

### Q: FFmpeg 不可用

**A**: 
- 开发环境：将 `ffmpeg.exe` 放在 `./ffmpeg/ffmpeg.exe`
- 调用 `CheckFFmpegAvailable()` 验证

### Q: 视频无法加载

**A**: 
- 确保视频路径正确
- 使用 `/output/xxx.mp4` 格式（通过 HTTP 访问）
- 检查 `main.go` 中的静态文件服务配置

### Q: 导出很慢

**A**: 
- 这是正常的，需要逐帧渲染
- 降低 FPS 可以加快速度
- 考虑使用 Web Worker（未实现）

### Q: 内存占用高

**A**:
- 正常现象，Canvas 渲染需要内存
- 完成后调用 `renderer.dispose()` 清理
- 考虑分批处理

## 性能优化

### 1. 避免频繁的 Go 调用

```javascript
// ❌ 不好
for (let i = 0; i < 1000; i++) {
  await window.go.main.App.SomeMethod(i);
}

// ✅ 好
const batch = Array.from({length: 1000}, (_, i) => i);
await window.go.main.App.ProcessBatch(batch);
```

### 2. 使用批量渲染

```javascript
// 每 30 帧更新一次 UI
if (frameIndex % 30 === 0) {
  this.progress = (frameIndex / totalFrames) * 100;
  await this.$nextTick(); // 让 UI 更新
}
```

### 3. 资源清理

```javascript
// 组件卸载时
beforeUnmount() {
  if (this.renderer) {
    this.renderer.dispose();
  }
  if (this.controller) {
    this.controller.cancel();
  }
}
```

## 进阶使用

### 自定义光标样式

修改 `CameraRenderer.drawCursor()`:

```javascript
drawCursor(x, y, eventType) {
  const ctx = this.ctx;
  
  // 自定义样式
  if (eventType.includes('down')) {
    // 点击时显示涟漪效果
    ctx.fillStyle = 'rgba(255, 0, 0, 0.2)';
    ctx.beginPath();
    ctx.arc(x, y, 40, 0, Math.PI * 2);
    ctx.fill();
  }
  
  // 绘制自定义光标图标
  // ...
}
```

### 添加导出预览

```vue
<template>
  <div>
    <canvas ref="previewCanvas"></canvas>
    <button @click="startExport">开始导出</button>
  </div>
</template>

<script>
export default {
  methods: {
    async startExport() {
      const canvas = this.$refs.previewCanvas;
      const renderer = new CameraRenderer(canvas);
      
      // 渲染时更新预览
      for (const frame of frames) {
        await renderer.renderFrame(frame, ...);
        // Canvas 自动显示最新帧
        await new Promise(r => setTimeout(r, 10)); // 减速预览
      }
    }
  }
};
</script>
```

### 使用 Pinia 管理状态

```javascript
// stores/export.js
import { defineStore } from 'pinia';

export const useExportStore = defineStore('export', {
  state: () => ({
    isExporting: false,
    progress: 0,
    statusMessage: '',
    error: null
  }),
  
  actions: {
    async startExport(config) {
      this.isExporting = true;
      this.error = null;
      
      const controller = new ExportController();
      
      try {
        await controller.export(config, (progress, message) => {
          this.progress = progress;
          this.statusMessage = message;
        });
      } catch (error) {
        this.error = error.message;
      } finally {
        this.isExporting = false;
      }
    }
  }
});
```

## 参考资源

- [Wails 官方文档](https://wails.io/)
- [Vue 3 文档](https://vuejs.org/)
- 本项目文档：
  - `WAILS_INTEGRATION.md` - Wails 集成指南
  - `CAMERA_MOVEMENT.md` - 技术实现
  - `EXPORT_GUIDE.md` - 导出流程

## 下一步

1. 查看 `components/ExportPanel.vue` 了解完整 UI 实现
2. 运行 `examples/export-example.js` 中的示例
3. 阅读 `WAILS_INTEGRATION.md` 深入了解 Wails 架构
4. 开始开发你自己的功能！

祝开发愉快！ 🚀
