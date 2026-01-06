package hook

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
)

// KeyboardEvent 键盘事件
type KeyboardEvent struct {
	Timestamp int64    `json:"t"`         // 相对时间戳（毫秒）
	Key       string   `json:"key"`       // 按键名称 (如 "A", "Enter", "Esc")
	Rawcode   uint16   `json:"rawcode"`   // 原始按键代码
	Modifiers []string `json:"modifiers"` // 修饰键 (Ctrl, Shift, Alt, Win)
	EventType string   `json:"type"`      // key_down, key_up
}

// KeyboardHook 键盘钩子
type KeyboardHook struct {
	events       []KeyboardEvent
	eventsMu     sync.Mutex
	isRecording  bool
	isPaused     bool
	startTime    time.Time
	pausedTime   time.Duration
	pauseStart   time.Time
	stopChan     chan bool
	eventHandler func(KeyboardEvent) // 可选的事件处理器
}

// NewKeyboardHook 创建键盘钩子
func NewKeyboardHook() *KeyboardHook {
	return &KeyboardHook{
		events:   make([]KeyboardEvent, 0),
		stopChan: make(chan bool),
	}
}

// StartRecording 开始录制键盘事件
func (k *KeyboardHook) StartRecording() error {
	k.eventsMu.Lock()
	if k.isRecording {
		k.eventsMu.Unlock()
		return fmt.Errorf("键盘钩子已在运行")
	}

	k.isRecording = true
	k.isPaused = false
	k.startTime = time.Now()
	k.pausedTime = 0
	k.events = make([]KeyboardEvent, 0)
	k.eventsMu.Unlock()

	// 启动事件处理协程
	go k.processEvents()

	fmt.Println("键盘钩子已启动")
	return nil
}

// StopRecording 停止录制
func (k *KeyboardHook) StopRecording() error {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	if !k.isRecording {
		return fmt.Errorf("键盘钩子未运行")
	}

	k.isRecording = false
	k.isPaused = false

	// 发送停止信号
	select {
	case k.stopChan <- true:
	default:
	}

	fmt.Printf("键盘钩子已停止，记录了 %d 个事件\n", len(k.events))
	return nil
}

// PauseRecording 暂停录制
func (k *KeyboardHook) PauseRecording() error {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	if !k.isRecording || k.isPaused {
		return fmt.Errorf("无法暂停")
	}

	k.isPaused = true
	k.pauseStart = time.Now()
	return nil
}

// ResumeRecording 恢复录制
func (k *KeyboardHook) ResumeRecording() error {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	if !k.isPaused {
		return fmt.Errorf("未暂停")
	}

	k.pausedTime += time.Since(k.pauseStart)
	k.isPaused = false
	return nil
}

// processEvents 处理键盘事件
func (k *KeyboardHook) processEvents() {
	// 启动 gohook 事件监听
	eventChan := hook.Start()
	defer hook.End()

	fmt.Println("键盘事件处理协程已启动")

	for {
		select {
		case <-k.stopChan:
			fmt.Println("收到停止信号，退出键盘事件处理")
			return

		case ev := <-eventChan:
			// 只处理键盘事件
			if ev.Kind == hook.KeyDown || ev.Kind == hook.KeyUp {
				k.handleKeyEvent(ev)
			}
		}
	}
}

// handleKeyEvent 处理单个键盘事件
func (k *KeyboardHook) handleKeyEvent(ev hook.Event) {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	if !k.isRecording || k.isPaused {
		return
	}

	// 计算相对时间戳
	elapsed := time.Since(k.startTime) - k.pausedTime
	timestamp := elapsed.Milliseconds()

	// 确定事件类型
	eventType := "key_down"
	if ev.Kind == hook.KeyUp {
		eventType = "key_up"
	}

	// 获取修饰键
	modifiers := k.getModifiers(ev)

	// 获取按键名称
	keyName := k.getKeyName(ev.Rawcode, ev.Keychar)

	// 创建事件
	event := KeyboardEvent{
		Timestamp: timestamp,
		Key:       keyName,
		Rawcode:   ev.Rawcode,
		Modifiers: modifiers,
		EventType: eventType,
	}

	// 保存事件
	k.events = append(k.events, event)

	// 调用事件处理器（如果有）
	if k.eventHandler != nil {
		k.eventHandler(event)
	}
}

// getModifiers 获取修饰键
func (k *KeyboardHook) getModifiers(ev hook.Event) []string {
	modifiers := []string{}

	// 检查修饰键状态
	// 注意：这里需要根据实际情况检测
	// gohook 的 ev.Mask 包含修饰键信息

	if ev.Mask&hook.Ctrl != 0 {
		modifiers = append(modifiers, "Ctrl")
	}
	if ev.Mask&hook.Shift != 0 {
		modifiers = append(modifiers, "Shift")
	}
	if ev.Mask&hook.Alt != 0 {
		modifiers = append(modifiers, "Alt")
	}
	if ev.Mask&hook.Cmd != 0 {
		modifiers = append(modifiers, "Win")
	}

	return modifiers
}

// getKeyName 获取按键名称
func (k *KeyboardHook) getKeyName(rawcode uint16, keychar rune) string {
	// 特殊键映射
	specialKeys := map[uint16]string{
		8:   "Backspace",
		9:   "Tab",
		13:  "Enter",
		27:  "Esc",
		32:  "Space",
		33:  "PageUp",
		34:  "PageDown",
		35:  "End",
		36:  "Home",
		37:  "Left",
		38:  "Up",
		39:  "Right",
		40:  "Down",
		45:  "Insert",
		46:  "Delete",
		112: "F1",
		113: "F2",
		114: "F3",
		115: "F4",
		116: "F5",
		117: "F6",
		118: "F7",
		119: "F8",
		120: "F9",
		121: "F10",
		122: "F11",
		123: "F12",
		160: "LShift",
		161: "RShift",
		162: "LCtrl",
		163: "RCtrl",
		164: "LAlt",
		165: "RAlt",
	}

	// 检查是否是特殊键
	if name, ok := specialKeys[rawcode]; ok {
		return name
	}

	// 普通字符键
	if keychar != 0 && keychar >= 32 && keychar < 127 {
		return string(keychar)
	}

	// 未知键，返回原始代码
	return fmt.Sprintf("Key%d", rawcode)
}

// GetEvents 获取所有事件
func (k *KeyboardHook) GetEvents() []KeyboardEvent {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	// 返回副本
	events := make([]KeyboardEvent, len(k.events))
	copy(events, k.events)
	return events
}

// SaveToFile 保存事件到文件
func (k *KeyboardHook) SaveToFile(filename string) error {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	data, err := json.MarshalIndent(k.events, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化键盘事件失败: %w", err)
	}

	if err := saveToFile(filename, data); err != nil {
		return fmt.Errorf("保存键盘事件失败: %w", err)
	}

	fmt.Printf("键盘事件已保存到: %s (%d 个事件)\n", filename, len(k.events))
	return nil
}

// LoadFromFile 从文件加载事件
func (k *KeyboardHook) LoadFromFile(filename string) error {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	data, err := readFromFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var events []KeyboardEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("解析键盘事件失败: %w", err)
	}

	k.events = events
	fmt.Printf("已加载 %d 个键盘事件\n", len(events))
	return nil
}

// SetEventHandler 设置事件处理器
func (k *KeyboardHook) SetEventHandler(handler func(KeyboardEvent)) {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()
	k.eventHandler = handler
}

// GetLastEvent 获取最后一个事件
func (k *KeyboardHook) GetLastEvent() *KeyboardEvent {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()

	if len(k.events) == 0 {
		return nil
	}

	return &k.events[len(k.events)-1]
}

// GetEventCount 获取事件数量
func (k *KeyboardHook) GetEventCount() int {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()
	return len(k.events)
}

// IsRecording 检查是否正在录制
func (k *KeyboardHook) IsRecording() bool {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()
	return k.isRecording
}

// IsPaused 检查是否暂停
func (k *KeyboardHook) IsPaused() bool {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()
	return k.isPaused
}

// Clear 清除所有事件
func (k *KeyboardHook) Clear() {
	k.eventsMu.Lock()
	defer k.eventsMu.Unlock()
	k.events = make([]KeyboardEvent, 0)
}

// FormatKeyCombo 格式化按键组合（用于显示）
func FormatKeyCombo(event KeyboardEvent) string {
	if len(event.Modifiers) == 0 {
		return event.Key
	}

	combo := ""
	for _, mod := range event.Modifiers {
		combo += mod + " + "
	}
	combo += event.Key

	return combo
}
