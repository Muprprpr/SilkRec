<template>
  <div class="export-panel">
    <h2>导出视频 - 带相机运动效果</h2>
    
    <!-- 导出配置 -->
    <div class="config-section" v-if="!isExporting">
      <h3>配置</h3>
      
      <div class="form-group">
        <label>输入视频路径:</label>
        <input v-model="config.videoPath" type="text" placeholder="output/recording.mp4" />
      </div>
      
      <div class="form-group">
        <label>鼠标数据路径:</label>
        <input v-model="config.mouseDataPath" type="text" placeholder="output/mouse_events.json" />
      </div>
      
      <div class="form-group">
        <label>输出路径:</label>
        <input v-model="config.outputPath" type="text" placeholder="output/export.mp4" />
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label>屏幕宽度:</label>
          <input v-model.number="config.screenWidth" type="number" />
        </div>
        
        <div class="form-group">
          <label>屏幕高度:</label>
          <input v-model.number="config.screenHeight" type="number" />
        </div>
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label>帧率 (FPS):</label>
          <input v-model.number="config.fps" type="number" min="15" max="60" />
        </div>
        
        <div class="form-group checkbox">
          <label>
            <input v-model="config.showCursor" type="checkbox" />
            显示光标
          </label>
        </div>
      </div>
      
      <!-- 导出按钮 -->
      <div class="actions">
        <button @click="startExport" class="btn-primary" :disabled="!canExport">
          开始导出
        </button>
        
        <button @click="testConnection" class="btn-secondary">
          测试连接
        </button>
      </div>
    </div>
    
    <!-- 导出进度 -->
    <div class="progress-section" v-if="isExporting">
      <h3>导出中...</h3>
      
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: progress + '%' }"></div>
      </div>
      
      <div class="progress-info">
        <span>{{ progress.toFixed(1) }}%</span>
        <span>{{ statusMessage }}</span>
      </div>
      
      <button @click="cancelExport" class="btn-danger">
        取消导出
      </button>
    </div>
    
    <!-- 导出结果 -->
    <div class="result-section" v-if="exportResult">
      <h3>导出完成!</h3>
      
      <div class="result-info">
        <p><strong>输出文件:</strong> {{ exportResult.outputPath }}</p>
        <p><strong>总帧数:</strong> {{ exportResult.totalFrames }}</p>
      </div>
      
      <button @click="reset" class="btn-primary">
        导出新视频
      </button>
    </div>
    
    <!-- 错误信息 -->
    <div class="error-section" v-if="error">
      <h3>错误</h3>
      <p class="error-message">{{ error }}</p>
      <button @click="clearError" class="btn-secondary">
        清除错误
      </button>
    </div>
    
    <!-- 导出信息预览 -->
    <div class="info-section" v-if="exportInfo">
      <h3>导出信息</h3>
      <pre>{{ JSON.stringify(exportInfo, null, 2) }}</pre>
    </div>
  </div>
</template>

<script>
import { ExportController } from '../utils/exporter.js';

export default {
  name: 'ExportPanel',
  
  data() {
    return {
      // 导出配置
      config: {
        videoPath: 'output/recording.mp4',
        mouseDataPath: 'output/mouse_events.json',
        outputPath: 'output/export.mp4',
        screenWidth: 1920,
        screenHeight: 1080,
        fps: 30,
        showCursor: true
      },
      
      // 状态
      isExporting: false,
      progress: 0,
      statusMessage: '',
      error: null,
      exportResult: null,
      exportInfo: null,
      
      // 导出控制器
      controller: null
    };
  },
  
  computed: {
    canExport() {
      return this.config.videoPath && 
             this.config.mouseDataPath && 
             this.config.outputPath &&
             this.config.screenWidth > 0 &&
             this.config.screenHeight > 0;
    }
  },
  
  mounted() {
    // 初始化导出控制器
    this.controller = new ExportController();
    
    // 获取屏幕信息（通过 Wails 绑定的 Go 方法）
    this.getScreenInfo();
  },
  
  methods: {
    /**
     * 获取屏幕信息
     */
    async getScreenInfo() {
      try {
        // 调用 Wails 绑定的 Go 方法
        const [width, height, dpi] = await window.go.main.App.GetScreenInfo();
        this.config.screenWidth = width;
        this.config.screenHeight = height;
        console.log('屏幕信息:', { width, height, dpi });
      } catch (error) {
        console.error('获取屏幕信息失败:', error);
      }
    },
    
    /**
     * 开始导出
     */
    async startExport() {
      this.clearError();
      this.isExporting = true;
      this.progress = 0;
      this.statusMessage = '准备中...';
      this.exportResult = null;
      
      try {
        // 执行导出
        const result = await this.controller.export(
          this.config,
          (progress, message) => {
            this.progress = progress;
            this.statusMessage = message;
          }
        );
        
        this.exportResult = result;
        this.isExporting = false;
        
        console.log('导出成功:', result);
        
      } catch (error) {
        this.error = error.message || String(error);
        this.isExporting = false;
        console.error('导出失败:', error);
      }
    },
    
    /**
     * 取消导出
     */
    async cancelExport() {
      try {
        await this.controller.cancel();
        this.isExporting = false;
        this.statusMessage = '已取消';
      } catch (error) {
        console.error('取消导出失败:', error);
      }
    },
    
    /**
     * 测试 Wails 连接
     */
    async testConnection() {
      try {
        // 测试基本的 Wails 绑定
        const result = await window.go.main.App.Greet('Wails');
        alert('连接成功! ' + result);
        
        // 测试 FFmpeg 可用性
        const ffmpegAvailable = await window.go.main.App.CheckFFmpegAvailable();
        console.log('FFmpeg 可用:', ffmpegAvailable);
        
        if (!ffmpegAvailable) {
          alert('警告: FFmpeg 不可用! 请确保 ffmpeg.exe 在正确位置。');
        }
        
        // 获取导出信息
        const info = await window.go.main.App.GetExportInfo();
        this.exportInfo = info;
        console.log('导出信息:', info);
        
      } catch (error) {
        this.error = '连接测试失败: ' + (error.message || String(error));
        console.error('测试失败:', error);
      }
    },
    
    /**
     * 重置状态
     */
    reset() {
      this.isExporting = false;
      this.progress = 0;
      this.statusMessage = '';
      this.exportResult = null;
      this.exportInfo = null;
      this.clearError();
    },
    
    /**
     * 清除错误
     */
    clearError() {
      this.error = null;
    }
  }
};
</script>

<style scoped>
.export-panel {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

h2 {
  color: #2c3e50;
  margin-bottom: 20px;
}

h3 {
  color: #34495e;
  margin: 20px 0 15px 0;
  font-size: 18px;
}

.config-section,
.progress-section,
.result-section,
.error-section,
.info-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #555;
  font-weight: 500;
}

.form-group input[type="text"],
.form-group input[type="number"] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input[type="text"]:focus,
.form-group input[type="number"]:focus {
  outline: none;
  border-color: #3498db;
}

.form-row {
  display: flex;
  gap: 15px;
}

.form-row .form-group {
  flex: 1;
}

.form-group.checkbox {
  display: flex;
  align-items: center;
}

.form-group.checkbox label {
  margin: 0;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.form-group.checkbox input[type="checkbox"] {
  margin-right: 8px;
  cursor: pointer;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

button {
  padding: 12px 24px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
}

.btn-danger {
  background: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background: #c0392b;
}

.progress-bar {
  width: 100%;
  height: 30px;
  background: #ecf0f1;
  border-radius: 15px;
  overflow: hidden;
  margin: 20px 0;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3498db, #2ecc71);
  transition: width 0.3s ease;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  color: #555;
  margin-bottom: 15px;
}

.result-info p {
  margin: 10px 0;
  color: #555;
}

.error-section {
  background: #fff5f5;
  border: 1px solid #e74c3c;
}

.error-message {
  color: #c0392b;
  margin: 10px 0;
  font-family: monospace;
  white-space: pre-wrap;
}

.info-section pre {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  color: #333;
}
</style>
