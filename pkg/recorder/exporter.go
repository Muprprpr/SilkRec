package recorder

import (
	"SmoothScreen/pkg/ffmpeg"
	"SmoothScreen/pkg/hook"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ExportConfig represents export configuration
type ExportConfig struct {
	VideoPath     string  // Input video path
	MouseDataPath string  // Mouse data JSON path
	OutputPath    string  // Output video path
	FPS           int     // Output frame rate
	EnableZoom    bool    // Enable zoom on click
	ZoomLevel     float64 // Zoom level when clicking
	SmoothFactor  float64 // Camera smoothness (0.0-1.0)
	ShowCursor    bool    // Show cursor in export
	CursorSize    int     // Cursor size in pixels
	ScreenWidth   int     // Screen width
	ScreenHeight  int     // Screen height
}

// DefaultExportConfig returns default export configuration
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		FPS:          30,
		EnableZoom:   true,
		ZoomLevel:    1.5,
		SmoothFactor: 0.15,
		ShowCursor:   true,
		CursorSize:   32,
	}
}

// Exporter handles video export with camera movements
type Exporter struct {
	config        ExportConfig
	ffmpegManager *ffmpeg.FFmpegManager
	mouseEvents   []hook.MouseEvent
	cameraFrames  []CameraFrame
}

// NewExporter creates a new exporter
func NewExporter(ffmpegManager *ffmpeg.FFmpegManager, config ExportConfig) *Exporter {
	return &Exporter{
		config:        config,
		ffmpegManager: ffmpegManager,
	}
}

// LoadMouseData loads mouse data from JSON file
func (e *Exporter) LoadMouseData(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read mouse data: %w", err)
	}

	if err := json.Unmarshal(data, &e.mouseEvents); err != nil {
		return fmt.Errorf("failed to parse mouse data: %w", err)
	}

	fmt.Printf("Loaded %d mouse events\n", len(e.mouseEvents))
	return nil
}

// GenerateCameraPath generates smooth camera movement path
func (e *Exporter) GenerateCameraPath() error {
	if len(e.mouseEvents) == 0 {
		return fmt.Errorf("no mouse events loaded")
	}

	// Generate camera frames based on mouse events
	e.cameraFrames = GenerateCameraPath(
		e.mouseEvents,
		e.config.ScreenWidth,
		e.config.ScreenHeight,
		e.config.FPS,
	)

	fmt.Printf("Generated %d camera frames\n", len(e.cameraFrames))
	return nil
}

// SaveCameraPath saves camera path to JSON file for debugging
func (e *Exporter) SaveCameraPath(path string) error {
	data, err := json.MarshalIndent(e.cameraFrames, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal camera path: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write camera path: %w", err)
	}

	fmt.Printf("Saved camera path to: %s\n", path)
	return nil
}

// GetCameraFrames returns the generated camera frames
func (e *Exporter) GetCameraFrames() []CameraFrame {
	return e.cameraFrames
}

// GetMouseEvents returns the loaded mouse events
func (e *Exporter) GetMouseEvents() []hook.MouseEvent {
	return e.mouseEvents
}

// PrepareExport prepares everything needed for export
func (e *Exporter) PrepareExport() error {
	// Load mouse data if not already loaded
	if len(e.mouseEvents) == 0 && e.config.MouseDataPath != "" {
		if err := e.LoadMouseData(e.config.MouseDataPath); err != nil {
			return fmt.Errorf("failed to load mouse data: %w", err)
		}
	}

	// Generate camera path
	if err := e.GenerateCameraPath(); err != nil {
		return fmt.Errorf("failed to generate camera path: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(e.config.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// GetExportInfo returns information about the export
func (e *Exporter) GetExportInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["mouseEventCount"] = len(e.mouseEvents)
	info["cameraFrameCount"] = len(e.cameraFrames)
	info["fps"] = e.config.FPS
	info["enableZoom"] = e.config.EnableZoom
	info["zoomLevel"] = e.config.ZoomLevel
	info["smoothFactor"] = e.config.SmoothFactor
	info["showCursor"] = e.config.ShowCursor

	if len(e.cameraFrames) > 0 {
		duration := float64(e.cameraFrames[len(e.cameraFrames)-1].Timestamp-e.cameraFrames[0].Timestamp) / 1000.0
		info["duration"] = duration
		info["estimatedFrames"] = int(duration * float64(e.config.FPS))
	}

	return info
}
