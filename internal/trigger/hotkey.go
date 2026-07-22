package trigger

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

// HotkeyTrigger 管理一个全局热键的注册与监听。
type HotkeyTrigger struct {
	hk *hotkey.Hotkey
}

// NewHotkeyTrigger 根据修饰键名称列表与主键名称解析并创建一个HotkeyTrigger（尚未注册）。
func NewHotkeyTrigger(modifierNames []string, keyName string) (*HotkeyTrigger, error) {
	mods := make([]hotkey.Modifier, 0, len(modifierNames))
	for _, name := range modifierNames {
		m, err := parseModifier(name)
		if err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}

	key, err := parseKey(keyName)
	if err != nil {
		return nil, err
	}

	return &HotkeyTrigger{hk: hotkey.New(mods, key)}, nil
}

// Start 注册热键并启动监听，每次按下热键都会调用onTrigger。
// 返回的stop函数用于停止监听并注销热键。
//
// 注意（macOS）：注册热键需要"辅助功能"权限，未授权时Register会返回错误；
// 且必须在mainthread.Init启动的主线程环境中调用，详见 cmd/screenocr/main.go。
func (t *HotkeyTrigger) Start(onTrigger func()) (stop func(), err error) {
	if err := t.hk.Register(); err != nil {
		return nil, fmt.Errorf("注册全局热键失败: %w", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-t.hk.Keydown():
				onTrigger()
			case <-done:
				return
			}
		}
	}()

	stop = func() {
		close(done)
		_ = t.hk.Unregister()
	}
	return stop, nil
}

func parseModifier(name string) (hotkey.Modifier, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ctrl", "control":
		return hotkey.ModCtrl, nil
	case "shift":
		return hotkey.ModShift, nil
	case "alt", "option":
		return altModifier(), nil
	case "win", "cmd", "command", "super", "meta":
		return superModifier(), nil
	default:
		return 0, fmt.Errorf("不支持的修饰键: %s（支持 ctrl/shift/alt/win(cmd)）", name)
	}
}

func parseKey(name string) (hotkey.Key, error) {
	key := strings.ToUpper(strings.TrimSpace(name))

	if len(key) == 1 {
		c := key[0]
		if c >= 'A' && c <= 'Z' {
			if k, ok := letterKeys[c]; ok {
				return k, nil
			}
		}
		if c >= '0' && c <= '9' {
			if k, ok := digitKeys[c]; ok {
				return k, nil
			}
		}
	}

	switch key {
	case "`", "~", "GRAVE", "BACKTICK", "TILDE":
		return backtickKey(), nil
	}

	if k, ok := namedKeys[key]; ok {
		return k, nil
	}

	return 0, fmt.Errorf("不支持的按键: %s", name)
}

var letterKeys = map[byte]hotkey.Key{
	'A': hotkey.KeyA, 'B': hotkey.KeyB, 'C': hotkey.KeyC, 'D': hotkey.KeyD,
	'E': hotkey.KeyE, 'F': hotkey.KeyF, 'G': hotkey.KeyG, 'H': hotkey.KeyH,
	'I': hotkey.KeyI, 'J': hotkey.KeyJ, 'K': hotkey.KeyK, 'L': hotkey.KeyL,
	'M': hotkey.KeyM, 'N': hotkey.KeyN, 'O': hotkey.KeyO, 'P': hotkey.KeyP,
	'Q': hotkey.KeyQ, 'R': hotkey.KeyR, 'S': hotkey.KeyS, 'T': hotkey.KeyT,
	'U': hotkey.KeyU, 'V': hotkey.KeyV, 'W': hotkey.KeyW, 'X': hotkey.KeyX,
	'Y': hotkey.KeyY, 'Z': hotkey.KeyZ,
}

var digitKeys = map[byte]hotkey.Key{
	'0': hotkey.Key0, '1': hotkey.Key1, '2': hotkey.Key2, '3': hotkey.Key3,
	'4': hotkey.Key4, '5': hotkey.Key5, '6': hotkey.Key6, '7': hotkey.Key7,
	'8': hotkey.Key8, '9': hotkey.Key9,
}

var namedKeys = map[string]hotkey.Key{
	"SPACE":  hotkey.KeySpace,
	"TAB":    hotkey.KeyTab,
	"ESC":    hotkey.KeyEscape,
	"ESCAPE": hotkey.KeyEscape,
	"ENTER":  hotkey.KeyReturn,
	"RETURN": hotkey.KeyReturn,
	"DELETE": hotkey.KeyDelete,
	"LEFT":   hotkey.KeyLeft,
	"RIGHT":  hotkey.KeyRight,
	"UP":     hotkey.KeyUp,
	"DOWN":   hotkey.KeyDown,
	"F1":     hotkey.KeyF1,
	"F2":     hotkey.KeyF2,
	"F3":     hotkey.KeyF3,
	"F4":     hotkey.KeyF4,
	"F5":     hotkey.KeyF5,
	"F6":     hotkey.KeyF6,
	"F7":     hotkey.KeyF7,
	"F8":     hotkey.KeyF8,
	"F9":     hotkey.KeyF9,
	"F10":    hotkey.KeyF10,
	"F11":    hotkey.KeyF11,
	"F12":    hotkey.KeyF12,
	"F13":    hotkey.KeyF13,
	"F14":    hotkey.KeyF14,
	"F15":    hotkey.KeyF15,
	"F16":    hotkey.KeyF16,
	"F17":    hotkey.KeyF17,
	"F18":    hotkey.KeyF18,
	"F19":    hotkey.KeyF19,
	"F20":    hotkey.KeyF20,
}
