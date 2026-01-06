/**
 * SilkRec 导出功能使用示例
 * 
 * 这个文件展示如何在 Wails 架构下使用相机运动导出功能
 * 可以直接在浏览器控制台运行这些示例
 */

/**
 * 示例 1: 基础导出流程
 * 
 * 这是最简单的导出示例，展示完整流程
 */
export async function basicExportExample() {
  console.log('=== 基础导出示例 ===');
  
  try {
    // Step 1: 检查 FFmpeg 是否可用
    const ffmpegAvailable = await window.go.main.App.CheckFFmpegAvailable();
    if (!ffmpegAvailable) {
      throw new Error('FFmpeg 不可用！请确保 ffmpeg.exe 在正确位置。');
    }
    console.log('✓ FFmpeg 可用');
    
    // Step 2: 获取屏幕信息
    const [screenWidth, screenHeight, dpi] = await window.go.main.App.GetScreenInfo();
    console.log(`✓ 屏幕信息: ${screenWidth}x${screenHeight}, DPI: ${dpi}`);
    
    // Step 3: 准备导出
    console.log('准备导出...');
    const exportInfo = await window.go.main.App.PrepareExport(
      'output/recording.mp4',      // 输入视频
      'output/mouse_events.json',  // 鼠标数据
      'output/export.mp4',         // 输出路径
      screenWidth,
      screenHeight,
      30                           // FPS
    );
    console.log('✓ 导出信息:', exportInfo);
    
    // Step 4: 获取相机帧
    console.log('获取相机帧...');
    const framesJSON = await window.go.main.App.GetCameraFrames();
    const frames = JSON.parse(framesJSON);
    console.log(`✓ 获取到 ${frames.length} 个相机帧`);
    
    // Step 5: 渲染和导出（这里仅展示逻辑，完整实现见 exporter.js）
    console.log('开始渲染和导出...');
    console.log('注意: 完整的渲染实现请参考 frontend/src/utils/exporter.js');
    console.log('示例完成！');
    
    return {
      success: true,
      exportInfo,
      totalFrames: frames.length
    };
    
  } catch (error) {
    console.error('❌ 导出失败:', error);
    throw error;
  }
}

/**
 * 示例 2: 使用 ExportController 完整导出
 * 
 * 使用我们封装的 ExportController 类进行完整导出
 */
export async function fullExportExample() {
  console.log('=== 完整导出示例 ===');
  
  // 动态导入 ExportController
  const { ExportController } = await import('../utils/exporter.js');
  
  const controller = new ExportController();
  
  try {
    // 获取屏幕信息
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    // 配置
    const config = {
      videoPath: 'output/recording.mp4',
      mouseDataPath: 'output/mouse_events.json',
      outputPath: 'output/export_full.mp4',
      screenWidth,
      screenHeight,
      fps: 30,
      showCursor: true
    };
    
    // 执行导出，带进度回调
    const result = await controller.export(config, (progress, message) => {
      console.log(`[${progress.toFixed(1)}%] ${message}`);
    });
    
    console.log('✓ 导出完成:', result);
    return result;
    
  } catch (error) {
    console.error('❌ 导出失败:', error);
    throw error;
  }
}

/**
 * 示例 3: 调试 - 保存相机路径
 * 
 * 生成并保存相机路径到文件，用于调试
 */
export async function debugCameraPathExample() {
  console.log('=== 调试相机路径示例 ===');
  
  try {
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    // 准备导出（生成相机路径）
    const exportInfo = await window.go.main.App.PrepareExport(
      'output/recording.mp4',
      'output/mouse_events.json',
      'output/debug_export.mp4',
      screenWidth,
      screenHeight,
      30
    );
    console.log('导出信息:', exportInfo);
    
    // 保存相机路径到文件
    await window.go.main.App.SaveCameraPath('output/camera_debug.json');
    console.log('✓ 相机路径已保存到: output/camera_debug.json');
    
    // 获取相机帧并显示前几个
    const framesJSON = await window.go.main.App.GetCameraFrames();
    const frames = JSON.parse(framesJSON);
    
    console.log('前 5 个相机帧:');
    frames.slice(0, 5).forEach((frame, i) => {
      console.log(`  帧 ${i}:`, {
        timestamp: frame.Timestamp + 'ms',
        camera: `(${frame.X.toFixed(1)}, ${frame.Y.toFixed(1)})`,
        zoom: frame.Zoom.toFixed(2),
        mouse: `(${frame.MouseX}, ${frame.MouseY})`,
        event: frame.EventType
      });
    });
    
    return { success: true, frameCount: frames.length };
    
  } catch (error) {
    console.error('❌ 调试失败:', error);
    throw error;
  }
}

/**
 * 示例 4: 测试所有 Wails 绑定
 * 
 * 验证所有导出相关的 API 是否正常工作
 */
export async function testAllBindings() {
  console.log('=== 测试所有 Wails 绑定 ===');
  
  const tests = [];
  
  // Test 1: Greet
  try {
    const greeting = await window.go.main.App.Greet('Wails');
    console.log('✓ Greet:', greeting);
    tests.push({ name: 'Greet', passed: true });
  } catch (error) {
    console.error('✗ Greet 失败:', error);
    tests.push({ name: 'Greet', passed: false, error });
  }
  
  // Test 2: CheckFFmpegAvailable
  try {
    const available = await window.go.main.App.CheckFFmpegAvailable();
    console.log('✓ CheckFFmpegAvailable:', available);
    tests.push({ name: 'CheckFFmpegAvailable', passed: true, result: available });
  } catch (error) {
    console.error('✗ CheckFFmpegAvailable 失败:', error);
    tests.push({ name: 'CheckFFmpegAvailable', passed: false, error });
  }
  
  // Test 3: GetScreenInfo
  try {
    const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
    console.log('✓ GetScreenInfo:', { width, height, dpi });
    tests.push({ name: 'GetScreenInfo', passed: true });
  } catch (error) {
    console.error('✗ GetScreenInfo 失败:', error);
    tests.push({ name: 'GetScreenInfo', passed: false, error });
  }
  
  // Test 4: GetExportInfo
  try {
    const info = await window.go.main.App.GetExportInfo();
    console.log('✓ GetExportInfo:', info);
    tests.push({ name: 'GetExportInfo', passed: true });
  } catch (error) {
    console.error('✗ GetExportInfo 失败:', error);
    tests.push({ name: 'GetExportInfo', passed: false, error });
  }
  
  // 总结
  const passed = tests.filter(t => t.passed).length;
  const total = tests.length;
  console.log(`\n测试结果: ${passed}/${total} 通过`);
  
  if (passed === total) {
    console.log('✓ 所有测试通过！Wails 绑定工作正常。');
  } else {
    console.warn('⚠ 部分测试失败，请检查 Wails 配置。');
  }
  
  return { passed, total, tests };
}

/**
 * 示例 5: 单帧渲染演示
 * 
 * 展示如何渲染单个相机帧（不导出）
 */
export async function renderSingleFrameExample() {
  console.log('=== 单帧渲染示例 ===');
  
  try {
    const [screenWidth, screenHeight] = await window.go.main.App.GetScreenInfo();
    
    // 准备导出
    await window.go.main.App.PrepareExport(
      'output/recording.mp4',
      'output/mouse_events.json',
      'output/temp.mp4',
      screenWidth,
      screenHeight,
      30
    );
    
    // 获取相机帧
    const framesJSON = await window.go.main.App.GetCameraFrames();
    const frames = JSON.parse(framesJSON);
    
    if (frames.length === 0) {
      throw new Error('没有相机帧');
    }
    
    // 选择中间的一帧
    const frame = frames[Math.floor(frames.length / 2)];
    console.log('选择的帧:', frame);
    
    // 创建 canvas
    const canvas = document.createElement('canvas');
    canvas.width = screenWidth;
    canvas.height = screenHeight;
    canvas.style.border = '2px solid red';
    const ctx = canvas.getContext('2d');
    
    // 加载视频
    const video = document.createElement('video');
    video.src = '/output/recording.mp4';
    
    await new Promise((resolve) => {
      video.onloadedmetadata = resolve;
    });
    
    // 跳转到指定时间
    video.currentTime = frame.Timestamp / 1000.0;
    
    await new Promise((resolve) => {
      video.onseeked = resolve;
    });
    
    // 计算视口
    const viewportWidth = screenWidth / frame.Zoom;
    const viewportHeight = screenHeight / frame.Zoom;
    const viewportX = frame.X - viewportWidth / 2;
    const viewportY = frame.Y - viewportHeight / 2;
    
    // 渲染
    ctx.save();
    ctx.scale(frame.Zoom, frame.Zoom);
    ctx.translate(-viewportX, -viewportY);
    ctx.drawImage(video, 0, 0, screenWidth, screenHeight);
    ctx.restore();
    
    // 绘制光标
    ctx.fillStyle = 'white';
    ctx.strokeStyle = 'black';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(frame.MouseX, frame.MouseY);
    ctx.lineTo(frame.MouseX + 8, frame.MouseY + 20);
    ctx.lineTo(frame.MouseX + 14, frame.MouseY + 14);
    ctx.lineTo(frame.MouseX + 20, frame.MouseY + 8);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
    
    // 显示在页面上
    document.body.appendChild(canvas);
    
    console.log('✓ 帧已渲染并添加到页面');
    console.log('Canvas 已添加到页面底部，滚动查看');
    
    return { success: true, canvas };
    
  } catch (error) {
    console.error('❌ 渲染失败:', error);
    throw error;
  }
}

/**
 * 快速测试函数
 * 在控制台运行: quickTest()
 */
export async function quickTest() {
  console.log('=== 快速测试 ===\n');
  
  try {
    // 测试绑定
    console.log('1. 测试 Wails 绑定...');
    await testAllBindings();
    
    console.log('\n2. 测试相机路径生成...');
    await debugCameraPathExample();
    
    console.log('\n✓ 快速测试完成！');
    console.log('运行完整导出: fullExportExample()');
    console.log('渲染单帧: renderSingleFrameExample()');
    
  } catch (error) {
    console.error('✗ 测试失败:', error);
  }
}

// 导出所有示例函数
export default {
  basicExportExample,
  fullExportExample,
  debugCameraPathExample,
  testAllBindings,
  renderSingleFrameExample,
  quickTest
};

/**
 * 使用说明:
 * 
 * 1. 在 Wails 应用中打开浏览器控制台 (F12)
 * 
 * 2. 导入示例:
 *    import examples from './examples/export-example.js'
 * 
 * 3. 运行示例:
 *    examples.quickTest()                  // 快速测试
 *    examples.testAllBindings()            // 测试所有 API
 *    examples.debugCameraPathExample()     // 生成相机路径
 *    examples.fullExportExample()          // 完整导出
 *    examples.renderSingleFrameExample()   // 渲染单帧
 * 
 * 或者直接运行:
 *    window.examples = examples
 *    examples.quickTest()
 */
