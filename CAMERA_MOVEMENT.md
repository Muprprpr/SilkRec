# Camera Movement & Export System

## Overview

SilkRec implements smooth camera movements (Carmen effect) that follow the mouse cursor with intelligent zoom and easing animations. This creates professional-looking screen recordings that automatically focus on areas of interest.

## Features

### 1. Smooth Camera Tracking
- **Interpolated Movement**: The virtual camera smoothly follows the mouse cursor using linear interpolation (lerp)
- **Configurable Smoothness**: Adjust the `smoothFactor` (0.0-1.0) to control how responsive the camera is
  - Lower values (e.g., 0.05): Fast, immediate tracking
  - Higher values (e.g., 0.25): Very smooth, delayed tracking
  - Default: 0.15 (balanced)

### 2. Automatic Zoom on Click
- **Zoom In on Click**: Camera zooms in (default 1.5x) when mouse buttons are pressed
- **Zoom Out on Release**: Camera zooms back out when mouse buttons are released
- **Hold Detection**: Maintains zoom during long clicks (hold events)
- **Configurable**: Enable/disable and adjust zoom levels

### 3. Viewport Calculation
- **Dynamic Viewport**: Calculates the visible screen area based on camera position and zoom
- **Boundary Clamping**: Ensures the viewport never goes outside the screen bounds
- **Resolution Independence**: Works with any screen resolution

## Architecture

### Core Components

#### 1. `CameraController` (`pkg/recorder/camera.go`)
Manages the virtual camera state and smooth transitions.

**Key Methods:**
- `Update(event)`: Updates camera based on mouse events
- `GetViewport()`: Returns the current viewport rectangle
- `GetTransform()`: Returns scale and offset for rendering
- `SetSmoothFactor(factor)`: Adjusts camera smoothness

**State:**
- `currentState`: Current interpolated camera position
- `targetState`: Target position based on latest mouse event
- `smoothFactor`: Interpolation factor (0.0-1.0)

#### 2. `Exporter` (`pkg/recorder/exporter.go`)
Handles video export with camera movements.

**Key Methods:**
- `PrepareExport()`: Loads mouse data and generates camera path
- `GenerateCameraPath()`: Creates camera frames synchronized with video
- `GetCameraFrames()`: Returns all generated camera frames
- `SaveCameraPath(path)`: Saves camera data for debugging

#### 3. `CameraFrame` Structure
Represents camera state at a specific timestamp:
```go
type CameraFrame struct {
    Timestamp int64   // Timestamp in milliseconds
    X         float64 // Camera X position
    Y         float64 // Camera Y position
    Zoom      float64 // Zoom level
    MouseX    int16   // Mouse X position
    MouseY    int16   // Mouse Y position
    EventType string  // Event type that triggered this frame
}
```

## Usage

### Backend API (Go)

#### 1. Prepare Export with Camera Movements
```go
// From frontend, call:
PrepareExport(
    videoPath,      // Input video path
    mouseDataPath,  // Mouse events JSON
    outputPath,     // Output video path
    screenWidth,    // e.g., 1920
    screenHeight,   // e.g., 1080
    fps             // e.g., 30
)
```

Returns export information including:
- Mouse event count
- Camera frame count
- Duration
- Estimated total frames

#### 2. Get Camera Frames for Rendering
```go
// Get camera frames as JSON string
cameraFramesJSON, err := GetCameraFrames()

// Parse in frontend to render each frame
```

#### 3. Save Camera Path for Debugging
```go
// Save camera path to JSON file
SaveCameraPath("output/camera_path.json")
```

### Frontend Integration

The frontend should:

1. **Call PrepareExport** to generate camera path
2. **Get camera frames** to know camera position at each timestamp
3. **Render each frame** by:
   - Reading the video frame
   - Applying camera transform (zoom + pan)
   - Drawing the viewport
   - Optionally drawing cursor at mouse position
4. **Export frames** using the existing pipe writer system

Example workflow:
```javascript
// 1. Prepare export
const exportInfo = await PrepareExport(
    videoPath, 
    mouseDataPath, 
    outputPath, 
    1920, 
    1080, 
    30
);

// 2. Get camera frames
const cameraFramesJSON = await GetCameraFrames();
const cameraFrames = JSON.parse(cameraFramesJSON);

// 3. For each frame, render with camera transform
for (const frame of cameraFrames) {
    // Calculate viewport
    const viewportWidth = screenWidth / frame.Zoom;
    const viewportHeight = screenHeight / frame.Zoom;
    const viewportX = frame.X - viewportWidth / 2;
    const viewportY = frame.Y - viewportHeight / 2;
    
    // Draw video frame with transform
    ctx.save();
    ctx.scale(frame.Zoom, frame.Zoom);
    ctx.translate(-viewportX, -viewportY);
    ctx.drawImage(videoFrame, 0, 0);
    
    // Draw cursor if enabled
    if (showCursor) {
        drawCursor(frame.MouseX, frame.MouseY);
    }
    
    ctx.restore();
    
    // Export frame to video
    await WriteExportFrame(canvas.toDataURL());
}
```

## Configuration

### Export Configuration Options

```go
type ExportConfig struct {
    VideoPath      string  // Input video path
    MouseDataPath  string  // Mouse data JSON path
    OutputPath     string  // Output video path
    FPS            int     // Output frame rate (default: 30)
    EnableZoom     bool    // Enable zoom on click (default: true)
    ZoomLevel      float64 // Zoom level when clicking (default: 1.5)
    SmoothFactor   float64 // Camera smoothness 0.0-1.0 (default: 0.15)
    ShowCursor     bool    // Show cursor in export (default: true)
    CursorSize     int     // Cursor size in pixels (default: 32)
    ScreenWidth    int     // Screen width
    ScreenHeight   int     // Screen height
}
```

### Recommended Settings

**For Tutorial Videos:**
- EnableZoom: true
- ZoomLevel: 1.5-2.0
- SmoothFactor: 0.15-0.20 (smooth but responsive)

**For Demo Videos:**
- EnableZoom: false or true
- ZoomLevel: 1.2-1.3 (subtle zoom)
- SmoothFactor: 0.10-0.15 (more responsive)

**For Fast-Paced Content:**
- EnableZoom: false
- SmoothFactor: 0.05-0.10 (very responsive)

## Algorithm Details

### Smooth Interpolation (Lerp)

The camera uses linear interpolation to smoothly transition between states:

```go
current = current + (target - current) * smoothFactor
```

This creates an exponential ease-out effect where:
- The camera moves quickly when far from target
- The camera slows down as it approaches target
- Movement is continuous and smooth

### Zoom Transition

Zoom transitions use a slower smoothing factor (0.5x) to avoid jarring zoom changes:

```go
currentZoom = currentZoom + (targetZoom - currentZoom) * (smoothFactor * 0.5)
```

### Frame Generation

Camera frames are generated at the specified FPS by:
1. Iterating through timestamps at frame intervals
2. Applying all mouse events up to that timestamp
3. Recording the resulting camera state
4. Continuing until all events are processed

## Debugging

### Save Camera Path
Save the generated camera path to inspect it:
```go
SaveCameraPath("output/camera_debug.json")
```

This creates a JSON file with all camera frames, useful for:
- Verifying smooth transitions
- Debugging coordinate issues
- Understanding zoom behavior
- Analyzing camera timing

### Check Export Info
Get export statistics:
```go
info := GetExportInfo()
// Returns: mouseEventCount, cameraFrameCount, fps, duration, etc.
```

## Performance Considerations

- **Frame Generation**: O(n) where n = number of mouse events
- **Memory**: Minimal - camera frames are lightweight structs
- **CPU**: Interpolation is fast; rendering is the bottleneck
- **Optimization**: Pre-generate camera frames once, reuse for rendering

## Future Enhancements

Potential improvements:
- [ ] Bezier curve smoothing for even smoother movements
- [ ] Predictive camera movement (look-ahead)
- [ ] Smart zoom based on UI element detection
- [ ] Customizable easing functions
- [ ] Camera shake/bounce effects
- [ ] Focus transitions between multiple points of interest
- [ ] Automatic speed adjustments based on mouse velocity
