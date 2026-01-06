/**
 * GPU 加速导出工具
 * 
 * 使用 FFmpeg 硬件加速和滤镜链直接处理视频
 * 无需前端渲染，效率最高
 */

/**
 * GPU 导出管理器
 * 简化的 API，直接调用后端 GPU 加速导出
 */
export class GPUExportManager {
  constructor() {
    this.isExporting = false;
  }

  /**
   * GPU 加速导出（推荐）
   * 
   * 这是最高效的导出方法，完全在 GPU 上处理
   * 使用 FFmpeg 硬件加速和滤镜链，无需前端渲染
   * 
   * @param {Object} config - 导出配置
   * @param {string} config.videoPath - 输入视频路径
   * @param {string} config.mouseDataPath - 鼠标数据 JSON 路径
   * @param {string} config.outputPath - 输出视频路径
   * @param {number} config.screenWidth - 屏幕宽度
   * @param {number} config.screenHeight - 屏幕高度
   * @param {number} config.fps - 帧率（默认 30）
   * @returns {Promise<void>}
   */
  async exportWithGPU(config) {
    try {
      console.log('🚀 开始 GPU 加速导出...', config);
      
      this.isExporting = true;
      
      // 调用后端 GPU 加速导出
      await window.go.main.App.ExportWithGPU(
        config.videoPath,
        config.mouseDataPath,
        config.outputPath,
        config.screenWidth,
        config.screenHeight,
        config.fps || 30
      );
      
      this.isExporting = false;
      
      console.log('✅ GPU 加速导出完成!');
      
      return {
        success: true,
        outputPath: config.outputPath
      };
      
    } catch (error) {
      this.isExporting = false;
      console.error('❌ GPU 导出失败:', error);
      throw new Error(`GPU 导出失败: ${error.message || error}`);
    }
  }

  /**
   * GPU 加速分段导出
   * 
   * 使用分段处理获得更精确的相机控制
   * 适合长视频或需要精确相机运动的场景
   * 
   * @param {Object} config - 导出配置（同 exportWithGPU）
   * @returns {Promise<void>}
   */
  async exportWithGPUSegmented(config) {
    try {
      console.log('🚀 开始 GPU 加速分段导出...', config);
      
      this.isExporting = true;
      
      // 调用后端 GPU 加速分段导出
      await window.go.main.App.ExportWithGPUSegmented(
        config.videoPath,
        config.mouseDataPath,
        config.outputPath,
        config.screenWidth,
        config.screenHeight,
        config.fps || 30
      );
      
      this.isExporting = false;
      
      console.log('✅ GPU 加速分段导出完成!');
      
      return {
        success: true,
        outputPath: config.outputPath
      };
      
    } catch (error) {
      this.isExporting = false;
      console.error('❌ GPU 分段导出失败:', error);
      throw new Error(`GPU 分段导出失败: ${error.message || error}`);
    }
  }

  /**
   * 停止 GPU 导出
   */
  async stop() {
    try {
      await window.go.main.App.StopGPUExport();
      this.isExporting = false;
      console.log('⏹️ GPU 导出已停止');
    } catch (error) {
      console.error('停止 GPU 导出失败:', error);
    }
  }

  /**
   * 获取导出进度
   * 注意：目前返回估算值，实际进度需要解析 FFmpeg 输出
   * @returns {Promise<number>}
   */
  async getProgress() {
    try {
      const progress = await window.go.main.App.GetGPUExportProgress();
      return progress;
    } catch (error) {
      console.error('获取进度失败:', error);
      return 0;
    }
  }

  /**
   * 检查是否正在导出
   */
  isExportingNow() {
    return this.isExporting;
  }
}

/**
 * 快速导出函数
 * 最简单的使用方式
 */
export async function quickExportWithGPU(videoPath, mouseDataPath, outputPath) {
  // 获取屏幕信息
  const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
  
  const manager = new GPUExportManager();
  
  return await manager.exportWithGPU({
    videoPath,
    mouseDataPath,
    outputPath,
    screenWidth,
    screenHeight,
    fps: 30
  });
}

/**
 * GPU 导出控制器（带进度监控）
 * 模拟进度反馈，因为 FFmpeg 输出解析较复杂
 */
export class GPUExportController {
  constructor() {
    this.manager = new GPUExportManager();
    this.progressInterval = null;
    this.estimatedDuration = 0;
    this.startTime = 0;
  }

  /**
   * 执行 GPU 导出（带进度估算）
   * 
   * @param {Object} config - 导出配置
   * @param {Function} onProgress - 进度回调 (progress: 0-100, message: string)
   * @param {boolean} useSegmented - 是否使用分段导出（默认 false）
   * @returns {Promise<Object>}
   */
  async export(config, onProgress, useSegmented = false) {
    try {
      // 估算时长（假设处理速度）
      // 实际速度取决于 GPU 性能，这里保守估计
      this.estimatedDuration = 10000; // 10 秒
      this.startTime = Date.now();
      
      // 开始模拟进度
      this.startProgressSimulation(onProgress);
      
      // 执行导出
      const result = useSegmented
        ? await this.manager.exportWithGPUSegmented(config)
        : await this.manager.exportWithGPU(config);
      
      // 停止进度模拟
      this.stopProgressSimulation();
      
      // 报告完成
      if (onProgress) {
        onProgress(100, '导出完成!');
      }
      
      return result;
      
    } catch (error) {
      this.stopProgressSimulation();
      throw error;
    }
  }

  /**
   * 开始进度模拟
   */
  startProgressSimulation(onProgress) {
    if (!onProgress) return;
    
    let lastProgress = 0;
    
    this.progressInterval = setInterval(() => {
      const elapsed = Date.now() - this.startTime;
      
      // 使用对数函数模拟进度（开始快，后面慢）
      let progress = Math.min(95, (Math.log(elapsed + 1) / Math.log(this.estimatedDuration + 1)) * 100);
      
      if (progress > lastProgress) {
        lastProgress = progress;
        
        // 生成状态消息
        let message = '处理中...';
        if (progress < 20) {
          message = '初始化 GPU 编码器...';
        } else if (progress < 50) {
          message = '应用相机变换...';
        } else if (progress < 80) {
          message = '硬件加速编码中...';
        } else {
          message = '最后处理...';
        }
        
        onProgress(progress, message);
      }
    }, 100); // 每 100ms 更新一次
  }

  /**
   * 停止进度模拟
   */
  stopProgressSimulation() {
    if (this.progressInterval) {
      clearInterval(this.progressInterval);
      this.progressInterval = null;
    }
  }

  /**
   * 取消导出
   */
  async cancel() {
    this.stopProgressSimulation();
    await this.manager.stop();
  }
}

/**
 * 性能对比工具
 * 比较 GPU 导出和 CPU 导出的性能差异
 */
export class ExportPerformanceComparator {
  constructor() {
    this.results = [];
  }

  /**
   * 测试 GPU 导出性能
   */
  async testGPUExport(config) {
    console.log('⏱️ 测试 GPU 导出性能...');
    
    const startTime = Date.now();
    const manager = new GPUExportManager();
    
    try {
      await manager.exportWithGPU(config);
      const duration = Date.now() - startTime;
      
      const result = {
        method: 'GPU',
        duration,
        success: true,
        speed: 'Fast'
      };
      
      this.results.push(result);
      console.log(`✅ GPU 导出完成: ${(duration / 1000).toFixed(2)}秒`);
      
      return result;
    } catch (error) {
      console.error('❌ GPU 导出失败:', error);
      return {
        method: 'GPU',
        duration: Date.now() - startTime,
        success: false,
        error: error.message
      };
    }
  }

  /**
   * 获取对比结果
   */
  getResults() {
    return this.results;
  }

  /**
   * 清除结果
   */
  clearResults() {
    this.results = [];
  }
}

// 默认导出
export default {
  GPUExportManager,
  GPUExportController,
  ExportPerformanceComparator,
  quickExportWithGPU
};

/**
 * 使用示例:
 * 
 * 1. 简单使用:
 *    import { quickExportWithGPU } from '@/utils/gpu-exporter.js';
 *    await quickExportWithGPU('input.mp4', 'mouse.json', 'output.mp4');
 * 
 * 2. 带进度:
 *    import { GPUExportController } from '@/utils/gpu-exporter.js';
 *    const controller = new GPUExportController();
 *    await controller.export(config, (progress, message) => {
 *      console.log(`${progress}%: ${message}`);
 *    });
 * 
 * 3. 分段导出:
 *    await controller.export(config, onProgress, true);
 */
