package sys

import (
	"fmt"
	"syscall"
	"unsafe"
)

// ScreenInfo 屏幕信息
type ScreenInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	DPI    int `json:"dpi"`
}

// Windows API 常量
const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
	LOGPIXELSX  = 88
	LOGPIXELSY  = 90
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procGetDeviceCaps    = gdi32.NewProc("GetDeviceCaps")
)

// GetScreenInfo 获取屏幕信息
func GetScreenInfo() (*ScreenInfo, error) {
	width := getSystemMetrics(SM_CXSCREEN)
	height := getSystemMetrics(SM_CYSCREEN)
	dpi := getDPI()

	return &ScreenInfo{
		Width:  width,
		Height: height,
		DPI:    dpi,
	}, nil
}

// getSystemMetrics 获取系统度量值
func getSystemMetrics(nIndex int) int {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(nIndex))
	return int(ret)
}

// getDPI 获取屏幕DPI
func getDPI() int {
	hdc := syscall.Handle(0)
	ret, _, _ := procGetDeviceCaps.Call(uintptr(hdc), uintptr(LOGPIXELSX))
	return int(ret)
}

// SetWindowAlwaysOnTop 设置窗口置顶
func SetWindowAlwaysOnTop(hwnd uintptr, top bool) error {
	// HWND_TOPMOST = -1, HWND_NOTOPMOST = -2
	// 在Go中uintptr是无符号的，需要使用十六进制表示
	hwndInsertAfter := uintptr(0xFFFFFFFE) // HWND_NOTOPMOST (-2)
	if top {
		hwndInsertAfter = uintptr(0xFFFFFFFF) // HWND_TOPMOST (-1)
	}

	// SWP_NOSIZE | SWP_NOMOVE | SWP_SHOWWINDOW
	flags := uintptr(0x0001 | 0x0002 | 0x0040)

	_, _, err := user32.NewProc("SetWindowPos").Call(
		hwnd,
		hwndInsertAfter,
		0, 0, 0, 0,
		flags,
	)

	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("设置窗口置顶失败: %w", err)
	}

	return nil
}

// GetWindowHandle 获取窗口句柄（通过窗口标题）
func GetWindowHandle(title string) (uintptr, error) {
	// 这里需要实现查找窗口的逻辑
	// 由于Windows API比较复杂，这里先返回0
	// 实际使用时可能需要调用FindWindow或EnumWindows
	return 0, fmt.Errorf("未实现")
}

// HideWindow 隐藏窗口
func HideWindow(hwnd uintptr) error {
	_, _, err := user32.NewProc("ShowWindow").Call(hwnd, 0) // SW_HIDE
	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("隐藏窗口失败: %w", err)
	}
	return nil
}

// ShowWindow 显示窗口
func ShowWindow(hwnd uintptr) error {
	_, _, err := user32.NewProc("ShowWindow").Call(hwnd, 1) // SW_SHOWNORMAL
	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("显示窗口失败: %w", err)
	}
	return nil
}

// MinimizeWindow 最小化窗口
func MinimizeWindow(hwnd uintptr) error {
	_, _, err := user32.NewProc("ShowWindow").Call(hwnd, 2) // SW_SHOWMINIMIZED
	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("最小化窗口失败: %w", err)
	}
	return nil
}

// MaximizeWindow 最大化窗口
func MaximizeWindow(hwnd uintptr) error {
	_, _, err := user32.NewProc("ShowWindow").Call(hwnd, 3) // SW_SHOWMAXIMIZED
	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("最大化窗口失败: %w", err)
	}
	return nil
}

// MoveWindow 移动窗口
func MoveWindow(hwnd uintptr, x, y, width, height int) error {
	_, _, err := user32.NewProc("MoveWindow").Call(
		hwnd,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		1, // bRepaint = TRUE
	)
	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("移动窗口失败: %w", err)
	}
	return nil
}

// GetWindowRect 获取窗口矩形
func GetWindowRect(hwnd uintptr) (int, int, int, int, error) {
	var rect struct {
		Left   int32
		Top    int32
		Right  int32
		Bottom int32
	}

	_, _, err := user32.NewProc("GetWindowRect").Call(
		hwnd,
		uintptr(unsafe.Pointer(&rect)),
	)

	if err != nil && err != syscall.Errno(0) {
		return 0, 0, 0, 0, fmt.Errorf("获取窗口矩形失败: %w", err)
	}

	return int(rect.Left), int(rect.Top), int(rect.Right), int(rect.Bottom), nil
}

// SetWindowPos 设置窗口位置和大小
func SetWindowPos(hwnd uintptr, x, y, width, height int) error {
	hwndInsertAfter := uintptr(0)     // HWND_TOP
	flags := uintptr(0x0001 | 0x0002) // SWP_NOSIZE | SWP_NOMOVE

	if width > 0 && height > 0 {
		flags &^= 0x0001 // 清除SWP_NOSIZE
	}

	if x >= 0 && y >= 0 {
		flags &^= 0x0002 // 清除SWP_NOMOVE
	}

	_, _, err := user32.NewProc("SetWindowPos").Call(
		hwnd,
		hwndInsertAfter,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		flags,
	)

	if err != nil && err != syscall.Errno(0) {
		return fmt.Errorf("设置窗口位置失败: %w", err)
	}

	return nil
}
