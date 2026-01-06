package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MouseEvent 表示一个鼠标事件
type MouseEvent struct {
	Timestamp int64  `json:"t"`        // 相对时间戳（毫秒，相对于录制开始）
	X         int16  `json:"x"`        // X坐标
	Y         int16  `json:"y"`        // Y坐标
	EventType string `json:"type"`     // 事件类型: move, l_down, l_up, r_down, r_up, m_down, m_up, scroll
	Duration  int    `json:"duration"` // hold持续时间(ms)
	Delta     int    `json:"delta"`    // 滚动增量
	Button    string `json:"button"`   // 按钮标识: left, right, middle
}

// MouseHook 管理鼠标钩子
type MouseHook struct {
	ctx           context.Context
	eventChan     chan hook.Event
	mouseData     []MouseEvent
	mouseDataMu   sync.Mutex
	isRecording   bool
	lastX         int16
	lastY         int16
	mouseDownTime map[string]time.Time // 记录鼠标按下时间
	mouseDownMu   sync.Mutex
	ticker        *time.Ticker
	immediateChan chan MouseEvent // 用于立即发送关键时刻事件
	startTime     time.Time       // 录制开始时间
}

// NewMouseHook 创建新的鼠标钩子
func NewMouseHook(ctx context.Context) *MouseHook {
	return &MouseHook{
		ctx:           ctx,
		eventChan:     hook.Start(),
		mouseData:     make([]MouseEvent, 0),
		mouseDownTime: make(map[string]time.Time),
		ticker:        time.NewTicker(50 * time.Millisecond), // 20Hz = 50ms
		immediateChan: make(chan MouseEvent, 100),            // 用于立即发送关键时刻事件
	}
}

// Start 开始监听鼠标事件
func (m *MouseHook) Start() {
	fmt.Println("启动增强鼠标监听...")

	go m.processEvents()
	go m.emitImmediateEvents() // 处理立即发送的事件
	go m.emitData()
}

// Stop 停止监听
func (m *MouseHook) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	hook.End()
}

// processEvents 处理原始鼠标事件
func (m *MouseHook) processEvents() {
	for {
		select {
		case ev := <-m.eventChan:
			m.handleEvent(ev)

		case <-m.ctx.Done():
			return
		}
	}
}

// handleEvent 处理单个事件
func (m *MouseHook) handleEvent(ev hook.Event) {
	switch ev.Kind {
	case hook.MouseMove:
		m.lastX = ev.X
		m.lastY = ev.Y
		if m.isRecording {
			m.addMouseEvent(MouseEvent{
				Timestamp: m.getRelativeTimestamp(),
				X:         ev.X,
				Y:         ev.Y,
				EventType: "move",
			})
		}

	case hook.MouseDown, hook.MouseUp:
		// 根据鼠标按键判断是左键、右键还是中键
		button := "left"
		if ev.Button == 2 {
			button = "right"
		} else if ev.Button == 3 {
			button = "middle"
		}

		if ev.Kind == hook.MouseDown {
			// 记录按下时间
			m.mouseDownMu.Lock()
			m.mouseDownTime[button] = time.Now()
			m.mouseDownMu.Unlock()

			if m.isRecording {
				event := MouseEvent{
					Timestamp: m.getRelativeTimestamp(),
					X:         ev.X,
					Y:         ev.Y,
					EventType: getMouseDownType(button),
					Button:    button,
				}
				m.addMouseEvent(event)
				// 立即发送给前端
				select {
				case m.immediateChan <- event:
				default:
				}
			}
		} else if ev.Kind == hook.MouseUp {
			// 计算持续时间
			m.mouseDownMu.Lock()
			if downTime, exists := m.mouseDownTime[button]; exists {
				duration := int(time.Since(downTime).Milliseconds())
				delete(m.mouseDownTime, button)

				if m.isRecording {
					// 如果持续时间较长，记录为hold事件
					if duration > 200 {
						event := MouseEvent{
							Timestamp: m.getRelativeTimestamp(),
							X:         ev.X,
							Y:         ev.Y,
							EventType: "hold",
							Duration:  duration,
							Button:    button,
						}
						m.addMouseEvent(event)
						// 立即发送给前端
						select {
						case m.immediateChan <- event:
						default:
						}
					}

					// 记录点击事件
					event := MouseEvent{
						Timestamp: m.getRelativeTimestamp(),
						X:         ev.X,
						Y:         ev.Y,
						EventType: getMouseUpType(button),
						Duration:  duration,
						Button:    button,
					}
					m.addMouseEvent(event)
					// 立即发送给前端
					select {
					case m.immediateChan <- event:
					default:
					}
				}
			}
			m.mouseDownMu.Unlock()
		}

	case hook.MouseWheel:
		if m.isRecording {
			event := MouseEvent{
				Timestamp: m.getRelativeTimestamp(),
				X:         m.lastX,
				Y:         m.lastY,
				EventType: "scroll",
				Delta:     int(ev.Amount),
			}
			m.addMouseEvent(event)
			// 立即发送给前端
			select {
			case m.immediateChan <- event:
			default:
			}
		}
	}
}

// addMouseEvent 添加鼠标事件到数据列表
func (m *MouseHook) addMouseEvent(event MouseEvent) {
	m.mouseDataMu.Lock()
	defer m.mouseDataMu.Unlock()
	m.mouseData = append(m.mouseData, event)
}

// emitImmediateEvents 立即发送关键时刻事件到前端
func (m *MouseHook) emitImmediateEvents() {
	for {
		select {
		case event := <-m.immediateChan:
			// 立即发送关键时刻事件
			if m.ctx != nil {
				runtime.EventsEmit(m.ctx, "mouse-position", event.X, event.Y, event.EventType, event.Button)
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// emitData 定时发送数据到前端
func (m *MouseHook) emitData() {
	for {
		select {
		case <-m.ticker.C:
			// 发送当前鼠标位置（用于实时显示）
			// 默认发送move事件类型
			if m.ctx != nil {
				runtime.EventsEmit(m.ctx, "mouse-position", m.lastX, m.lastY, "move", "")
			}

		case <-m.ctx.Done():
			return
		}
	}
}

// StartRecording 开始录制
func (m *MouseHook) StartRecording() {
	m.mouseDataMu.Lock()
	m.mouseData = make([]MouseEvent, 0)
	m.startTime = time.Now()
	m.isRecording = true
	m.mouseDataMu.Unlock()
	fmt.Println("开始录制鼠标数据...")
}

// StopRecording 停止录制
func (m *MouseHook) StopRecording() {
	m.mouseDataMu.Lock()
	m.isRecording = false
	m.mouseDataMu.Unlock()
	fmt.Printf("录制结束，共捕获 %d 个鼠标事件\n", len(m.mouseData))
}

// GetMouseData 获取录制的鼠标数据
func (m *MouseHook) GetMouseData() []MouseEvent {
	m.mouseDataMu.Lock()
	defer m.mouseDataMu.Unlock()
	return m.mouseData
}

// GetMouseDataJSON 获取鼠标数据的JSON格式
func (m *MouseHook) GetMouseDataJSON() (string, error) {
	data := m.GetMouseData()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化鼠标数据失败: %w", err)
	}
	return string(jsonData), nil
}

// ClearMouseData 清空鼠标数据
func (m *MouseHook) ClearMouseData() {
	m.mouseDataMu.Lock()
	m.mouseData = make([]MouseEvent, 0)
	m.mouseDataMu.Unlock()
}

// getRelativeTimestamp 获取相对于录制开始的时间戳（毫秒）
func (m *MouseHook) getRelativeTimestamp() int64 {
	return time.Since(m.startTime).Milliseconds()
}

// getMouseDownType 获取鼠标按下事件类型
func getMouseDownType(button string) string {
	switch button {
	case "left":
		return "l_down"
	case "right":
		return "r_down"
	case "middle":
		return "m_down"
	default:
		return "l_down"
	}
}

// getMouseUpType 获取鼠标释放事件类型
func getMouseUpType(button string) string {
	switch button {
	case "left":
		return "l_up"
	case "right":
		return "r_up"
	case "middle":
		return "m_up"
	default:
		return "l_up"
	}
}
