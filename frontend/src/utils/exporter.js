/**
 * SilkRec Export Utility
 * Wails 前端导出工具类
 * 
 * 该文件提供了与 Go 后端交互的导出功能封装
 * 所有的方法都通过 Wails 的运行时绑定调用后端 API
 */

// Wails 运行时 - 这些会在 Wails 构建时自动生成
// 在开发环境下，这些绑定在 window.go 对象中
// 生产环境下，通过 wailsjs 模块导入

/**
 * 导出管理器类
 * 处理视频导出的完整流程，包括相机运动效果
 */
export class ExportManager {
  constructor() {
    this.isExporting = false;
    this.currentProgress = 0;
    this.totalFrames = 0;
    this.cameraFrames = [];
  }

  /**
   * 准备导出 - 加载鼠标数据并生成相机路径
   * @param {string} videoPath - 输入视频路径
   * @param {string} mouseDataPath - 鼠标数据 JSON 文件路径
   * @param {string} outputPath - 输出视频路径
   * @param {number} screenWidth - 屏幕宽度
   * @param {number} screenHeight - 屏幕高度
   * @param {number} fps - 帧率（默认 30）
   * @returns {Promise<Object>} 导出信息
   */
  async prepareExport(videoPath, mouseDataPath, outputPath, screenWidth, screenHeight, fps = 30) {
    try {
      console.log('准备导出...', { videoPath, mouseDataPath, outputPath, screenWidth, screenHeight, fps });
      
      // 调用 Wails 绑定的 Go 方法
      // window.go.main.App 是 Wails 自动生成的命名空间
      const exportInfo = await window.go.main.App.PrepareExport(
        videoPath,
        mouseDataPath,
        outputPath,
        screenWidth,
        screenHeight,
        fps
      );
      
      console.log('导出准备完成:', exportInfo);
      this.totalFrames = exportInfo.cameraFrameCount || 0;
      
      return exportInfo;
    } catch (error) {
      console.error('准备导出失败:', error);
      throw new Error(`导出准备失败: ${error.message || error}`);
    }
  }

  /**
   * 获取生成的相机帧
   * @returns {Promise<Array>} 相机帧数组
   */
  async getCameraFrames() {
    try {
      console.log('获取相机帧...');
      
      // 调用 Go 方法获取 JSON 字符串
      const framesJSON = await window.go.main.App.GetCameraFrames();
      
      // 解析 JSON
      this.cameraFrames = JSON.parse(framesJSON);
      
      console.log(`获取到 ${this.cameraFrames.length} 个相机帧`);
      
      return this.cameraFrames;
    } catch (error) {
      console.error('获取相机帧失败:', error);
      throw new Error(`获取相机帧失败: ${error.message || error}`);
    }
  }

  /**
   * 保存相机路径用于调试
   * @param {string} outputPath - 输出文件路径
   */
  async saveCameraPath(outputPath) {
    try {
      await window.go.main.App.SaveCameraPath(outputPath);
      console.log('相机路径已保存:', outputPath);
    } catch (error) {
      console.error('保存相机路径失败:', error);
      throw error;
    }
  }

  /**
   * 获取导出信息
   * @returns {Promise<Object>} 导出统计信息
   */
  async getExportInfo() {
    try {
      const info = await window.go.main.App.GetExportInfo();
      return info;
    } catch (error) {
      console.error('获取导出信息失败:', error);
      throw error;
    }
  }

  /**
   * 开始导出流程（启动 FFmpeg 管道）
   * @param {string} outputPath - 输出文件路径
   * @param {number} frameRate - 帧率
   */
  async startExport(outputPath, frameRate = 30) {
    try {
      await window.go.main.App.StartExport(outputPath, frameRate);
      this.isExporting = true;
      this.currentProgress = 0;
      console.log('导出已启动');
    } catch (error) {
      console.error('启动导出失败:', error);
      throw error;
    }
  }

  /**
   * 写入一帧到导出流
   * @param {string} base64Data - Base64 编码的图像数据
   */
  async writeFrame(base64Data) {
    try {
      await window.go.main.App.WriteExportFrame(base64Data);
      this.currentProgress++;
    } catch (error) {
      console.error('写入帧失败:', error);
      throw error;
    }
  }

  /**
   * 完成导出
   */
  async finishExport() {
    try {
      await window.go.main.App.FinishExport();
      this.isExporting = false;
      console.log('导出完成');
    } catch (error) {
      console.error('完成导出失败:', error);
      throw error;
    }
  }

  /**
   * 停止导出
   */
  async stopExport() {
    try {
      await window.go.main.App.StopExport();
      this.isExporting = false;
      console.log('导出已停止');
    } catch (error) {
      console.error('停止导出失败:', error);
      throw error;
    }
  }

  /**
   * 获取当前进度百分比
   * @returns {number} 进度（0-100）
   */
  getProgress() {
    if (this.totalFrames === 0) return 0;
    return Math.min(100, (this.currentProgress / this.totalFrames) * 100);
  }

  /**
   * 重置导出状态
   */
  reset() {
    this.isExporting = false;
    this.currentProgress = 0;
    this.totalFrames = 0;
    this.cameraFrames = [];
  }
}

/**
 * 相机渲染器类
 * 处理视频帧的相机变换和渲染
 */
export class CameraRenderer {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.video = null;
  }

  /**
   * 加载视频
   * @param {string} videoPath - 视频文件路径
   * @returns {Promise<HTMLVideoElement>}
   */
  async loadVideo(videoPath) {
    return new Promise((resolve, reject) => {
      this.video = document.createElement('video');
      this.video.src = videoPath;
      
      this.video.onloadedmetadata = () => {
        console.log('视频已加载:', {
          duration: this.video.duration,
          width: this.video.videoWidth,
          height: this.video.videoHeight
        });
        resolve(this.video);
      };
      
      this.video.onerror = (error) => {
        reject(new Error('视频加载失败: ' + error));
      };
    });
  }

  /**
   * 渲染单个相机帧
   * @param {Object} cameraFrame - 相机帧数据
   * @param {number} screenWidth - 屏幕宽度
   * @param {number} screenHeight - 屏幕高度
   * @param {boolean} showCursor - 是否显示光标
   * @returns {Promise<string>} Base64 编码的图像数据
   */
  async renderFrame(cameraFrame, screenWidth, screenHeight, showCursor = true) {
    // 跳转到视频的对应时间
    this.video.currentTime = cameraFrame.Timestamp / 1000.0;
    await this.waitForSeek();

    // 清空画布
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

    // 计算视口
    const viewportWidth = screenWidth / cameraFrame.Zoom;
    const viewportHeight = screenHeight / cameraFrame.Zoom;
    const viewportX = cameraFrame.X - viewportWidth / 2;
    const viewportY = cameraFrame.Y - viewportHeight / 2;

    // 应用相机变换
    this.ctx.save();
    
    // 缩放
    this.ctx.scale(cameraFrame.Zoom, cameraFrame.Zoom);
    
    // 平移（负值因为我们移动的是画布，不是视口）
    this.ctx.translate(-viewportX, -viewportY);
    
    // 绘制视频帧
    this.ctx.drawImage(this.video, 0, 0, screenWidth, screenHeight);
    
    this.ctx.restore();

    // 绘制光标
    if (showCursor) {
      this.drawCursor(cameraFrame.MouseX, cameraFrame.MouseY, cameraFrame.EventType);
    }

    // 返回 Base64 数据（移除 data:image/png;base64, 前缀）
    return this.canvas.toDataURL('image/png').split(',')[1];
  }

  /**
   * 绘制光标
   * @param {number} x - 光标 X 坐标
   * @param {number} y - 光标 Y 坐标
   * @param {string} eventType - 事件类型
   */
  drawCursor(x, y, eventType) {
    const ctx = this.ctx;
    
    // 根据事件类型选择光标样式
    if (eventType && eventType.includes('down')) {
      // 点击状态 - 红色高亮
      ctx.fillStyle = 'rgba(255, 0, 0, 0.3)';
      ctx.beginPath();
      ctx.arc(x, y, 30, 0, Math.PI * 2);
      ctx.fill();
    }
    
    // 绘制箭头光标
    ctx.fillStyle = 'white';
    ctx.strokeStyle = 'black';
    ctx.lineWidth = 2;
    
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(x + 8, y + 20);
    ctx.lineTo(x + 14, y + 14);
    ctx.lineTo(x + 20, y + 8);
    ctx.closePath();
    
    ctx.fill();
    ctx.stroke();
  }

  /**
   * 等待视频跳转完成
   * @returns {Promise<void>}
   */
  waitForSeek() {
    return new Promise((resolve) => {
      const onSeeked = () => {
        this.video.removeEventListener('seeked', onSeeked);
        resolve();
      };
      this.video.addEventListener('seeked', onSeeked);
    });
  }

  /**
   * 清理资源
   */
  dispose() {
    if (this.video) {
      this.video.pause();
      this.video.src = '';
      this.video = null;
    }
  }
}

/**
 * 完整的导出流程控制器
 */
export class ExportController {
  constructor() {
    this.exportManager = new ExportManager();
    this.renderer = null;
    this.onProgress = null; // 进度回调
  }

  /**
   * 执行完整的导出流程
   * @param {Object} config - 导出配置
   * @param {Function} onProgress - 进度回调 (progress: 0-100)
   */
  async export(config, onProgress) {
    this.onProgress = onProgress;

    try {
      // Step 1: 准备导出
      this.updateProgress(0, '准备导出...');
      const exportInfo = await this.exportManager.prepareExport(
        config.videoPath,
        config.mouseDataPath,
        config.outputPath,
        config.screenWidth,
        config.screenHeight,
        config.fps || 30
      );

      // Step 2: 获取相机帧
      this.updateProgress(5, '生成相机路径...');
      const cameraFrames = await this.exportManager.getCameraFrames();

      if (cameraFrames.length === 0) {
        throw new Error('没有生成相机帧');
      }

      // Step 3: 创建渲染器
      this.updateProgress(10, '初始化渲染器...');
      const canvas = document.createElement('canvas');
      canvas.width = config.screenWidth;
      canvas.height = config.screenHeight;
      this.renderer = new CameraRenderer(canvas);

      // Step 4: 加载视频
      this.updateProgress(15, '加载视频...');
      await this.renderer.loadVideo(config.videoPath);

      // Step 5: 启动导出
      this.updateProgress(20, '启动导出流程...');
      await this.exportManager.startExport(config.outputPath, config.fps || 30);

      // Step 6: 渲染并导出每一帧
      const totalFrames = cameraFrames.length;
      for (let i = 0; i < totalFrames; i++) {
        const frame = cameraFrames[i];
        
        // 渲染帧
        const frameData = await this.renderer.renderFrame(
          frame,
          config.screenWidth,
          config.screenHeight,
          config.showCursor !== false
        );

        // 写入帧
        await this.exportManager.writeFrame(frameData);

        // 更新进度 (20% - 95%)
        const progress = 20 + (i / totalFrames) * 75;
        this.updateProgress(progress, `渲染中: ${i + 1}/${totalFrames}`);
      }

      // Step 7: 完成导出
      this.updateProgress(95, '完成导出...');
      await this.exportManager.finishExport();

      this.updateProgress(100, '导出完成！');

      return {
        success: true,
        outputPath: config.outputPath,
        totalFrames: totalFrames
      };

    } catch (error) {
      console.error('导出失败:', error);
      
      // 尝试停止导出
      try {
        await this.exportManager.stopExport();
      } catch (e) {
        // 忽略停止时的错误
      }

      throw error;
    } finally {
      // 清理资源
      if (this.renderer) {
        this.renderer.dispose();
        this.renderer = null;
      }
    }
  }

  /**
   * 更新进度
   * @param {number} progress - 进度 (0-100)
   * @param {string} message - 状态消息
   */
  updateProgress(progress, message) {
    console.log(`[${progress.toFixed(1)}%] ${message}`);
    if (this.onProgress) {
      this.onProgress(progress, message);
    }
  }

  /**
   * 取消导出
   */
  async cancel() {
    try {
      await this.exportManager.stopExport();
      if (this.renderer) {
        this.renderer.dispose();
      }
    } catch (error) {
      console.error('取消导出失败:', error);
    }
  }
}

// 默认导出
export default {
  ExportManager,
  CameraRenderer,
  ExportController
};
