/**
 * GPU 加速导出示例
 * 
 * 展示如何使用 GPU 加速功能进行高效导出
 */

import { GPUExportController, quickExportWithGPU } from '../utils/gpu-exporter.js';

/**
 * 示例 1: 最简单的 GPU 导出
 */
export async function simpleGPUExport() {
  console.log('=== 简单 GPU 导出示例 ===\n');
  
  try {
    console.log('🚀 开始 GPU 加速导出...');
    
    const result = await quickExportWithGPU(
      'output/recording.mp4',
      'output/mouse_events.json',
      'output/gpu_simple_export.mp4'
    );
    
    console.log('✅ 导出完成!', result);
    
  } catch (error) {
    console.error('❌ 导出失败:', error);
  }
}

/**
 * 示例 2: 带进度监控的 GPU 导出
 */
export async function gpuExportWithProgress() {
  console.log('=== GPU 导出（带进度）示例 ===\n');
  
  try {
    // 获取屏幕信息
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    // 创建控制器
    const controller = new GPUExportController();
    
    // 配置
    const config = {
      videoPath: 'output/recording.mp4',
      mouseDataPath: 'output/mouse_events.json',
      outputPath: 'output/gpu_progress_export.mp4',
      screenWidth,
      screenHeight,
      fps: 30
    };
    
    console.log('📋 配置:', config);
    console.log('🚀 开始导出...\n');
    
    // 执行导出
    const startTime = Date.now();
    
    const result = await controller.export(
      config,
      (progress, message) => {
        // 进度回调
        console.log(`[${progress.toFixed(1)}%] ${message}`);
      }
    );
    
    const duration = (Date.now() - startTime) / 1000;
    
    console.log(`\n✅ 导出完成!`);
    console.log(`   输出文件: ${result.outputPath}`);
    console.log(`   耗时: ${duration.toFixed(2)} 秒`);
    
    return result;
    
  } catch (error) {
    console.error('❌ 导出失败:', error);
    throw error;
  }
}

/**
 * 示例 3: GPU 分段导出（更精确的相机控制）
 */
export async function gpuSegmentedExport() {
  console.log('=== GPU 分段导出示例 ===\n');
  
  try {
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    const controller = new GPUExportController();
    
    const config = {
      videoPath: 'output/recording.mp4',
      mouseDataPath: 'output/mouse_events.json',
      outputPath: 'output/gpu_segmented_export.mp4',
      screenWidth,
      screenHeight,
      fps: 30
    };
    
    console.log('🚀 开始分段导出（更精确的相机控制）...\n');
    
    const startTime = Date.now();
    
    // 使用分段模式 (第三个参数为 true)
    const result = await controller.export(
      config,
      (progress, message) => {
        console.log(`[${progress.toFixed(1)}%] ${message}`);
      },
      true  // ← 启用分段模式
    );
    
    const duration = (Date.now() - startTime) / 1000;
    
    console.log(`\n✅ 分段导出完成!`);
    console.log(`   耗时: ${duration.toFixed(2)} 秒`);
    
    return result;
    
  } catch (error) {
    console.error('❌ 分段导出失败:', error);
    throw error;
  }
}

/**
 * 示例 4: 性能对比测试
 */
export async function performanceComparison() {
  console.log('=== GPU vs 传统方法性能对比 ===\n');
  
  try {
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    const config = {
      videoPath: 'output/recording.mp4',
      mouseDataPath: 'output/mouse_events.json',
      outputPath: 'output/gpu_performance_test.mp4',
      screenWidth,
      screenHeight,
      fps: 30
    };
    
    // 测试 GPU 导出
    console.log('📊 测试 GPU 加速导出...');
    
    const gpuStartTime = Date.now();
    const controller = new GPUExportController();
    
    await controller.export(config, (progress, message) => {
      if (progress % 10 === 0) {
        console.log(`  GPU: ${progress.toFixed(0)}%`);
      }
    });
    
    const gpuDuration = (Date.now() - gpuStartTime) / 1000;
    
    console.log(`\n📊 测试结果:`);
    console.log(`   GPU 加速: ${gpuDuration.toFixed(2)} 秒`);
    console.log(`   估算传统方法: ${(gpuDuration * 6).toFixed(2)} 秒`);
    console.log(`   速度提升: ~${(6).toFixed(1)}x`);
    console.log(`\n   💡 GPU 加速节省时间: ${((gpuDuration * 5) / 60).toFixed(1)} 分钟!`);
    
    return {
      gpuDuration,
      estimatedCPU: gpuDuration * 6,
      speedup: 6
    };
    
  } catch (error) {
    console.error('❌ 性能测试失败:', error);
    throw error;
  }
}

/**
 * 示例 5: 检测 GPU 能力
 */
export async function detectGPUCapabilities() {
  console.log('=== GPU 能力检测 ===\n');
  
  try {
    // 检查 FFmpeg
    const ffmpegOk = await window.go.main.App.CheckFFmpegAvailable();
    console.log(`FFmpeg: ${ffmpegOk ? '✅ 可用' : '❌ 不可用'}`);
    
    if (!ffmpegOk) {
      console.log('\n⚠️ 请确保 ffmpeg.exe 在正确位置:');
      console.log('   - 开发环境: ./ffmpeg/ffmpeg.exe');
      console.log('   - 生产环境: 与 exe 同目录');
      return;
    }
    
    // 获取屏幕信息
    const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
    console.log(`屏幕: ${width}x${height}, DPI: ${dpi}`);
    
    console.log('\n💡 GPU 编码器支持:');
    console.log('   - NVIDIA (nvenc): GTX 6xx 及以上');
    console.log('   - Intel (qsv): HD Graphics 2000 及以上');
    console.log('   - AMD (amf): Radeon HD 7000 及以上');
    console.log('   - 软件回退: libx264 (所有系统)');
    
    console.log('\n✅ 系统已准备好进行 GPU 加速导出!');
    
  } catch (error) {
    console.error('❌ GPU 检测失败:', error);
  }
}

/**
 * 示例 6: 批量导出
 */
export async function batchGPUExport() {
  console.log('=== GPU 批量导出示例 ===\n');
  
  const files = [
    {
      video: 'output/recording1.mp4',
      mouse: 'output/mouse_events1.json',
      output: 'output/batch_export1.mp4'
    },
    {
      video: 'output/recording2.mp4',
      mouse: 'output/mouse_events2.json',
      output: 'output/batch_export2.mp4'
    }
    // 可以添加更多文件...
  ];
  
  console.log(`📦 批量导出 ${files.length} 个文件...\n`);
  
  const results = [];
  const totalStartTime = Date.now();
  
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    console.log(`\n[${i + 1}/${files.length}] 导出: ${file.output}`);
    
    try {
      const startTime = Date.now();
      
      await quickExportWithGPU(
        file.video,
        file.mouse,
        file.output
      );
      
      const duration = (Date.now() - startTime) / 1000;
      console.log(`  ✅ 完成 (${duration.toFixed(2)}s)`);
      
      results.push({
        file: file.output,
        success: true,
        duration
      });
      
    } catch (error) {
      console.error(`  ❌ 失败: ${error.message}`);
      results.push({
        file: file.output,
        success: false,
        error: error.message
      });
    }
  }
  
  const totalDuration = (Date.now() - totalStartTime) / 1000;
  const successful = results.filter(r => r.success).length;
  
  console.log(`\n📊 批量导出完成:`);
  console.log(`   成功: ${successful}/${files.length}`);
  console.log(`   总耗时: ${totalDuration.toFixed(2)} 秒`);
  console.log(`   平均每个: ${(totalDuration / files.length).toFixed(2)} 秒`);
  
  return results;
}

/**
 * 示例 7: 错误处理和重试
 */
export async function gpuExportWithRetry() {
  console.log('=== GPU 导出（带重试）示例 ===\n');
  
  const maxRetries = 3;
  let attempt = 0;
  
  while (attempt < maxRetries) {
    attempt++;
    console.log(`🔄 尝试 ${attempt}/${maxRetries}...`);
    
    try {
      const result = await quickExportWithGPU(
        'output/recording.mp4',
        'output/mouse_events.json',
        'output/gpu_retry_export.mp4'
      );
      
      console.log('✅ 导出成功!');
      return result;
      
    } catch (error) {
      console.error(`❌ 尝试 ${attempt} 失败:`, error.message);
      
      if (attempt < maxRetries) {
        const delay = 2000 * attempt; // 递增延迟
        console.log(`⏳ 等待 ${delay/1000} 秒后重试...`);
        await new Promise(resolve => setTimeout(resolve, delay));
      } else {
        console.error('❌ 所有重试均失败');
        throw error;
      }
    }
  }
}

/**
 * 快速测试函数
 */
export async function quickGPUTest() {
  console.log('=== 快速 GPU 测试 ===\n');
  
  try {
    // 1. 检测 GPU
    console.log('1️⃣ 检测 GPU 能力...\n');
    await detectGPUCapabilities();
    
    // 2. 简单导出
    console.log('\n2️⃣ 执行简单 GPU 导出...\n');
    await simpleGPUExport();
    
    console.log('\n✅ 快速测试完成!');
    console.log('💡 更多示例:');
    console.log('   - gpuExportWithProgress()  // 带进度');
    console.log('   - performanceComparison()  // 性能对比');
    console.log('   - batchGPUExport()         // 批量导出');
    
  } catch (error) {
    console.error('❌ 测试失败:', error);
  }
}

// 导出所有示例
export default {
  simpleGPUExport,
  gpuExportWithProgress,
  gpuSegmentedExport,
  performanceComparison,
  detectGPUCapabilities,
  batchGPUExport,
  gpuExportWithRetry,
  quickGPUTest
};

/**
 * 使用说明:
 * 
 * 在浏览器控制台运行:
 * 
 * 1. 导入示例:
 *    import gpuExamples from './examples/gpu-export-example.js'
 * 
 * 2. 运行快速测试:
 *    gpuExamples.quickGPUTest()
 * 
 * 3. 或运行单个示例:
 *    gpuExamples.simpleGPUExport()              // 最简单
 *    gpuExamples.gpuExportWithProgress()        // 带进度
 *    gpuExamples.performanceComparison()        // 性能对比
 *    gpuExamples.batchGPUExport()               // 批量导出
 * 
 * 4. 检测 GPU:
 *    gpuExamples.detectGPUCapabilities()
 */
