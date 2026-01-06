<template>
  <div class="video-editor">
    <div class="editor-header">
      <h1>🎬 SilkRec 视频编辑器</h1>
      <div class="header-actions">
        <button @click="loadProject" class="btn-secondary">加载项目</button>
        <button @click="exportVideo" class="btn-primary" :disabled="!canExport">
          导出视频
        </button>
      </div>
    </div>

    <!-- 预览区域 -->
    <div class="preview-section">
      <div class="preview-container" :style="previewContainerStyle">
        <canvas 
          ref="previewCanvas" 
          :width="config.screenWidth" 
          :height="config.screenHeight"
          :style="canvasStyle"
        ></canvas>
        
        <!-- 鼠标光标预览 -->
        <div 
          class="cursor-preview" 
          v-if="showCursorPreview"
          :style="cursorPreviewStyle"
        >
          <img 
            :src="currentCursorImage" 
            :style="cursorImageStyle"
            alt="光标"
          />
        </div>
      </div>
      
      <div class="preview-controls">
        <button @click="playPreview" v-if="!isPlaying">▶️ 播放预览</button>
        <button @click="pausePreview" v-else>⏸️ 暂停</button>
        <button @click="stopPreview">⏹️ 停止</button>
        <span class="timecode">{{ currentTime }} / {{ totalDuration }}</span>
      </div>
    </div>

    <!-- 参数控制面板 -->
    <div class="parameters-panel">
      <h3>⚙️ 参数控制</h3>
      
      <div class="parameter-section">
        <h4>相机运动</h4>
        
        <div class="parameter-item">
          <label>
            <span>平滑强度</span>
            <span class="value">{{ animationParams.smoothness.toFixed(2) }}</span>
          </label>
          <input 
            type="range" 
            v-model.number="animationParams.smoothness"
            min="0.01" 
            max="0.5" 
            step="0.01"
            class="slider"
          />
          <div class="param-hint">值越小越平滑，越大越灵敏</div>
        </div>

        <div class="parameter-item">
          <label>
            <span>缩放强度</span>
            <span class="value">{{ animationParams.zoomLevel.toFixed(2) }}x</span>
          </label>
          <input 
            type="range" 
            v-model.number="animationParams.zoomLevel"
            min="1.0" 
            max="3.0" 
            step="0.1"
            class="slider"
          />
          <div class="param-hint">点击时的缩放倍数</div>
        </div>

        <div class="parameter-item">
          <label>
            <span>运动速度</span>
            <span class="value">{{ animationParams.speed.toFixed(2) }}</span>
          </label>
          <input 
            type="range" 
            v-model.number="animationParams.speed"
            min="0.5" 
            max="2.0" 
            step="0.1"
            class="slider"
          />
          <div class="param-hint">相机跟随速度</div>
        </div>
      </div>

      <div class="parameter-section">
        <h4>视频大小</h4>
        
        <div class="parameter-item">
          <label>
            <span>录屏画面大小</span>
            <span class="value">{{ animationParams.videoScale * 100 }}%</span>
          </label>
          <div class="scale-buttons">
            <button 
              @click="animationParams.videoScale = 0.8"
              :class="{ active: animationParams.videoScale === 0.8 }"
            >
              80%
            </button>
            <button 
              @click="animationParams.videoScale = 0.9"
              :class="{ active: animationParams.videoScale === 0.9 }"
            >
              90%
            </button>
            <button 
              @click="animationParams.videoScale = 1.0"
              :class="{ active: animationParams.videoScale === 1.0 }"
            >
              100%
            </button>
          </div>
          <div class="param-hint">录屏内容在画面中的占比</div>
        </div>
      </div>

      <div class="parameter-section">
        <h4>光标设置</h4>
        
        <div class="parameter-item">
          <label>
            <span>光标样式</span>
          </label>
          <div class="cursor-selector">
            <div 
              v-for="cursor in cursorOptions" 
              :key="cursor.id"
              @click="selectCursor(cursor)"
              :class="['cursor-option', { active: selectedCursor.id === cursor.id }]"
            >
              <img :src="cursor.preview" :alt="cursor.name" />
              <span>{{ cursor.name }}</span>
            </div>
            <div class="cursor-option upload" @click="uploadCustomCursor">
              <div class="upload-icon">+</div>
              <span>上传</span>
            </div>
          </div>
        </div>

        <div class="parameter-item">
          <label>
            <span>光标大小</span>
            <span class="value">{{ animationParams.cursorSize }}px</span>
          </label>
          <input 
            type="range" 
            v-model.number="animationParams.cursorSize"
            min="16" 
            max="64" 
            step="4"
            class="slider"
          />
        </div>

        <div class="parameter-item">
          <label>
            <input type="checkbox" v-model="animationParams.showClickEffect" />
            显示点击效果
          </label>
        </div>
      </div>

      <div class="parameter-section">
        <h4>背景设置</h4>
        
        <div class="parameter-item">
          <label>
            <span>背景类型</span>
          </label>
          <div class="background-type-selector">
            <button 
              @click="backgroundType = 'solid'"
              :class="{ active: backgroundType === 'solid' }"
            >
              纯色
            </button>
            <button 
              @click="backgroundType = 'gradient'"
              :class="{ active: backgroundType === 'gradient' }"
            >
              渐变
            </button>
            <button 
              @click="backgroundType = 'image'"
              :class="{ active: backgroundType === 'image' }"
            >
              图片
            </button>
          </div>
        </div>

        <div class="parameter-item" v-if="backgroundType === 'solid'">
          <label>背景颜色</label>
          <input type="color" v-model="backgroundColor" class="color-picker" />
        </div>

        <div class="parameter-item" v-if="backgroundType === 'gradient'">
          <label>渐变颜色 1</label>
          <input type="color" v-model="gradientColor1" class="color-picker" />
          <label>渐变颜色 2</label>
          <input type="color" v-model="gradientColor2" class="color-picker" />
        </div>

        <div class="parameter-item" v-if="backgroundType === 'image'">
          <label>背景图片</label>
          <div class="image-upload">
            <img v-if="backgroundImage" :src="backgroundImage" class="bg-preview" />
            <button @click="uploadBackgroundImage" class="upload-btn">
              {{ backgroundImage ? '更换图片' : '上传图片' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 时间轴编辑器 -->
    <div class="timeline-editor">
      <h3>🎞️ 时间轴编辑器</h3>
      
      <div class="timeline-container">
        <!-- 时间刻度 -->
        <div class="timeline-ruler">
          <div 
            v-for="n in timelineSegments" 
            :key="n"
            class="time-marker"
            :style="{ left: (n / timelineSegments * 100) + '%' }"
          >
            {{ formatTime(n / timelineSegments * totalDurationMs) }}
          </div>
        </div>

        <!-- 视频轨道 -->
        <div class="track video-track">
          <div class="track-header">
            <span class="track-icon">🎥</span>
            <span class="track-name">视频轨道</span>
            <button @click="toggleTrack('video')" class="track-toggle">
              {{ tracks.video.enabled ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>
          <div class="track-content">
            <div 
              class="track-clip video-clip"
              :style="{ width: '100%' }"
            >
              <span class="clip-name">{{ videoFileName }}</span>
              <div class="clip-duration">{{ totalDuration }}</div>
            </div>
          </div>
        </div>

        <!-- 效果轨道 -->
        <div class="track effect-track">
          <div class="track-header">
            <span class="track-icon">✨</span>
            <span class="track-name">效果轨道</span>
            <button @click="toggleTrack('effect')" class="track-toggle">
              {{ tracks.effect.enabled ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>
          <div class="track-content">
            <!-- 相机运动效果 -->
            <div 
              v-for="effect in effects" 
              :key="effect.id"
              class="track-clip effect-clip"
              :style="effectClipStyle(effect)"
              @mousedown="startDragEffect($event, effect)"
            >
              <span class="clip-name">{{ effect.name }}</span>
              <div class="effect-indicator" :class="effect.type"></div>
            </div>
            
            <!-- 添加效果按钮 -->
            <button @click="addEffect" class="add-effect-btn">
              + 添加效果
            </button>
          </div>
        </div>

        <!-- 鼠标轨道（只读，显示鼠标事件） -->
        <div class="track mouse-track">
          <div class="track-header">
            <span class="track-icon">🖱️</span>
            <span class="track-name">鼠标事件</span>
          </div>
          <div class="track-content">
            <div 
              v-for="(event, index) in mouseEventMarkers" 
              :key="index"
              class="mouse-event-marker"
              :style="{ left: (event.timestamp / totalDurationMs * 100) + '%' }"
              :class="event.type"
              :title="`${event.type} at ${formatTime(event.timestamp)}`"
            ></div>
          </div>
        </div>

        <!-- 播放头 -->
        <div 
          class="playhead" 
          :style="{ left: playheadPosition + '%' }"
          @mousedown="startDragPlayhead"
        ></div>
      </div>
    </div>

    <!-- 导出设置对话框 -->
    <div v-if="showExportDialog" class="modal-overlay" @click="showExportDialog = false">
      <div class="modal-content" @click.stop>
        <h3>导出设置</h3>
        
        <div class="form-group">
          <label>输出路径</label>
          <input v-model="exportConfig.outputPath" type="text" />
        </div>

        <div class="form-group">
          <label>帧率 (FPS)</label>
          <input v-model.number="exportConfig.fps" type="number" min="15" max="60" />
        </div>

        <div class="form-group">
          <label>质量</label>
          <select v-model="exportConfig.quality">
            <option value="high">高质量</option>
            <option value="medium">中等质量</option>
            <option value="low">低质量（快速）</option>
          </select>
        </div>

        <div class="form-group">
          <label>
            <input type="checkbox" v-model="exportConfig.useGPU" />
            使用 GPU 加速
          </label>
        </div>

        <div class="modal-actions">
          <button @click="showExportDialog = false" class="btn-secondary">取消</button>
          <button @click="confirmExport" class="btn-primary">开始导出</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'VideoEditor',
  
  data() {
    return {
      // 项目配置
      config: {
        videoPath: 'output/recording.mp4',
        mouseDataPath: 'output/mouse_events.json',
        screenWidth: 1920,
        screenHeight: 1080,
      },
      
      // 动画参数
      animationParams: {
        smoothness: 0.15,      // 平滑强度 (0.01-0.5)
        zoomLevel: 1.5,        // 缩放倍数 (1.0-3.0)
        speed: 1.0,            // 运动速度 (0.5-2.0)
        videoScale: 1.0,       // 视频画面大小 (0.8 or 1.0)
        cursorSize: 32,        // 光标大小 (16-64)
        showClickEffect: true, // 显示点击效果
      },
      
      // 光标选项
      cursorOptions: [
        { id: 'default', name: '默认', preview: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0iTTAgMGwyNCAyNC0xMC0yLTQgMTB6IiBmaWxsPSIjZmZmIi8+PC9zdmc+' },
        { id: 'pointer', name: '手型', preview: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMTIiIGN5PSIxMiIgcj0iMTAiIGZpbGw9IiNmZmYiLz48L3N2Zz4=' },
        { id: 'circle', name: '圆形', preview: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMTIiIGN5PSIxMiIgcj0iOCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZmYwMDAwIiBzdHJva2Utd2lkdGg9IjIiLz48L3N2Zz4=' },
      ],
      selectedCursor: null,
      
      // 背景设置
      backgroundType: 'gradient', // 'solid', 'gradient', 'image'
      backgroundColor: '#1a2a6c',
      gradientColor1: '#1a2a6c',
      gradientColor2: '#2c3e50',
      backgroundImage: null,
      
      // 时间轴
      tracks: {
        video: { enabled: true },
        effect: { enabled: true },
      },
      effects: [
        { id: 1, name: '相机跟随', type: 'camera', start: 0, duration: 100 },
        { id: 2, name: '点击缩放', type: 'zoom', start: 20, duration: 30 },
      ],
      mouseEventMarkers: [],
      
      // 播放控制
      isPlaying: false,
      currentTime: '00:00',
      totalDuration: '00:00',
      totalDurationMs: 0,
      playheadPosition: 0,
      timelineSegments: 10,
      
      // 预览
      showCursorPreview: false,
      videoFileName: 'recording.mp4',
      
      // 导出
      showExportDialog: false,
      exportConfig: {
        outputPath: 'output/final_export.mp4',
        fps: 30,
        quality: 'high',
        useGPU: true,
      },
    };
  },
  
  computed: {
    canExport() {
      return this.config.videoPath && this.config.mouseDataPath;
    },
    
    previewContainerStyle() {
      return {
        background: this.getBackgroundStyle(),
      };
    },
    
    canvasStyle() {
      const scale = this.animationParams.videoScale;
      return {
        transform: `scale(${scale})`,
        transformOrigin: 'center',
      };
    },
    
    currentCursorImage() {
      return this.selectedCursor ? this.selectedCursor.preview : this.cursorOptions[0].preview;
    },
    
    cursorPreviewStyle() {
      return {
        width: this.animationParams.cursorSize + 'px',
        height: this.animationParams.cursorSize + 'px',
      };
    },
    
    cursorImageStyle() {
      return {
        width: '100%',
        height: '100%',
      };
    },
  },
  
  mounted() {
    this.selectedCursor = this.cursorOptions[0];
    this.loadInitialData();
  },
  
  methods: {
    async loadInitialData() {
      try {
        // 获取屏幕信息
        const [width, height] = await window.go.main.App.GetScreenInfo();
        this.config.screenWidth = width;
        this.config.screenHeight = height;
        
        // 加载鼠标事件标记
        // await this.loadMouseEvents();
      } catch (error) {
        console.error('加载初始数据失败:', error);
      }
    },
    
    async loadMouseEvents() {
      try {
        const eventsJSON = await window.go.main.App.GetMouseEvents();
        const events = JSON.parse(eventsJSON);
        
        // 转换为标记
        this.mouseEventMarkers = events
          .filter(e => e.type !== 'move')
          .map(e => ({
            timestamp: e.t,
            type: e.type,
          }));
          
        if (events.length > 0) {
          this.totalDurationMs = events[events.length - 1].t;
          this.totalDuration = this.formatTime(this.totalDurationMs);
        }
      } catch (error) {
        console.error('加载鼠标事件失败:', error);
      }
    },
    
    getBackgroundStyle() {
      if (this.backgroundType === 'solid') {
        return this.backgroundColor;
      } else if (this.backgroundType === 'gradient') {
        return `linear-gradient(135deg, ${this.gradientColor1}, ${this.gradientColor2})`;
      } else if (this.backgroundType === 'image' && this.backgroundImage) {
        return `url(${this.backgroundImage}) center/cover`;
      }
      return '#1a2a6c';
    },
    
    selectCursor(cursor) {
      this.selectedCursor = cursor;
    },
    
    uploadCustomCursor() {
      // 创建文件输入
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = 'image/png,image/svg+xml';
      
      input.onchange = (e) => {
        const file = e.target.files[0];
        if (file) {
          const reader = new FileReader();
          reader.onload = (event) => {
            const customCursor = {
              id: 'custom_' + Date.now(),
              name: '自定义',
              preview: event.target.result,
            };
            this.cursorOptions.push(customCursor);
            this.selectedCursor = customCursor;
          };
          reader.readAsDataURL(file);
        }
      };
      
      input.click();
    },
    
    uploadBackgroundImage() {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = 'image/*';
      
      input.onchange = (e) => {
        const file = e.target.files[0];
        if (file) {
          const reader = new FileReader();
          reader.onload = (event) => {
            this.backgroundImage = event.target.result;
          };
          reader.readAsDataURL(file);
        }
      };
      
      input.click();
    },
    
    toggleTrack(trackName) {
      this.tracks[trackName].enabled = !this.tracks[trackName].enabled;
    },
    
    effectClipStyle(effect) {
      return {
        left: effect.start + '%',
        width: effect.duration + '%',
      };
    },
    
    addEffect() {
      const newEffect = {
        id: Date.now(),
        name: '新效果',
        type: 'camera',
        start: this.playheadPosition,
        duration: 20,
      };
      this.effects.push(newEffect);
    },
    
    startDragEffect(event, effect) {
      // 实现效果拖拽逻辑
      console.log('拖拽效果:', effect);
    },
    
    startDragPlayhead(event) {
      // 实现播放头拖拽逻辑
      const timeline = event.currentTarget.parentElement;
      const rect = timeline.getBoundingClientRect();
      
      const onMove = (e) => {
        const x = e.clientX - rect.left;
        const percent = Math.max(0, Math.min(100, (x / rect.width) * 100));
        this.playheadPosition = percent;
        this.currentTime = this.formatTime(this.totalDurationMs * percent / 100);
      };
      
      const onUp = () => {
        document.removeEventListener('mousemove', onMove);
        document.removeEventListener('mouseup', onUp);
      };
      
      document.addEventListener('mousemove', onMove);
      document.addEventListener('mouseup', onUp);
    },
    
    playPreview() {
      this.isPlaying = true;
      // 实现预览播放逻辑
    },
    
    pausePreview() {
      this.isPlaying = false;
    },
    
    stopPreview() {
      this.isPlaying = false;
      this.playheadPosition = 0;
      this.currentTime = '00:00';
    },
    
    loadProject() {
      alert('加载项目功能待实现');
    },
    
    exportVideo() {
      this.showExportDialog = true;
    },
    
    async confirmExport() {
      this.showExportDialog = false;
      
      try {
        // 调用导出 API，传递所有参数
        const result = await window.go.main.App.ExportWithCustomParams(
          this.config.videoPath,
          this.config.mouseDataPath,
          this.exportConfig.outputPath,
          this.config.screenWidth,
          this.config.screenHeight,
          this.exportConfig.fps,
          JSON.stringify(this.animationParams),
          JSON.stringify({
            backgroundType: this.backgroundType,
            backgroundColor: this.backgroundColor,
            gradientColor1: this.gradientColor1,
            gradientColor2: this.gradientColor2,
            backgroundImage: this.backgroundImage,
          }),
          this.selectedCursor.preview
        );
        
        alert('导出完成！');
      } catch (error) {
        alert('导出失败: ' + error.message);
      }
    },
    
    formatTime(ms) {
      const seconds = Math.floor(ms / 1000);
      const minutes = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    },
  },
};
</script>

<style scoped>
.video-editor {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1e1e1e;
  color: #e0e0e0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: #252525;
  border-bottom: 1px solid #333;
}

.editor-header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 预览区域 */
.preview-section {
  padding: 20px;
  background: #2a2a2a;
  border-bottom: 1px solid #333;
}

.preview-container {
  position: relative;
  width: 100%;
  max-width: 960px;
  aspect-ratio: 16 / 9;
  margin: 0 auto;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-container canvas {
  max-width: 100%;
  max-height: 100%;
  transition: transform 0.3s;
}

.cursor-preview {
  position: absolute;
  pointer-events: none;
  transition: all 0.1s;
}

.preview-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 15px;
}

.timecode {
  font-family: monospace;
  font-size: 14px;
  color: #999;
}

/* 参数面板 */
.parameters-panel {
  padding: 20px;
  background: #252525;
  max-height: 400px;
  overflow-y: auto;
  border-bottom: 1px solid #333;
}

.parameters-panel h3 {
  margin: 0 0 20px 0;
  font-size: 16px;
  color: #fff;
}

.parameter-section {
  margin-bottom: 25px;
}

.parameter-section h4 {
  margin: 0 0 15px 0;
  font-size: 14px;
  color: #aaa;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.parameter-item {
  margin-bottom: 20px;
}

.parameter-item label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  color: #ccc;
}

.parameter-item .value {
  font-weight: 600;
  color: #4a9eff;
}

.slider {
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: #444;
  outline: none;
  -webkit-appearance: none;
}

.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #4a9eff;
  cursor: pointer;
  transition: all 0.2s;
}

.slider::-webkit-slider-thumb:hover {
  background: #6bb3ff;
  transform: scale(1.2);
}

.param-hint {
  margin-top: 5px;
  font-size: 11px;
  color: #888;
}

.scale-buttons {
  display: flex;
  gap: 10px;
}

.scale-buttons button {
  flex: 1;
  padding: 8px;
  border: 1px solid #444;
  background: #333;
  color: #ccc;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.scale-buttons button.active {
  background: #4a9eff;
  border-color: #4a9eff;
  color: white;
}

.cursor-selector {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 10px;
  margin-top: 10px;
}

.cursor-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px;
  border: 2px solid #444;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  background: #333;
}

.cursor-option:hover {
  border-color: #666;
  background: #3a3a3a;
}

.cursor-option.active {
  border-color: #4a9eff;
  background: #2a4a6a;
}

.cursor-option img {
  width: 32px;
  height: 32px;
  margin-bottom: 5px;
}

.cursor-option span {
  font-size: 11px;
  color: #ccc;
}

.cursor-option.upload {
  background: #2a2a2a;
  border-style: dashed;
}

.upload-icon {
  font-size: 32px;
  color: #666;
}

.background-type-selector {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.background-type-selector button {
  flex: 1;
  padding: 8px;
  border: 1px solid #444;
  background: #333;
  color: #ccc;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.background-type-selector button.active {
  background: #4a9eff;
  border-color: #4a9eff;
  color: white;
}

.color-picker {
  width: 100%;
  height: 40px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 8px;
}

.image-upload {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}

.bg-preview {
  width: 100%;
  height: 100px;
  object-fit: cover;
  border-radius: 4px;
}

.upload-btn {
  padding: 10px;
  background: #4a9eff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-btn:hover {
  background: #6bb3ff;
}

/* 时间轴 */
.timeline-editor {
  flex: 1;
  padding: 20px;
  background: #1e1e1e;
  overflow-y: auto;
}

.timeline-editor h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
}

.timeline-container {
  position: relative;
  background: #252525;
  border-radius: 8px;
  padding: 15px;
  min-height: 300px;
}

.timeline-ruler {
  position: relative;
  height: 30px;
  border-bottom: 1px solid #444;
  margin-bottom: 10px;
}

.time-marker {
  position: absolute;
  font-size: 11px;
  color: #888;
  transform: translateX(-50%);
}

.track {
  margin-bottom: 15px;
  border-radius: 6px;
  overflow: hidden;
}

.track-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px;
  background: #2a2a2a;
  border-bottom: 1px solid #333;
}

.track-icon {
  font-size: 16px;
}

.track-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
}

.track-toggle {
  background: none;
  border: none;
  color: #ccc;
  cursor: pointer;
  font-size: 16px;
  padding: 0;
}

.track-content {
  position: relative;
  min-height: 60px;
  background: #333;
  padding: 10px;
}

.track-clip {
  position: absolute;
  height: 40px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  padding: 0 10px;
  font-size: 12px;
  cursor: move;
  transition: all 0.2s;
}

.video-clip {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.effect-clip {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.clip-name {
  flex: 1;
  font-weight: 500;
}

.clip-duration {
  font-size: 10px;
  opacity: 0.8;
}

.effect-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: white;
  margin-left: 8px;
}

.mouse-event-marker {
  position: absolute;
  width: 3px;
  height: 40px;
  background: #ff6b6b;
  top: 10px;
  cursor: pointer;
}

.mouse-event-marker.l_down,
.mouse-event-marker.l_up {
  background: #4a9eff;
}

.mouse-event-marker.r_down,
.mouse-event-marker.r_up {
  background: #51cf66;
}

.add-effect-btn {
  padding: 8px 12px;
  background: #3a3a3a;
  border: 1px dashed #555;
  color: #ccc;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.add-effect-btn:hover {
  background: #444;
  border-color: #666;
}

.playhead {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 2px;
  background: #f00;
  cursor: ew-resize;
  z-index: 10;
}

.playhead::before {
  content: '';
  position: absolute;
  top: -5px;
  left: -5px;
  width: 12px;
  height: 12px;
  background: #f00;
  border-radius: 50%;
}

/* 按钮 */
button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-secondary {
  background: #3a3a3a;
  color: #ccc;
}

.btn-secondary:hover {
  background: #444;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #2a2a2a;
  border-radius: 12px;
  padding: 30px;
  width: 500px;
  max-width: 90%;
}

.modal-content h3 {
  margin: 0 0 20px 0;
  color: #fff;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #ccc;
  font-size: 13px;
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group select {
  width: 100%;
  padding: 10px;
  background: #333;
  border: 1px solid #444;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 14px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 25px;
}
</style>
