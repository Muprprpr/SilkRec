# Export with Camera Movement Guide

## Quick Start

This guide explains how to export recordings with smooth camera movements and proper mouse tracking.

## Backend Setup (Already Implemented)

The backend now includes:
1. **Camera Controller** - Handles smooth camera movements
2. **Exporter** - Generates camera paths from mouse events
3. **Export API** - New methods in `app.go` for frontend integration

## Export Workflow

### Step 1: Record Video and Mouse Data

When recording, the system captures:
- Video file (via FFmpeg screen capture)
- Mouse events JSON (position, clicks, scrolls)

These files are automatically saved:
- `output/recording_TIMESTAMP.mp4`
- `output/mouse_events.json`

### Step 2: Prepare Export

Call the `PrepareExport` API to generate camera path:

```javascript
// From your frontend
const exportInfo = await window.go.main.App.PrepareExport(
    "output/recording_1234567890.mp4",  // Input video path
    "output/mouse_events.json",         // Mouse data path
    "output/final_export.mp4",          // Output path
    1920,                               // Screen width
    1080,                               // Screen height
    30                                  // FPS
);

console.log(exportInfo);
// {
//     mouseEventCount: 1543,
//     cameraFrameCount: 900,
//     fps: 30,
//     enableZoom: true,
//     zoomLevel: 1.5,
//     smoothFactor: 0.15,
//     duration: 30.0,
//     estimatedFrames: 900
// }
```

### Step 3: Get Camera Frames

Retrieve the generated camera path:

```javascript
const cameraFramesJSON = await window.go.main.App.GetCameraFrames();
const cameraFrames = JSON.parse(cameraFramesJSON);

// cameraFrames is an array of:
// [
//     {
//         Timestamp: 0,        // milliseconds
//         X: 960.0,           // camera center X
//         Y: 540.0,           // camera center Y
//         Zoom: 1.0,          // zoom level
//         MouseX: 960,        // mouse X position
//         MouseY: 540,        // mouse Y position
//         EventType: "move"   // event type
//     },
//     ...
// ]
```

### Step 4: Render and Export Frames

For each camera frame, render the transformed video:

```javascript
// 1. Start export pipeline
await window.go.main.App.StartExport("output/final_export.mp4", 30);

// 2. Load video
const video = document.createElement('video');
video.src = "output/recording_1234567890.mp4";
await video.play();

// 3. Create canvas for rendering
const canvas = document.createElement('canvas');
canvas.width = 1920;
canvas.height = 1080;
const ctx = canvas.getContext('2d');

// 4. Render each frame
for (let i = 0; i < cameraFrames.length; i++) {
    const frame = cameraFrames[i];
    
    // Seek video to correct timestamp
    video.currentTime = frame.Timestamp / 1000.0;
    await waitForSeek(video);
    
    // Calculate viewport
    const viewportWidth = 1920 / frame.Zoom;
    const viewportHeight = 1080 / frame.Zoom;
    const viewportX = frame.X - viewportWidth / 2;
    const viewportY = frame.Y - viewportHeight / 2;
    
    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    // Draw video with camera transform
    ctx.save();
    ctx.scale(frame.Zoom, frame.Zoom);
    ctx.translate(-viewportX, -viewportY);
    ctx.drawImage(video, 0, 0, 1920, 1080);
    ctx.restore();
    
    // Draw cursor
    drawCursor(ctx, frame.MouseX, frame.MouseY);
    
    // Export frame
    const imageData = canvas.toDataURL('image/png').split(',')[1];
    await window.go.main.App.WriteExportFrame(imageData);
    
    // Progress callback
    if (i % 30 === 0) {
        console.log(`Progress: ${(i / cameraFrames.length * 100).toFixed(1)}%`);
    }
}

// 5. Finish export
await window.go.main.App.FinishExport();
console.log("Export complete!");
```

## Helper Functions

### Wait for Video Seek
```javascript
function waitForSeek(video) {
    return new Promise((resolve) => {
        const onSeeked = () => {
            video.removeEventListener('seeked', onSeeked);
            resolve();
        };
        video.addEventListener('seeked', onSeeked);
    });
}
```

### Draw Cursor
```javascript
function drawCursor(ctx, x, y) {
    // Simple arrow cursor
    ctx.fillStyle = 'white';
    ctx.strokeStyle = 'black';
    ctx.lineWidth = 2;
    
    // Draw arrow shape
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(x + 5, y + 15);
    ctx.lineTo(x + 10, y + 10);
    ctx.lineTo(x + 15, y + 5);
    ctx.closePath();
    
    ctx.fill();
    ctx.stroke();
}
```

## Coordinate System

The coordinate system uses standard screen coordinates:
- Origin (0, 0) is at the top-left corner
- X increases to the right
- Y increases downward
- Camera center (X, Y) is the focus point
- Viewport is calculated around the camera center

### Example Coordinate Calculations

For a 1920x1080 screen with camera at (960, 540) and zoom 1.5:

```javascript
// Viewport size
viewportWidth = 1920 / 1.5 = 1280
viewportHeight = 1080 / 1.5 = 720

// Viewport top-left corner
viewportX = 960 - 1280/2 = 320
viewportY = 540 - 720/2 = 180

// This viewport shows a 1280x720 region of the screen
// centered at (960, 540), then scaled up to fill 1920x1080
```

## Common Issues and Solutions

### Issue 1: Preview Focuses on Bottom-Right

**Problem**: When rendering, the view is stuck at the bottom-right corner.

**Solution**: Ensure camera initialization is correct:
```javascript
// Camera should start at screen center
const centerX = screenWidth / 2;
const centerY = screenHeight / 2;

// Verify first camera frame
console.log(cameraFrames[0]);
// Should show X ≈ centerX, Y ≈ centerY
```

**Backend Fix**: The camera controller now properly initializes to screen center and clamps viewport bounds correctly.

### Issue 2: Mouse Movement Not Synced

**Problem**: Mouse cursor doesn't match recorded movements.

**Solution**: Ensure timestamps are properly synchronized:
```javascript
// Use camera frame timestamp for video seek
video.currentTime = frame.Timestamp / 1000.0;

// Use mouse position from camera frame
drawCursor(ctx, frame.MouseX, frame.MouseY);
```

### Issue 3: Jerky Camera Movement

**Problem**: Camera moves too fast or has jumpy transitions.

**Solution**: Adjust smooth factor in export config:
```javascript
// Lower = faster, more responsive (0.05 - 0.10)
// Higher = smoother, more delayed (0.15 - 0.25)
// Default: 0.15

// To change, you would need to modify ExportConfig
// before calling PrepareExport (backend modification required)
```

### Issue 4: No Zoom on Click

**Problem**: Camera doesn't zoom when mouse clicks occur.

**Solution**: Verify zoom is enabled in export config:
- Check that `EnableZoom: true` in `ExportConfig`
- Ensure mouse events include click events (l_down, l_up, etc.)
- Verify zoom level is appropriate (default: 1.5)

## Advanced Customization

### Custom Zoom Levels

To customize zoom behavior, modify the backend `ExportConfig`:

```go
// In your frontend, request custom config via new API method
// (Would need to add ConfigureExport method to app.go)

config := recorder.DefaultExportConfig()
config.EnableZoom = true
config.ZoomLevel = 2.0          // 2x zoom on click
config.SmoothFactor = 0.10      // More responsive
```

### Custom Cursor Drawing

Create custom cursor styles:

```javascript
function drawCustomCursor(ctx, x, y, eventType) {
    // Different cursor for different events
    switch (eventType) {
        case 'l_down':
            // Clicked cursor (larger, colored)
            ctx.fillStyle = 'rgba(255, 0, 0, 0.5)';
            ctx.beginPath();
            ctx.arc(x, y, 20, 0, Math.PI * 2);
            ctx.fill();
            break;
        
        case 'scroll':
            // Scroll indicator
            ctx.fillStyle = 'rgba(0, 255, 0, 0.5)';
            ctx.fillRect(x - 10, y - 20, 20, 40);
            break;
        
        default:
            // Normal cursor
            drawCursor(ctx, x, y);
    }
}
```

## Performance Optimization

### Batch Rendering
Process frames in batches to improve performance:

```javascript
const BATCH_SIZE = 30;
for (let batch = 0; batch < cameraFrames.length; batch += BATCH_SIZE) {
    const batchFrames = cameraFrames.slice(batch, batch + BATCH_SIZE);
    
    for (const frame of batchFrames) {
        // Render frame...
    }
    
    // Update progress
    const progress = (batch / cameraFrames.length) * 100;
    updateProgressUI(progress);
    
    // Allow UI to update
    await new Promise(resolve => setTimeout(resolve, 0));
}
```

### Memory Management
Clear resources between batches:

```javascript
// After processing batch
canvas.getContext('2d').clearRect(0, 0, canvas.width, canvas.height);
if (global.gc) global.gc(); // Force garbage collection (if available)
```

## Testing

### Debug Camera Path
Save camera path for inspection:

```javascript
await window.go.main.App.SaveCameraPath("output/camera_debug.json");
```

This creates a JSON file you can inspect to verify:
- Camera starts at screen center
- Camera follows mouse smoothly
- Zoom transitions are correct
- Timestamps are sequential

### Verify Export Info
Check export statistics:

```javascript
const info = await window.go.main.App.GetExportInfo();
console.log("Export Info:", info);
// Verify mouseEventCount > 0
// Verify cameraFrameCount matches expected duration * fps
```

## Complete Example

See `examples/export_with_camera.html` for a complete working example.

## Troubleshooting

1. **Check FFmpeg availability**:
   ```javascript
   const available = await window.go.main.App.CheckFFmpegAvailable();
   console.log("FFmpeg available:", available);
   ```

2. **Verify mouse data was recorded**:
   ```javascript
   const mouseData = await window.go.main.App.GetMouseData();
   console.log("Mouse events:", JSON.parse(mouseData).length);
   ```

3. **Check recording status**:
   ```javascript
   const status = await window.go.main.App.GetRecordingStatus();
   console.log("Recording status:", status);
   ```

## Next Steps

- Implement frontend export UI
- Add progress indicators
- Add export configuration options in UI
- Add export presets (tutorial, demo, fast-paced)
- Add export quality settings
