package recorder

import (
	"SmoothScreen/pkg/hook"
	"math"
)

// CameraState represents the virtual camera state
type CameraState struct {
	X      float64 // Camera center X position
	Y      float64 // Camera center Y position
	Zoom   float64 // Zoom level (1.0 = no zoom, 2.0 = 2x zoom)
	Width  int     // Viewport width
	Height int     // Viewport height
}

// CameraController manages smooth camera movements following the mouse
type CameraController struct {
	currentState CameraState
	targetState  CameraState
	screenWidth  int
	screenHeight int
	smoothFactor float64 // 0.0 - 1.0, higher = smoother but slower response
	zoomOnClick  bool
	clickZoom    float64
	defaultZoom  float64
}

// NewCameraController creates a new camera controller
func NewCameraController(screenWidth, screenHeight int) *CameraController {
	return &CameraController{
		currentState: CameraState{
			X:      float64(screenWidth) / 2,
			Y:      float64(screenHeight) / 2,
			Zoom:   1.0,
			Width:  screenWidth,
			Height: screenHeight,
		},
		targetState: CameraState{
			X:      float64(screenWidth) / 2,
			Y:      float64(screenHeight) / 2,
			Zoom:   1.0,
			Width:  screenWidth,
			Height: screenHeight,
		},
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		smoothFactor: 0.15, // Smooth camera movement (85% previous, 15% target)
		zoomOnClick:  true,
		clickZoom:    1.5, // Zoom to 1.5x when clicking
		defaultZoom:  1.0,
	}
}

// Update updates the camera state based on mouse events
// Returns true if the camera state changed
func (c *CameraController) Update(event hook.MouseEvent) bool {
	changed := false

	// Update target position based on mouse position
	c.targetState.X = float64(event.X)
	c.targetState.Y = float64(event.Y)

	// Update zoom based on event type
	if c.zoomOnClick {
		switch event.EventType {
		case "l_down", "r_down", "m_down":
			// Zoom in on click
			c.targetState.Zoom = c.clickZoom
			changed = true
		case "l_up", "r_up", "m_up":
			// Zoom out on release
			c.targetState.Zoom = c.defaultZoom
			changed = true
		case "hold":
			// Keep zoom during hold
			c.targetState.Zoom = c.clickZoom
		}
	}

	// Apply smooth interpolation (easing)
	prevX := c.currentState.X
	prevY := c.currentState.Y
	prevZoom := c.currentState.Zoom

	c.currentState.X = lerp(c.currentState.X, c.targetState.X, c.smoothFactor)
	c.currentState.Y = lerp(c.currentState.Y, c.targetState.Y, c.smoothFactor)
	c.currentState.Zoom = lerp(c.currentState.Zoom, c.targetState.Zoom, c.smoothFactor*0.5) // Slower zoom transition

	// Check if camera actually moved
	if math.Abs(c.currentState.X-prevX) > 0.1 ||
		math.Abs(c.currentState.Y-prevY) > 0.1 ||
		math.Abs(c.currentState.Zoom-prevZoom) > 0.01 {
		changed = true
	}

	return changed
}

// GetViewport calculates the viewport rectangle for the current camera state
// Returns x, y (top-left corner), width, and height of the viewport
func (c *CameraController) GetViewport() (x, y, width, height int) {
	// Calculate viewport size based on zoom
	viewportWidth := float64(c.screenWidth) / c.currentState.Zoom
	viewportHeight := float64(c.screenHeight) / c.currentState.Zoom

	// Calculate top-left corner of viewport (centered on camera position)
	x = int(c.currentState.X - viewportWidth/2)
	y = int(c.currentState.Y - viewportHeight/2)

	// Clamp to screen bounds to prevent rendering outside screen area
	// This ensures the viewport stays within [0, screenWidth) x [0, screenHeight)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Ensure viewport doesn't extend beyond right edge
	maxX := c.screenWidth - int(viewportWidth)
	if maxX < 0 {
		maxX = 0 // Handle case where viewport is larger than screen (zoom < 1.0)
	}
	if x > maxX {
		x = maxX
	}

	// Ensure viewport doesn't extend beyond bottom edge
	maxY := c.screenHeight - int(viewportHeight)
	if maxY < 0 {
		maxY = 0 // Handle case where viewport is larger than screen (zoom < 1.0)
	}
	if y > maxY {
		y = maxY
	}

	return x, y, int(viewportWidth), int(viewportHeight)
}

// GetTransform returns the transform matrix for rendering
// Returns scale and offset for transforming screen coordinates to viewport
func (c *CameraController) GetTransform() (scale, offsetX, offsetY float64) {
	x, y, _, _ := c.GetViewport()
	return c.currentState.Zoom, -float64(x), -float64(y)
}

// SetSmoothFactor sets the camera smoothness (0.0 = instant, 1.0 = very smooth)
func (c *CameraController) SetSmoothFactor(factor float64) {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	c.smoothFactor = factor
}

// SetZoomOnClick enables or disables zoom on click
func (c *CameraController) SetZoomOnClick(enabled bool) {
	c.zoomOnClick = enabled
}

// SetClickZoom sets the zoom level when clicking
func (c *CameraController) SetClickZoom(zoom float64) {
	c.clickZoom = zoom
}

// GetState returns the current camera state
func (c *CameraController) GetState() CameraState {
	return c.currentState
}

// Reset resets the camera to default state
func (c *CameraController) Reset() {
	c.currentState = CameraState{
		X:      float64(c.screenWidth) / 2,
		Y:      float64(c.screenHeight) / 2,
		Zoom:   1.0,
		Width:  c.screenWidth,
		Height: c.screenHeight,
	}
	c.targetState = c.currentState
}

// lerp performs linear interpolation between two values
func lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

// EaseInOutCubic applies cubic easing to a value (smoother acceleration/deceleration)
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// EaseOutQuad applies quadratic easing out (fast start, slow end)
func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

// CameraFrame represents a camera state at a specific time
type CameraFrame struct {
	Timestamp int64   // Timestamp in milliseconds
	X         float64 // Camera X position
	Y         float64 // Camera Y position
	Zoom      float64 // Zoom level
	MouseX    int16   // Mouse X position
	MouseY    int16   // Mouse Y position
	EventType string  // Event type that triggered this frame
}

// GenerateCameraPath generates smooth camera frames from mouse events
func GenerateCameraPath(mouseEvents []hook.MouseEvent, screenWidth, screenHeight int, fps int) []CameraFrame {
	if len(mouseEvents) == 0 {
		return []CameraFrame{}
	}

	controller := NewCameraController(screenWidth, screenHeight)
	frames := make([]CameraFrame, 0)

	// Get time range
	startTime := mouseEvents[0].Timestamp
	endTime := mouseEvents[len(mouseEvents)-1].Timestamp
	frameDuration := int64(1000 / fps) // Frame duration in ms

	eventIndex := 0
	for timestamp := startTime; timestamp <= endTime; timestamp += frameDuration {
		// Find all events within this frame
		for eventIndex < len(mouseEvents) && mouseEvents[eventIndex].Timestamp <= timestamp {
			controller.Update(mouseEvents[eventIndex])
			eventIndex++
		}

		// Record camera frame
		state := controller.GetState()
		frame := CameraFrame{
			Timestamp: timestamp,
			X:         state.X,
			Y:         state.Y,
			Zoom:      state.Zoom,
		}

		// Add mouse position from last event
		if eventIndex > 0 {
			lastEvent := mouseEvents[eventIndex-1]
			frame.MouseX = lastEvent.X
			frame.MouseY = lastEvent.Y
			frame.EventType = lastEvent.EventType
		}

		frames = append(frames, frame)
	}

	return frames
}
